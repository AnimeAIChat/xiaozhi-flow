import { SimpleEventEmitter } from '../../plugins/utils/ProcessManager';
import {
  NodeDefinition,
  ParameterDefinition,
  ConfigNode
} from '../../plugins/types';
import { Point } from 'reactflow';
import { v4 as uuidv4 } from 'uuid';

interface NodeTemplate {
  id: string;
  definition: NodeDefinition;
  defaultData: any;
  preview?: string;
}

interface NodeCreationOptions {
  position?: Point;
  data?: any;
  parentId?: string;
  skipValidation?: boolean;
  autoId?: boolean;
}

interface NodeCloneOptions {
  position?: Point;
  offset?: Point;
  deepClone?: boolean;
  newId?: string;
}

interface ValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
}

export class DynamicNodeFactory extends SimpleEventEmitter {
  private templates: Map<string, NodeTemplate> = new Map();
  private nodeCounters: Map<string, number> = new Map();

  constructor() {
    super();
  }

  /**
   * 从节点定义创建节点
   */
  createNode(
    nodeId: string,
    definition: NodeDefinition,
    options: NodeCreationOptions = {}
  ): ConfigNode {
    try {
      // 验证节点定义
      if (!options.skipValidation) {
        this.validateNodeDefinition(definition);
      }

      // 生成节点ID
      const finalNodeId = options.autoId ? this.generateNodeId(nodeId) : nodeId;

      // 准备节点数据
      const nodeData = this.prepareNodeData(definition, options.data);

      // 创建节点对象
      const node: ConfigNode = {
        id: finalNodeId,
        type: 'plugin',
        position: options.position || { x: 0, y: 0 },
        data: {
          ...nodeData,
          pluginId: this.extractPluginId(definition.id),
          nodeDefinition: definition,
          dynamicParameters: this.extractDynamicParameters(definition, options.data),
          serviceStatus: 'idle'
        }
      };

      // 更新节点计数器
      this.updateNodeCounter(nodeId);

      this.emit('node-created', {
        nodeId: finalNodeId,
        node,
        definition,
        options
      });

      return node;

    } catch (error) {
      this.emit('node-creation-error', {
        nodeId,
        definition,
        error: error instanceof Error ? error.message : 'Unknown error'
      });
      throw error;
    }
  }

  /**
   * 从模板创建节点
   */
  createNodeFromTemplate(
    templateId: string,
    options: NodeCreationOptions = {}
  ): ConfigNode {
    const template = this.templates.get(templateId);
    if (!template) {
      throw new Error(`Template ${templateId} not found`);
    }

    // 合并默认数据和提供的数据
    const mergedData = {
      ...template.defaultData,
      ...options.data
    };

    return this.createNode(templateId, template.definition, {
      ...options,
      data: mergedData
    });
  }

  /**
   * 克隆节点
   */
  cloneNode(
    sourceNode: ConfigNode,
    options: NodeCloneOptions = {}
  ): ConfigNode {
    try {
      const offset = options.offset || { x: 50, y: 50 };
      const newPosition = {
        x: sourceNode.position.x + offset.x,
        y: sourceNode.position.y + offset.y
      };

      const clonedData = options.deepClone
        ? JSON.parse(JSON.stringify(sourceNode.data))
        : { ...sourceNode.data };

      // 生成新ID
      const newId = options.newId || this.generateNodeId(sourceNode.data.key || 'node');

      // 克隆节点
      const clonedNode: ConfigNode = {
        ...sourceNode,
        id: newId,
        position: options.position || newPosition,
        data: {
          ...clonedData,
          key: newId,
          label: `${clonedData.label} (Copy)`,
          serviceStatus: 'idle'
        }
      };

      this.emit('node-cloned', {
        sourceNode,
        clonedNode,
        options
      });

      return clonedNode;

    } catch (error) {
      this.emit('node-clone-error', {
        sourceNode,
        error: error instanceof Error ? error.message : 'Unknown error'
      });
      throw error;
    }
  }

  /**
   * 批量创建节点
   */
  createNodesBatch(
    nodeSpecs: Array<{
      nodeId: string;
      definition: NodeDefinition;
      options?: NodeCreationOptions;
    }>
  ): ConfigNode[] {
    const results: ConfigNode[] = [];
    const errors: Array<{ index: number; error: string }> = [];

    nodeSpecs.forEach((spec, index) => {
      try {
        const node = this.createNode(spec.nodeId, spec.definition, spec.options);
        results.push(node);
      } catch (error) {
        errors.push({
          index,
          error: error instanceof Error ? error.message : 'Unknown error'
        });
      }
    });

    this.emit('batch-creation-completed', {
      total: nodeSpecs.length,
      success: results.length,
      errors,
      nodes: results
    });

    if (errors.length > 0) {
      console.warn(`Node batch creation completed with ${errors.length} errors:`, errors);
    }

    return results;
  }

  /**
   * 注册节点模板
   */
  registerTemplate(template: NodeTemplate): void {
    this.templates.set(template.id, template);
    this.emit('template-registered', { template });
  }

  /**
   * 注销节点模板
   */
  unregisterTemplate(templateId: string): boolean {
    const removed = this.templates.delete(templateId);
    if (removed) {
      this.emit('template-unregistered', { templateId });
    }
    return removed;
  }

  /**
   * 获取节点模板
   */
  getTemplate(templateId: string): NodeTemplate | undefined {
    return this.templates.get(templateId);
  }

  /**
   * 获取所有模板
   */
  getAllTemplates(): NodeTemplate[] {
    return Array.from(this.templates.values());
  }

  /**
   * 创建节点模板
   */
  createTemplate(
    templateId: string,
    definition: NodeDefinition,
    defaultData: any = {}
  ): NodeTemplate {
    const template: NodeTemplate = {
      id: templateId,
      definition,
      defaultData: this.prepareNodeData(definition, defaultData)
    };

    this.registerTemplate(template);
    return template;
  }

  /**
   * 从现有节点创建模板
   */
  createTemplateFromNode(
    templateId: string,
    node: ConfigNode,
    includeData = true
  ): NodeTemplate {
    if (!node.data.nodeDefinition) {
      throw new Error('Node does not have a node definition');
    }

    const definition = node.data.nodeDefinition;
    const defaultData = includeData ? { ...node.data } : {};

    return this.createTemplate(templateId, definition, defaultData);
  }

  /**
   * 验证节点定义
   */
  validateNodeDefinition(definition: NodeDefinition): ValidationResult {
    const errors: string[] = [];
    const warnings: string[] = [];

    // 基本字段验证
    if (!definition.id) {
      errors.push('Node definition must have an id');
    }

    if (!definition.displayName) {
      errors.push('Node definition must have a display name');
    }

    if (!definition.category) {
      warnings.push('Node definition should have a category');
    }

    // 参数验证
    if (!definition.parameters || !Array.isArray(definition.parameters)) {
      errors.push('Node definition must have parameters array');
    } else {
      definition.parameters.forEach((param, index) => {
        if (!param.id) {
          errors.push(`Parameter ${index} must have an id`);
        }

        if (!param.name) {
          errors.push(`Parameter ${index} must have a name`);
        }

        if (!param.type) {
          errors.push(`Parameter ${index} must have a type`);
        }

        if (param.required && param.defaultValue !== undefined) {
          warnings.push(`Parameter ${param.id} is required but has a default value`);
        }

        // 验证动态配置
        if (param.dynamic) {
          if (param.dynamic.options && typeof param.dynamic.options !== 'function') {
            errors.push(`Parameter ${param.id} dynamic.options must be a function`);
          }

          if (param.dynamic.validation && typeof param.dynamic.validation !== 'function') {
            errors.push(`Parameter ${param.id} dynamic.validation must be a function`);
          }

          if (param.dynamic.visible && typeof param.dynamic.visible !== 'function') {
            errors.push(`Parameter ${param.id} dynamic.visible must be a function`);
          }
        }
      });
    }

    // 端点验证
    if (definition.endpoints && Array.isArray(definition.endpoints)) {
      definition.endpoints.forEach((endpoint, index) => {
        if (!endpoint.id) {
          errors.push(`Endpoint ${index} must have an id`);
        }

        if (!endpoint.path) {
          errors.push(`Endpoint ${index} must have a path`);
        }

        if (!endpoint.method) {
          errors.push(`Endpoint ${index} must have a method`);
        }
      });
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * 更新节点参数
   */
  updateNodeParameters(
    node: ConfigNode,
    parameterUpdates: Record<string, any>,
    validate = true
  ): ConfigNode {
    try {
      if (!node.data.nodeDefinition) {
        throw new Error('Node does not have a node definition');
      }

      const definition = node.data.nodeDefinition;
      const updatedData = { ...node.data };

      // 应用参数更新
      for (const [paramId, value] of Object.entries(parameterUpdates)) {
        const param = definition.parameters.find(p => p.id === paramId);
        if (!param) {
          throw new Error(`Parameter ${paramId} not found in node definition`);
        }

        // 验证参数值
        if (validate) {
          this.validateParameterValue(param, value);
        }

        // 应用转换（如果有）
        const transformedValue = param.backendMapping?.transform
          ? param.backendMapping.transform(value)
          : value;

        updatedData.value = updatedData.value || {};
        updatedData.value[paramId] = transformedValue;
      }

      // 更新动态参数
      if (node.data.dynamicParameters) {
        updatedData.dynamicParameters = {
          ...node.data.dynamicParameters,
          ...parameterUpdates
        };
      }

      const updatedNode = {
        ...node,
        data: updatedData
      };

      this.emit('node-updated', {
        node: updatedNode,
        parameterUpdates,
        validate
      });

      return updatedNode;

    } catch (error) {
      this.emit('node-update-error', {
        node,
        parameterUpdates,
        error: error instanceof Error ? error.message : 'Unknown error'
      });
      throw error;
    }
  }

  /**
   * 验证参数值
   */
  private validateParameterValue(param: ParameterDefinition, value: any): void {
    // 类型检查
    switch (param.type) {
      case 'number':
        if (typeof value !== 'number' && isNaN(Number(value))) {
          throw new Error(`Parameter ${param.id} must be a number`);
        }
        break;

      case 'boolean':
        if (typeof value !== 'boolean') {
          throw new Error(`Parameter ${param.id} must be a boolean`);
        }
        break;

      case 'string':
        if (typeof value !== 'string') {
          throw new Error(`Parameter ${param.id} must be a string`);
        }
        break;

      case 'array':
        if (!Array.isArray(value)) {
          throw new Error(`Parameter ${param.id} must be an array`);
        }
        break;

      case 'object':
        if (typeof value !== 'object' || value === null || Array.isArray(value)) {
          throw new Error(`Parameter ${param.id} must be an object`);
        }
        break;
    }

    // 约束检查
    if (param.constraints) {
      const numValue = Number(value);

      if (param.constraints.min !== undefined && numValue < param.constraints.min) {
        throw new Error(`Parameter ${param.id} must be >= ${param.constraints.min}`);
      }

      if (param.constraints.max !== undefined && numValue > param.constraints.max) {
        throw new Error(`Parameter ${param.id} must be <= ${param.constraints.max}`);
      }

      if (param.constraints.minLength !== undefined && typeof value === 'string' && value.length < param.constraints.minLength) {
        throw new Error(`Parameter ${param.id} must have at least ${param.constraints.minLength} characters`);
      }

      if (param.constraints.maxLength !== undefined && typeof value === 'string' && value.length > param.constraints.maxLength) {
        throw new Error(`Parameter ${param.id} must have at most ${param.constraints.maxLength} characters`);
      }

      if (param.constraints.pattern && typeof value === 'string' && !new RegExp(param.constraints.pattern).test(value)) {
        throw new Error(`Parameter ${param.id} does not match required pattern`);
      }
    }
  }

  /**
   * 准备节点数据
   */
  private prepareNodeData(definition: NodeDefinition, customData?: any): any {
    const data: any = {
      key: definition.id,
      label: definition.displayName,
      description: definition.description,
      category: definition.category,
      dataType: 'object',
      required: false,
      editable: true,
      icon: definition.category,
      color: definition.color,
      value: {}
    };

    // 设置默认参数值
    for (const param of definition.parameters) {
      if (param.defaultValue !== undefined) {
        data.value[param.id] = param.defaultValue;
      }
    }

    // 应用自定义数据
    if (customData) {
      Object.assign(data, customData);
    }

    return data;
  }

  /**
   * 提取动态参数
   */
  private extractDynamicParameters(definition: NodeDefinition, data?: any): Record<string, any> {
    const dynamicParams: Record<string, any> = {};

    for (const param of definition.parameters) {
      if (param.dynamic) {
        dynamicParams[param.id] = data?.value?.[param.id] || param.defaultValue;
      }
    }

    return dynamicParams;
  }

  /**
   * 提取插件ID
   */
  private extractPluginId(nodeId: string): string {
    const parts = nodeId.split('-');
    return parts.length > 1 ? parts[0] : nodeId;
  }

  /**
   * 生成节点ID
   */
  private generateNodeId(baseId: string): string {
    const counter = this.nodeCounters.get(baseId) || 0;
    this.nodeCounters.set(baseId, counter + 1);
    return counter === 0 ? baseId : `${baseId}-${counter}`;
  }

  /**
   * 更新节点计数器
   */
  private updateNodeCounter(nodeId: string): void {
    const baseId = nodeId.split('-')[0];
    const counter = this.nodeCounters.get(baseId) || 0;
    this.nodeCounters.set(baseId, Math.max(counter, this.extractNodeNumber(nodeId)));
  }

  /**
   * 提取节点编号
   */
  private extractNodeNumber(nodeId: string): number {
    const parts = nodeId.split('-');
    if (parts.length > 1) {
      const num = parseInt(parts[parts.length - 1]);
      return isNaN(num) ? 1 : num;
    }
    return 1;
  }

  /**
   * 清理工厂
   */
  cleanup(): void {
    this.templates.clear();
    this.nodeCounters.clear();
    this.emit('factory-cleared');
  }
}

// 全局动态节点工厂实例
export const dynamicNodeFactory = new DynamicNodeFactory();