import React, { useState, useRef, useCallback, useEffect } from 'react';
import { Card, Button, Select, Space, Tooltip, Modal, Form, Input, InputNumber, App, Badge, Progress } from 'antd';
import { DatabaseOutlined, ApiOutlined, RobotOutlined, CloudOutlined, SettingOutlined, PlusOutlined, DeleteOutlined, EditOutlined, PlayCircleOutlined, PauseCircleOutlined, StopOutlined, ReloadOutlined } from '@ant-design/icons';
import { BaseNode, NodeData } from './nodes';

const { Option } = Select;

interface SimpleNode extends BaseNode {
  x: number;
  y: number;
  id: string;
  data: NodeData;
}

interface SimpleConnection {
  id: string;
  from: string;
  to: string;
  fromOutput: string;
  toInput: string;
}

export const SimpleReteEditor: React.FC<{
  onNodesChange?: (nodes: SimpleNode[]) => void;
  onConnectionsChange?: (connections: SimpleConnection[]) => void;
  onExecute?: () => void;
  workflowId?: string;
  autoConnect?: boolean;
  adapter?: any;
}> = ({
  onNodesChange,
  onConnectionsChange,
  onExecute,
  workflowId = 'xiaozhi-flow-default-startup',
  autoConnect = true,
  adapter
}) => {
  const { message } = App.useApp();
  const [nodes, setNodes] = useState<SimpleNode[]>([]);
  const [connections, setConnections] = useState<SimpleConnection[]>([]);
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  const [editingNode, setEditingNode] = useState<SimpleNode | null>(null);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [form] = Form.useForm();
  const [isDragging, setIsDragging] = useState(false);
  const [draggedNode, setDraggedNode] = useState<string | null>(null);
  const [mousePos, setMousePos] = useState({ x: 0, y: 0 });
  const canvasRef = useRef<HTMLDivElement>(null);

  // 启动流程相关状态
  const [startupAdapter] = useState(() => {
    if (adapter) return adapter;
    throw new Error("Adapter is required");
  });
  const [isLoading, setIsLoading] = useState(false);
  const [execution, setExecution] = useState<any>(null);
  const [executionStats, setExecutionStats] = useState<any>(null);
  const [pollingInterval, setPollingInterval] = useState<NodeJS.Timeout | null>(null);
  const [availableNodeTypes, setAvailableNodeTypes] = useState<any[]>([]);

  // 使用 useRef 来防止重复初始化
  const hasInitialized = useRef(false);

  // 加载可用节点类型
  useEffect(() => {
    if (adapter && adapter.getCapabilities) {
      adapter.getCapabilities().then((caps: any[]) => {
        setAvailableNodeTypes(Array.isArray(caps) ? caps : []);
      }).catch((err: any) => {
        console.error('Failed to load capabilities:', err);
        setAvailableNodeTypes([]);
      });
    }
  }, [adapter]);

  // 初始化和连接管理
  useEffect(() => {
    if (!autoConnect || hasInitialized.current) return;

    const initializeStartupFlow = async () => {
      try {
        setIsLoading(true);
        message.loading('正在加载启动流程...', 0);

        // 获取工作流数据
        const workflow = await startupAdapter.getWorkflow(workflowId);
        if (workflow) {
          const workflowNodes = startupAdapter.convertWorkflowToEditorNodes(workflow);
          const workflowConnections = startupAdapter.convertWorkflowToEditorConnections(workflow);

          setNodes(workflowNodes);
          setConnections(workflowConnections);

          // 使用最新的回调函数
          onNodesChange?.(workflowNodes);
          onConnectionsChange?.(workflowConnections);

          message.success('启动流程数据加载成功');
        }
      } catch (error) {
        console.error('初始化启动流程失败:', error);
        message.error('启动流程加载失败，使用模拟数据');

        // 降级到模拟数据
        const fallbackNodes: SimpleNode[] = [
          {
            id: 'storage-init',
            data: {
              label: '存储初始化',
              type: 'database',
              status: 'stopped',
              description: '初始化数据库连接和存储系统',
              metrics: { '关键节点': '是', '超时时间': '30s' }
            },
            x: 100,
            y: 100
          } as SimpleNode,
          {
            id: 'config-load',
            data: {
              label: '配置加载',
              type: 'config',
              status: 'stopped',
              description: '加载系统配置和环境变量',
              metrics: { '关键节点': '是', '超时时间': '10s' }
            },
            x: 400,
            y: 100
          } as SimpleNode
        ];

        setNodes(fallbackNodes);
        onNodesChange?.(fallbackNodes);
        setConnections([]);
        onConnectionsChange?.([]);
      } finally {
        setIsLoading(false);
        message.destroy();
      }
    };

    hasInitialized.current = true;
    initializeStartupFlow();

    return () => {
      // 清理轮询
      if (pollingInterval) {
        clearInterval(pollingInterval);
        setPollingInterval(null);
      }
    };

    const initializeMockData = () => {
      const initialNodes: SimpleNode[] = [
        {
          id: 'storage-init',
          data: {
            label: '存储初始化',
            type: 'database',
            status: 'stopped',
            description: '初始化数据库连接和存储系统',
            metrics: { '关键节点': '是', '超时时间': '30s' }
          },
          x: 100,
          y: 100
        } as SimpleNode,
        {
          id: 'config-load',
          data: {
            label: '配置加载',
            type: 'config',
            status: 'stopped',
            description: '加载系统配置和环境变量',
            metrics: { '关键节点': '是', '超时时间': '10s' }
          },
          x: 400,
          y: 100
        } as SimpleNode,
        {
          id: 'service-start',
          data: {
            label: '服务启动',
            type: 'api',
            status: 'stopped',
            description: '启动核心服务组件',
            metrics: { '关键节点': '是', '超时时间': '60s' }
          },
          x: 700,
          y: 100
        } as SimpleNode,
        {
          id: 'auth-setup',
          data: {
            label: '认证设置',
            type: 'api',
            status: 'stopped',
            description: '配置认证和授权系统',
            metrics: { '关键节点': '是', '超时时间': '20s' }
          },
          x: 100,
          y: 300
        } as SimpleNode,
        {
          id: 'plugin-load',
          data: {
            label: '插件加载',
            type: 'cloud',
            status: 'stopped',
            description: '加载和初始化插件系统',
            metrics: { '可选节点': '是', '超时时间': '30s' }
          },
          x: 400,
          y: 300
        } as SimpleNode
      ];

      setNodes(initialNodes);

      const initialConnections: SimpleConnection[] = [
        {
          id: 'conn-1',
          from: 'storage-init',
          to: 'config-load',
          fromOutput: 'data',
          toInput: 'input'
        },
        {
          id: 'conn-2',
          from: 'config-load',
          to: 'service-start',
          fromOutput: 'response',
          toInput: 'prompt'
        },
        {
          id: 'conn-3',
          from: 'storage-init',
          to: 'auth-setup',
          fromOutput: 'data',
          toInput: 'input'
        },
        {
          id: 'conn-4',
          from: 'config-load',
          to: 'plugin-load',
          fromOutput: 'response',
          toInput: 'prompt'
        },
        {
          id: 'conn-5',
          from: 'auth-setup',
          to: 'plugin-load',
          fromOutput: 'response',
          toInput: 'context'
        }
      ];

      setConnections(initialConnections);
    };

    initializeStartupFlow();

    return () => {
      // 清理轮询定时器
      if (pollingInterval) {
        clearInterval(pollingInterval);
        setPollingInterval(null);
      }
    };
  }, [workflowId, autoConnect]); // 移除会导致重新执行的不必要依赖项

  // 添加新节点
  const addNode = useCallback((nodeType: NodeData['type']) => {
    const newNode: SimpleNode = {
      id: `${nodeType}-${Date.now()}`,
      data: {
        label: `New ${nodeType.charAt(0).toUpperCase() + nodeType.slice(1)}`,
        type: nodeType,
        status: 'stopped',
        description: `${nodeType} 节点`,
        metrics: {}
      },
      x: 300,
      y: 200
    };

    setNodes(prev => {
      const updated = [...prev, newNode];
      onNodesChange?.(updated);
      return updated;
    });
  }, [onNodesChange]);

  // 删除节点
  const deleteNode = useCallback((nodeId: string) => {
    setNodes(prev => {
      const updated = prev.filter(n => n.id !== nodeId);
      onNodesChange?.(updated);
      return updated;
    });

    setConnections(prev => {
      const updated = prev.filter(c => c.from !== nodeId && c.to !== nodeId);
      onConnectionsChange?.(updated);
      return updated;
    });
  }, [onNodesChange, onConnectionsChange]);

  // 编辑节点
  const startEditNode = useCallback((node: SimpleNode) => {
    setEditingNode(node);
    setEditModalVisible(true);
    form.setFieldsValue({
      label: node.data.label,
      description: node.data.description
    });
  }, [form]);

  // 保存节点编辑
  const saveNodeEdit = useCallback(() => {
    if (!editingNode) return;

    form.validateFields().then(values => {
      setNodes(prev => {
        const updated = prev.map(node =>
          node.id === editingNode.id
            ? {
                ...node,
                data: {
                  ...node.data,
                  label: values.label,
                  description: values.description
                }
              }
            : node
        );
        onNodesChange?.(updated);
        return updated;
      });

      setEditModalVisible(false);
      setEditingNode(null);
    });
  }, [editingNode, form, onNodesChange]);

  // 拖拽功能
  const handleMouseDown = useCallback((e: React.MouseEvent, nodeId: string) => {
    e.preventDefault();
    setIsDragging(true);
    setDraggedNode(nodeId);
    setSelectedNode(nodeId);

    const rect = canvasRef.current?.getBoundingClientRect();
    if (rect) {
      setMousePos({
        x: e.clientX - rect.left,
        y: e.clientY - rect.top
      });
    }
  }, []);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!isDragging || !draggedNode) return;

    const rect = canvasRef.current?.getBoundingClientRect();
    if (rect) {
      const currentX = e.clientX - rect.left;
      const currentY = e.clientY - rect.top;

      setNodes(prev => prev.map(node =>
        node.id === draggedNode
          ? {
              ...node,
              x: node.x + (currentX - mousePos.x),
              y: node.y + (currentY - mousePos.y)
            }
          : node
      ));

      setMousePos({ x: currentX, y: currentY });
    }
  }, [isDragging, draggedNode, mousePos]);

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
    setDraggedNode(null);
  }, []);

  // 创建连接
  const createConnection = useCallback((fromNode: string, toNode: string) => {
    const newConnection: SimpleConnection = {
      id: `conn-${Date.now()}`,
      from: fromNode,
      to: toNode,
      fromOutput: 'data',
      toInput: 'input'
    };

    setConnections(prev => {
      const updated = [...prev, newConnection];
      onConnectionsChange?.(updated);
      return updated;
    });
  }, [onConnectionsChange]);

  // 执行工作流
  const executeWorkflow = useCallback(async () => {
    try {
      // 清理之前的轮询
      if (pollingInterval) {
        clearInterval(pollingInterval);
        setPollingInterval(null);
      }

      message.loading('正在执行启动流程...', 0);

      const executionId = await startupAdapter.executeWorkflow(workflowId);
      message.success(`启动流程执行开始: ${executionId}`);

      // 创建初始执行对象
      const mockExecution = startupAdapter.createMockExecution(workflowId, nodes);
      setExecution(mockExecution);
      const initialStats = startupAdapter.generateExecutionStats(mockExecution);
      setExecutionStats(initialStats);

      // 开始轮询执行状态
      startupAdapter.pollExecutionStatus(executionId, (status) => {
        const updatedExecution = { ...execution, ...status };
        setExecution(updatedExecution);

        // 更新节点状态
        const updatedNodes = startupAdapter.updateNodesByExecution(nodes, updatedExecution);
        const updatedConnections = startupAdapter.updateConnectionsByExecution(connections, updatedExecution);

        setNodes(updatedNodes);
        setConnections(updatedConnections);

        // 更新统计信息
        const stats = startupAdapter.generateExecutionStats(updatedExecution);
        setExecutionStats(stats);

        onNodesChange?.(updatedNodes);
        onConnectionsChange?.(updatedConnections);

        // 执行完成时停止轮询
        if (status.status === 'completed' || status.status === 'failed' || status.status === 'cancelled') {
          if (pollingInterval) {
            clearInterval(pollingInterval);
            setPollingInterval(null);
          }
          message.success(`启动流程${status.status === 'completed' ? '执行完成' : '执行结束'}`);
        }
      });

      onExecute?.();
    } catch (error) {
      console.error('执行启动流程失败:', error);
      message.error('执行启动流程失败，使用模拟执行');
      simulateExecution();
    }
  }, [startupAdapter, workflowId, nodes, connections, onExecute, execution, pollingInterval, onNodesChange, onConnectionsChange]);

  // 模拟执行（用于演示或降级）
  const simulateExecution = useCallback(() => {
    let currentNodeIndex = 0;
    const executeNextNode = () => {
      if (currentNodeIndex >= nodes.length) {
        message.success('模拟执行完成');
        return;
      }

      const node = nodes[currentNodeIndex];

      // 设置为运行中
      setNodes(prev => prev.map(n =>
        n.id === node.id
          ? { ...n, data: { ...n.data, status: 'running' } }
          : n
      ));

      // 模拟执行进度
      let progress = 0;
      const progressInterval = setInterval(() => {
        progress += 0.1;
        if (progress >= 1) {
          clearInterval(progressInterval);

          // 设置为完成
          setNodes(prev => prev.map(n =>
            n.id === node.id
              ? { ...n, data: { ...n.data, status: 'stopped', metrics: { ...n.data.metrics, '进度': '100%' } } }
              : n
          ));

          currentNodeIndex++;
          setTimeout(executeNextNode, 500);
        } else {
          setNodes(prev => prev.map(n =>
            n.id === node.id
              ? { ...n, data: { ...n.data, metrics: { ...n.data.metrics, '进度': `${Math.round(progress * 100)}%` } } }
              : n
          ));
        }
      }, 200);
    };

    executeNextNode();
    onExecute?.();
  }, [nodes, onExecute]);

  // 取消执行
  const cancelExecution = useCallback(async () => {
    if (!execution) {
      message.warning('没有正在执行的流程');
      return;
    }

    try {
      await startupAdapter.cancelExecution(execution.id);
      message.success('流程已取消');
    } catch (error) {
      console.error('取消执行失败:', error);
      message.error('取消执行失败');
    }
  }, [execution, startupAdapter]);

  // 暂停执行
  const pauseExecution = useCallback(async () => {
    if (!execution) {
      message.warning('没有正在执行的流程');
      return;
    }

    try {
      await startupAdapter.pauseExecution(execution.id);
      message.success('流程已暂停');
    } catch (error) {
      console.error('暂停执行失败:', error);
      message.error('暂停执行失败');
    }
  }, [execution, startupAdapter]);

  // 恢复执行
  const resumeExecution = useCallback(async () => {
    if (!execution) {
      message.warning('没有正在执行的流程');
      return;
    }

    try {
      await startupAdapter.resumeExecution(execution.id);
      message.success('流程已恢复');
    } catch (error) {
      console.error('恢复执行失败:', error);
      message.error('恢复执行失败');
    }
  }, [execution, startupAdapter]);

  // 保存工作流
  const handleSave = useCallback(async () => {
    if (adapter && adapter.saveWorkflow) {
      try {
        await adapter.saveWorkflow(workflowId, nodes, connections);
        message.success('工作流保存成功');
      } catch (error) {
        console.error('保存工作流失败:', error);
        message.error('保存工作流失败');
      }
    } else {
      message.warning('当前适配器不支持保存功能');
    }
  }, [adapter, workflowId, nodes, connections]);

  // 获取节点图标
  const getNodeIcon = (type: NodeData['type']) => {
    switch (type) {
      case 'database': return <DatabaseOutlined style={{ color: '#1890ff', fontSize: '20px' }} />;
      case 'api': return <ApiOutlined style={{ color: '#52c41a', fontSize: '20px' }} />;
      case 'ai': return <RobotOutlined style={{ color: '#722ed1', fontSize: '20px' }} />;
      case 'cloud': return <CloudOutlined style={{ color: '#fa8c16', fontSize: '20px' }} />;
      case 'config': return <SettingOutlined style={{ color: '#eb2f96', fontSize: '20px' }} />;
      default: return <SettingOutlined style={{ fontSize: '20px' }} />;
    }
  };

  // 获取状态颜色
  const getStatusColor = (status: NodeData['status']) => {
    switch (status) {
      case 'running': return '#52c41a';
      case 'warning': return '#faad14';
      case 'stopped': return '#ff4d4f';
      default: return '#d9d9d9';
    }
  };

  // 获取连接路径
  const getConnectionPath = (connection: SimpleConnection) => {
    const fromNode = nodes.find(n => n.id === connection.from);
    const toNode = nodes.find(n => n.id === connection.to);

    if (!fromNode || !toNode) return '';

    const x1 = fromNode.x + 150; // 节点宽度的一半
    const y1 = fromNode.y + 50;  // 节点高度的一半
    const x2 = toNode.x;
    const y2 = toNode.y + 50;

    const dx = x2 - x1;
    const controlPoint1X = x1 + dx * 0.5;
    const controlPoint1Y = y1;
    const controlPoint2X = x2 - dx * 0.5;
    const controlPoint2Y = y2;

    return `M ${x1} ${y1} C ${controlPoint1X} ${controlPoint1Y}, ${controlPoint2X} ${controlPoint2Y}, ${x2} ${y2}`;
  };

  return (
    <div className="w-full h-full relative bg-gray-50 overflow-hidden">

      {/* 状态信息 */}
      <div className="absolute bottom-4 left-4 z-10 bg-white rounded-lg shadow-lg p-4">
        <div className="text-sm font-semibold text-gray-600 mb-2">工作流信息</div>
        <div className="text-xs text-gray-600 space-y-1">
          <div>工作流ID: {workflowId}</div>
          <div>节点数量: {nodes.length}</div>
          <div>连接数量: {connections.length}</div>
          <div>选中: {selectedNode ? '1' : '0'} 个节点</div>
          {executionStats && (
            <>
              <div className="border-t pt-1 mt-2">
                <div>执行进度: {executionStats.progress}%</div>
                <div>已完成: {executionStats.completed}</div>
                <div>失败: {executionStats.failed}</div>
                <div>执行时间: {executionStats.duration}s</div>
              </div>
            </>
          )}
        </div>
      </div>

      {/* 画布 */}
      <div
        ref={canvasRef}
        className="w-full h-full"
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        onClick={() => setSelectedNode(null)}
      >
        {/* SVG 连接线 */}
        <svg className="absolute inset-0 pointer-events-none" style={{ width: '100%', height: '100%' }}>
          {connections.map(connection => (
            <path
              key={connection.id}
              d={getConnectionPath(connection)}
              fill="none"
              stroke="#1890ff"
              strokeWidth="2"
              strokeDasharray={nodes.find(n => n.id === connection.from)?.data.status === 'running' ? '5,5' : '0'}
              opacity="0.6"
            />
          ))}
        </svg>

        {/* 节点 */}
        {nodes.map(node => (
          <Card
            key={node.id}
            className={`absolute cursor-move transition-all duration-200 ${
              selectedNode === node.id ? 'ring-2 ring-blue-500' : ''
            }`}
            style={{
              left: `${node.x}px`,
              top: `${node.y}px`,
              width: '300px',
              borderLeft: `4px solid ${getStatusColor(node.data.status)}`,
              boxShadow: selectedNode === node.id ? '0 4px 12px rgba(24, 144, 255, 0.15)' : '0 2px 8px rgba(0,0,0,0.1)'
            }}
            onMouseDown={(e) => handleMouseDown(e, node.id)}
            onClick={(e) => {
              e.stopPropagation();
              setSelectedNode(node.id);
            }}
            onDoubleClick={(e) => {
              e.stopPropagation();
              startEditNode(node);
            }}
            size="small"
            hoverable
          >
            <div className="flex items-start space-x-3">
              <div className="flex-shrink-0">
                {getNodeIcon(node.data.type)}
              </div>
              <div className="flex-grow min-w-0">
                <div className="font-semibold text-gray-800 text-sm truncate">{node.data.label}</div>
                {node.data.description && (
                  <div className="text-xs text-gray-500 mt-1 truncate">{node.data.description}</div>
                )}
                <div className="flex items-center justify-between mt-2">
                  <div className="flex items-center space-x-2">
                    <div
                      className="w-2 h-2 rounded-full"
                      style={{ backgroundColor: getStatusColor(node.data.status) }}
                    />
                    <span className="text-xs text-gray-500">{node.data.status}</span>
                  </div>
                  <div className="flex space-x-1">
                    {node.data.metrics && Object.keys(node.data.metrics).length > 0 && (
                      <div className="text-xs text-gray-400">
                        {Object.keys(node.data.metrics).length} 指标
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
};