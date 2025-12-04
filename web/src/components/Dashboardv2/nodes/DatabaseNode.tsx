import { BaseNode, NodeData, renderNode } from './BaseNode';

export class DatabaseNode extends BaseNode {
  constructor(id: string, label: string, data?: Partial<NodeData>) {
    super(id, label, 'database', {
      description: '数据库连接节点',
      metrics: {
        connections: 0,
        queries: 0
      },
      ...data
    });
  }
}

// 注册节点组件
export const DatabaseNodeComponent = renderNode;