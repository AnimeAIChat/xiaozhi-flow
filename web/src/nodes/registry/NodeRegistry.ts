import type { NodeProps } from '@xyflow/react';
import type { ComponentType } from 'react';
import {
  type IPlugin,
  type NodeDefinition,
  type ParameterDefinition,
  RuntimeAdapter,
} from '../../plugins/types';
import { SimpleEventEmitter } from '../../plugins/utils/ProcessManager';

interface NodeRegistration {
  definition: NodeDefinition;
  component?: ComponentType<NodeProps>;
  pluginId?: string;
  registeredAt: Date;
  category: string;
  tags: string[];
}

interface NodeSearchQuery {
  query?: string;
  category?: string;
  tags?: string[];
  pluginId?: string;
  runtime?: string;
  limit?: number;
  offset?: number;
}

interface NodeRegistrationOptions {
  override?: boolean;
  validate?: boolean;
  autoCategorize?: boolean;
}

export class NodeRegistry extends SimpleEventEmitter {
  private nodes: Map<string, NodeRegistration> = new Map();
  private categories: Set<string> = new Set();
  private tags: Set<string> = new Set();
  private components: Map<string, ComponentType<NodeProps>> = new Map();

  constructor() {
    super();
    this.setupDefaultCategories();
  }

  /**
   * 注册单个节点类型
   */
  registerNode(
    definition: NodeDefinition,
    component?: ComponentType<NodeProps>,
    options: NodeRegistrationOptions = {},
  ): boolean {
    try {
      // 验证节点定义
      if (options.validate !== false) {
        this.validateNodeDefinition(definition);
      }

      const nodeId = definition.id;

      // 检查是否已存在
      if (this.nodes.has(nodeId) && !options.override) {
        throw new Error(
          `Node type ${nodeId} is already registered. Use override option to replace.`,
        );
      }

      // 自动分类
      if (options.autoCategorize !== false) {
        this.autoCategorizeNode(definition);
      }

      // 创建注册记录
      const registration: NodeRegistration = {
        definition,
        component,
        registeredAt: new Date(),
        category: definition.category,
        tags: [...definition.tags, definition.category],
      };

      // 注册节点
      this.nodes.set(nodeId, registration);

      // 注册组件
      if (component) {
        this.components.set(nodeId, component);
      }

      // 更新分类和标签
      this.categories.add(definition.category);
      definition.tags.forEach((tag) => this.tags.add(tag));

      this.emit('node-registered', {
        nodeId,
        definition,
        component,
        registration,
      });

      return true;
    } catch (error) {
      this.emit('node-registration-error', {
        nodeId: definition.id,
        error: error instanceof Error ? error.message : 'Unknown error',
      });
      throw error;
    }
  }

  /**
   * 批量注册节点类型
   */
  registerNodes(
    definitions: NodeDefinition[],
    components?: Map<string, ComponentType<NodeProps>>,
    options: NodeRegistrationOptions = {},
  ): number {
    let successCount = 0;
    const errors: Array<{ nodeId: string; error: string }> = [];

    for (const definition of definitions) {
      try {
        const component = components?.get(definition.id);
        if (this.registerNode(definition, component, options)) {
          successCount++;
        }
      } catch (error) {
        errors.push({
          nodeId: definition.id,
          error: error instanceof Error ? error.message : 'Unknown error',
        });
      }
    }

    this.emit('nodes-registered', {
      total: definitions.length,
      success: successCount,
      errors,
    });

    return successCount;
  }

  /**
   * 从插件加载节点
   */
  loadFromPlugin(plugin: IPlugin): number {
    if (!plugin.nodeDefinition) {
      return 0;
    }

    const nodeId = `${plugin.id}-${plugin.nodeDefinition.id}`;
    const definition: NodeDefinition = {
      ...plugin.nodeDefinition,
      id: nodeId,
      category: plugin.metadata.category,
      tags: [...plugin.metadata.tags, plugin.runtime],
    };

    try {
      this.registerNode(definition, plugin.nodeDefinition.customComponent, {
        validate: true,
        autoCategorize: true,
        override: true,
      });

      this.emit('plugin-nodes-loaded', {
        pluginId: plugin.id,
        nodeId,
        definition,
      });

      return 1;
    } catch (error) {
      this.emit('plugin-node-load-error', {
        pluginId: plugin.id,
        nodeId,
        error: error instanceof Error ? error.message : 'Unknown error',
      });
      return 0;
    }
  }

  /**
   * 卸载节点类型
   */
  unregisterNode(nodeId: string): boolean {
    const registration = this.nodes.get(nodeId);
    if (!registration) {
      return false;
    }

    // 移除节点
    this.nodes.delete(nodeId);
    this.components.delete(nodeId);

    // 更新分类和标签
    this.updateCategoriesAndTags();

    this.emit('node-unregistered', {
      nodeId,
      registration,
    });

    return true;
  }

  /**
   * 卸载插件的所有节点
   */
  unloadPluginNodes(pluginId: string): number {
    let removedCount = 0;

    for (const [nodeId, registration] of this.nodes) {
      if (registration.pluginId === pluginId) {
        if (this.unregisterNode(nodeId)) {
          removedCount++;
        }
      }
    }

    this.emit('plugin-nodes-unloaded', {
      pluginId,
      count: removedCount,
    });

    return removedCount;
  }

  /**
   * 获取节点定义
   */
  getNodeDefinition(nodeId: string): NodeDefinition | undefined {
    return this.nodes.get(nodeId)?.definition;
  }

  /**
   * 获取节点组件
   */
  getNodeComponent(nodeId: string): ComponentType<NodeProps> | undefined {
    return this.components.get(nodeId);
  }

  /**
   * 获取节点注册信息
   */
  getNodeRegistration(nodeId: string): NodeRegistration | undefined {
    return this.nodes.get(nodeId);
  }

  /**
   * 搜索节点
   */
  searchNodes(query: NodeSearchQuery): NodeDefinition[] {
    const results: NodeDefinition[] = [];
    const {
      query: searchQuery,
      category,
      tags,
      pluginId,
      runtime,
      limit,
      offset = 0,
    } = query;

    for (const registration of this.nodes.values()) {
      const { definition } = registration;

      // 插件ID过滤
      if (pluginId && registration.pluginId !== pluginId) {
        continue;
      }

      // 运行时过滤
      if (runtime && !definition.tags.includes(runtime)) {
        continue;
      }

      // 分类过滤
      if (category && definition.category !== category) {
        continue;
      }

      // 标签过滤
      if (tags && tags.length > 0) {
        const hasAllTags = tags.every((tag) => definition.tags.includes(tag));
        if (!hasAllTags) {
          continue;
        }
      }

      // 文本搜索
      if (searchQuery) {
        const searchText = searchQuery.toLowerCase();
        const searchTextIn = [
          definition.id,
          definition.displayName,
          definition.description,
          definition.category,
          ...definition.tags,
        ]
          .join(' ')
          .toLowerCase();

        if (!searchTextIn.includes(searchText)) {
          continue;
        }
      }

      results.push(definition);
    }

    // 排序（按相关性和注册时间）
    results.sort((a, b) => {
      const regA = this.nodes.get(a.id)!;
      const regB = this.nodes.get(b.id)!;

      // 搜索查询优先匹配ID和名称
      if (searchQuery) {
        const queryLower = searchQuery.toLowerCase();
        const aMatch =
          a.id.toLowerCase().includes(queryLower) ||
          a.displayName.toLowerCase().includes(queryLower);
        const bMatch =
          b.id.toLowerCase().includes(queryLower) ||
          b.displayName.toLowerCase().includes(queryLower);

        if (aMatch && !bMatch) return -1;
        if (!aMatch && bMatch) return 1;
      }

      // 按注册时间排序（新的在前）
      return regB.registeredAt.getTime() - regA.registeredAt.getTime();
    });

    // 应用偏移和限制
    const start = offset;
    const end = limit ? start + limit : undefined;
    return results.slice(start, end);
  }

  /**
   * 获取所有节点定义
   */
  getAllNodeDefinitions(): NodeDefinition[] {
    return Array.from(this.nodes.values()).map((reg) => reg.definition);
  }

  /**
   * 按分类获取节点
   */
  getNodesByCategory(category: string): NodeDefinition[] {
    return this.searchNodes({ category });
  }

  /**
   * 获取所有分类
   */
  getCategories(): string[] {
    return Array.from(this.categories).sort();
  }

  /**
   * 获取所有标签
   */
  getTags(): string[] {
    return Array.from(this.tags).sort();
  }

  /**
   * 获取分类统计
   */
  getCategoryStats(): Map<string, number> {
    const stats = new Map<string, number>();

    for (const registration of this.nodes.values()) {
      const category = registration.definition.category;
      stats.set(category, (stats.get(category) || 0) + 1);
    }

    return stats;
  }

  /**
   * 获取标签统计
   */
  getTagStats(): Map<string, number> {
    const stats = new Map<string, number>();

    for (const registration of this.nodes.values()) {
      for (const tag of registration.definition.tags) {
        stats.set(tag, (stats.get(tag) || 0) + 1);
      }
    }

    return stats;
  }

  /**
   * 获取注册的组件列表
   */
  getRegisteredComponents(): Map<string, ComponentType<NodeProps>> {
    return new Map(this.components);
  }

  /**
   * 检查节点是否已注册
   */
  isRegistered(nodeId: string): boolean {
    return this.nodes.has(nodeId);
  }

  /**
   * 获取注册统计
   */
  getRegistryStats(): {
    totalNodes: number;
    totalCategories: number;
    totalTags: number;
    pluginNodes: number;
    customComponents: number;
  } {
    const pluginNodes = Array.from(this.nodes.values()).filter(
      (reg) => reg.pluginId !== undefined,
    ).length;

    const customComponents = Array.from(this.nodes.values()).filter(
      (reg) => reg.component !== undefined,
    ).length;

    return {
      totalNodes: this.nodes.size,
      totalCategories: this.categories.size,
      totalTags: this.tags.size,
      pluginNodes,
      customComponents,
    };
  }

  /**
   * 清理注册表
   */
  clear(): void {
    this.nodes.clear();
    this.components.clear();
    this.categories.clear();
    this.tags.clear();
    this.emit('registry-cleared');
  }

  /**
   * 验证节点定义
   */
  private validateNodeDefinition(definition: NodeDefinition): void {
    if (!definition.id || definition.id.trim() === '') {
      throw new Error('Node definition must have a valid id');
    }

    if (!definition.displayName || definition.displayName.trim() === '') {
      throw new Error('Node definition must have a valid display name');
    }

    if (!definition.category || definition.category.trim() === '') {
      throw new Error('Node definition must have a valid category');
    }

    if (!Array.isArray(definition.parameters)) {
      throw new Error('Node definition must have a parameters array');
    }

    if (!Array.isArray(definition.tags)) {
      throw new Error('Node definition must have a tags array');
    }

    // 验证参数定义
    for (const param of definition.parameters) {
      this.validateParameterDefinition(param);
    }

    // 验证端口定义
    if (definition.inputs) {
      for (const input of definition.inputs) {
        if (!input.id || !input.type || !input.dataType) {
          throw new Error('Input port must have id, type, and dataType');
        }
      }
    }

    if (definition.outputs) {
      for (const output of definition.outputs) {
        if (!output.id || !output.type || !output.dataType) {
          throw new Error('Output port must have id, type, and dataType');
        }
      }
    }
  }

  /**
   * 验证参数定义
   */
  private validateParameterDefinition(param: ParameterDefinition): void {
    if (!param.id || param.id.trim() === '') {
      throw new Error('Parameter definition must have a valid id');
    }

    if (!param.name || param.name.trim() === '') {
      throw new Error('Parameter definition must have a valid name');
    }

    if (!param.type) {
      throw new Error('Parameter definition must have a valid type');
    }
  }

  /**
   * 自动分类节点
   */
  private autoCategorizeNode(definition: NodeDefinition): void {
    // 如果没有分类，根据标签自动分类
    if (!definition.category || definition.category.trim() === '') {
      if (definition.tags.includes('LLM') || definition.tags.includes('AI')) {
        definition.category = 'AI';
      } else if (
        definition.tags.includes('database') ||
        definition.tags.includes('storage')
      ) {
        definition.category = 'Data';
      } else if (
        definition.tags.includes('api') ||
        definition.tags.includes('service')
      ) {
        definition.category = 'Service';
      } else if (
        definition.tags.includes('utility') ||
        definition.tags.includes('tool')
      ) {
        definition.category = 'Utility';
      } else {
        definition.category = 'General';
      }
    }

    // 确保分类在分类列表中
    this.categories.add(definition.category);
  }

  /**
   * 更新分类和标签
   */
  private updateCategoriesAndTags(): void {
    this.categories.clear();
    this.tags.clear();

    for (const registration of this.nodes.values()) {
      this.categories.add(registration.definition.category);
      registration.definition.tags.forEach((tag) => this.tags.add(tag));
    }
  }

  /**
   * 设置默认分类
   */
  private setupDefaultCategories(): void {
    const defaultCategories = [
      'AI',
      'Data',
      'Service',
      'Utility',
      'General',
      'Custom',
    ];

    defaultCategories.forEach((category) => this.categories.add(category));
  }
}

// 全局节点注册表实例
export const nodeRegistry = new NodeRegistry();
