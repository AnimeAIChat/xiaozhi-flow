import { BaseNode, NodeData, renderNode } from './BaseNode';

export class CloudNode extends BaseNode {
  constructor(id: string, label: string, data?: Partial<NodeData>) {
    super(id, label, 'cloud', {
      description: '云服务节点',
      metrics: {
        bandwidth: '0MB/s',
        storage: '0GB',
        cpu: '0%'
      },
      ...data
    });
  }
}

// 注册节点组件
export const CloudNodeComponent = renderNode;