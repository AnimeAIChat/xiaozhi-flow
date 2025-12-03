import { useState, useEffect, useRef, useCallback } from 'react';
import { Editor, Node } from 'rete';
import { useAppStore } from '../stores/useAppStore';
import { useWebSocket } from '../services/startupWebSocket';
import { ReteNode } from '../utils/reteDataConverter';
import { log } from '../utils/logger';

// 启动执行状态接口
export interface StartupExecution {
  id: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'paused';
  progress: number;
  current_nodes: string[];
  nodes: StartupNodeStatus[];
  start_time: string;
  end_time?: string;
  error?: string;
}

// 节点执行状态
export interface StartupNodeStatus {
  id: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped';
  progress: number;
  start_time?: string;
  end_time?: string;
  duration?: number;
  logs: string[];
  error?: string;
}

// 工作流定义接口
export interface StartupWorkflow {
  id: string;
  name: string;
  description?: string;
  nodes: ReteNode[];
  connections: any[];
  created_at: string;
  updated_at: string;
}

/**
 * Rete.js 工作流状态管理 Hook
 */
export const useReteWorkflowState = (editor: Editor | null) => {
  // 状态管理
  const [currentWorkflow, setCurrentWorkflow] = useState<StartupWorkflow | null>(null);
  const [execution, setExecution] = useState<StartupExecution | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedNodes, setSelectedNodes] = useState<Set<string>>(new Set());
  const [nodeDetails, setNodeDetails] = useState<Map<string, any>>(new Map());

  // 全局状态
  const { setDashboardState } = useAppStore();

  // WebSocket 连接
  const {
    subscribe,
    executeWorkflow,
    cancelExecution,
    pauseExecution,
    resumeExecution
  } = useWebSocket();

  // 防抖更新引用
  const updateTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const batchUpdatesRef = useRef<Map<string, Partial<StartupNodeStatus>>>(new Map());

  /**
   * 批量更新节点状态
   */
  const batchUpdateNodeStatus = useCallback((updates: Map<string, Partial<StartupNodeStatus>>) => {
    if (!editor || !currentWorkflow) return;

    try {
      // 批量处理更新
      updates.forEach((update, nodeId) => {
        const node = editor.getNode(nodeId);
        if (node) {
          // 更新节点数据
          node.data = {
            ...node.data,
            ...update
          };
        }
      });

      // 触发重新渲染
      editor.trigger('update');

      log.debug('批量更新节点状态', {
        nodeCount: updates.size
      }, 'ui', 'ReteWorkflowState');
    } catch (error) {
      log.error('批量更新节点状态失败', { error }, 'system', 'ReteWorkflowState');
    }
  }, [editor, currentWorkflow]);

  /**
   * 更新单个节点状态
   */
  const updateNodeStatus = useCallback((
    nodeId: string,
    status: StartupNodeStatus['status'],
    error?: string,
    logs?: string[]
  ) => {
    if (!editor) return;

    // 添加到批量更新队列
    batchUpdatesRef.current.set(nodeId, { status, error });

    // 防抖处理
    if (updateTimeoutRef.current) {
      clearTimeout(updateTimeoutRef.current);
    }

    updateTimeoutRef.current = setTimeout(() => {
      batchUpdateNodeStatus(batchUpdatesRef.current);
      batchUpdatesRef.current.clear();
    }, 100); // 100ms 防抖

    log.debug('更新节点状态', { nodeId, status, error }, 'ui', 'ReteWorkflowState');
  }, [editor, batchUpdateNodeStatus]);

  /**
   * 处理 WebSocket 事件
   */
  const handleWebSocketEvent = useCallback((event: any) => {
    try {
      switch (event.type) {
        case 'execution_start':
          setExecution(event.data);
          setDashboardState({ isExecuting: true });
          log.info('工作流执行开始', { executionId: event.data.id }, 'system', 'ReteWorkflowState');
          break;

        case 'execution_end':
          setExecution(prev => prev ? { ...prev, ...event.data } : event.data);
          setDashboardState({ isExecuting: false });
          log.info('工作流执行结束', { executionId: event.data.id }, 'system', 'ReteWorkflowState');
          break;

        case 'node_start':
          updateNodeStatus(event.data.nodeId, 'running');
          break;

        case 'node_complete':
          updateNodeStatus(event.data.nodeId, 'completed', undefined, event.data.logs);
          break;

        case 'node_error':
          updateNodeStatus(event.data.nodeId, 'failed', event.data.error, event.data.logs);
          break;

        case 'node_progress':
          // 处理进度更新
          if (editor && event.data.nodeId) {
            const node = editor.getNode(event.data.nodeId);
            if (node) {
              node.data.progress = event.data.progress;
              editor.trigger('update');
            }
          }
          break;

        case 'execution_cancelled':
          setExecution(prev => prev ? { ...prev, status: 'failed' } : null);
          setDashboardState({ isExecuting: false });
          log.warn('工作流执行被取消', null, 'system', 'ReteWorkflowState');
          break;

        default:
          log.debug('未处理的 WebSocket 事件', { eventType: event.type }, 'ui', 'ReteWorkflowState');
      }
    } catch (error) {
      log.error('处理 WebSocket 事件失败', { eventType: event.type, error }, 'system', 'ReteWorkflowState');
    }
  }, [updateNodeStatus, setDashboardState, editor]);

  /**
   * 初始化 WebSocket 连接
   */
  useEffect(() => {
    const unsubscribe = subscribe(handleWebSocketEvent);
    return unsubscribe;
  }, [handleWebSocketEvent, subscribe]);

  /**
   * 清理定时器
   */
  useEffect(() => {
    return () => {
      if (updateTimeoutRef.current) {
        clearTimeout(updateTimeoutRef.current);
      }
    };
  }, []);

  /**
   * 执行工作流
   */
  const handleExecuteWorkflow = useCallback(async (workflowId?: string) => {
    if (!currentWorkflow) {
      log.error('没有当前工作流可执行', null, 'system', 'ReteWorkflowState');
      return;
    }

    setIsLoading(true);
    try {
      const executionId = await executeWorkflow(workflowId || currentWorkflow.id);
      log.info('工作流执行请求已发送', { executionId }, 'system', 'ReteWorkflowState');
    } catch (error) {
      log.error('执行工作流失败', { error }, 'system', 'ReteWorkflowState');
    } finally {
      setIsLoading(false);
    }
  }, [currentWorkflow, executeWorkflow]);

  /**
   * 取消执行
   */
  const handleCancelExecution = useCallback(() => {
    if (!execution) return;

    try {
      cancelExecution(execution.id);
      log.info('取消工作流执行', { executionId: execution.id }, 'system', 'ReteWorkflowState');
    } catch (error) {
      log.error('取消执行失败', { error }, 'system', 'ReteWorkflowState');
    }
  }, [execution, cancelExecution]);

  /**
   * 暂停执行
   */
  const handlePauseExecution = useCallback(() => {
    if (!execution) return;

    try {
      pauseExecution(execution.id);
      log.info('暂停工作流执行', { executionId: execution.id }, 'system', 'ReteWorkflowState');
    } catch (error) {
      log.error('暂停执行失败', { error }, 'system', 'ReteWorkflowState');
    }
  }, [execution, pauseExecution]);

  /**
   * 恢复执行
   */
  const handleResumeExecution = useCallback(() => {
    if (!execution) return;

    try {
      resumeExecution(execution.id);
      log.info('恢复工作流执行', { executionId: execution.id }, 'system', 'ReteWorkflowState');
    } catch (error) {
      log.error('恢复执行失败', { error }, 'system', 'ReteWorkflowState');
    }
  }, [execution, resumeExecution]);

  /**
   * 选择节点
   */
  const handleNodeSelect = useCallback((nodeId: string, multiSelect = false) => {
    setSelectedNodes(prev => {
      const newSet = new Set(prev);
      if (multiSelect) {
        if (newSet.has(nodeId)) {
          newSet.delete(nodeId);
        } else {
          newSet.add(nodeId);
        }
      } else {
        newSet.clear();
        newSet.add(nodeId);
      }
      return newSet;
    });

    // 更新编辑器中的节点选择状态
    if (editor) {
      const node = editor.getNode(nodeId);
      if (node) {
        node.selected = true;
        editor.trigger('update');
      }
    }
  }, [editor]);

  /**
   * 清除选择
   */
  const handleClearSelection = useCallback(() => {
    setSelectedNodes(new Set());

    // 清除编辑器中的选择状态
    if (editor) {
      editor.getNodes().forEach(node => {
        node.selected = false;
      });
      editor.trigger('update');
    }
  }, [editor]);

  /**
   * 获取节点详情
   */
  const getNodeDetails = useCallback((nodeId: string) => {
    return nodeDetails.get(nodeId) || null;
  }, [nodeDetails]);

  /**
   * 设置节点详情
   */
  const setNodeDetailsInternal = useCallback((nodeId: string, details: any) => {
    setNodeDetails(prev => new Map(prev.set(nodeId, details)));
  }, []);

  return {
    // 状态
    currentWorkflow,
    execution,
    isLoading,
    selectedNodes,
    nodeDetails,

    // 操作方法
    setCurrentWorkflow,
    handleExecuteWorkflow,
    handleCancelExecution,
    handlePauseExecution,
    handleResumeExecution,
    handleNodeSelect,
    handleClearSelection,
    getNodeDetails,
    setNodeDetails: setNodeDetailsInternal,

    // 工具方法
    updateNodeStatus,
    batchUpdateNodeStatus
  };
};