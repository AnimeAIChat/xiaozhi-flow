export { BaseNode, renderNode } from './BaseNode';
export { DatabaseNode, DatabaseNodeComponent } from './DatabaseNode';
export { ApiNode, ApiNodeComponent } from './ApiNode';
export { AiNode, AiNodeComponent } from './AiNode';
export { CloudNode, CloudNodeComponent } from './CloudNode';
export { ConfigNode, ConfigNodeComponent } from './ConfigNode';

// 所有节点类型的映射
export const NODE_TYPES = {
  DatabaseNode,
  ApiNode,
  AiNode,
  CloudNode,
  ConfigNode
} as const;