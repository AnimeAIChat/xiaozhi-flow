/**
 * 配置管理服务
 * 处理配置记录的CRUD操作、画布状态管理、快照等功能
 */

import type {
  ConfigCanvasState,
  ConfigCategory,
  ConfigEdge,
  ConfigExport,
  ConfigFilter,
  ConfigNode,
  ConfigRecord,
  ConfigSnapshot,
  ConfigTemplate,
  ConfigUpdateOperation,
  ConfigValidation,
} from '../types/config';
import { log } from '../utils/logger';
import { apiService } from './api';

export class ConfigService {
  private baseUrl = '/admin/config';

  /**
   * 获取所有配置记录
   */
  async getConfigs(filter?: ConfigFilter): Promise<ConfigRecord[]> {
    try {
      const params = new URLSearchParams();
      if (filter?.category) params.append('category', filter.category);
      if (filter?.isActive !== undefined)
        params.append('isActive', String(filter.isActive));
      if (filter?.searchText) params.append('search', filter.searchText);

      // 设置大的limit来获取所有记录
      params.append('limit', '1000');
      params.append('page', '1'); // 确保从第一页开始

      console.log(
        'ConfigService: Getting configs with params:',
        params.toString(),
      );
      console.log('ConfigService: All parameters:', {
        category: filter?.category,
        isActive: filter?.isActive,
        searchText: filter?.searchText,
        limit: '1000',
        page: '1',
      });

      log.info('获取配置记录', { filter }, 'config', 'ConfigService');

      const response = await apiService.client.get(`${this.baseUrl}/records`, {
        params,
      });
      console.log('ConfigService: API response:', response.data);
      // console.log('ConfigService: Full response structure:', JSON.stringify(response.data, null, 2));
      console.log(
        'ConfigService: Extracted records:',
        response.data.data?.data?.records,
      );

      const result = response.data.data?.data?.records || [];
      console.log('ConfigService: Final result length:', result.length);
      return result;
    } catch (error) {
      console.error('ConfigService: Error getting configs:', error);
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

      const response = await apiService.client.get(
        `${this.baseUrl}/records/${key}`,
      );
      return response.data.data;
    } catch (error) {
      log.warn('获取配置失败', { key, error }, 'config', 'ConfigService');
      return null;
    }
  }

  /**
   * 创建配置记录
   */
  async createConfig(
    config: Omit<ConfigRecord, 'id' | 'created_at' | 'updated_at'>,
  ): Promise<ConfigRecord> {
    try {
      log.info(
        '创建配置记录',
        { key: config.key, category: config.category },
        'config',
        'ConfigService',
      );

      const response = await apiService.client.post(
        `${this.baseUrl}/records`,
        config,
      );
      return response.data.data;
    } catch (error) {
      log.error(
        '创建配置记录失败',
        { config, error },
        'config',
        'ConfigService',
      );
      throw error;
    }
  }

  /**
   * 更新配置记录
   */
  async updateConfig(
    key: string,
    updates: Partial<ConfigRecord>,
  ): Promise<ConfigRecord> {
    try {
      log.info('更新配置记录', { key, updates }, 'config', 'ConfigService');

      const response = await apiService.client.put(
        `${this.baseUrl}/records/${key}`,
        updates,
      );
      return response.data.data;
    } catch (error) {
      log.error(
        '更新配置记录失败',
        { key, updates, error },
        'config',
        'ConfigService',
      );
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

      const response = await apiService.client.get(
        `${this.baseUrl}/categories`,
      );
      return response.data.data || [];
    } catch (error) {
      log.error('获取配置分类失败', error, 'config', 'ConfigService');
      return [];
    }
  }

  /**
   * 批量更新配置
   */
  async batchUpdateConfigs(
    updates: Array<{ key: string; value: any }>,
  ): Promise<void> {
    try {
      log.info(
        '批量更新配置',
        { count: updates.length },
        'config',
        'ConfigService',
      );

      await apiService.client.post(`${this.baseUrl}/batch-update`, { updates });
    } catch (error) {
      log.error(
        '批量更新配置失败',
        { updates, error },
        'config',
        'ConfigService',
      );
      throw error;
    }
  }

  /**
   * 验证配置
   */
  async validateConfigs(configs?: ConfigRecord[]): Promise<ConfigValidation> {
    try {
      log.info(
        '验证配置',
        { count: configs?.length },
        'config',
        'ConfigService',
      );

      const response = await apiService.client.post(
        `${this.baseUrl}/validate`,
        { configs },
      );
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
      log.warn(
        '获取画布状态失败，使用默认状态',
        error,
        'config',
        'ConfigService',
      );
      return this.getDefaultCanvasState();
    }
  }

  /**
   * 保存配置画布状态
   */
  async saveCanvasState(state: ConfigCanvasState): Promise<void> {
    try {
      log.info(
        '保存配置画布状态',
        {
          nodeCount: state.nodes.length,
          edgeCount: state.edges.length,
        },
        'config',
        'ConfigService',
      );

      await apiService.client.post(`${this.baseUrl}/canvas`, state);
    } catch (error) {
      log.error('保存画布状态失败', error, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 创建配置快照
   */
  async createSnapshot(
    name: string,
    description?: string,
  ): Promise<ConfigSnapshot> {
    try {
      log.info(
        '创建配置快照',
        { name, description },
        'config',
        'ConfigService',
      );

      const response = await apiService.client.post(
        `${this.baseUrl}/snapshots`,
        {
          name,
          description,
        },
      );
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

      await apiService.client.post(
        `${this.baseUrl}/snapshots/${snapshotId}/restore`,
      );
    } catch (error) {
      log.error(
        '恢复配置快照失败',
        { snapshotId, error },
        'config',
        'ConfigService',
      );
      throw error;
    }
  }

  /**
   * 导出配置
   */
  async exportConfig(filter?: ConfigFilter): Promise<ConfigExport> {
    try {
      log.info('导出配置', { filter }, 'config', 'ConfigService');

      const response = await apiService.client.post(`${this.baseUrl}/export`, {
        filter,
      });
      return response.data.data;
    } catch (error) {
      log.error('导出配置失败', error, 'config', 'ConfigService');
      throw error;
    }
  }

  /**
   * 导入配置
   */
  async importConfig(
    exportData: ConfigExport,
    options?: { overwrite?: boolean; validateOnly?: boolean },
  ): Promise<any> {
    try {
      log.info(
        '导入配置',
        {
          configCount: exportData.configs.length,
          options,
        },
        'config',
        'ConfigService',
      );

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

      const response = await apiService.client.post(
        `${this.baseUrl}/templates/${templateId}/create`,
      );
      return response.data.data;
    } catch (error) {
      log.error(
        '从模板创建配置失败',
        { templateId, error },
        'config',
        'ConfigService',
      );
      throw error;
    }
  }

  /**
   * 获取配置变更历史
   */
  async getConfigHistory(
    key?: string,
    limit: number = 50,
  ): Promise<ConfigUpdateOperation[]> {
    try {
      log.debug('获取配置变更历史', { key, limit }, 'config', 'ConfigService');

      const params = new URLSearchParams();
      if (key) params.append('key', key);
      params.append('limit', String(limit));

      const response = await apiService.client.get(`${this.baseUrl}/history`, {
        params,
      });
      return response.data.data || [];
    } catch (error) {
      log.error('获取配置历史失败', { key, error }, 'config', 'ConfigService');
      return [];
    }
  }

  /**
   * 搜索配置
   */
  async searchConfigs(
    query: string,
    options?: { category?: string; dataType?: string },
  ): Promise<ConfigRecord[]> {
    try {
      log.info('搜索配置', { query, options }, 'config', 'ConfigService');

      const params = new URLSearchParams();
      params.append('q', query);
      if (options?.category) params.append('category', options.category);
      if (options?.dataType) params.append('dataType', options.dataType);

      const response = await apiService.client.get(`${this.baseUrl}/search`, {
        params,
      });
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
   * 将配置记录转换为画布节点（使用React Flow子流功能，智能分组）
   */
  configsToNodes(configs: ConfigRecord[]): ConfigNode[] {
    console.log('ConfigService: 开始分组配置记录，总数:', configs.length);

    // 按大类分组（第一段）
    const categoryGroups = new Map<string, ConfigRecord[]>();

    configs.forEach((config) => {
      const parts = config.key.split('.');
      const mainCategory = parts[0]; // 第一段作为大类

      if (!mainCategory) return; // 跳过空键

      if (!categoryGroups.has(mainCategory)) {
        categoryGroups.set(mainCategory, []);
      }
      categoryGroups.get(mainCategory)!.push(config);
    });

    console.log('ConfigService: 大类分组结果，分类数:', categoryGroups.size);

    const nodes: ConfigNode[] = [];
    let categoryY = 50; // 大类容器的Y坐标

    // 为每个大类创建节点结构
    categoryGroups.forEach((categoryConfigs, category) => {
      const categoryColor = this.getColorForCategory(category);
      const categoryIcon = this.getIconForCategory(category);

      // 分析配置层级结构
      const configStructure = this.analyzeConfigStructure(categoryConfigs);
      console.log(`ConfigService: ${category} 配置结构:`, configStructure);

      if (configStructure.hasThreeLevels) {
        // 三级结构：第一层用子流归拢b.c
        const { groupNode, childNodes } = this.createThreeLevelSubFlow(
          categoryConfigs,
          category,
          categoryColor,
          categoryIcon,
          { x: 100 + this.getCategoryColumn(category) * 350, y: categoryY },
        );
        nodes.push(groupNode, ...childNodes);
        categoryY += this.calculateCategoryHeight(categoryConfigs) + 50;
      } else {
        // 二级结构：全部视为一个节点
        const singleNode = this.createSingleLevelNode(
          categoryConfigs,
          category,
          categoryColor,
          categoryIcon,
          { x: 100 + this.getCategoryColumn(category) * 250, y: categoryY },
        );
        nodes.push(singleNode);
        categoryY += 120; // 单个节点固定间距
      }
    });

    console.log(
      'ConfigService: 生成的节点数（包含group和子节点）:',
      nodes.length,
    );
    return nodes;
  }

  /**
   * 分析配置的层级结构
   */
  private analyzeConfigStructure(configs: ConfigRecord[]): {
    hasThreeLevels: boolean;
    subCategories: string[];
  } {
    const subCategories = new Set<string>();
    let hasThreeLevels = false;

    configs.forEach((config) => {
      const parts = config.key.split('.');
      if (parts.length >= 2) {
        subCategories.add(parts[1]);
      }
      if (parts.length >= 3) {
        hasThreeLevels = true;
      }
    });

    return {
      hasThreeLevels,
      subCategories: Array.from(subCategories),
    };
  }

  /**
   * 创建三级结构子流：第一层a用子流归拢b.c
   */
  private createThreeLevelSubFlow(
    categoryConfigs: ConfigRecord[],
    category: string,
    categoryColor: string,
    categoryIcon: string,
    position: { x: number; y: number },
  ): { groupNode: ConfigNode; childNodes: ConfigNode[] } {
    const childNodes: ConfigNode[] = [];
    const configCount = categoryConfigs.length;

    // 计算实际的第二级分组数量，而不是总配置数量
    const uniqueBCKeys = new Set(
      categoryConfigs.map((config) => {
        const parts = config.key.split('.');
        return parts.slice(1, 2).join('.'); // 只取第二级b
      }),
    );

    const groupWidth = 300;
    const groupHeight = Math.max(200, uniqueBCKeys.size * 80 + 80);

    // 创建大group节点（第一层a）
    const groupNode: ConfigNode = {
      id: `group-${category}`,
      type: 'group',
      position,
      style: {
        width: groupWidth,
        height: groupHeight,
        backgroundColor: categoryColor + '10',
        border: `2px solid ${categoryColor}`,
        borderRadius: '8px',
      },
      data: {
        key: category,
        label: `${category} (${uniqueBCKeys.size}个服务)`,
        description: `${category} 服务组`,
        category,
        value: categoryConfigs,
        dataType: 'category-group',
        editable: false,
        icon: categoryIcon,
        color: categoryColor,
        configCount: uniqueBCKeys.size,
      },
    };

    // 按第二层b分组，每个b创建一个节点包含其所有c配置
    const bGroups = new Map<string, ConfigRecord[]>();
    categoryConfigs.forEach((config) => {
      const parts = config.key.split('.');
      const bKey = parts[1]; // 第二级b

      if (!bGroups.has(bKey)) {
        bGroups.set(bKey, []);
      }
      bGroups.get(bKey)!.push(config);
    });

    let bY = 40;

    // 为每个b创建一个简洁的节点，包含其所有配置
    bGroups.forEach((bConfigs, bKey) => {
      // 获取该b的所有第三级配置
      const cConfigs = bConfigs.map((config) => {
        const parts = config.key.split('.');
        return {
          key: parts[2], // 第三级配置名
          value: config.value,
          description: config.description,
          fullKey: config.key,
        };
      });

      // 创建b节点，包含其所有c配置
      const bNode: ConfigNode = {
        id: `node-${category}-${bKey}`,
        type: 'config',
        position: { x: 20, y: bY },
        parentId: `group-${category}`,
        data: {
          key: bKey,
          label: `${bKey} (${cConfigs.length}个配置)`,
          description: `${category}.${bKey} 服务配置`,
          category,
          value: cConfigs, // 存储该b下的所有c配置
          dataType: 'b-service-node',
          editable: false,
          icon: this.getIconForCategory(bKey),
          color: categoryColor,
          configCount: cConfigs.length,
          subCategory: bKey,
        },
      };

      childNodes.push(bNode);
      bY += 80; // 每个b节点80px高度
    });

    return { groupNode, childNodes };
  }

  /**
   * 创建单层节点：二级结构全部视为一个节点
   */
  private createSingleLevelNode(
    categoryConfigs: ConfigRecord[],
    category: string,
    categoryColor: string,
    categoryIcon: string,
    position: { x: number; y: number },
  ): ConfigNode {
    // 获取该分类下的所有配置值
    const configValues = categoryConfigs.map((config) => ({
      key: config.key,
      value: config.value,
      description: config.description,
    }));

    return {
      id: `node-${category}`,
      type: 'config',
      position,
      data: {
        key: category,
        label: `${category} (${categoryConfigs.length}个配置)`,
        description: `${category} 配置项集合`,
        category: category,
        value: configValues, // 存储该分类下的所有配置
        dataType: 'category-node',
        editable: false, // 分类节点不可直接编辑
        icon: categoryIcon,
        color: categoryColor,
        configCount: categoryConfigs.length,
      },
    };
  }

  /**
   * 创建二级结构：归拢成单个节点
   */
  private createSingleCategoryNode(
    categoryConfigs: ConfigRecord[],
    category: string,
    categoryColor: string,
    categoryIcon: string,
    position: { x: number; y: number },
  ): ConfigNode {
    // 获取该分类下的所有配置值
    const configValues = categoryConfigs.map((config) => ({
      key: config.key,
      value: config.value,
      description: config.description,
    }));

    return {
      id: `node-${category}`,
      type: 'config',
      position,
      data: {
        key: category,
        label: `${category} (${categoryConfigs.length}个配置)`,
        description: `${category} 配置项集合`,
        category: category,
        value: configValues, // 存储该分类下的所有配置
        dataType: 'category-node',
        editable: false, // 分类节点不可直接编辑
        icon: categoryIcon,
        color: categoryColor,
        configCount: categoryConfigs.length,
      },
    };
  }

  /**
   * 计算分类group的高度
   */
  private calculateCategoryHeight(configs: ConfigRecord[]): number {
    const structure = this.analyzeConfigStructure(configs);

    if (structure.hasThreeLevels) {
      // 三级结构需要更多空间
      const baseHeight = 250;
      const configHeight = configs.length * 35;
      const subGroupSpacing = structure.subCategories.length * 15;
      return Math.max(baseHeight, configHeight + subGroupSpacing + 80);
    } else {
      // 二级结构
      const baseHeight = 200;
      const configHeight = configs.length * 60;
      return Math.max(baseHeight, configHeight + 80);
    }
  }

  /**
   * 根据大类获取列位置
   */
  private getCategoryColumn(category: string): number {
    const columnMap: Record<string, number> = {
      ASR: 0, // 语音识别
      TTS: 1, // 语音合成
      LLM: 2, // 语言模型
      VLLM: 3, // 视觉语言模型
      server: 4, // 服务器配置
      web: 5, // Web服务
      transport: 6, // 传输层
      system: 7, // 系统配置
      audio: 8, // 音频配置
      database: 9, // 数据库配置
    };
    return columnMap[category] || Object.keys(columnMap).length;
  }

  /**
   * 根据大类获取图标
   */
  private getIconForCategory(category: string): string {
    const iconMap: Record<string, string> = {
      ASR: 'microphone',
      TTS: 'sound',
      LLM: 'robot',
      VLLM: 'eye',
      server: 'server',
      web: 'global',
      transport: 'api',
      system: 'setting',
      audio: 'audio',
      database: 'database',
    };
    return iconMap[category] || 'setting';
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
      ASR: '#fa8c16', // 语音识别 - 橙色
      TTS: '#52c41a', // 语音合成 - 绿色
      LLM: '#1890ff', // 语言模型 - 蓝色
      VLLM: '#722ed1', // 视觉语言模型 - 紫色
      server: '#13c2c2', // 服务器配置 - 青色
      web: '#eb2f96', // Web服务 - 粉色
      transport: '#faad14', // 传输层 - 黄色
      system: '#f5222d', // 系统配置 - 红色
      audio: '#a0d911', // 音频配置 - 浅绿
      database: '#2f54eb', // 数据库配置 - 深蓝
      user: '#52c41a', // 用户配置 - 绿色
      device: '#fa8c16', // 设备配置 - 橙色
      network: '#722ed1', // 网络配置 - 紫色
      media: '#eb2f96', // 媒体配置 - 粉色
      security: '#ff4d4f', // 安全配置 - 红色
      performance: '#13c2c2', // 性能配置 - 青色
    };
    return colors[category || ''] || '#666666';
  }
}

// 创建配置服务实例
export const configService = new ConfigService();
export default configService;
