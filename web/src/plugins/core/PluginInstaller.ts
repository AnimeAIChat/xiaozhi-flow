import { IPlugin, PluginSource, PluginInfo, PluginDetails } from '../types';
import { SimpleEventEmitter } from '../utils/ProcessManager';

interface InstallProgress {
  stage: 'downloading' | 'extracting' | 'validating' | 'installing' | 'configuring' | 'completed';
  progress: number; // 0-100
  message: string;
  error?: string;
}

export class PluginInstaller extends SimpleEventEmitter {
  private pluginsDirectory = '/plugins';
  private tempDirectory = '/temp/plugins';
  private installedPlugins: Map<string, IPlugin> = new Map();

  constructor() {
    super();
    this.ensureDirectories();
  }

  /**
   * 安装插件
   */
  async install(source: PluginSource): Promise<IPlugin> {
    const installId = this.generateInstallId();

    try {
      this.emit('install-started', { installId, source });

      // 根据源类型选择安装方法
      let plugin: IPlugin;

      switch (source.type) {
        case 'local':
          plugin = await this.installFromLocal(source.localPath!, installId);
          break;
        case 'url':
          plugin = await this.installFromURL(source.url!, installId);
          break;
        case 'market':
          plugin = await this.installFromMarket(source.marketId!, installId);
          break;
        case 'registry':
          plugin = await this.installFromRegistry(source.registry!, installId);
          break;
        default:
          throw new Error(`Unsupported plugin source type: ${source.type}`);
      }

      // 验证插件
      await this.validatePlugin(plugin);

      // 检查是否已安装
      if (this.installedPlugins.has(plugin.id) && !source.options?.overwrite) {
        throw new Error(`Plugin ${plugin.id} is already installed. Use overwrite option to reinstall.`);
      }

      // 安装插件文件
      await this.installPluginFiles(plugin, installId);

      // 注册插件
      this.installedPlugins.set(plugin.id, plugin);

      // 保存安装记录
      await this.saveInstallRecord(plugin, source);

      this.emit('install-completed', { installId, plugin });
      return plugin;

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      this.emit('install-error', { installId, source, error: errorMessage });
      throw new Error(`Failed to install plugin: ${errorMessage}`);
    } finally {
      // 清理临时文件
      await this.cleanupTempFiles(installId);
    }
  }

  /**
   * 卸载插件
   */
  async uninstall(pluginId: string): Promise<void> {
    try {
      this.emit('uninstall-started', { pluginId });

      const plugin = this.installedPlugins.get(pluginId);
      if (!plugin) {
        throw new Error(`Plugin ${pluginId} is not installed`);
      }

      // 停止插件（如果正在运行）
      await this.stopPlugin(pluginId);

      // 删除插件文件
      await this.removePluginFiles(pluginId);

      // 移除安装记录
      await this.removeInstallRecord(pluginId);

      // 从已安装列表中移除
      this.installedPlugins.delete(pluginId);

      this.emit('uninstall-completed', { pluginId });

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      this.emit('uninstall-error', { pluginId, error: errorMessage });
      throw new Error(`Failed to uninstall plugin ${pluginId}: ${errorMessage}`);
    }
  }

  /**
   * 更新插件
   */
  async update(pluginId: string, source?: PluginSource): Promise<IPlugin> {
    try {
      this.emit('update-started', { pluginId });

      const currentPlugin = this.installedPlugins.get(pluginId);
      if (!currentPlugin) {
        throw new Error(`Plugin ${pluginId} is not installed`);
      }

      // 如果没有提供源，尝试从当前源更新
      const updateSource = source || await this.getUpdateSource(currentPlugin);

      // 备份当前插件
      const backup = await this.backupPlugin(currentPlugin);

      try {
        // 安装新版本
        const updatedPlugin = await this.install({
          ...updateSource,
          options: { ...updateSource.options, overwrite: true }
        });

        this.emit('update-completed', { pluginId, oldVersion: currentPlugin.version, newVersion: updatedPlugin.version });
        return updatedPlugin;

      } catch (error) {
        // 如果更新失败，恢复备份
        await this.restorePlugin(backup);
        throw error;
      }

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      this.emit('update-error', { pluginId, error: errorMessage });
      throw new Error(`Failed to update plugin ${pluginId}: ${errorMessage}`);
    }
  }

  /**
   * 获取已安装的插件列表
   */
  getInstalledPlugins(): IPlugin[] {
    return Array.from(this.installedPlugins.values());
  }

  /**
   * 检查插件是否已安装
   */
  isInstalled(pluginId: string): boolean {
    return this.installedPlugins.has(pluginId);
  }

  /**
   * 获取已安装的插件
   */
  getInstalledPlugin(pluginId: string): IPlugin | undefined {
    return this.installedPlugins.get(pluginId);
  }

  /**
   * 检查插件更新
   */
  async checkUpdates(pluginId: string): Promise<PluginInfo | null> {
    const plugin = this.installedPlugins.get(pluginId);
    if (!plugin) {
      throw new Error(`Plugin ${pluginId} is not installed`);
    }

    try {
      // 这里应该检查各个源的更新
      // 例如检查市场、npm注册表等
      const latestInfo = await this.getLatestPluginInfo(plugin);

      if (this.isNewerVersion(latestInfo.version, plugin.version)) {
        return latestInfo;
      }

      return null;

    } catch (error) {
      console.warn(`Failed to check updates for plugin ${pluginId}:`, error);
      return null;
    }
  }

  /**
   * 从本地安装插件
   */
  private async installFromLocal(localPath: string, installId: string): Promise<IPlugin> {
    this.reportProgress(installId, 'downloading', 0, 'Reading local plugin...');

    try {
      // 检查路径是否存在
      if (!await this.pathExists(localPath)) {
        throw new Error(`Local path does not exist: ${localPath}`);
      }

      // 读取插件配置文件
      const configPath = this.joinPath(localPath, 'plugin.json');
      if (!await this.pathExists(configPath)) {
        throw new Error(`Plugin configuration file not found: ${configPath}`);
      }

      this.reportProgress(installId, 'validating', 30, 'Validating plugin configuration...');

      const configContent = await this.readFile(configPath);
      const pluginConfig = JSON.parse(configContent);

      // 验证配置结构
      this.validatePluginConfig(pluginConfig);

      // 创建插件对象
      const plugin: IPlugin = {
        ...pluginConfig,
        // 确保必要字段存在
        id: pluginConfig.id,
        name: pluginConfig.name,
        version: pluginConfig.version,
        description: pluginConfig.description || '',
        author: pluginConfig.author || '',
        type: pluginConfig.type || 'frontend',
        runtime: pluginConfig.runtime || 'javascript',
        nodeDefinition: pluginConfig.nodeDefinition,
        metadata: {
          category: pluginConfig.category || 'General',
          color: pluginConfig.color || '#1890ff',
          tags: pluginConfig.tags || [],
          ...pluginConfig.metadata
        }
      };

      this.reportProgress(installId, 'installing', 60, 'Installing plugin files...');

      // 复制插件文件到临时目录
      const tempPluginPath = this.joinPath(this.tempDirectory, installId);
      await this.copyDirectory(localPath, tempPluginPath);

      // 验证插件结构
      await this.validatePluginStructure(tempPluginPath, plugin);

      this.reportProgress(installId, 'completed', 100, 'Plugin installation completed');

      return plugin;

    } catch (error) {
      throw new Error(`Failed to install from local path: ${error}`);
    }
  }

  /**
   * 从URL安装插件
   */
  private async installFromURL(url: string, installId: string): Promise<IPlugin> {
    this.reportProgress(installId, 'downloading', 0, 'Downloading plugin from URL...');

    try {
      // 下载插件
      const tempPluginPath = this.joinPath(this.tempDirectory, installId);
      await this.downloadPlugin(url, tempPluginPath, (progress) => {
        this.reportProgress(installId, 'downloading', progress, `Downloading plugin... ${progress}%`);
      });

      // 从下载的文件安装
      return await this.installFromLocal(tempPluginPath, installId);

    } catch (error) {
      throw new Error(`Failed to install from URL: ${error}`);
    }
  }

  /**
   * 从市场安装插件
   */
  private async installFromMarket(marketId: string, installId: string): Promise<IPlugin> {
    this.reportProgress(installId, 'downloading', 0, 'Fetching plugin from marketplace...');

    try {
      // 获取插件详情
      const pluginDetails = await this.getPluginFromMarket(marketId);

      this.reportProgress(installId, 'downloading', 20, 'Downloading plugin files...');

      // 下载插件文件
      const tempPluginPath = this.joinPath(this.tempDirectory, installId);
      await this.downloadPlugin(pluginDetails.downloadUrl, tempPluginPath, (progress) => {
        this.reportProgress(installId, 'downloading', 20 + progress * 0.6, `Downloading plugin... ${20 + progress * 0.6}%`);
      });

      // 从下载的文件安装
      const plugin = await this.installFromLocal(tempPluginPath, installId);

      // 添加市场信息
      plugin.metadata = {
        ...plugin.metadata,
        marketplaceId: marketId,
        downloads: pluginDetails.downloads,
        rating: pluginDetails.rating
      };

      return plugin;

    } catch (error) {
      throw new Error(`Failed to install from marketplace: ${error}`);
    }
  }

  /**
   * 从注册表安装插件
   */
  private async installFromRegistry(registry: { name: string; version?: string }, installId: string): Promise<IPlugin> {
    this.reportProgress(installId, 'downloading', 0, 'Fetching plugin from registry...');

    try {
      // 从npm注册表获取包
      const packageInfo = await this.getPackageFromRegistry(registry.name, registry.version);

      this.reportProgress(installId, 'downloading', 20, 'Downloading package...');

      // 下载包
      const tempPluginPath = this.joinPath(this.tempDirectory, installId);
      await this.downloadPackage(packageInfo.tarball, tempPluginPath, (progress) => {
        this.reportProgress(installId, 'downloading', 20 + progress * 0.6, `Downloading package... ${20 + progress * 0.6}%`);
      });

      // 从下载的包安装
      return await this.installFromLocal(tempPluginPath, installId);

    } catch (error) {
      throw new Error(`Failed to install from registry: ${error}`);
    }
  }

  /**
   * 验证插件配置
   */
  private validatePluginConfig(config: any): void {
    const requiredFields = ['id', 'name', 'version', 'nodeDefinition'];

    for (const field of requiredFields) {
      if (!config[field]) {
        throw new Error(`Missing required field: ${field}`);
      }
    }

    // 验证ID格式
    if (!/^[a-z0-9-]+$/.test(config.id)) {
      throw new Error('Plugin ID must contain only lowercase letters, numbers, and hyphens');
    }

    // 验证版本格式
    if (!/^\d+\.\d+\.\d+/.test(config.version)) {
      throw new Error('Plugin version must follow semantic versioning (e.g., 1.0.0)');
    }

    // 验证节点定义
    if (!config.nodeDefinition.parameters || !Array.isArray(config.nodeDefinition.parameters)) {
      throw new Error('Node definition must have parameters array');
    }
  }

  /**
   * 验证插件结构
   */
  private async validatePluginStructure(pluginPath: string, plugin: IPlugin): Promise<void> {
    // 检查必要文件
    const requiredFiles = ['plugin.json'];

    for (const file of requiredFiles) {
      const filePath = this.joinPath(pluginPath, file);
      if (!await this.pathExists(filePath)) {
        throw new Error(`Required file not found: ${file}`);
      }
    }

    // 检查后端配置
    if (plugin.backend) {
      const backendPath = this.joinPath(pluginPath, plugin.backend.entryPoint);
      if (!await this.pathExists(backendPath)) {
        throw new Error(`Backend entry point not found: ${plugin.backend.entryPoint}`);
      }

      // 检查依赖文件
      if (plugin.backend.dependencies) {
        for (const depFile of plugin.backend.dependencies) {
          const depPath = this.joinPath(pluginPath, depFile);
          if (!await this.pathExists(depPath)) {
            throw new Error(`Dependency file not found: ${depFile}`);
          }
        }
      }
    }
  }

  /**
   * 安装插件文件
   */
  private async installPluginFiles(plugin: IPlugin, installId: string): Promise<void> {
    const tempPluginPath = this.joinPath(this.tempDirectory, installId);
    const pluginPath = this.joinPath(this.pluginsDirectory, plugin.id);

    this.reportProgress(installId, 'configuring', 80, 'Installing plugin files...');

    // 如果插件已存在，先删除
    if (await this.pathExists(pluginPath)) {
      await this.removeDirectory(pluginPath);
    }

    // 移动插件文件到最终位置
    await this.moveDirectory(tempPluginPath, pluginPath);

    // 设置权限
    await this.setPluginPermissions(pluginPath, plugin);
  }

  /**
   * 保存安装记录
   */
  private async saveInstallRecord(plugin: IPlugin, source: PluginSource): Promise<void> {
    const record = {
      pluginId: plugin.id,
      name: plugin.name,
      version: plugin.version,
      installedAt: new Date().toISOString(),
      source,
      plugin
    };

    const recordsPath = this.joinPath(this.pluginsDirectory, 'install-records.json');
    let records: any[] = [];

    if (await this.pathExists(recordsPath)) {
      const content = await this.readFile(recordsPath);
      records = JSON.parse(content);
    }

    // 移除旧记录（如果存在）
    records = records.filter((r: any) => r.pluginId !== plugin.id);
    records.push(record);

    await this.writeFile(recordsPath, JSON.stringify(records, null, 2));
  }

  /**
   * 移除安装记录
   */
  private async removeInstallRecord(pluginId: string): Promise<void> {
    const recordsPath = this.joinPath(this.pluginsDirectory, 'install-records.json');

    if (!await this.pathExists(recordsPath)) {
      return;
    }

    const content = await this.readFile(recordsPath);
    const records = JSON.parse(content);

    const filteredRecords = records.filter((r: any) => r.pluginId !== pluginId);

    await this.writeFile(recordsPath, JSON.stringify(filteredRecords, null, 2));
  }

  /**
   * 报告安装进度
   */
  private reportProgress(installId: string, stage: InstallProgress['stage'], progress: number, message: string): void {
    const progressInfo: InstallProgress = { stage, progress, message };
    this.emit('install-progress', { installId, progress: progressInfo });
  }

  /**
   * 生成安装ID
   */
  private generateInstallId(): string {
    return `install-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * 确保目录存在
   */
  private async ensureDirectories(): Promise<void> {
    await this.ensureDirectory(this.pluginsDirectory);
    await this.ensureDirectory(this.tempDirectory);
  }

  /**
   * 清理临时文件
   */
  private async cleanupTempFiles(installId: string): Promise<void> {
    const tempPath = this.joinPath(this.tempDirectory, installId);
    if (await this.pathExists(tempPath)) {
      await this.removeDirectory(tempPath);
    }
  }

  /**
   * 停止插件
   */
  private async stopPlugin(pluginId: string): Promise<void> {
    // 这里应该调用插件管理器来停止插件
    this.emit('plugin-stop-requested', { pluginId });
  }

  /**
   * 删除插件文件
   */
  private async removePluginFiles(pluginId: string): Promise<void> {
    const pluginPath = this.joinPath(this.pluginsDirectory, pluginId);
    if (await this.pathExists(pluginPath)) {
      await this.removeDirectory(pluginPath);
    }
  }

  /**
   * 检查版本是否更新
   */
  private isNewerVersion(newVersion: string, currentVersion: string): boolean {
    const newParts = newVersion.split('.').map(Number);
    const currentParts = currentVersion.split('.').map(Number);

    for (let i = 0; i < Math.max(newParts.length, currentParts.length); i++) {
      const newPart = newParts[i] || 0;
      const currentPart = currentParts[i] || 0;

      if (newPart > currentPart) return true;
      if (newPart < currentPart) return false;
    }

    return false;
  }

  /**
   * 获取更新源
   */
  private async getUpdateSource(plugin: IPlugin): Promise<PluginSource> {
    // 这里应该根据插件安装记录来确定更新源
    // 暂时返回市场源
    return {
      type: 'market',
      marketId: plugin.id
    };
  }

  /**
   * 备份插件
   */
  private async backupPlugin(plugin: IPlugin): Promise<any> {
    // 实现插件备份逻辑
    return { plugin, timestamp: Date.now() };
  }

  /**
   * 恢复插件
   */
  private async restorePlugin(backup: any): Promise<void> {
    // 实现插件恢复逻辑
    console.log('Restoring plugin from backup:', backup);
  }

  // 以下方法需要在实际环境中实现或调用相应的API

  private async pathExists(path: string): Promise<boolean> {
    // 实现路径检查
    return true; // 暂时返回true
  }

  private async readFile(path: string): Promise<string> {
    // 实现文件读取
    return '{}'; // 暂时返回空对象
  }

  private async writeFile(path: string, content: string): Promise<void> {
    // 实现文件写入
  }

  private async ensureDirectory(path: string): Promise<void> {
    // 实现目录创建
  }

  private async copyDirectory(src: string, dest: string): Promise<void> {
    // 实现目录复制
  }

  private async moveDirectory(src: string, dest: string): Promise<void> {
    // 实现目录移动
  }

  private async removeDirectory(path: string): Promise<void> {
    // 实现目录删除
  }

  private async setPluginPermissions(path: string, plugin: IPlugin): Promise<void> {
    // 实现权限设置
  }

  private joinPath(...paths: string[]): string {
    // 实现路径连接
    return paths.join('/');
  }

  private async downloadPlugin(url: string, dest: string, onProgress?: (progress: number) => void): Promise<void> {
    // 实现插件下载
    console.log('Downloading plugin from:', url, 'to:', dest);
  }

  private async downloadPackage(tarball: string, dest: string, onProgress?: (progress: number) => void): Promise<void> {
    // 实现包下载
    console.log('Downloading package from:', tarball, 'to:', dest);
  }

  private async getPluginFromMarket(marketId: string): Promise<any> {
    // 实现从市场获取插件信息
    return { downloadUrl: '', downloads: 0, rating: 0 };
  }

  private async getPackageFromRegistry(name: string, version?: string): Promise<any> {
    // 实现从注册表获取包信息
    return { tarball: '' };
  }

  private async getLatestPluginInfo(plugin: IPlugin): Promise<PluginInfo> {
    // 实现获取最新插件信息
    return {
      id: plugin.id,
      name: plugin.name,
      description: plugin.description,
      version: plugin.version,
      author: plugin.author,
      downloads: 0,
      rating: 0,
      tags: plugin.metadata.tags,
      lastUpdated: new Date()
    };
  }
}

// 全局插件安装器实例
export const pluginInstaller = new PluginInstaller();