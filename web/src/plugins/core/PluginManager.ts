import {
  type IPlugin,
  type PluginContext,
  PluginEvent,
  type PluginFilter,
  type PluginSource,
  type PluginStatus,
  type RuntimeAdapter,
  type ServiceInfo,
} from '../types';
import { SimpleEventEmitter } from '../utils/ProcessManager';

export class PluginManager extends SimpleEventEmitter {
  private plugins: Map<string, IPlugin> = new Map();
  private pluginStatuses: Map<string, PluginStatus> = new Map();
  private pluginContexts: Map<string, PluginContext> = new Map();
  private runtimeAdapters: Map<string, RuntimeAdapter> = new Map();
  private services: Map<string, ServiceInfo> = new Map();

  constructor() {
    super();
    this.setupEventHandlers();
  }

  /**
   * 注册运行时适配器
   */
  registerRuntimeAdapter(type: string, adapter: RuntimeAdapter): void {
    this.runtimeAdapters.set(type, adapter);
    this.emit('runtime-registered', { type, adapter });
  }

  /**
   * 从源加载插件
   */
  async loadPlugin(source: PluginSource): Promise<IPlugin> {
    try {
      this.emit('plugin-loading', { source });

      // 根据源类型加载插件
      const plugin = await this.loadPluginFromSource(source);

      // 验证插件
      this.validatePlugin(plugin);

      // 存储插件
      this.plugins.set(plugin.id, plugin);

      // 初始化插件状态
      const status: PluginStatus = {
        id: plugin.id,
        status: 'loaded',
        loadedAt: new Date(),
        services: [],
      };
      this.pluginStatuses.set(plugin.id, status);

      // 创建插件上下文
      const context = this.createPluginContext(plugin);
      this.pluginContexts.set(plugin.id, context);

      // 调用插件加载钩子
      if (plugin.onLoad) {
        await plugin.onLoad(context);
      }

      this.emit('plugin-loaded', { plugin });
      return plugin;
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      this.emit('plugin-error', { source, error: errorMessage });
      throw new Error(`Failed to load plugin: ${errorMessage}`);
    }
  }

  /**
   * 卸载插件
   */
  async unloadPlugin(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId);
    const status = this.pluginStatuses.get(pluginId);

    if (!plugin) {
      throw new Error(`Plugin ${pluginId} not found`);
    }

    try {
      this.emit('plugin-unloading', { plugin });

      // 如果插件处于活动状态，先停用
      if (status?.status === 'active') {
        await this.deactivatePlugin(pluginId);
      }

      // 调用插件卸载钩子
      if (plugin.onUnload) {
        await plugin.onUnload();
      }

      // 清理资源
      await this.cleanupPlugin(pluginId);

      // 移除插件
      this.plugins.delete(pluginId);
      this.pluginStatuses.delete(pluginId);
      this.pluginContexts.delete(pluginId);

      this.emit('plugin-unloaded', { plugin });
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      this.emit('plugin-error', { plugin, error: errorMessage });
      throw new Error(`Failed to unload plugin ${pluginId}: ${errorMessage}`);
    }
  }

  /**
   * 激活插件
   */
  async activatePlugin(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId);
    const status = this.pluginStatuses.get(pluginId);

    if (!plugin) {
      throw new Error(`Plugin ${pluginId} not found`);
    }

    if (status?.status === 'active') {
      return; // 已经激活
    }

    try {
      this.emit('plugin-activating', { plugin });

      // 更新状态
      this.updatePluginStatus(pluginId, 'activating');

      // 启动后端服务（如果有）
      if (plugin.backend) {
        await this.startBackendService(pluginId);
      }

      // 调用插件激活钩子
      if (plugin.onActivate) {
        const context = this.pluginContexts.get(pluginId);
        if (context) {
          await plugin.onActivate();
        }
      }

      // 更新状态为激活
      this.updatePluginStatus(pluginId, 'active', {
        activatedAt: new Date(),
      });

      this.emit('plugin-activated', { plugin });
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      this.updatePluginStatus(pluginId, 'error', { error: errorMessage });
      this.emit('plugin-error', { plugin, error: errorMessage });
      throw new Error(`Failed to activate plugin ${pluginId}: ${errorMessage}`);
    }
  }

  /**
   * 停用插件
   */
  async deactivatePlugin(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId);
    const status = this.pluginStatuses.get(pluginId);

    if (!plugin) {
      throw new Error(`Plugin ${pluginId} not found`);
    }

    if (status?.status !== 'active') {
      return; // 已经停用
    }

    try {
      this.emit('plugin-deactivating', { plugin });

      // 更新状态
      this.updatePluginStatus(pluginId, 'deactivating');

      // 调用插件停用钩子
      if (plugin.onDeactivate) {
        await plugin.onDeactivate();
      }

      // 停止后端服务（如果有）
      if (plugin.backend) {
        await this.stopBackendService(pluginId);
      }

      // 更新状态为未激活
      this.updatePluginStatus(pluginId, 'inactive');

      this.emit('plugin-deactivated', { plugin });
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      this.updatePluginStatus(pluginId, 'error', { error: errorMessage });
      this.emit('plugin-error', { plugin, error: errorMessage });
      throw new Error(
        `Failed to deactivate plugin ${pluginId}: ${errorMessage}`,
      );
    }
  }

  /**
   * 获取插件
   */
  getPlugin(pluginId: string): IPlugin | undefined {
    return this.plugins.get(pluginId);
  }

  /**
   * 获取插件状态
   */
  getPluginStatus(pluginId: string): PluginStatus | undefined {
    return this.pluginStatuses.get(pluginId);
  }

  /**
   * 获取所有插件
   */
  getPlugins(filter?: PluginFilter): IPlugin[] {
    let plugins = Array.from(this.plugins.values());

    if (filter) {
      plugins = plugins.filter((plugin) => {
        if (filter.type && plugin.type !== filter.type) return false;
        if (filter.runtime && plugin.runtime !== filter.runtime) return false;
        if (filter.category && plugin.metadata.category !== filter.category)
          return false;
        if (filter.author && plugin.author !== filter.author) return false;
        if (
          filter.tags &&
          !filter.tags.some((tag) => plugin.metadata.tags.includes(tag))
        )
          return false;

        if (filter.enabled !== undefined) {
          const status = this.pluginStatuses.get(plugin.id);
          const isLoaded = status?.status !== undefined;
          if (filter.enabled !== isLoaded) return false;
        }

        if (filter.active !== undefined) {
          const status = this.pluginStatuses.get(plugin.id);
          const isActive = status?.status === 'active';
          if (filter.active !== isActive) return false;
        }

        return true;
      });
    }

    return plugins;
  }

  /**
   * 获取所有插件状态
   */
  getPluginStatuses(): PluginStatus[] {
    return Array.from(this.pluginStatuses.values());
  }

  /**
   * 重新加载插件
   */
  async reloadPlugin(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId);
    if (!plugin) {
      throw new Error(`Plugin ${pluginId} not found`);
    }

    const source = this.extractPluginSource(plugin);
    await this.unloadPlugin(pluginId);
    await this.loadPlugin(source);
  }

  /**
   * 启动后端服务
   */
  private async startBackendService(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId);
    if (!plugin?.backend) return;

    const adapter = this.runtimeAdapters.get(plugin.runtime);
    if (!adapter) {
      throw new Error(`No runtime adapter found for ${plugin.runtime}`);
    }

    try {
      const serviceInfo = await adapter.start(plugin.backend);
      serviceInfo.pluginId = pluginId;

      this.services.set(`${pluginId}-${serviceInfo.id}`, serviceInfo);

      // 更新插件状态
      const status = this.pluginStatuses.get(pluginId);
      if (status) {
        status.services.push(serviceInfo);
      }

      this.emit('service-started', { pluginId, serviceInfo });
    } catch (error) {
      throw new Error(`Failed to start backend service: ${error}`);
    }
  }

  /**
   * 停止后端服务
   */
  private async stopBackendService(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId);
    if (!plugin?.backend) return;

    const adapter = this.runtimeAdapters.get(plugin.runtime);
    if (!adapter) return;

    const status = this.pluginStatuses.get(pluginId);
    if (!status) return;

    // 停止所有相关服务
    for (const serviceInfo of status.services) {
      try {
        await adapter.stop(serviceInfo);
        this.services.delete(`${pluginId}-${serviceInfo.id}`);
        this.emit('service-stopped', { pluginId, serviceInfo });
      } catch (error) {
        console.error(`Failed to stop service ${serviceInfo.id}:`, error);
      }
    }

    // 清空服务列表
    status.services = [];
  }

  /**
   * 从源加载插件
   */
  private async loadPluginFromSource(source: PluginSource): Promise<IPlugin> {
    switch (source.type) {
      case 'local':
        return this.loadFromLocal(source.localPath!);
      case 'url':
        return this.loadFromURL(source.url!);
      case 'market':
        return this.loadFromMarket(source.marketId!);
      case 'registry':
        return this.loadFromRegistry(source.registry!);
      default:
        throw new Error(`Unsupported plugin source type: ${source.type}`);
    }
  }

  /**
   * 从本地加载插件
   */
  private async loadFromLocal(localPath: string): Promise<IPlugin> {
    // 这里需要实现从本地文件系统加载插件的逻辑
    // 读取 plugin.json 配置文件，验证结构等
    throw new Error('Local plugin loading not implemented yet');
  }

  /**
   * 从URL加载插件
   */
  private async loadFromURL(url: string): Promise<IPlugin> {
    // 这里需要实现从URL下载和加载插件的逻辑
    throw new Error('URL plugin loading not implemented yet');
  }

  /**
   * 从市场加载插件
   */
  private async loadFromMarket(marketId: string): Promise<IPlugin> {
    // 这里需要实现从插件市场加载插件的逻辑
    throw new Error('Market plugin loading not implemented yet');
  }

  /**
   * 从注册表加载插件
   */
  private async loadFromRegistry(registry: {
    name: string;
    version?: string;
  }): Promise<IPlugin> {
    // 这里需要实现从npm注册表等加载插件的逻辑
    throw new Error('Registry plugin loading not implemented yet');
  }

  /**
   * 验证插件
   */
  private validatePlugin(plugin: IPlugin): void {
    if (!plugin.id || !plugin.name || !plugin.version) {
      throw new Error('Plugin must have id, name, and version');
    }

    if (!plugin.nodeDefinition || !plugin.nodeDefinition.parameters) {
      throw new Error('Plugin must have node definition with parameters');
    }

    // 检查ID是否已存在
    if (this.plugins.has(plugin.id)) {
      throw new Error(`Plugin with id ${plugin.id} already exists`);
    }
  }

  /**
   * 创建插件上下文
   */
  private createPluginContext(plugin: IPlugin): PluginContext {
    return {
      pluginId: plugin.id,
      workingDirectory: `/plugins/${plugin.id}`,
      api: this.createPluginAPI(plugin),
      storage: this.createPluginStorage(plugin.id),
      logger: this.createPluginLogger(plugin.id),
    };
  }

  /**
   * 创建插件API
   */
  private createPluginAPI(plugin: IPlugin) {
    return {
      registerNode: (definition: any) => {
        this.emit('node-registered', { pluginId: plugin.id, definition });
      },
      unregisterNode: (nodeId: string) => {
        this.emit('node-unregistered', { pluginId: plugin.id, nodeId });
      },
      registerService: (serviceId: string, config: any) => {
        this.emit('service-registered', {
          pluginId: plugin.id,
          serviceId,
          config,
        });
      },
      unregisterService: (serviceId: string) => {
        this.emit('service-unregistered', { pluginId: plugin.id, serviceId });
      },
      emit: (event: string, data: any) => {
        this.emit(`plugin:${plugin.id}:${event}`, data);
      },
      on: (event: string, handler: (data: any) => void) => {
        this.on(`plugin:${plugin.id}:${event}`, handler);
      },
      off: (event: string, handler: (data: any) => void) => {
        this.off(`plugin:${plugin.id}:${event}`, handler);
      },
      getConfig: (key: string) => {
        // 从全局配置中获取插件配置
        return this.getPluginConfig(plugin.id, key);
      },
      setConfig: (key: string, value: any) => {
        // 设置插件配置
        this.setPluginConfig(plugin.id, key, value);
      },
      showNotification: (
        message: string,
        type: 'info' | 'success' | 'warning' | 'error',
      ) => {
        this.emit('notification', { message, type, pluginId: plugin.id });
      },
    };
  }

  /**
   * 创建插件存储
   */
  private createPluginStorage(pluginId: string) {
    // 这里应该实现持久化存储，可以使用 localStorage 或 IndexedDB
    return {
      get: async (key: string) => {
        return localStorage.getItem(`plugin:${pluginId}:${key}`);
      },
      set: async (key: string, value: any) => {
        localStorage.setItem(
          `plugin:${pluginId}:${key}`,
          JSON.stringify(value),
        );
      },
      delete: async (key: string) => {
        localStorage.removeItem(`plugin:${pluginId}:${key}`);
      },
      clear: async () => {
        const keys = Object.keys(localStorage).filter((k) =>
          k.startsWith(`plugin:${pluginId}:`),
        );
        keys.forEach((k) => localStorage.removeItem(k));
      },
      keys: async () => {
        return Object.keys(localStorage).filter((k) =>
          k.startsWith(`plugin:${pluginId}:`),
        );
      },
    };
  }

  /**
   * 创建插件日志
   */
  private createPluginLogger(pluginId: string) {
    return {
      debug: (message: string, ...args: any[]) => {
        console.debug(`[Plugin:${pluginId}] ${message}`, ...args);
      },
      info: (message: string, ...args: any[]) => {
        console.info(`[Plugin:${pluginId}] ${message}`, ...args);
      },
      warn: (message: string, ...args: any[]) => {
        console.warn(`[Plugin:${pluginId}] ${message}`, ...args);
      },
      error: (message: string, ...args: any[]) => {
        console.error(`[Plugin:${pluginId}] ${message}`, ...args);
      },
    };
  }

  /**
   * 更新插件状态
   */
  private updatePluginStatus(
    pluginId: string,
    status: string,
    updates?: Partial<PluginStatus>,
  ): void {
    const currentStatus = this.pluginStatuses.get(pluginId);
    if (currentStatus) {
      currentStatus.status = status as any;
      if (updates) {
        Object.assign(currentStatus, updates);
      }
      this.emit('plugin-status-changed', { pluginId, status: currentStatus });
    }
  }

  /**
   * 清理插件资源
   */
  private async cleanupPlugin(pluginId: string): Promise<void> {
    // 停止所有相关服务
    await this.stopBackendService(pluginId);

    // 移除所有事件监听器
    this.removeAllListeners(`plugin:${pluginId}:*`);

    // 清理其他资源...
  }

  /**
   * 从插件提取源信息
   */
  private extractPluginSource(plugin: IPlugin): PluginSource {
    // 这里应该根据插件信息推断原始源
    // 暂时返回本地源
    return {
      type: 'local',
      localPath: `/plugins/${plugin.id}`,
    };
  }

  /**
   * 获取插件配置
   */
  private getPluginConfig(pluginId: string, key: string): any {
    const config = localStorage.getItem(`plugin-config:${pluginId}`);
    if (config) {
      try {
        const parsed = JSON.parse(config);
        return parsed[key];
      } catch {
        return undefined;
      }
    }
    return undefined;
  }

  /**
   * 设置插件配置
   */
  private setPluginConfig(pluginId: string, key: string, value: any): void {
    const configKey = `plugin-config:${pluginId}`;
    const existing = localStorage.getItem(configKey);
    const config = existing ? JSON.parse(existing) : {};
    config[key] = value;
    localStorage.setItem(configKey, JSON.stringify(config));
  }

  /**
   * 设置事件处理器
   */
  private setupEventHandlers(): void {
    // 处理未捕获的插件错误
    this.on('error', (error) => {
      console.error('Plugin Manager Error:', error);
    });
  }

  /**
   * 获取服务信息
   */
  getService(serviceId: string): ServiceInfo | undefined {
    return this.services.get(serviceId);
  }

  /**
   * 获取所有服务
   */
  getServices(): ServiceInfo[] {
    return Array.from(this.services.values());
  }

  /**
   * 获取插件的服务
   */
  getPluginServices(pluginId: string): ServiceInfo[] {
    return Array.from(this.services.values()).filter(
      (service) => service.pluginId === pluginId,
    );
  }
}

// 全局插件管理器实例
export const pluginManager = new PluginManager();
