export { AiNode, AiNodeComponent } from './AiNode';
export { ApiNode, ApiNodeComponent } from './ApiNode';
export { BaseNode, renderNode } from './BaseNode';
export { CloudNode, CloudNodeComponent } from './CloudNode';
export { ConfigNode, ConfigNodeComponent } from './ConfigNode';
export { DatabaseNode, DatabaseNodeComponent } from './DatabaseNode';

// 所有节点类型的映射
export const NODE_TYPES = {
  DatabaseNode,
  ApiNode,
  AiNode,
  CloudNode,
  ConfigNode,
} as const;
