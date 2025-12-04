import { BaseNode, NodeData, renderNode } from './BaseNode';

export class AiNode extends BaseNode {
  constructor(id: string, label: string, data?: Partial<NodeData>) {
    super(id, label, 'ai', {
      description: 'AI 服务节点',
      metrics: {
        tokens: 0,
        cost: '$0.00',
        latency: '0ms'
      },
      ...data
    });
  }
}

// 注册节点组件
export const AiNodeComponent = renderNode;