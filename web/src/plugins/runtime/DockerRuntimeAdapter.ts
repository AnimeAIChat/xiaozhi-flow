import {
  RuntimeAdapter,
  BackendConfig,
  ServiceInfo
} from '../types';
import { ProcessManager } from '../utils/ProcessManager';
import { SimpleEventEmitter } from '../utils/ProcessManager';

interface DockerEnvironment {
  version: string;
  apiVersion: string;
  host: string;
  certPath?: string;
}

interface DockerContainerInfo {
  id: string;
  name: string;
  status: string;
  ports: { [containerPort: string]: { [hostPort: string]: {} } };
  labels: { [key: string]: string };
  created: Date;
}

interface DockerServiceConfig extends BackendConfig {
  imageName?: string;
  dockerfile?: string;
  context?: string;
  buildArgs?: { [key: string]: string };
  containerName?: string;
  network?: string;
  volumes?: { [hostPath: string]: { bind: string; mode: string } };
  autoRemove?: boolean;
  restartPolicy?: 'no' | 'on-failure' | 'always' | 'unless-stopped';
  privileged?: boolean;
  user?: string;
  workingDir?: string;
}

export class DockerRuntimeAdapter extends SimpleEventEmitter implements RuntimeAdapter {
  type = 'docker';
  name = 'Docker Runtime';
  version = '1.0.0';

  private processManager: ProcessManager;
  private environment: DockerEnvironment | null = null;
  private containers: Map<string, DockerContainerInfo> = new Map();
  private services: Map<string, ServiceInfo> = new Map();

  constructor() {
    super();
    this.processManager = new ProcessManager();
  }

  /**
   * 启动Docker服务
   */
  async start(config: DockerServiceConfig): Promise<ServiceInfo> {
    try {
      this.emit('service-starting', { config });

      // 验证Docker环境
      await this.validateDockerEnvironment();

      // 构建或拉取镜像
      const imageName = await this.prepareImage(config);

      // 创建并启动容器
      const containerInfo = await this.startContainer(config, imageName);

      // 等待服务启动
      const serviceInfo = await this.waitForContainerReady(containerInfo, config);

      this.services.set(`${config.pluginId}:${serviceInfo.id}`, serviceInfo);
      this.containers.set(containerInfo.id, containerInfo);
      this.emit('service-started', { serviceInfo });

      return serviceInfo;

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      this.emit('service-error', { config, error: errorMessage });
      throw new Error(`Failed to start Docker service: ${errorMessage}`);
    }
  }

  /**
   * 停止Docker服务
   */
  async stop(serviceInfo: ServiceInfo): Promise<void> {
    try {
      this.emit('service-stopping', { serviceInfo });

      const serviceKey = `${serviceInfo.pluginId}:${serviceInfo.id}`;
      const service = this.services.get(serviceKey);

      if (!service) {
        throw new Error('Service not found');
      }

      // 查找对应的容器
      const containerInfo = Array.from(this.containers.values()).find(
        container => container.labels && container.labels['plugin-id'] === serviceInfo.pluginId
      );

      if (containerInfo) {
        await this.stopContainer(containerInfo.id);
        this.containers.delete(containerInfo.id);
      }

      this.services.delete(serviceKey);
      this.emit('service-stopped', { serviceInfo });

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      this.emit('service-error', { serviceInfo, error: errorMessage });
      throw new Error(`Failed to stop Docker service: ${errorMessage}`);
    }
  }

  /**
   * 重启Docker服务
   */
  async restart(serviceInfo: ServiceInfo): Promise<ServiceInfo> {
    const serviceKey = `${serviceInfo.pluginId}:${serviceInfo.id}`;
    const service = this.services.get(serviceKey);

    if (!service) {
      throw new Error('Service not found');
    }

    // 保存当前配置
    const currentConfig = service.config as DockerServiceConfig;

    // 停止服务
    await this.stop(serviceInfo);

    // 重新启动服务
    return await this.start(currentConfig);
  }

  /**
   * 获取服务状态
   */
  async getStatus(serviceInfo: ServiceInfo): Promise<ServiceInfo> {
    const serviceKey = `${serviceInfo.pluginId}:${serviceInfo.id}`;
    const service = this.services.get(serviceKey);

    if (!service) {
      throw new Error('Service not found');
    }

    // 查找对应的容器
    const containerInfo = Array.from(this.containers.values()).find(
      container => container.labels && container.labels['plugin-id'] === serviceInfo.pluginId
    );

    if (containerInfo) {
      // 获取容器最新状态
      const updatedContainer = await this.getContainerInfo(containerInfo.id);
      if (updatedContainer) {
        this.containers.set(containerInfo.id, updatedContainer);

        // 更新服务状态
        if (updatedContainer.status.startsWith('Up')) {
          service.status = 'running';
        } else {
          service.status = 'stopped';
          service.error = `Container status: ${updatedContainer.status}`;
        }

        // 获取容器统计信息
        const stats = await this.getContainerStats(containerInfo.id);
        if (stats) {
          service.memoryUsage = stats.memoryUsage;
          service.cpuUsage = stats.cpuUsage;
        }
      }
    }

    // 检查服务健康状态
    if (service.status === 'running') {
      const isHealthy = await this.healthCheck(serviceInfo);
      service.healthStatus = isHealthy ? 'healthy' : 'unhealthy';
      service.lastHealthCheck = new Date();
    }

    return service;
  }

  /**
   * 执行健康检查
   */
  async healthCheck(serviceInfo: ServiceInfo): Promise<boolean> {
    try {
      if (!serviceInfo.baseUrl) {
        return false;
      }

      const response = await fetch(`${serviceInfo.baseUrl}/health`, {
        method: 'GET',
        signal: AbortSignal.timeout(5000)
      });

      return response.ok;

    } catch (error) {
      return false;
    }
  }

  /**
   * 验证Docker环境
   */
  private async validateDockerEnvironment(): Promise<DockerEnvironment> {
    if (this.environment) {
      return this.environment;
    }

    try {
      // 检查Docker是否安装
      const versionResult = await this.executeDockerCommand(['--version']);
      const versionMatch = versionResult.stdout?.match(/Docker version (\d+\.\d+\.\d+)/);
      const version = versionMatch ? versionMatch[1] : 'unknown';

      // 获取Docker API版本
      const infoResult = await this.executeDockerCommand(['version', '--format', '{{.Server.APIVersion}}']);
      const apiVersion = infoResult.stdout?.trim() || 'unknown';

      this.environment = {
        version,
        apiVersion,
        host: process.env.DOCKER_HOST || 'unix:///var/run/docker.sock'
      };

      return this.environment;

    } catch (error) {
      throw new Error('Docker environment not available. Please ensure Docker is installed and running.');
    }
  }

  /**
   * 准备Docker镜像
   */
  private async prepareImage(config: DockerServiceConfig): Promise<string> {
    if (config.imageName) {
      // 如果指定了镜像名称，检查是否存在，不存在则拉取
      const imageExists = await this.checkImageExists(config.imageName);
      if (!imageExists) {
        await this.pullImage(config.imageName);
      }
      return config.imageName;
    } else {
      // 如果没有指定镜像，构建镜像
      return await this.buildImage(config);
    }
  }

  /**
   * 构建Docker镜像
   */
  private async buildImage(config: DockerServiceConfig): Promise<string> {
    const dockerfile = config.dockerfile || 'Dockerfile';
    const context = config.context || config.workingDirectory || '.';
    const imageName = `plugin-${config.pluginId}:latest`;

    const buildArgs = [
      'build',
      '-f', dockerfile,
      '-t', imageName
    ];

    // 添加构建参数
    if (config.buildArgs) {
      Object.entries(config.buildArgs).forEach(([key, value]) => {
        buildArgs.push('--build-arg', `${key}=${value}`);
      });
    }

    buildArgs.push(context);

    await this.executeDockerCommand(buildArgs, {
      cwd: config.workingDirectory,
      timeout: 600000 // 10分钟超时
    });

    return imageName;
  }

  /**
   * 启动Docker容器
   */
  private async startContainer(config: DockerServiceConfig, imageName: string): Promise<DockerContainerInfo> {
    const containerName = config.containerName || `plugin-${config.pluginId}-${Date.now()}`;

    const runArgs = [
      'run',
      '-d', // 后台运行
      '--name', containerName,
      '--label', `plugin-id=${config.pluginId}`,
      '--label', `service-type=docker`
    ];

    // 添加端口映射
    if (config.port) {
      runArgs.push('-p', `${config.port}:${config.port}`);
    }

    // 添加环境变量
    if (config.envVars) {
      Object.entries(config.envVars).forEach(([key, value]) => {
        runArgs.push('-e', `${key}=${value}`);
      });
    }

    // 添加网络配置
    if (config.network) {
      runArgs.push('--network', config.network);
    }

    // 添加卷挂载
    if (config.volumes) {
      Object.entries(config.volumes).forEach(([hostPath, mountConfig]) => {
        runArgs.push('-v', `${hostPath}:${mountConfig.bind}:${mountConfig.mode}`);
      });
    }

    // 添加重启策略
    if (config.restartPolicy) {
      runArgs.push('--restart', config.restartPolicy);
    }

    // 添加用户配置
    if (config.user) {
      runArgs.push('--user', config.user);
    }

    // 添加工作目录
    if (config.workingDir) {
      runArgs.push('-w', config.workingDir);
    }

    // 添加特权模式
    if (config.privileged) {
      runArgs.push('--privileged');
    }

    // 自动移除容器
    if (config.autoRemove) {
      runArgs.push('--rm');
    }

    // 添加镜像名称
    runArgs.push(imageName);

    // 启动容器
    const result = await this.executeDockerCommand(runArgs, {
      timeout: 30000
    });

    const containerId = result.stdout?.trim();
    if (!containerId) {
      throw new Error('Failed to start container');
    }

    // 获取容器信息
    return await this.getContainerInfo(containerId);
  }

  /**
   * 等待容器就绪
   */
  private async waitForContainerReady(containerInfo: DockerContainerInfo, config: DockerServiceConfig): Promise<ServiceInfo> {
    const serviceInfo: ServiceInfo = {
      id: containerInfo.name,
      pluginId: config.pluginId || 'unknown',
      name: 'Docker Service',
      status: 'starting',
      port: config.port || 0,
      host: '127.0.0.1',
      baseUrl: config.port ? `http://127.0.0.1:${config.port}` : '',
      startTime: containerInfo.created,
      runtime: 'docker',
      healthStatus: 'unknown',
      config
    };

    // 获取端口映射
    if (!config.port && containerInfo.ports) {
      const portMapping = Object.keys(containerInfo.ports)[0];
      if (portMapping) {
        const hostPorts = Object.keys(containerInfo.ports[portMapping]);
        if (hostPorts.length > 0) {
          serviceInfo.port = parseInt(hostPorts[0]);
          serviceInfo.baseUrl = `http://127.0.0.1:${serviceInfo.port}`;
        }
      }
    }

    // 等待服务启动
    const timeout = config.startupTimeout || 30000;
    const startTime = Date.now();
    const maxTime = startTime + timeout;

    while (Date.now() < maxTime) {
      try {
        if (serviceInfo.port > 0) {
          const response = await fetch(`${serviceInfo.baseUrl}/health`, {
            method: 'GET',
            signal: AbortSignal.timeout(2000)
          });

          if (response.ok) {
            serviceInfo.status = 'running';
            serviceInfo.healthStatus = 'healthy';
            return serviceInfo;
          }
        }
      } catch (error) {
        // 服务还未就绪，继续等待
      }

      // 检查容器状态
      const updatedContainer = await this.getContainerInfo(containerInfo.id);
      if (updatedContainer && !updatedContainer.status.startsWith('Up')) {
        throw new Error(`Container stopped unexpectedly: ${updatedContainer.status}`);
      }

      await new Promise(resolve => setTimeout(resolve, 2000));
    }

    throw new Error(`Service not ready after ${timeout}ms`);
  }

  /**
   * 检查镜像是否存在
   */
  private async checkImageExists(imageName: string): Promise<boolean> {
    try {
      await this.executeDockerCommand(['inspect', imageName], { timeout: 10000 });
      return true;
    } catch (error) {
      return false;
    }
  }

  /**
   * 拉取Docker镜像
   */
  private async pullImage(imageName: string): Promise<void> {
    await this.executeDockerCommand(['pull', imageName], {
      timeout: 600000 // 10分钟超时
    });
  }

  /**
   * 停止容器
   */
  private async stopContainer(containerId: string): Promise<void> {
    try {
      // 尝试优雅停止
      await this.executeDockerCommand(['stop', containerId], {
        timeout: 30000
      });
    } catch (error) {
      // 如果优雅停止失败，强制删除
      try {
        await this.executeDockerCommand(['rm', '-f', containerId], {
          timeout: 10000
        });
      } catch (rmError) {
        throw new Error(`Failed to stop and remove container: ${rmError}`);
      }
    }
  }

  /**
   * 获取容器信息
   */
  private async getContainerInfo(containerId: string): Promise<DockerContainerInfo | null> {
    try {
      const result = await this.executeDockerCommand([
        'inspect',
        containerId,
        '--format', '{{json .}}'
      ], { timeout: 10000 });

      const containerData = JSON.parse(result.stdout || 'null');
      if (!containerData) {
        return null;
      }

      return {
        id: containerData.Id,
        name: containerData.Name.replace(/^\//, ''),
        status: containerData.State.Status,
        ports: containerData.NetworkSettings.Ports || {},
        labels: containerData.Config.Labels || {},
        created: new Date(containerData.Created)
      };

    } catch (error) {
      return null;
    }
  }

  /**
   * 获取容器统计信息
   */
  private async getContainerStats(containerId: string): Promise<{
    memoryUsage: number;
    cpuUsage: number;
  } | null> {
    try {
      const result = await this.executeDockerCommand([
        'stats',
        containerId,
        '--no-stream',
        '--format', '{{json .}}'
      ], { timeout: 5000 });

      const statsData = JSON.parse(result.stdout || 'null');
      if (!statsData) {
        return null;
      }

      return {
        memoryUsage: statsData.MemUsage || 0,
        cpuUsage: statsData.CPUPerc || 0
      };

    } catch (error) {
      return null;
    }
  }

  /**
   * 执行Docker命令
   */
  private async executeDockerCommand(
    args: string[],
    options: {
      cwd?: string;
      timeout?: number;
      env?: Record<string, string>;
    } = {}
  ): Promise<{ stdout: string; stderr: string; code: number | null }> {
    return await this.processManager.executeCommand({
      command: 'docker',
      args,
      cwd: options.cwd,
      env: { ...process.env, ...options.env },
      timeout: options.timeout || 30000
    });
  }

  /**
   * 清理资源
   */
  async cleanup(serviceInfo?: ServiceInfo): Promise<void> {
    if (serviceInfo) {
      await this.stop(serviceInfo);
    } else {
      // 清理所有容器
      const stopPromises = Array.from(this.containers.values()).map(
        container => this.stopContainer(container.id).catch(error =>
          console.error(`Failed to stop container ${container.id}:`, error)
        )
      );
      await Promise.all(stopPromises);

      this.containers.clear();
    }

    this.services.clear();
  }
}