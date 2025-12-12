import type { BackendConfig, RuntimeAdapter, ServiceInfo } from '../types';
import { ProcessManager, SimpleEventEmitter } from '../utils/ProcessManager';

interface PythonEnvironment {
  path: string;
  version: string;
  virtualEnv: string | null;
  packages: string[];
}

interface PythonServiceConfig extends BackendConfig {
  pythonPath?: string;
  requirementsPath?: string;
  virtualEnvPath?: string;
  pythonArgs?: string[];
  pipArgs?: string[];
}

export class PythonRuntimeAdapter
  extends SimpleEventEmitter
  implements RuntimeAdapter
{
  type = 'python';
  name = 'Python Runtime';
  version = '1.0.0';

  private processManager: ProcessManager;
  private environments: Map<string, PythonEnvironment> = new Map();
  private services: Map<string, ServiceInfo> = new Map();

  constructor() {
    super();
    this.processManager = new ProcessManager();
  }

  /**
   * 启动Python服务
   */
  async start(config: PythonServiceConfig): Promise<ServiceInfo> {
    try {
      this.emit('service-starting', { config });

      // 验证Python环境
      const pythonEnv = await this.validateEnvironment(config);

      // 准备虚拟环境
      const venvPath = await this.prepareVirtualEnvironment(config);

      // 安装依赖
      if (config.dependencies && config.dependencies.length > 0) {
        await this.installDependencies(config, venvPath);
      }

      // 启动Python进程
      const serviceInfo = await this.startPythonProcess(
        config,
        pythonEnv,
        venvPath,
      );

      // 等待服务启动
      await this.waitForServiceReady(
        serviceInfo,
        config.startupTimeout || 30000,
      );

      this.services.set(`${config.pluginId}:${serviceInfo.id}`, serviceInfo);
      this.emit('service-started', { serviceInfo });

      return serviceInfo;
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      this.emit('service-error', { config, error: errorMessage });
      throw new Error(`Failed to start Python service: ${errorMessage}`);
    }
  }

  /**
   * 停止Python服务
   */
  async stop(serviceInfo: ServiceInfo): Promise<void> {
    try {
      this.emit('service-stopping', { serviceInfo });

      const serviceKey = `${serviceInfo.pluginId}:${serviceInfo.id}`;
      const service = this.services.get(serviceKey);

      if (!service) {
        throw new Error('Service not found');
      }

      // 停止Python进程
      if (service.pid) {
        await this.processManager.killProcess(service.pid, 'SIGTERM');

        // 等待进程优雅退出
        await this.waitForProcessExit(service.pid, 5000);

        // 如果进程还在运行，强制杀死
        const isRunning = await this.processManager.isProcessRunning(
          service.pid,
        );
        if (isRunning) {
          await this.processManager.killProcess(service.pid, 'SIGKILL');
        }
      }

      this.services.delete(serviceKey);
      this.emit('service-stopped', { serviceInfo });
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      this.emit('service-error', { serviceInfo, error: errorMessage });
      throw new Error(`Failed to stop Python service: ${errorMessage}`);
    }
  }

  /**
   * 重启Python服务
   */
  async restart(serviceInfo: ServiceInfo): Promise<ServiceInfo> {
    const serviceKey = `${serviceInfo.pluginId}:${serviceInfo.id}`;
    const service = this.services.get(serviceKey);

    if (!service) {
      throw new Error('Service not found');
    }

    // 保存当前配置
    const currentConfig = service.config as PythonServiceConfig;

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
        const processInfo = await this.processManager.getProcessInfo(
          service.pid,
        );
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
        signal: AbortSignal.timeout(5000),
      });

      return response.ok;
    } catch (error) {
      return false;
    }
  }

  /**
   * 验证Python环境
   */
  async validateEnvironment(): Promise<PythonEnvironment> {
    const pythonPath = await this.findPythonExecutable();
    const version = await this.getPythonVersion(pythonPath);

    return {
      path: pythonPath,
      version,
      virtualEnv: null,
      packages: [],
    };
  }

  /**
   * 准备虚拟环境
   */
  private async prepareVirtualEnvironment(
    config: PythonServiceConfig,
  ): Promise<string> {
    const pluginId = config.pluginId || 'unknown';
    const venvPath = config.virtualEnvPath || `/tmp/python-venvs/${pluginId}`;

    // 检查虚拟环境是否存在
    const venvExists = await this.pathExists(venvPath);

    if (!venvExists) {
      // 创建虚拟环境
      await this.createVirtualEnvironment(venvPath);
    }

    return venvPath;
  }

  /**
   * 创建虚拟环境
   */
  private async createVirtualEnvironment(venvPath: string): Promise<void> {
    const pythonPath = await this.findPythonExecutable();

    await this.processManager.executeCommand({
      command: pythonPath,
      args: ['-m', 'venv', venvPath],
      cwd: '/',
      timeout: 60000, // 1分钟超时
    });
  }

  /**
   * 安装Python依赖
   */
  private async installDependencies(
    config: PythonServiceConfig,
    venvPath: string,
  ): Promise<void> {
    const pipPath = this.getPipPath(venvPath);

    for (const depFile of config.dependencies || []) {
      if (depFile.endsWith('.txt')) {
        // 安装requirements.txt文件
        await this.processManager.executeCommand({
          command: pipPath,
          args: ['install', '-r', depFile],
          cwd: config.workingDirectory || '/',
          env: {
            ...process.env,
            PYTHONPATH: venvPath + '/lib/python*/site-packages',
          },
          timeout: 300000, // 5分钟超时
        });
      } else {
        // 安装单个包
        await this.processManager.executeCommand({
          command: pipPath,
          args: ['install', depFile, ...(config.pipArgs || [])],
          cwd: config.workingDirectory || '/',
          env: {
            ...process.env,
            PYTHONPATH: venvPath + '/lib/python*/site-packages',
          },
          timeout: 180000, // 3分钟超时
        });
      }
    }
  }

  /**
   * 启动Python进程
   */
  private async startPythonProcess(
    config: PythonServiceConfig,
    pythonEnv: PythonEnvironment,
    venvPath: string,
  ): Promise<ServiceInfo> {
    const pythonPath = this.getPythonPath(venvPath);
    const entryPoint = config.entryPoint || 'main.py';

    const args = [
      entryPoint,
      '--port',
      config.port?.toString() || '0',
      '--host',
      '127.0.0.1',
      ...(config.pythonArgs || []),
    ];

    const env = {
      ...process.env,
      ...config.envVars,
      PYTHONPATH: config.workingDirectory || '.',
      VIRTUAL_ENV: venvPath,
      PATH: `${venvPath}/bin:${process.env.PATH}`,
    };

    const processInfo = await this.processManager.startProcess({
      command: pythonPath,
      args,
      cwd: config.workingDirectory || '/',
      env,
      stdio: ['pipe', 'pipe', 'pipe'],
    });

    const port = config.port || 0;
    const serviceInfo: ServiceInfo = {
      id: 'main',
      pluginId: config.pluginId || 'unknown',
      name: 'Python Service',
      status: 'starting',
      port,
      host: '127.0.0.1',
      baseUrl: `http://127.0.0.1:${port}`,
      pid: processInfo.pid,
      startTime: new Date(),
      runtime: 'python',
      healthStatus: 'unknown',
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
   * 等待服务就绪
   */
  private async waitForServiceReady(
    serviceInfo: ServiceInfo,
    timeout: number,
  ): Promise<void> {
    const startTime = Date.now();
    const maxTime = startTime + timeout;

    while (Date.now() < maxTime) {
      try {
        if (serviceInfo.port > 0) {
          const response = await fetch(`${serviceInfo.baseUrl}/health`, {
            method: 'GET',
            signal: AbortSignal.timeout(2000),
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

      await new Promise((resolve) => setTimeout(resolve, 1000));
    }

    throw new Error(`Service not ready after ${timeout}ms`);
  }

  /**
   * 等待进程退出
   */
  private async waitForProcessExit(
    pid: number,
    timeout: number,
  ): Promise<void> {
    const startTime = Date.now();
    const maxTime = startTime + timeout;

    while (Date.now() < maxTime) {
      const isRunning = await this.processManager.isProcessRunning(pid);
      if (!isRunning) {
        return;
      }
      await new Promise((resolve) => setTimeout(resolve, 100));
    }
  }

  /**
   * 查找Python可执行文件
   */
  private async findPythonExecutable(): Promise<string> {
    const candidates = [
      'python3',
      'python',
      '/usr/bin/python3',
      '/usr/bin/python',
    ];

    for (const candidate of candidates) {
      try {
        await this.processManager.executeCommand({
          command: candidate,
          args: ['--version'],
          timeout: 5000,
        });
        return candidate;
      } catch (error) {}
    }

    throw new Error('Python executable not found');
  }

  /**
   * 获取Python版本
   */
  private async getPythonVersion(pythonPath: string): Promise<string> {
    try {
      const result = await this.processManager.executeCommand({
        command: pythonPath,
        args: ['--version'],
        timeout: 5000,
      });

      const output = result.stdout || result.stderr || '';
      const match = output.match(/Python (\d+\.\d+\.\d+)/);
      return match ? match[1] : 'unknown';
    } catch (error) {
      return 'unknown';
    }
  }

  /**
   * 虚拟环境中的Python路径
   */
  private getPythonPath(venvPath: string): string {
    return `${venvPath}/bin/python`;
  }

  /**
   * 虚拟环境中的pip路径
   */
  private getPipPath(venvPath: string): string {
    return `${venvPath}/bin/pip`;
  }

  /**
   * 检查路径是否存在
   */
  private async pathExists(path: string): Promise<boolean> {
    try {
      await this.processManager.executeCommand({
        command: 'test',
        args: ['-d', path],
        timeout: 5000,
      });
      return true;
    } catch (error) {
      return false;
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
      const stopPromises = Array.from(this.services.values()).map((service) =>
        this.stop(service).catch((error) =>
          console.error(`Failed to stop service ${service.id}:`, error),
        ),
      );
      await Promise.all(stopPromises);
    }

    this.environments.clear();
    this.services.clear();
  }
}
