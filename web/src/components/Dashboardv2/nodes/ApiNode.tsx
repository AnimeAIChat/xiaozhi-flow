import { BaseNode, type NodeData, renderNode } from './BaseNode';

export class ApiNode extends BaseNode {
  constructor(id: string, label: string, data?: Partial<NodeData>) {
    super(id, label, 'api', {
      description: 'REST API 调用节点',
      metrics: {
        requests: 0,
        errors: 0,
        avgResponseTime: '0ms',
      },
      ...data,
    });
  }
}

// 注册节点组件
export const ApiNodeComponent = renderNode;
