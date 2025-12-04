import { BaseNode, NodeData, renderNode } from './BaseNode';

export class ConfigNode extends BaseNode {
  constructor(id: string, label: string, data?: Partial<NodeData>) {
    super(id, label, 'config', {
      description: '配置管理节点',
      metrics: {
        variables: 0,
        lastUpdated: new Date().toISOString()
      },
      ...data
    });
  }
}

// 注册节点组件
export const ConfigNodeComponent = renderNode;