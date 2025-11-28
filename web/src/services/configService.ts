/**
 * 配置管理服务
 * 处理配置记录的CRUD操作、画布状态管理、快照等功能
 */

import { apiService } from './api';
import { log } from '../utils/logger';
import type {
  ConfigRecord,
  ConfigCategory,
  ConfigCanvasState,
  ConfigNode,
  ConfigEdge,
  ConfigFilter,
  ConfigSnapshot,
  ConfigTemplate,
  ConfigValidation,
  ConfigExport,
  ConfigUpdateOperation,
} from '../types/config';

export class ConfigService {
  private baseUrl = '/admin/config';

  /**
   * 获取所有配置记录
   */
  async getConfigs(filter?: ConfigFilter): Promise<ConfigRecord[]> {
    try {
      const params = new URLSearchParams();
      if (filter?.category) params.append('category', filter.category);
      if (filter?.isActive !== undefined) params.append('isActive', String(filter.isActive));
      if (filter?.searchText) params.append('search', filter.searchText);

      log.info('获取配置记录', { filter }, 'config', 'ConfigService');

      const response = await apiService.client.get(`${this.baseUrl}/records`, { params });
      return response.data.data || [];
    } catch (error) {
      log.error('获取配置记录失败', error, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 获取单个配置记录
   */
  async getConfig(key: string): Promise<ConfigRecord | null> {
    try {
      log.debug('获取单个配置', { key }, 'config', 'ConfigService');

      const response = await apiService.client.get(`${this.baseUrl}/records/${key}`);
      return response.data.data;
    } catch (error) {
      log.warn('获取配置失败', { key, error }, 'config', 'ConfigService');
      return null;
    }
  }

  /**
   * 创建配置记录
   */
  async createConfig(config: Omit<ConfigRecord, 'id' | 'created_at' | 'updated_at'>): Promise<ConfigRecord> {
    try {
      log.info('创建配置记录', { key: config.key, category: config.category }, 'config', 'ConfigService');

      const response = await apiService.client.post(`${this.baseUrl}/records`, config);
      return response.data.data;
    } catch (error) {
      log.error('创建配置记录失败', { config, error }, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 更新配置记录
   */
  async updateConfig(key: string, updates: Partial<ConfigRecord>): Promise<ConfigRecord> {
    try {
      log.info('更新配置记录', { key, updates }, 'config', 'ConfigService');

      const response = await apiService.client.put(`${this.baseUrl}/records/${key}`, updates);
      return response.data.data;
    } catch (error) {
      log.error('更新配置记录失败', { key, updates, error }, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 删除配置记录
   */
  async deleteConfig(key: string): Promise<void> {
    try {
      log.warn('删除配置记录', { key }, 'config', 'ConfigService');

      await apiService.client.delete(`${this.baseUrl}/records/${key}`);
    } catch (error) {
      log.error('删除配置记录失败', { key, error }, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 获取配置分类
   */
  async getConfigCategories(): Promise<ConfigCategory[]> {
    try {
      log.debug('获取配置分类', null, 'config', 'ConfigService');

      const response = await apiService.client.get(`${this.baseUrl}/categories`);
      return response.data.data || [];
    } catch (error) {
      log.error('获取配置分类失败', error, 'config', 'ConfigService');
      return [];
    }
  }

  /**
   * 批量更新配置
   */
  async batchUpdateConfigs(updates: Array<{ key: string; value: any }>): Promise<void> {
    try {
      log.info('批量更新配置', { count: updates.length }, 'config', 'ConfigService');

      await apiService.client.post(`${this.baseUrl}/batch-update`, { updates });
    } catch (error) {
      log.error('批量更新配置失败', { updates, error }, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 验证配置
   */
  async validateConfigs(configs?: ConfigRecord[]): Promise<ConfigValidation> {
    try {
      log.info('验证配置', { count: configs?.length }, 'config', 'ConfigService');

      const response = await apiService.client.post(`${this.baseUrl}/validate`, { configs });
      return response.data.data;
    } catch (error) {
      log.error('配置验证失败', error, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 获取配置画布状态
   */
  async getCanvasState(): Promise<ConfigCanvasState> {
    try {
      log.debug('获取配置画布状态', null, 'config', 'ConfigService');

      const response = await apiService.client.get(`${this.baseUrl}/canvas`);
      return response.data.data || this.getDefaultCanvasState();
    } catch (error) {
      log.warn('获取画布状态失败，使用默认状态', error, 'config', 'ConfigService');
      return this.getDefaultCanvasState();
    }
  }

  /**
   * 保存配置画布状态
   */
  async saveCanvasState(state: ConfigCanvasState): Promise<void> {
    try {
      log.info('保存配置画布状态', {
        nodeCount: state.nodes.length,
        edgeCount: state.edges.length
      }, 'config', 'ConfigService');

      await apiService.client.post(`${this.baseUrl}/canvas`, state);
    } catch (error) {
      log.error('保存画布状态失败', error, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 创建配置快照
   */
  async createSnapshot(name: string, description?: string): Promise<ConfigSnapshot> {
    try {
      log.info('创建配置快照', { name, description }, 'config', 'ConfigService');

      const response = await apiService.client.post(`${this.baseUrl}/snapshots`, {
        name,
        description,
      });
      return response.data.data;
    } catch (error) {
      log.error('创建配置快照失败', { name, error }, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 获取配置快照列表
   */
  async getSnapshots(): Promise<ConfigSnapshot[]> {
    try {
      log.debug('获取配置快照列表', null, 'config', 'ConfigService');

      const response = await apiService.client.get(`${this.baseUrl}/snapshots`);
      return response.data.data || [];
    } catch (error) {
      log.error('获取快照列表失败', error, 'config', 'ConfigService');
      return [];
    }
  }

  /**
   * 恢复配置快照
   */
  async restoreSnapshot(snapshotId: string): Promise<void> {
    try {
      log.info('恢复配置快照', { snapshotId }, 'config', 'ConfigService');

      await apiService.client.post(`${this.baseUrl}/snapshots/${snapshotId}/restore`);
    } catch (error) {
      log.error('恢复配置快照失败', { snapshotId, error }, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 导出配置
   */
  async exportConfig(filter?: ConfigFilter): Promise<ConfigExport> {
    try {
      log.info('导出配置', { filter }, 'config', 'ConfigService');

      const response = await apiService.client.post(`${this.baseUrl}/export`, { filter });
      return response.data.data;
    } catch (error) {
      log.error('导出配置失败', error, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 导入配置
   */
  async importConfig(exportData: ConfigExport, options?: { overwrite?: boolean; validateOnly?: boolean }): Promise<any> {
    try {
      log.info('导入配置', {
        configCount: exportData.configs.length,
        options
      }, 'config', 'ConfigService');

      const response = await apiService.client.post(`${this.baseUrl}/import`, {
        exportData,
        options,
      });
      return response.data.data;
    } catch (error) {
      log.error('导入配置失败', error, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 获取配置模板
   */
  async getTemplates(): Promise<ConfigTemplate[]> {
    try {
      log.debug('获取配置模板', null, 'config', 'ConfigService');

      const response = await apiService.client.get(`${this.baseUrl}/templates`);
      return response.data.data || [];
    } catch (error) {
      log.error('获取配置模板失败', error, 'config', 'ConfigService');
      return [];
    }
  }

  /**
   * 从模板创建配置
   */
  async createFromTemplate(templateId: string): Promise<ConfigCanvasState> {
    try {
      log.info('从模板创建配置', { templateId }, 'config', 'ConfigService');

      const response = await apiService.client.post(`${this.baseUrl}/templates/${templateId}/create`);
      return response.data.data;
    } catch (error) {
      log.error('从模板创建配置失败', { templateId, error }, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 获取配置变更历史
   */
  async getConfigHistory(key?: string, limit: number = 50): Promise<ConfigUpdateOperation[]> {
    try {
      log.debug('获取配置变更历史', { key, limit }, 'config', 'ConfigService');

      const params = new URLSearchParams();
      if (key) params.append('key', key);
      params.append('limit', String(limit));

      const response = await apiService.client.get(`${this.baseUrl}/history`, { params });
      return response.data.data || [];
    } catch (error) {
      log.error('获取配置历史失败', { key, error }, 'config', 'ConfigService');
      return [];
    }
  }

  /**
   * 搜索配置
   */
  async searchConfigs(query: string, options?: { category?: string; dataType?: string }): Promise<ConfigRecord[]> {
    try {
      log.info('搜索配置', { query, options }, 'config', 'ConfigService');

      const params = new URLSearchParams();
      params.append('q', query);
      if (options?.category) params.append('category', options.category);
      if (options?.dataType) params.append('dataType', options.dataType);

      const response = await apiService.client.get(`${this.baseUrl}/search`, { params });
      return response.data.data || [];
    } catch (error) {
      log.error('搜索配置失败', { query, error }, 'config', 'ConfigService');
      return [];
    }
  }

  /**
   * 获取默认画布状态
   */
  private getDefaultCanvasState(): ConfigCanvasState {
    return {
      nodes: [],
      edges: [],
      viewport: {
        x: 0,
        y: 0,
        zoom: 1,
      },
      history: {
        past: [],
        present: {
          nodes: [],
          edges: [],
          viewport: { x: 0, y: 0, zoom: 1 },
          history: { past: [], present: {} as ConfigCanvasState, future: [] },
        },
        future: [],
      },
    };
  }

  /**
   * 将配置记录转换为画布节点
   */
  configsToNodes(configs: ConfigRecord[]): ConfigNode[] {
    return configs.map((config, index) => ({
      id: config.id.toString(),
      type: 'config' as const,
      position: {
        x: 100 + (index % 5) * 200,
        y: 100 + Math.floor(index / 5) * 150,
      },
      data: {
        key: config.key,
        label: config.key,
        description: config.description,
        category: config.category,
        value: config.value,
        dataType: Array.isArray(config.value) ? 'array' : typeof config.value,
        editable: true,
        icon: this.getIconForDataType(config.value),
        color: this.getColorForCategory(config.category),
      },
    }));
  }

  /**
   * 根据数据类型获取图标
   */
  private getIconForDataType(value: any): string {
    if (Array.isArray(value)) return 'array';
    if (typeof value === 'object' && value !== null) return 'object';
    if (typeof value === 'string') return 'string';
    if (typeof value === 'number') return 'number';
    if (typeof value === 'boolean') return 'boolean';
    return 'unknown';
  }

  /**
   * 根据分类获取颜色
   */
  private getColorForCategory(category?: string): string {
    const colors: Record<string, string> = {
      'system': '#1890ff',
      'user': '#52c41a',
      'device': '#faad14',
      'network': '#722ed1',
      'media': '#eb2f96',
      'security': '#ff4d4f',
      'performance': '#13c2c2',
    };
    return colors[category || ''] || '#666666';
  }
}

// 创建配置服务实例
export const configService = new ConfigService();
export default configService;