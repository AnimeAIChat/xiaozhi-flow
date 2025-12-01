import {
  BackendConfig,
  ServiceInfo,
  RuntimeAdapter,
  ServiceProxy
} from '../types';
import { SimpleEventEmitter } from '../utils/ProcessManager';
import { pluginManager } from './PluginManager';
import { PythonRuntimeAdapter } from '../runtime/PythonRuntimeAdapter';
import { GoRuntimeAdapter } from '../runtime/GoRuntimeAdapter';
import { DockerRuntimeAdapter } from '../runtime/DockerRuntimeAdapter';

interface ServiceEvent {
  type: 'started' | 'stopped' | 'error' | 'health-check';
  serviceId: string;
  pluginId: string;
  data?: any;
  timestamp: Date;
}

export class BackendServiceManager extends SimpleEventEmitter {
  private services: Map<string, ServiceInfo> = new Map();
  private runtimeAdapters: Map<string, RuntimeAdapter> = new Map();
  private serviceProxies: Map<string, ServiceProxy> = new Map();
  private healthCheckIntervals: Map<string, NodeJS.Timeout> = new Map();
  private portAllocator: PortAllocator;

  constructor() {
    super();
    this.portAllocator = new PortAllocator();
    this.registerDefaultRuntimeAdapters();
    this.setupEventHandlers();
  }

  /**
   * 注册默认运行时适配器
   */
  private registerDefaultRuntimeAdapters(): void {
    // 注册Python运行时适配器
    const pythonAdapter = new PythonRuntimeAdapter();
    this.registerRuntimeAdapter('python', pythonAdapter);

    // 注册Go运行时适配器
    const goAdapter = new GoRuntimeAdapter();
    this.registerRuntimeAdapter('go', goAdapter);

    // 注册Docker运行时适配器
    const dockerAdapter = new DockerRuntimeAdapter();
    this.registerRuntimeAdapter('docker', dockerAdapter);

    // 注册别名
    this.registerRuntimeAdapter('python3', pythonAdapter);
    this.registerRuntimeAdapter('golang', goAdapter);
  }

  /**
   * 注册运行时适配器
   */
  registerRuntimeAdapter(type: string, adapter: RuntimeAdapter): void {
    this.runtimeAdapters.set(type, adapter);
    this.emit('runtime-adapter-registered', { type, adapter });
  }

  /**
   * 启动后端服务
   */
  async startService(
    pluginId: string,
    serviceId: string,
    config: BackendConfig
  ): Promise<ServiceInfo> {
    const fullServiceId = `${pluginId}:${serviceId}`;

    try {
      this.emit('service-starting', { serviceId: fullServiceId, pluginId, config });

      // 检查服务是否已启动
      if (this.services.has(fullServiceId)) {
        const existingService = this.services.get(fullServiceId)!;
        if (existingService.status === 'running') {
          return existingService;
        }
      }

      // 获取运行时适配器
      const adapter = this.runtimeAdapters.get(config.runtime || 'python');
      if (!adapter) {
        throw new Error(`No runtime adapter found for runtime: ${config.runtime || 'python'}`);
      }

      // 分配端口
      const port = config.port || await this.portAllocator.allocatePort();

      // 准备配置
      const serviceConfig: BackendConfig = {
        ...config,
        port,
        envVars: {
          ...config.envVars,
          SERVICE_ID: serviceId,
          PLUGIN_ID: pluginId,
          PORT: port.toString()
        }
      };

      // 启动服务
      const serviceInfo: ServiceInfo = await adapter.start(serviceConfig);
      serviceInfo.id = serviceId;
      serviceInfo.pluginId = pluginId;
      serviceInfo.status = 'running';
      serviceInfo.startTime = new Date();
      serviceInfo.config = serviceConfig;

      // 存储服务信息
      this.services.set(fullServiceId, serviceInfo);

      // 创建服务代理
      const proxy = new ServiceProxy(serviceInfo);
      this.serviceProxies.set(fullServiceId, proxy);

      // 启动健康检查
      this.startHealthCheck(pluginId, serviceId, serviceInfo, serviceConfig);

      this.emit('service-started', { serviceId: fullServiceId, pluginId, serviceInfo });
      return serviceInfo;

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      this.emit('service-error', {
        serviceId: fullServiceId,
        pluginId,
        error: errorMessage
      });
      throw new Error(`Failed to start service ${fullServiceId}: ${errorMessage}`);
    }
  }

  /**
   * 停止后端服务
   */
  async stopService(pluginId: string, serviceId: string): Promise<void> {
    const fullServiceId = `${pluginId}:${serviceId}`;
    const serviceInfo = this.services.get(fullServiceId);

    if (!serviceInfo) {
      throw new Error(`Service ${fullServiceId} not found`);
    }

    try {
      this.emit('service-stopping', { serviceId: fullServiceId, pluginId });

      // 停止健康检查
      this.stopHealthCheck(fullServiceId);

      // 获取运行时适配器
      const adapter = this.runtimeAdapters.get(serviceInfo.runtime || 'python');
      if (adapter) {
        await adapter.stop(serviceInfo);
      }

      // 更新状态
      serviceInfo.status = 'stopped';
      serviceInfo.pid = undefined;

      // 移除服务代理
      this.serviceProxies.delete(fullServiceId);

      // 释放端口
      if (serviceInfo.port) {
        this.portAllocator.releasePort(serviceInfo.port);
      }

      this.emit('service-stopped', { serviceId: fullServiceId, pluginId });

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      this.emit('service-error', {
        serviceId: fullServiceId,
        pluginId,
        error: errorMessage
      });
      throw new Error(`Failed to stop service ${fullServiceId}: ${errorMessage}`);
    }
  }

  /**
   * 重启后端服务
   */
  async restartService(pluginId: string, serviceId: string, config?: BackendConfig): Promise<ServiceInfo> {
    const fullServiceId = `${pluginId}:${serviceId}`;
    const serviceInfo = this.services.get(fullServiceId);

    if (serviceInfo && config) {
      // 使用新配置重启
      await this.stopService(pluginId, serviceId);
      return await this.startService(pluginId, serviceId, config);
    } else if (serviceInfo) {
      // 使用原配置重启
      const adapter = this.runtimeAdapters.get(serviceInfo.runtime || 'python');
      if (adapter) {
        const restartedService = await adapter.restart(serviceInfo);
        restartedService.status = 'running';
        this.services.set(fullServiceId, restartedService);
        return restartedService;
      }
    }

    throw new Error(`Cannot restart service ${fullServiceId}: service not found or no configuration provided`);
  }

  /**
   * 获取服务信息
   */
  getService(pluginId: string, serviceId: string): ServiceInfo | undefined {
    const fullServiceId = `${pluginId}:${serviceId}`;
    return this.services.get(fullServiceId);
  }

  /**
   * 获取所有服务
   */
  getServices(): ServiceInfo[] {
    return Array.from(this.services.values());
  }

  /**
   * 获取插件的所有服务
   */
  getPluginServices(pluginId: string): ServiceInfo[] {
    return Array.from(this.services.values()).filter(
      service => service.pluginId === pluginId
    );
  }

  /**
   * 获取服务代理
   */
  getServiceProxy(pluginId: string, serviceId: string): ServiceProxy | undefined {
    const fullServiceId = `${pluginId}:${serviceId}`;
    return this.serviceProxies.get(fullServiceId);
  }

  /**
   * 检查服务状态
   */
  async checkServiceStatus(pluginId: string, serviceId: string): Promise<ServiceInfo> {
    const fullServiceId = `${pluginId}:${serviceId}`;
    const serviceInfo = this.services.get(fullServiceId);

    if (!serviceInfo) {
      throw new Error(`Service ${fullServiceId} not found`);
    }

    const adapter = this.runtimeAdapters.get(serviceInfo.runtime || 'python');
    if (adapter) {
      const updatedInfo = await adapter.getStatus(serviceInfo);
      this.services.set(fullServiceId, updatedInfo);
      return updatedInfo;
    }

    return serviceInfo;
  }

  /**
   * 执行健康检查
   */
  async performHealthCheck(pluginId: string, serviceId: string): Promise<boolean> {
    const fullServiceId = `${pluginId}:${serviceId}`;
    const serviceInfo = this.services.get(fullServiceId);

    if (!serviceInfo) {
      return false;
    }

    try {
      const adapter = this.runtimeAdapters.get(serviceInfo.runtime || 'python');
      if (adapter) {
        const isHealthy = await adapter.healthCheck(serviceInfo);

        // 更新健康状态
        serviceInfo.healthStatus = isHealthy ? 'healthy' : 'unhealthy';
        serviceInfo.lastHealthCheck = new Date();

        this.emit('service-health-checked', {
          serviceId: fullServiceId,
          pluginId,
          isHealthy,
          serviceInfo
        });

        return isHealthy;
      }

      return false;
    } catch (error) {
      serviceInfo.healthStatus = 'unhealthy';
      serviceInfo.lastHealthCheck = new Date();
      serviceInfo.error = error instanceof Error ? error.message : 'Health check failed';

      this.emit('service-health-check-failed', {
        serviceId: fullServiceId,
        pluginId,
        error: serviceInfo.error
      });

      return false;
    }
  }

  /**
   * 停止插件的所有服务
   */
  async stopPluginServices(pluginId: string): Promise<void> {
    const pluginServices = this.getPluginServices(pluginId);

    const stopPromises = pluginServices.map(service =>
      this.stopService(pluginId, service.id).catch(error => {
        console.error(`Failed to stop service ${service.id}:`, error);
      })
    );

    await Promise.all(stopPromises);
  }

  /**
   * 启动健康检查
   */
  private startHealthCheck(
    pluginId: string,
    serviceId: string,
    serviceInfo: ServiceInfo,
    config: BackendConfig
  ): void {
    const fullServiceId = `${pluginId}:${serviceId}`;

    if (!config.healthCheck) {
      return;
    }

    const interval = setInterval(async () => {
      try {
        await this.performHealthCheck(pluginId, serviceId);
      } catch (error) {
        console.error(`Health check failed for service ${fullServiceId}:`, error);
      }
    }, config.healthCheck.interval);

    this.healthCheckIntervals.set(fullServiceId, interval);
  }

  /**
   * 停止健康检查
   */
  private stopHealthCheck(fullServiceId: string): void {
    const interval = this.healthCheckIntervals.get(fullServiceId);
    if (interval) {
      clearInterval(interval);
      this.healthCheckIntervals.delete(fullServiceId);
    }
  }

  /**
   * 设置事件处理器
   */
  private setupEventHandlers(): void {
    // 监听插件卸载事件，停止相关服务
    pluginManager.on('plugin-unloaded', ({ plugin }) => {
      this.stopPluginServices(plugin.id).catch(error => {
        console.error(`Failed to stop services for plugin ${plugin.id}:`, error);
      });
    });

    // 监听插件停用事件，停止相关服务
    pluginManager.on('plugin-deactivated', ({ plugin }) => {
      this.stopPluginServices(plugin.id).catch(error => {
        console.error(`Failed to stop services for plugin ${plugin.id}:`, error);
      });
    });

    // 处理未捕获的错误
    this.on('error', (error) => {
      console.error('Backend Service Manager Error:', error);
    });
  }

  /**
   * 清理资源
   */
  async cleanup(): Promise<void> {
    // 停止所有健康检查
    for (const [serviceId, interval] of this.healthCheckIntervals) {
      clearInterval(interval);
    }
    this.healthCheckIntervals.clear();

    // 停止所有服务
    const stopPromises = Array.from(this.services.values()).map(service =>
      this.stopService(service.pluginId, service.id).catch(error => {
        console.error(`Failed to stop service ${service.id} during cleanup:`, error);
      })
    );

    await Promise.all(stopPromises);

    // 清理数据结构
    this.services.clear();
    this.serviceProxies.clear();
    this.portAllocator.cleanup();
  }
}

/**
 * 端口分配器
 */
class PortAllocator {
  private usedPorts: Set<number> = new Set();
  private portRange = { min: 8000, max: 9000 };

  /**
   * 分配端口
   */
  async allocatePort(): Promise<number> {
    for (let port = this.portRange.min; port <= this.portRange.max; port++) {
      if (!this.usedPorts.has(port) && await this.isPortAvailable(port)) {
        this.usedPorts.add(port);
        return port;
      }
    }

    throw new Error('No available ports in the specified range');
  }

  /**
   * 释放端口
   */
  releasePort(port: number): void {
    this.usedPorts.delete(port);
  }

  /**
   * 检查端口是否可用
   */
  private async isPortAvailable(port: number): Promise<boolean> {
    try {
      // 这里应该实现真正的端口检查逻辑
      // 例如尝试连接端口或检查系统端口占用情况
      return true; // 暂时返回true
    } catch {
      return false;
    }
  }

  /**
   * 清理
   */
  cleanup(): void {
    this.usedPorts.clear();
  }
}

/**
 * 服务代理类
 */
export class ServiceProxy {
  private baseUrl: string;
  private timeout: number;

  constructor(serviceInfo: ServiceInfo, timeout = 30000) {
    this.baseUrl = `http://${serviceInfo.host}:${serviceInfo.port}`;
    this.timeout = timeout;
  }

  /**
   * 调用服务API
   */
  async callService(endpoint: string, params: any = {}, options: RequestInit = {}): Promise<any> {
    const url = `${this.baseUrl}${endpoint}`;
    const controller = new AbortController();

    const timeoutId = setTimeout(() => {
      controller.abort();
    }, this.timeout);

    try {
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...options.headers
        },
        body: JSON.stringify(params),
        signal: controller.signal,
        ...options
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`Service responded with status: ${response.status}`);
      }

      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        return await response.json();
      } else {
        return await response.text();
      }

    } catch (error) {
      clearTimeout(timeoutId);

      if (error instanceof Error && error.name === 'AbortError') {
        throw new Error(`Service call timeout after ${this.timeout}ms`);
      }

      throw error;
    }
  }

  /**
   * 流式响应
   */
  async *streamResponse(endpoint: string, params: any = {}): AsyncIterable<any> {
    const url = `${this.baseUrl}${endpoint}`;
    const controller = new AbortController();

    const timeoutId = setTimeout(() => {
      controller.abort();
    }, this.timeout);

    try {
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(params),
        signal: controller.signal
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`Service responded with status: ${response.status}`);
      }

      if (!response.body) {
        throw new Error('Response body is null');
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value, { stream: true });
        const lines = chunk.split('\n').filter(line => line.trim());

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6);
            try {
              yield JSON.parse(data);
            } catch {
              yield data;
            }
          }
        }
      }

    } catch (error) {
      clearTimeout(timeoutId);

      if (error instanceof Error && error.name === 'AbortError') {
        throw new Error(`Stream response timeout after ${this.timeout}ms`);
      }

      throw error;
    }
  }

  /**
   * 健康检查
   */
  async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/health`, {
        method: 'GET',
        signal: AbortSignal.timeout(5000)
      });
      return response.ok;
    } catch {
      return false;
    }
  }
}

// 全局后端服务管理器实例
export const backendServiceManager = new BackendServiceManager();