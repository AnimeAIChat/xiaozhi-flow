import { useCallback, useEffect, useState, useRef } from 'react';
import { useNodesState, useEdgesState, addEdge, Node, Edge, Connection } from '@xyflow/react';
import { WorkflowNodeData } from '../types';
import { apiService } from '../../../services/api';
import { startupWebSocketManager, StartupExecution, StartupWorkflow } from '../../../services/startupWebSocket';
import {
  convertStartupWorkflowToReactFlowNodes,
  convertStartupWorkflowToReactFlowEdges,
  updateNodeStyleByExecution,
  updateEdgesAnimation,
  calculateNodeLayout
} from '../../../utils/startupDataConverter';
import { log } from '../../../utils/logger';

// 保留静态数据作为fallback
import { workflowNodes, workflowEdges } from '../data';

interface UseWorkflowStateOptions {
  autoConnect?: boolean;  // 是否自动连接WebSocket
  workflowId?: string;    // 指定工作流ID
  executionId?: string;   // 指定执行ID
}

export const useWorkflowState = (options: UseWorkflowStateOptions = {}) => {
  const {
    autoConnect = true,
    workflowId: initialWorkflowId = 'xiaozhi-flow-default-startup',
    executionId: initialExecutionId
  } = options;

  // 状态管理
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [workflow, setWorkflow] = useState<StartupWorkflow | null>(null);
  const [execution, setExecution] = useState<StartupExecution | null>(null);
  const [workflowId, setWorkflowId] = useState(initialWorkflowId);
  const [executionId, setExecutionId] = useState(initialExecutionId || null);
  const [isConnected, setIsConnected] = useState(false);

  // 使用静态数据作为初始状态
  const [nodes, setNodes, onNodesChange] = useNodesState<Node<WorkflowNodeData>>(workflowNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(workflowEdges);

  // WebSocket消息处理器引用
  const handlersRef = useRef<{
    executionStart: ((message: any) => void) | null;
    executionProgress: ((message: any) => void) | null;
    executionEnd: ((message: any) => void) | null;
    nodeStart: ((message: any) => void) | null;
    nodeProgress: ((message: any) => void) | null;
    nodeComplete: ((message: any) => void) | null;
    nodeError: ((message: any) => void) | null;
  }>({
    executionStart: null,
    executionProgress: null,
    executionEnd: null,
    nodeStart: null,
    nodeProgress: null,
    nodeComplete: null,
    nodeError: null,
  });

  // 连接状态
  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge({ ...params, type: 'smoothstep' }, eds)),
    [setEdges]
  );

  // 加载工作流定义
  const loadWorkflow = useCallback(async (id: string) => {
    try {
      setIsLoading(true);
      setError(null);

      log.info('加载启动工作流', { workflow_id: id }, 'workflow', 'useWorkflowState');

      // 尝试从API获取工作流定义
      try {
        console.log('🔄 正在获取启动工作流数据...', { workflow_id: id });
        const workflowData = await apiService.getStartupWorkflow(id);
        console.log('✅ 成功获取工作流数据:', workflowData);
        setWorkflow(workflowData);

        // 计算节点布局
        const nodesWithLayout = workflowData.nodes ?
          calculateNodeLayout(workflowData.nodes) : workflowData.nodes;

        const workflowWithLayout = {
          ...workflowData,
          nodes: nodesWithLayout
        };

        // 转换为ReactFlow格式
        console.log('🔄 转换数据格式...', { nodes_count: workflowWithLayout.nodes?.length });
        const reactFlowNodes = convertStartupWorkflowToReactFlowNodes(workflowWithLayout);
        const reactFlowEdges = convertStartupWorkflowToReactFlowEdges(workflowWithLayout);

        console.log('✅ 转换完成:', {
          nodes_count: reactFlowNodes.length,
          edges_count: reactFlowEdges.length,
          first_node: reactFlowNodes[0]?.data?.label
        });

        setNodes(reactFlowNodes);
        setEdges(reactFlowEdges);

        log.info('成功加载启动工作流', {
          workflow_id: id,
          nodes_count: reactFlowNodes.length,
          edges_count: reactFlowEdges.length
        }, 'workflow', 'useWorkflowState');

      } catch (apiError) {
        console.error('❌ API加载工作流失败:', apiError);
        log.warn('API加载工作流失败，使用静态数据', {
          workflow_id: id,
          error: apiError.message
        }, 'workflow', 'useWorkflowState');

        // API失败时使用静态数据
        console.warn('⚠️ Fallback到静态数据');
        setNodes(workflowNodes);
        setEdges(workflowEdges);
      }

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '加载工作流失败';
      setError(errorMessage);
      log.error('加载工作流失败', {
        workflow_id: id,
        error: errorMessage
      }, 'workflow', 'useWorkflowState');

      // 出错时使用静态数据
      setNodes(workflowNodes);
      setEdges(workflowEdges);
    } finally {
      setIsLoading(false);
    }
  }, [setNodes, setEdges]);

  // 连接WebSocket
  const connectWebSocket = useCallback(async () => {
    try {
      if (startupWebSocketManager.isConnected()) {
        setIsConnected(true);
        return;
      }

      log.info('连接启动流程WebSocket', null, 'workflow', 'useWorkflowState');
      await startupWebSocketManager.connect();
      setIsConnected(true);

      // 如果有执行ID，订阅该执行的事件
      if (executionId) {
        startupWebSocketManager.subscribe(executionId);
      }

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'WebSocket连接失败';
      log.warn('WebSocket连接失败', { error: errorMessage }, 'workflow', 'useWorkflowState');
      setIsConnected(false);
    }
  }, [executionId]);

  // 设置WebSocket消息处理器
  const setupWebSocketHandlers = useCallback(() => {
    // 移除旧的处理器
    Object.values(handlersRef.current).forEach(handler => {
      if (handler) {
        startupWebSocketManager.off('*', handler);
      }
    });

    // 执行开始
    const handleExecutionStart = (message: any) => {
      log.info('工作流执行开始', message.data, 'workflow', 'useWorkflowState');
      if (message.data.execution_id) {
        setExecutionId(message.data.execution_id);
        startupWebSocketManager.subscribe(message.data.execution_id);
      }
    };

    // 执行进度更新
    const handleExecutionProgress = (message: any) => {
      log.debug('工作流执行进度', message.data, 'workflow', 'useWorkflowState');
      setExecution(prev => {
        const updated = { ...prev, ...message.data };

        // 更新节点和边的样式
        if (nodes.length > 0) {
          const updatedNodes = updateNodeStyleByExecution(nodes, updated);
          const updatedEdges = updateEdgesAnimation(edges, updated);
          setNodes(updatedNodes);
          setEdges(updatedEdges);
        }

        return updated;
      });
    };

    // 执行结束
    const handleExecutionEnd = (message: any) => {
      log.info('工作流执行结束', message.data, 'workflow', 'useWorkflowState');
      setExecution(prev => ({ ...prev, ...message.data }));
    };

    // 节点开始
    const handleNodeStart = (message: any) => {
      log.debug('节点开始执行', message.data, 'workflow', 'useWorkflowState');
      // 可以在这里添加节点级别的动画效果
    };

    // 节点进度
    const handleNodeProgress = (message: any) => {
      log.debug('节点执行进度', message.data, 'workflow', 'useWorkflowState');
    };

    // 节点完成
    const handleNodeComplete = (message: any) => {
      log.debug('节点执行完成', message.data, 'workflow', 'useWorkflowState');
    };

    // 节点错误
    const handleNodeError = (message: any) => {
      log.error('节点执行错误', message.data, 'workflow', 'useWorkflowState');
    };

    // 注册处理器
    startupWebSocketManager.on('execution_start', handleExecutionStart);
    startupWebSocketManager.on('execution_progress', handleExecutionProgress);
    startupWebSocketManager.on('execution_end', handleExecutionEnd);
    startupWebSocketManager.on('node_start', handleNodeStart);
    startupWebSocketManager.on('node_progress', handleNodeProgress);
    startupWebSocketManager.on('node_complete', handleNodeComplete);
    startupWebSocketManager.on('node_error', handleNodeError);

    // 保存处理器引用
    handlersRef.current = {
      executionStart: handleExecutionStart,
      executionProgress: handleExecutionProgress,
      executionEnd: handleExecutionEnd,
      nodeStart: handleNodeStart,
      nodeProgress: handleNodeProgress,
      nodeComplete: handleNodeComplete,
      nodeError: handleNodeError,
    };
  }, [nodes, edges, setNodes, setEdges]);

  // 执行工作流
  const executeWorkflow = useCallback(async (inputs?: Record<string, any>) => {
    if (!isConnected) {
      throw new Error('WebSocket未连接');
    }

    try {
      log.info('执行启动工作流', { workflow_id: workflowId }, 'workflow', 'useWorkflowState');
      const newExecutionId = await startupWebSocketManager.executeWorkflow(workflowId, inputs);
      setExecutionId(newExecutionId);
      return newExecutionId;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '执行工作流失败';
      log.error('执行工作流失败', {
        workflow_id: workflowId,
        error: errorMessage
      }, 'workflow', 'useWorkflowState');
      throw err;
    }
  }, [workflowId, isConnected]);

  // 获取执行状态
  const getExecutionStatus = useCallback(async (id?: string) => {
    const targetId = id || executionId;
    if (!targetId) {
      throw new Error('未指定执行ID');
    }

    try {
      const status = await apiService.getStartupExecutionStatus(targetId);
      setExecution(status);
      return status;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '获取执行状态失败';
      log.error('获取执行状态失败', {
        execution_id: targetId,
        error: errorMessage
      }, 'workflow', 'useWorkflowState');
      throw err;
    }
  }, [executionId]);

  // 取消执行
  const cancelExecution = useCallback(async (id?: string) => {
    const targetId = id || executionId;
    if (!targetId) {
      throw new Error('未指定执行ID');
    }

    try {
      if (isConnected) {
        startupWebSocketManager.cancelExecution(targetId);
      } else {
        await apiService.cancelStartupExecution(targetId);
      }

      if (targetId === executionId) {
        setExecution(null);
        setExecutionId(null);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '取消执行失败';
      log.error('取消执行失败', {
        execution_id: targetId,
        error: errorMessage
      }, 'workflow', 'useWorkflowState');
      throw err;
    }
  }, [executionId, isConnected]);

  // 暂停执行
  const pauseExecution = useCallback(async (id?: string) => {
    const targetId = id || executionId;
    if (!targetId) {
      throw new Error('未指定执行ID');
    }

    try {
      if (isConnected) {
        startupWebSocketManager.pauseExecution(targetId);
      } else {
        await apiService.pauseStartupExecution(targetId);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '暂停执行失败';
      log.error('暂停执行失败', {
        execution_id: targetId,
        error: errorMessage
      }, 'workflow', 'useWorkflowState');
      throw err;
    }
  }, [executionId, isConnected]);

  // 恢复执行
  const resumeExecution = useCallback(async (id?: string) => {
    const targetId = id || executionId;
    if (!targetId) {
      throw new Error('未指定执行ID');
    }

    try {
      if (isConnected) {
        startupWebSocketManager.resumeExecution(targetId);
      } else {
        await apiService.resumeStartupExecution(targetId);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '恢复执行失败';
      log.error('恢复执行失败', {
        execution_id: targetId,
        error: errorMessage
      }, 'workflow', 'useWorkflowState');
      throw err;
    }
  }, [executionId, isConnected]);

  // 切换工作流
  const switchWorkflow = useCallback(async (id: string) => {
    setWorkflowId(id);
    setExecution(null);
    setExecutionId(null);
    await loadWorkflow(id);
  }, [loadWorkflow]);

  // 初始化
  useEffect(() => {
    if (autoConnect) {
      connectWebSocket();
    }
    loadWorkflow(workflowId);
  }, []);

  // WebSocket连接状态变化时设置处理器
  useEffect(() => {
    if (isConnected) {
      setupWebSocketHandlers();
    }
  }, [isConnected, setupWebSocketHandlers]);

  // 清理
  useEffect(() => {
    return () => {
      // 移除WebSocket处理器
      Object.values(handlersRef.current).forEach(handler => {
        if (handler) {
          startupWebSocketManager.off('*', handler);
        }
      });
    };
  }, []);

  return {
    // 基础数据
    nodes,
    edges,
    workflow,
    execution,

    // 状态
    isLoading,
    error,
    isConnected,
    workflowId,
    executionId,

    // ReactFlow回调
    onNodesChange,
    onEdgesChange,
    onConnect,

    // 工作流操作
    loadWorkflow,
    switchWorkflow,
    executeWorkflow,
    getExecutionStatus,
    cancelExecution,
    pauseExecution,
    resumeExecution,
    connectWebSocket,

    // 额外方法
    refresh: () => loadWorkflow(workflowId),
  };
};