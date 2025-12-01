import {
  RuntimeAdapter,
  BackendConfig,
  ServiceInfo
} from '../types';
import { ProcessManager } from '../utils/ProcessManager';
import { SimpleEventEmitter } from '../utils/ProcessManager';

interface GoEnvironment {
  path: string;
  version: string;
  goRoot: string;
  goPath: string;
  goproxy?: string;
}

interface GoServiceConfig extends BackendConfig {
  goPath?: string;
  goModFile?: string;
  buildArgs?: string[];
  runArgs?: string[];
  binaryPath?: string;
  buildOutput?: string;
  crossCompile?: {
    goos: string;
    goarch: string;
  };
}

export class GoRuntimeAdapter extends SimpleEventEmitter implements RuntimeAdapter {
  type = 'go';
  name = 'Go Runtime';
  version = '1.0.0';

  private processManager: ProcessManager;
  private environments: Map<string, GoEnvironment> = new Map();
  private services: Map<string, ServiceInfo> = new Map();
  private buildCache: Map<string, string> = new Map();

  constructor() {
    super();
    this.processManager = new ProcessManager();
  }

  /**
   * 启动Go服务
   */
  async start(config: GoServiceConfig): Promise<ServiceInfo> {
    try {
      this.emit('service-starting', { config });

      // 验证Go环境
      const goEnv = await this.validateEnvironment(config);

      // 准备Go模块环境
      await this.prepareGoModule(config);

      // 构建Go应用
      const binaryPath = await this.buildGoApp(config, goEnv);

      // 启动Go进程
      const serviceInfo = await this.startGoProcess(config, binaryPath);

      // 等待服务启动
      await this.waitForServiceReady(serviceInfo, config.startupTimeout || 30000);

      this.services.set(`${config.pluginId}:${serviceInfo.id}`, serviceInfo);
      this.emit('service-started', { serviceInfo });

      return serviceInfo;

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      this.emit('service-error', { config, error: errorMessage });
      throw new Error(`Failed to start Go service: ${errorMessage}`);
    }
  }

  /**
   * 停止Go服务
   */
  async stop(serviceInfo: ServiceInfo): Promise<void> {
    try {
      this.emit('service-stopping', { serviceInfo });

      const serviceKey = `${serviceInfo.pluginId}:${serviceInfo.id}`;
      const service = this.services.get(serviceKey);

      if (!service) {
        throw new Error('Service not found');
      }

      // 停止Go进程
      if (service.pid) {
        await this.processManager.killProcess(service.pid, 'SIGTERM');

        // 等待进程优雅退出
        await this.waitForProcessExit(service.pid, 5000);

        // 如果进程还在运行，强制杀死
        const isRunning = await this.processManager.isProcessRunning(service.pid);
        if (isRunning) {
          await this.processManager.killProcess(service.pid, 'SIGKILL');
        }
      }

      this.services.delete(serviceKey);
      this.emit('service-stopped', { serviceInfo });

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      this.emit('service-error', { serviceInfo, error: errorMessage });
      throw new Error(`Failed to stop Go service: ${errorMessage}`);
    }
  }

  /**
   * 重启Go服务
   */
  async restart(serviceInfo: ServiceInfo): Promise<ServiceInfo> {
    const serviceKey = `${serviceInfo.pluginId}:${serviceInfo.id}`;
    const service = this.services.get(serviceKey);

    if (!service) {
      throw new Error('Service not found');
    }

    // 保存当前配置
    const currentConfig = service.config as GoServiceConfig;

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

    // 检查进程是否还在运行
    if (service.pid) {
      const isRunning = await this.processManager.isProcessRunning(service.pid);
      if (!isRunning) {
        service.status = 'stopped';
        service.error = 'Process terminated unexpectedly';
      } else {
        // 获取进程信息
        const processInfo = await this.processManager.getProcessInfo(service.pid);
        service.memoryUsage = processInfo.memoryUsage;
        service.cpuUsage = processInfo.cpuUsage;
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
   * 验证Go环境
   */
  async validateEnvironment(): Promise<GoEnvironment> {
    const goPath = await this.findGoExecutable();
    const version = await this.getGoVersion(goPath);
    const goRoot = process.env.GOROOT || await this.getGoRoot(goPath);
    const goPathEnv = process.env.GOPATH || await this.getGoPath(goPath);

    return {
      path: goPath,
      version,
      goRoot,
      goPath: goPathEnv,
      goproxy: process.env.GOPROXY
    };
  }

  /**
   * 准备Go模块环境
   */
  private async prepareGoModule(config: GoServiceConfig): Promise<void> {
    const workingDir = config.workingDirectory || '/';

    // 检查是否存在go.mod文件
    const goModPath = config.goModFile || `${workingDir}/go.mod`;

    if (!(await this.pathExists(goModPath))) {
      // 初始化Go模块
      await this.processManager.executeCommand({
        command: 'go',
        args: ['mod', 'init', config.pluginId || 'service'],
        cwd: workingDir,
        env: this.buildGoEnvironment(config),
        timeout: 30000
      });
    }

    // 下载依赖
    await this.processManager.executeCommand({
      command: 'go',
      args: ['mod', 'download'],
      cwd: workingDir,
      env: this.buildGoEnvironment(config),
      timeout: 120000 // 2分钟超时
    });

    // 整理依赖
    await this.processManager.executeCommand({
      command: 'go',
      args: ['mod', 'tidy'],
      cwd: workingDir,
      env: this.buildGoEnvironment(config),
      timeout: 60000
    });
  }

  /**
   * 构建Go应用
   */
  private async buildGoApp(config: GoServiceConfig, goEnv: GoEnvironment): Promise<string> {
    const workingDir = config.workingDirectory || '/';
    const buildOutput = config.buildOutput || `/tmp/go-binaries/${config.pluginId || 'service'}`;

    // 确保输出目录存在
    await this.ensureDirectoryExists(buildOutput);

    const buildArgs = [
      'build',
      '-o', buildOutput,
      ...(config.buildArgs || [])
    ];

    // 如果是交叉编译
    if (config.crossCompile) {
      buildArgs.push(
        '-ldflags', '-s -w', // 去除调试信息，减小二进制大小
        '-trimpath'
      );
    }

    // 构建Go应用
    await this.processManager.executeCommand({
      command: 'go',
      args: buildArgs,
      cwd: workingDir,
      env: this.buildGoEnvironment(config, config.crossCompile),
      timeout: 300000 // 5分钟超时
    });

    // 检查二进制文件是否生成成功
    if (!(await this.pathExists(buildOutput))) {
      throw new Error(`Failed to build Go binary: ${buildOutput}`);
    }

    // 设置执行权限
    await this.processManager.executeCommand({
      command: 'chmod',
      args: ['+x', buildOutput],
      timeout: 5000
    });

    return buildOutput;
  }

  /**
   * 启动Go进程
   */
  private async startGoProcess(config: GoServiceConfig, binaryPath: string): Promise<ServiceInfo> {
    const args = [
      '--port', config.port?.toString() || '0',
      '--host', '127.0.0.1',
      ...(config.runArgs || [])
    ];

    const env = {
      ...process.env,
      ...config.envVars,
      PORT: config.port?.toString() || '0',
      HOST: '127.0.0.1'
    };

    const processInfo = await this.processManager.startProcess({
      command: binaryPath,
      args,
      cwd: config.workingDirectory || '/',
      env,
      stdio: ['pipe', 'pipe', 'pipe']
    });

    const port = config.port || 0;
    const serviceInfo: ServiceInfo = {
      id: 'main',
      pluginId: config.pluginId || 'unknown',
      name: 'Go Service',
      status: 'starting',
      port,
      host: '127.0.0.1',
      baseUrl: `http://127.0.0.1:${port}`,
      pid: processInfo.pid,
      startTime: new Date(),
      runtime: 'go',
      healthStatus: 'unknown',
      config
    };

    // 监听进程输出
    if (processInfo.stdout) {
      processInfo.stdout.on('data', (data) => {
        const output = data.toString();
        this.emit('service-stdout', { serviceInfo, output });

        // 解析端口信息（如果使用端口0自动分配）
        if (output.includes('Server running on port')) {
          const match = output.match(/port (\d+)/);
          if (match) {
            serviceInfo.port = parseInt(match[1]);
            serviceInfo.baseUrl = `http://127.0.0.1:${serviceInfo.port}`;
          }
        }
      });
    }

    if (processInfo.stderr) {
      processInfo.stderr.on('data', (data) => {
        const output = data.toString();
        this.emit('service-stderr', { serviceInfo, output });
      });
    }

    processInfo.on('exit', (code, signal) => {
      serviceInfo.status = 'stopped';
      serviceInfo.error = `Process exited with code ${code}`;
      this.emit('service-exited', { serviceInfo, code, signal });
    });

    return serviceInfo;
  }

  /**
   * 构建Go环境变量
   */
  private buildGoEnvironment(config: GoServiceConfig, crossCompile?: GoServiceConfig['crossCompile']): Record<string, string> {
    const env = {
      ...process.env,
      ...config.envVars,
      GO111MODULE: 'on',
      CGO_ENABLED: '0' // 静态编译
    };

    if (crossCompile) {
      env.GOOS = crossCompile.goos;
      env.GOARCH = crossCompile.goarch;
    }

    if (config.goproxy) {
      env.GOPROXY = config.goproxy;
    }

    if (config.goPath) {
      env.GOPATH = config.goPath;
    }

    return env;
  }

  /**
   * 等待服务就绪
   */
  private async waitForServiceReady(serviceInfo: ServiceInfo, timeout: number): Promise<void> {
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
            return;
          }
        }
      } catch (error) {
        // 服务还未就绪，继续等待
      }

      await new Promise(resolve => setTimeout(resolve, 1000));
    }

    throw new Error(`Service not ready after ${timeout}ms`);
  }

  /**
   * 等待进程退出
   */
  private async waitForProcessExit(pid: number, timeout: number): Promise<void> {
    const startTime = Date.now();
    const maxTime = startTime + timeout;

    while (Date.now() < maxTime) {
      const isRunning = await this.processManager.isProcessRunning(pid);
      if (!isRunning) {
        return;
      }
      await new Promise(resolve => setTimeout(resolve, 100));
    }
  }

  /**
   * 查找Go可执行文件
   */
  private async findGoExecutable(): Promise<string> {
    const candidates = ['go', '/usr/local/go/bin/go', '/usr/bin/go'];

    for (const candidate of candidates) {
      try {
        await this.processManager.executeCommand({
          command: candidate,
          args: ['version'],
          timeout: 5000
        });
        return candidate;
      } catch (error) {
        continue;
      }
    }

    throw new Error('Go executable not found');
  }

  /**
   * 获取Go版本
   */
  private async getGoVersion(goPath: string): Promise<string> {
    try {
      const result = await this.processManager.executeCommand({
        command: goPath,
        args: ['version'],
        timeout: 5000
      });

      const output = result.stdout || result.stderr || '';
      const match = output.match(/go version go(\d+\.\d+\.\d+)/);
      return match ? match[1] : 'unknown';
    } catch (error) {
      return 'unknown';
    }
  }

  /**
   * 获取Go根目录
   */
  private async getGoRoot(goPath: string): Promise<string> {
    try {
      const result = await this.processManager.executeCommand({
        command: goPath,
        args: ['env', 'GOROOT'],
        timeout: 5000
      });

      return result.stdout?.trim() || '';
    } catch (error) {
      return '';
    }
  }

  /**
   * 获取Go路径
   */
  private async getGoPath(goPath: string): Promise<string> {
    try {
      const result = await this.processManager.executeCommand({
        command: goPath,
        args: ['env', 'GOPATH'],
        timeout: 5000
      });

      return result.stdout?.trim() || '';
    } catch (error) {
      return '';
    }
  }

  /**
   * 检查路径是否存在
   */
  private async pathExists(path: string): Promise<boolean> {
    try {
      await this.processManager.executeCommand({
        command: 'test',
        args: ['-f', path],
        timeout: 5000
      });
      return true;
    } catch (error) {
      return false;
    }
  }

  /**
   * 确保目录存在
   */
  private async ensureDirectoryExists(dirPath: string): Promise<void> {
    try {
      await this.processManager.executeCommand({
        command: 'mkdir',
        args: ['-p', dirPath],
        timeout: 5000
      });
    } catch (error) {
      // 目录可能已存在，忽略错误
    }
  }

  /**
   * 清理资源
   */
  async cleanup(serviceInfo?: ServiceInfo): Promise<void> {
    if (serviceInfo) {
      await this.stop(serviceInfo);
    } else {
      // 清理所有服务
      const stopPromises = Array.from(this.services.values()).map(
        service => this.stop(service).catch(error =>
          console.error(`Failed to stop service ${service.id}:`, error)
        )
      );
      await Promise.all(stopPromises);
    }

    this.environments.clear();
    this.services.clear();
    this.buildCache.clear();
  }
}