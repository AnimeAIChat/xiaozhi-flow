/**
 * 启动流程数据转换工具
 * 负责将后端启动流程数据转换为前端ReactFlow可用的格式
 */

import { Node, Edge } from 'reactflow';
import { StartupWorkflow, StartupExecution, StartupWorkflowNode } from '../services/startupWebSocket';
import { WorkflowNodeData } from '../components/Dashboard/types';

// 节点类型映射
const NODE_TYPE_MAP: Record<string, string> = {
  'storage': 'database',
  'config': 'config',
  'service': 'api',
  'auth': 'api',
  'plugin': 'cloud',
  'default': 'api'
};

// 状态映射
const STATUS_MAP: Record<string, 'running' | 'warning' | 'stopped'> = {
  'pending': 'stopped',
  'running': 'running',
  'completed': 'running',
  'failed': 'warning',
  'paused': 'warning',
  'cancelled': 'stopped'
};

// 节点颜色映射
const NODE_COLOR_MAP: Record<string, string> = {
  'storage': '#1890ff',  // 蓝色
  'config': '#52c41a',   // 绿色
  'service': '#722ed1',  // 紫色
  'auth': '#fa8c16',     // 橙色
  'plugin': '#13c2c2',   // 青色
  'default': '#d9d9d9'   // 灰色
};

/**
 * 将后端启动节点转换为ReactFlow节点
 */
export function convertStartupNodeToReactFlow(
  node: StartupWorkflowNode,
  execution?: StartupExecution
): Node<WorkflowNodeData> {
  // 从执行中获取节点状态
  let nodeStatus = node.status;
  let startTime: string | undefined;
  let endTime: string | undefined;
  let duration: number | undefined;
  let error: string | undefined;
  let progress: number | undefined;
  let metrics: Record<string, any> = {};

  if (execution) {
    const executionNode = execution.nodes.find(n => n.id === node.id);
    if (executionNode) {
      nodeStatus = executionNode.status;
      startTime = executionNode.start_time;
      endTime = executionNode.end_time;
      duration = executionNode.duration;
      error = executionNode.error;
      progress = executionNode.progress;
      metrics = executionNode.metrics || {};
    }
  }

  // 构建指标数据
  const displayMetrics: Record<string, string> = {
    ...metrics,
  };

  // 添加基本指标
  if (node.critical) {
    displayMetrics['类型'] = '关键';
  }
  if (node.optional) {
    displayMetrics['类型'] = '可选';
  }
  if (node.timeout) {
    displayMetrics['超时'] = `${Math.round(node.timeout / 1000)}s`;
  }
  if (progress !== undefined) {
    displayMetrics['进度'] = `${Math.round(progress * 100)}%`;
  }
  if (duration) {
    displayMetrics['耗时'] = `${Math.round(duration / 1000)}s`;
  }
  if (error) {
    displayMetrics['错误'] = error.substring(0, 50) + (error.length > 50 ? '...' : '');
  }

  return {
    id: node.id,
    type: 'custom',
    position: { x: node.position.x, y: node.position.y },
    data: {
      label: node.name,
      type: NODE_TYPE_MAP[node.type] || NODE_TYPE_MAP.default,
      status: STATUS_MAP[nodeStatus] || 'stopped',
      description: node.description,
      metrics: displayMetrics,
      // 扩展数据，供高级功能使用
      startupNode: {
        ...node,
        status: nodeStatus,
        start_time: startTime,
        end_time: endTime,
        duration,
        error,
        progress
      }
    },
  };
}

/**
 * 将后端启动工作流转换为ReactFlow节点数组
 */
export function convertStartupWorkflowToReactFlowNodes(
  workflow: StartupWorkflow,
  execution?: StartupExecution
): Node<WorkflowNodeData>[] {
  return workflow.nodes.map(node => convertStartupNodeToReactFlow(node, execution));
}

/**
 * 将后端启动工作流边转换为ReactFlow边数组
 */
export function convertStartupWorkflowToReactFlowEdges(
  workflow: StartupWorkflow
): Edge[] {
  return workflow.edges.map(edge => ({
    id: edge.id,
    source: edge.from,
    target: edge.to,
    type: 'smoothstep',
    label: edge.label,
    animated: true,
    style: {
      strokeWidth: 2,
      stroke: '#b1b1b7'
    },
    markerEnd: {
      type: 'arrow',
      color: '#b1b1b7'
    }
  }));
}

/**
 * 根据执行状态更新节点样式
 */
export function updateNodeStyleByExecution(
  nodes: Node<WorkflowNodeData>[],
  execution: StartupExecution
): Node<WorkflowNodeData>[] {
  return nodes.map(node => {
    const executionNode = execution.nodes.find(n => n.id === node.id);
    if (!executionNode) return node;

    const updatedNode = { ...node };

    // 更新状态
    updatedNode.data.status = STATUS_MAP[executionNode.status] || 'stopped';

    // 更新指标
    if (updatedNode.data.startupNode) {
      updatedNode.data.startupNode.status = executionNode.status;
      updatedNode.data.startupNode.start_time = executionNode.start_time;
      updatedNode.data.startupNode.end_time = executionNode.end_time;
      updatedNode.data.startupNode.duration = executionNode.duration;
      updatedNode.data.startupNode.error = executionNode.error;
      updatedNode.data.startupNode.progress = executionNode.progress;
    }

    // 更新显示指标
    const newMetrics = { ...updatedNode.data.metrics };

    if (executionNode.progress !== undefined) {
      newMetrics['进度'] = `${Math.round(executionNode.progress * 100)}%`;
    }

    if (executionNode.duration) {
      newMetrics['耗时'] = `${Math.round(executionNode.duration / 1000)}s`;
    }

    if (executionNode.error) {
      newMetrics['错误'] = executionNode.error.substring(0, 50) +
        (executionNode.error.length > 50 ? '...' : '');
    }

    updatedNode.data.metrics = newMetrics;

    // 更新节点颜色（根据状态）
    if (executionNode.status === 'running') {
      updatedNode.style = {
        ...updatedNode.style,
        border: '2px solid #52c41a',
        boxShadow: '0 0 10px rgba(82, 196, 26, 0.3)'
      };
    } else if (executionNode.status === 'failed') {
      updatedNode.style = {
        ...updatedNode.style,
        border: '2px solid #ff4d4f',
        boxShadow: '0 0 10px rgba(255, 77, 79, 0.3)'
      };
    } else if (executionNode.status === 'completed') {
      updatedNode.style = {
        ...updatedNode.style,
        border: '2px solid #1890ff',
        boxShadow: '0 0 10px rgba(24, 144, 255, 0.3)'
      };
    }

    return updatedNode;
  });
}

/**
 * 更新边的动画状态
 */
export function updateEdgesAnimation(
  edges: Edge[],
  execution: StartupExecution
): Edge[] {
  const currentNodes = new Set(execution.current_nodes);

  return edges.map(edge => {
    const isSourceActive = currentNodes.has(edge.source);
    const isTargetActive = currentNodes.has(edge.target);

    return {
      ...edge,
      animated: isSourceActive || isTargetActive,
      style: {
        ...edge.style,
        stroke: isSourceActive || isTargetActive ? '#52c41a' : '#b1b1b7',
        strokeWidth: isSourceActive || isTargetActive ? 3 : 2
      },
      markerEnd: {
        type: 'arrow',
        color: isSourceActive || isTargetActive ? '#52c41a' : '#b1b1b7'
      }
    };
  });
}

/**
 * 生成启动流程统计信息
 */
export function generateStartupStats(execution: StartupExecution) {
  const statusCount = execution.nodes.reduce((acc, node) => {
    acc[node.status] = (acc[node.status] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  return {
    total: execution.total_nodes,
    completed: execution.completed_nodes,
    failed: execution.failed_nodes,
    running: statusCount.running || 0,
    pending: statusCount.pending || 0,
    paused: statusCount.paused || 0,
    progress: Math.round(execution.progress * 100),
    duration: Math.round(execution.duration / 1000),
    startTime: new Date(execution.start_time).toLocaleString(),
    endTime: execution.end_time ? new Date(execution.end_time).toLocaleString() : null,
    status: execution.status,
    criticalFailed: execution.nodes.filter(n => n.critical && n.status === 'failed').length,
    optionalSkipped: execution.nodes.filter(n => n.optional && n.status === 'pending' && execution.status === 'completed').length
  };
}

/**
 * 格式化节点类型显示名称
 */
export function formatNodeType(type: string): string {
  const typeNames: Record<string, string> = {
    'storage': '存储',
    'config': '配置',
    'service': '服务',
    'auth': '认证',
    'plugin': '插件'
  };
  return typeNames[type] || type;
}

/**
 * 格式化状态显示名称
 */
export function formatNodeStatus(status: string): string {
  const statusNames: Record<string, string> = {
    'pending': '等待中',
    'running': '运行中',
    'completed': '已完成',
    'failed': '失败',
    'paused': '已暂停',
    'cancelled': '已取消'
  };
  return statusNames[status] || status;
}

/**
 * 计算节点布局（自动布局算法）
 */
export function calculateNodeLayout(nodes: StartupWorkflowNode[]): StartupWorkflowNode[] {
  // 简单的分层布局算法
  const levels: StartupWorkflowNode[][] = [];
  const processed = new Set<string>();

  // 找到没有依赖的节点作为第一层
  const firstLevel = nodes.filter(node => node.depends_on.length === 0);
  levels.push(firstLevel);
  firstLevel.forEach(node => processed.add(node.id));

  // 逐层计算
  while (processed.size < nodes.length) {
    const currentLevel: StartupWorkflowNode[] = [];

    nodes.forEach(node => {
      if (processed.has(node.id)) return;

      // 检查所有依赖是否都已处理
      const allDepsProcessed = node.depends_on.every(dep => processed.has(dep));

      if (allDepsProcessed) {
        currentLevel.push(node);
        processed.add(node.id);
      }
    });

    if (currentLevel.length === 0) {
      // 避免死循环，将剩余节点放在当前层
      nodes.forEach(node => {
        if (!processed.has(node.id)) {
          currentLevel.push(node);
          processed.add(node.id);
        }
      });
    }

    if (currentLevel.length > 0) {
      levels.push(currentLevel);
    }
  }

  // 计算位置
  const nodeWidth = 200;
  const nodeHeight = 100;
  const horizontalSpacing = 50;
  const verticalSpacing = 150;

  levels.forEach((level, levelIndex) => {
    const levelWidth = level.length * nodeWidth + (level.length - 1) * horizontalSpacing;
    const startX = (levelWidth / 2) * -1; // 居中对齐
    const y = levelIndex * (nodeHeight + verticalSpacing);

    level.forEach((node, nodeIndex) => {
      node.position.x = startX + nodeIndex * (nodeWidth + horizontalSpacing);
      node.position.y = y;
    });
  });

  return nodes;
}