import { useCallback, useEffect, useState } from 'react';
import { useNodesState, useEdgesState, Node, Edge, Connection } from 'reactflow';
import { WorkflowNodeData } from '../types';
import { startupWebSocketManager, StartupExecution } from '../../../services/startupWebSocket';
import {
  convertStartupWorkflowToReactFlowNodes,
  convertStartupWorkflowToReactFlowEdges,
  updateNodeStyleByExecution,
  updateEdgesAnimation
} from '../../../utils/startupDataConverter';
import { log } from '../../../utils/logger';

// 保留静态数据作为fallback
import { workflowNodes, workflowEdges } from '../data';

interface UseWorkflowStateOptions {
  autoConnect?: boolean;  // 是否自动连接WebSocket
  workflowId?: string;    // 指定工作流ID
  onStart?: (executionId: string) => void;
  onComplete?: (execution: StartupExecution) => void;
  onProgress?: (progress: number) => void;
}

export const useWorkflowState = (options: UseWorkflowStateOptions = {}) => {
  const {
    autoConnect = false,
    workflowId = 'xiaozhi-flow-default-startup',
    onStart,
    onComplete,
    onProgress
  } = options;

  // ReactFlow 状态
  const [nodes, setNodes, onNodesChange] = useNodesState<WorkflowNodeData[]>(workflowNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(workflowEdges);

  // 工作流状态
  const [workflow, setWorkflow] = useState<any>(null);
  const [execution, setExecution] = useState<StartupExecution | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [executionId, setExecutionId] = useState<string>(workflowId);

  // WebSocket 连接管理
  useEffect(() => {
    if (!autoConnect) return;

    let mounted = true;

    const connectWebSocket = async () => {
      try {
        setIsLoading(true);
        setError(null);

        // 连接 WebSocket
        await startupWebSocketManager.connect();

        if (mounted) {
          setIsConnected(true);
          log.info('WebSocket 连接成功', { workflowId }, 'ui', 'useWorkflowState');
        }
      } catch (err) {
        if (mounted) {
          const errorMessage = err instanceof Error ? err.message : 'WebSocket 连接失败';
          setError(errorMessage);
          setIsConnected(false);
          log.error('WebSocket 连接失败', { workflowId, error: errorMessage }, 'ui', 'useWorkflowState');
        }
      } finally {
        if (mounted) {
          setIsLoading(false);
        }
      }
    };

    connectWebSocket();

    // 清理函数
    return () => {
      mounted = false;
    };
  }, [autoConnect, workflowId]);

  // 获取工作流数据
  useEffect(() => {
    if (!isConnected) return;

    const loadWorkflow = async () => {
      try {
        const workflowData = await startupWebSocketManager.getWorkflow(workflowId);
        if (workflowData) {
          setWorkflow(workflowData);

          // 转换为 ReactFlow 格式
          const reactFlowNodes = convertStartupWorkflowToReactFlowNodes(workflowData);
          const reactFlowEdges = convertStartupWorkflowToReactFlowEdges(workflowData);

          setNodes(reactFlowNodes);
          setEdges(reactFlowEdges);

          log.info('工作流数据加载成功', {
            workflowId,
            nodeCount: reactFlowNodes.length,
            edgeCount: reactFlowEdges.length
          }, 'ui', 'useWorkflowState');
        }
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : '加载工作流失败';
        setError(errorMessage);
        log.error('加载工作流失败', { workflowId, error: errorMessage }, 'ui', 'useWorkflowState');
      }
    };

    loadWorkflow();
  }, [isConnected, workflowId, setNodes, setEdges]);

  // WebSocket 事件订阅
  useEffect(() => {
    if (!isConnected) return;

    const handleExecutionStart = (data: any) => {
      setExecution(data);
      setExecutionId(data.id);
      onStart?.( data.id);
      log.info('工作流执行开始', { executionId: data.id }, 'ui', 'useWorkflowState');
    };

    const handleExecutionProgress = (data: any) => {
      if (execution) {
        const updatedExecution = { ...execution, ...data };
        setExecution(updatedExecution);
        onProgress?.( data.progress );

        // 更新节点样式
        const updatedNodes = updateNodeStyleByExecution(nodes, updatedExecution);
        const updatedEdges = updateEdgesAnimation(edges, updatedExecution);

        setNodes(updatedNodes);
        setEdges(updatedEdges);
      }
    };

    const handleExecutionEnd = (data: any) => {
      if (execution) {
        const updatedExecution = { ...execution, ...data };
        setExecution(updatedExecution);
        onComplete?.( updatedExecution );

        // 最终更新节点样式
        const updatedNodes = updateNodeStyleByExecution(nodes, updatedExecution);
        const updatedEdges = updateEdgesAnimation(edges, updatedExecution);

        setNodes(updatedNodes);
        setEdges(updatedEdges);
      }

      log.info('工作流执行结束', { executionId: execution?.id }, 'ui', 'useWorkflowState');
    };

    // 订阅事件
    const unsubscribers = [
      startupWebSocketManager.on('execution_start', handleExecutionStart),
      startupWebSocketManager.on('execution_progress', handleExecutionProgress),
      startupWebSocketManager.on('execution_end', handleExecutionEnd)
    ];

    return () => {
      unsubscribers.forEach(unsub => unsub());
    };
  }, [isConnected, execution, nodes, edges, onStart, onComplete, onProgress, setExecution, setNodes, setEdges]);

  // 处理连接
  const onConnect = useCallback((params: Connection) => {
    log.info('创建新连接', { params }, 'ui', 'useWorkflowState');
    // 这里可以添加连接处理的业务逻辑
  }, []);

  // 执行工作流
  const executeWorkflow = useCallback(async (inputs?: Record<string, any>) => {
    if (!isConnected) {
      throw new Error('WebSocket 未连接');
    }

    if (execution && (execution.status === 'running' || execution.status === 'paused')) {
      throw new Error('工作流正在执行中');
    }

    try {
      const newExecutionId = await startupWebSocketManager.executeWorkflow(workflowId, inputs);
      setExecutionId(newExecutionId);

      log.info('工作流执行请求成功', {
        workflowId,
        executionId: newExecutionId,
        inputs
      }, 'ui', 'useWorkflowState');

      return newExecutionId;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '执行工作流失败';
      setError(errorMessage);
      log.error('执行工作流失败', {
        workflowId,
        inputs,
        error: errorMessage
      }, 'ui', 'useWorkflowState');
      throw err;
    }
  }, [isConnected, execution, workflowId, setExecutionId, setError]);

  // 取消执行
  const cancelExecution = useCallback(async () => {
    if (!executionId) {
      throw new Error('没有正在执行的工作流');
    }

    try {
      await startupWebSocketManager.cancelExecution(executionId);
      log.info('工作流取消成功', { executionId }, 'ui', 'useWorkflowState');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '取消执行失败';
      setError(errorMessage);
      log.error('取消执行失败', {
        executionId,
        error: errorMessage
      }, 'ui', 'useWorkflowState');
      throw err;
    }
  }, [executionId, setError]);

  // 暂停执行
  const pauseExecution = useCallback(async () => {
    if (!executionId) {
      throw new Error('没有正在执行的工作流');
    }

    try {
      await startupWebSocketManager.pauseExecution(executionId);
      log.info('工作流暂停成功', { executionId }, 'ui', 'useWorkflowState');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '暂停执行失败';
      setError(errorMessage);
      log.error('暂停执行失败', {
        executionId,
        error: errorMessage
      }, 'ui', 'useWorkflowState');
      throw err;
    }
  }, [executionId, setError]);

  // 恢复执行
  const resumeExecution = useCallback(async () => {
    if (!executionId) {
      throw new Error('没有正在执行的工作流');
    }

    try {
      await startupWebSocketManager.resumeExecution(executionId);
      log.info('工作流恢复成功', { executionId }, 'ui', 'useWorkflowState');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '恢复执行失败';
      setError(errorMessage);
      log.error('恢复执行失败', {
        executionId,
        error: errorMessage
      }, 'ui', 'useWorkflowState');
      throw err;
    }
  }, [executionId, setError]);

  // 切换工作流
  const switchWorkflow = useCallback(async (newWorkflowId: string) => {
    if (execution && (execution.status === 'running' || execution.status === 'paused')) {
      throw new Error('请先停止当前执行');
    }

    try {
      setWorkflowId(newWorkflowId);

      // 重新加载工作流数据
      const workflowData = await startupWebSocketManager.getWorkflow(newWorkflowId);
      if (workflowData) {
        setWorkflow(workflowData);

        // 转换为 ReactFlow 格式
        const reactFlowNodes = convertStartupWorkflowToReactFlowNodes(workflowData);
        const reactFlowEdges = convertStartupWorkflowToReactFlowEdges(workflowData);

        setNodes(reactFlowNodes);
        setEdges(reactFlowEdges);
      }

      log.info('工作流切换成功', {
        fromWorkflowId: workflowId,
        toWorkflowId: newWorkflowId
      }, 'ui', 'useWorkflowState');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '切换工作流失败';
      setError(errorMessage);
      log.error('切换工作流失败', {
        fromWorkflowId: workflowId,
        toWorkflowId: newWorkflowId,
        error: errorMessage
      }, 'ui', 'useWorkflowState');
      throw err;
    }
  }, [execution, workflowId, setWorkflowId, setNodes, setEdges]);

  // 刷新数据
  const refresh = useCallback(async () => {
    if (!isConnected) return;

    try {
      setError(null);

      // 重新加载工作流数据
      const workflowData = await startupWebSocketManager.getWorkflow(workflowId);
      if (workflowData) {
        setWorkflow(workflowData);

        // 转换为 ReactFlow 格式
        const reactFlowNodes = convertStartupWorkflowToReactFlowNodes(workflowData);
        const reactFlowEdges = convertStartupWorkflowToReactFlowEdges(workflowData);

        setNodes(reactFlowNodes);
        setEdges(reactFlowEdges);
      }

      log.info('工作流数据刷新成功', { workflowId }, 'ui', 'useWorkflowState');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '刷新失败';
      setError(errorMessage);
      log.error('刷新工作流失败', { workflowId, error: errorMessage }, 'ui', 'useWorkflowState');
    }
  }, [isConnected, workflowId, setWorkflow, setNodes, setEdges, setError]);

  // 获取执行状态
  const getExecutionStatus = useCallback(() => {
    if (!executionId) return null;
    return startupWebSocketManager.getExecutionStatus(executionId);
  }, [executionId]);

  return {
    // ReactFlow 状态
    nodes,
    edges,
    onNodesChange,
    onEdgesChange,
    onConnect,

    // 工作流状态
    workflow,
    execution,
    executionId,
    isLoading,
    error,
    isConnected,

    // 操作方法
    executeWorkflow,
    cancelExecution,
    pauseExecution,
    resumeExecution,
    switchWorkflow,
    refresh,
    getExecutionStatus
  };
};