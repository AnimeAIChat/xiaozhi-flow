// ReactFlow 到 Rete.js 的数据转换器

import type {
  Edge as ReactFlowEdge,
  Node as ReactFlowNode,
} from '@xyflow/react';
import type { Connection } from 'rete';
import type { NodeStatus, NodeType } from './nodeUtils';

// Rete.js 节点数据接口
export interface ReteNodeData {
  label: string;
  type: NodeType;
  status: NodeStatus;
  description?: string;
  metrics?: Record<string, string | number>;
}

// Rete.js 节点定义
export interface ReteNode {
  id: string;
  x: number;
  y: number;
  label: string;
  data: ReteNodeData;
}

/**
 * 将 ReactFlow 节点转换为 Rete.js 节点
 */
export const convertReactFlowNodeToRete = (
  reactNode: ReactFlowNode,
): ReteNode => {
  const nodeData: ReteNodeData = {
    label: reactNode.data.label || reactNode.id,
    type: reactNode.data.type || 'api',
    status: reactNode.data.status || 'stopped',
    description: reactNode.data.description,
    metrics: reactNode.data.metrics,
  };

  const reteNode: ReteNode = {
    id: reactNode.id,
    x: reactNode.position.x,
    y: reactNode.position.y,
    label: nodeData.label,
    data: nodeData,
  };

  return reteNode;
};

/**
 * 将 ReactFlow 边转换为 Rete.js 连接
 */
export const convertReactFlowEdgeToRete = (
  reactEdge: ReactFlowEdge,
  nodes: ReteNode[],
): Connection | null => {
  try {
    const sourceNode = nodes.find((n) => n.id === reactEdge.source);
    const targetNode = nodes.find((n) => n.id === reactEdge.target);

    if (!sourceNode || !targetNode) {
      console.warn(
        `无法找到连接的节点: ${reactEdge.source} -> ${reactEdge.target}`,
      );
      return null;
    }

    const connection: Connection = {
      source: reactEdge.source,
      target: reactEdge.target,
      sourceOutput: reactEdge.sourceHandle || 'output',
      targetInput: reactEdge.targetHandle || 'input',
    };

    return connection;
  } catch (error) {
    console.error('转换 ReactFlow 边失败:', error);
    return null;
  }
};

/**
 * 批量转换 ReactFlow 数据到 Rete.js 格式
 */
export const convertReactFlowToRete = (
  nodes: ReactFlowNode[],
  edges: ReactFlowEdge[],
): { nodes: ReteNode[]; connections: Connection[] } => {
  try {
    // 转换节点
    const reteNodes: ReteNode[] = nodes.map((node) =>
      convertReactFlowNodeToRete(node),
    );

    // 转换连接
    const reteConnections: Connection[] = edges
      .map((edge) => convertReactFlowEdgeToRete(edge, reteNodes))
      .filter((connection): connection is Connection => connection !== null);

    return {
      nodes: reteNodes,
      connections: reteConnections,
    };
  } catch (error) {
    console.error('批量转换 ReactFlow 数据失败:', error);
    return {
      nodes: [],
      connections: [],
    };
  }
};

/**
 * 将 Rete.js 节点转换为 ReactFlow 节点
 */
export const convertReteNodeToReactFlow = (
  reteNode: ReteNode,
): ReactFlowNode => {
  return {
    id: reteNode.id,
    type: 'default', // 使用自定义节点类型
    position: {
      x: reteNode.x || 0,
      y: reteNode.y || 0,
    },
    data: {
      label: reteNode.data.label,
      type: reteNode.data.type,
      status: reteNode.data.status,
      description: reteNode.data.description,
      metrics: reteNode.data.metrics,
    },
  };
};

/**
 * 将 Rete.js 连接转换为 ReactFlow 边
 */
export const convertReteConnectionToReactFlow = (
  connection: Connection,
): ReactFlowEdge => {
  return {
    id: `${connection.source}-${connection.target}`,
    source: connection.source,
    target: connection.target,
    sourceHandle: connection.sourceOutput,
    targetHandle: connection.targetInput,
    type: 'smoothstep',
    animated: true,
    style: {
      stroke: '#3b82f6',
      strokeWidth: 2,
    },
  };
};

/**
 * 批量转换 Rete.js 数据到 ReactFlow 格式
 */
export const convertReteToReactFlow = (
  nodes: ReteNode[],
  connections: Connection[],
): { nodes: ReactFlowNode[]; edges: ReactFlowEdge[] } => {
  try {
    const reactFlowNodes: ReactFlowNode[] = nodes.map((node) =>
      convertReteNodeToReactFlow(node),
    );
    const reactFlowEdges: ReactFlowEdge[] = connections.map((conn) =>
      convertReteConnectionToReactFlow(conn),
    );

    return {
      nodes: reactFlowNodes,
      edges: reactFlowEdges,
    };
  } catch (error) {
    console.error('批量转换 Rete.js 数据失败:', error);
    return {
      nodes: [],
      edges: [],
    };
  }
};

/**
 * 验证节点类型是否兼容
 */
export const validateNodeType = (type: string): type is NodeType => {
  return ['database', 'api', 'ai', 'cloud', 'config'].includes(type);
};

/**
 * 验证节点状态是否兼容
 */
export const validateNodeStatus = (status: string): status is NodeStatus => {
  return ['running', 'stopped', 'warning'].includes(status);
};
