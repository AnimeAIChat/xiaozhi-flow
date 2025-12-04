/**
 * 配置视图组件
 * 基于ConfigEditor的ConfigCanvas，适配为Dashboard的视图模式
 * 集成侧边栏和悬浮组件库面板
 */

import React, { useCallback, useEffect, useState, useRef, useMemo } from 'react';
import {
  ReactFlow,
  Node,
  Edge,
  addEdge,
  Connection,
  useNodesState,
  useEdgesState,
  Controls,
  MiniMap,
  Background,
  BackgroundVariant,
  Position,
  Handle,
  NodeProps,
  ReactFlowProvider,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import {
  Card,
  Button,
  Space,
  Input,
  Select,
  Switch,
  Form,
  InputNumber,
  Tooltip,
  Badge,
  Typography,
  message,
  Tag,
  Popconfirm,
} from 'antd';
import {
  SettingOutlined,
  SaveOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  EyeInvisibleOutlined,
  FolderOpenOutlined,
  CodeOutlined,
  DatabaseOutlined,
  PlayCircleOutlined,
  StopOutlined,
  ApiOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import { configService } from '../../../../services/configService';
import { log } from '../../../../utils/logger';
import type { ConfigNode, ConfigEdge, ConfigEditMode, ConfigRecord } from '../../../../types/config';
import type { ConfigViewProps } from '../../types';
import { ConfigSidebar } from './Sidebar';
import { FloatingComponentLibrary } from './FloatingPanel';
import DragHandler from './DragHandler/DragHandler';
import DynamicNodeRenderer from '../../../../nodes/renderer/DynamicNodeRenderer';
import { nodeRegistry } from '../../../../nodes/registry/NodeRegistry';
import { dynamicNodeFactory } from '../../../../nodes/factory/DynamicNodeFactory';
import { backendServiceManager } from '../../../../plugins/core/BackendServiceManager';

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

// 自定义配置节点组件
const ConfigNodeComponent: React.FC<NodeProps<ConfigNode>> = ({ data, selected, id }) => {
  const [editing, setEditing] = useState(false);
  const [editValue, setEditValue] = useState<any>(data.value);
  const [expanded, setExpanded] = useState(false);

  const getNodeIcon = (dataType: string) => {
    switch (dataType) {
      case 'object':
        return <CodeOutlined style={{ color: '#722ed1' }} />;
      case 'array':
        return <DatabaseOutlined style={{ color: '#13c2c2' }} />;
      case 'string':
        return <ApiOutlined style={{ color: '#1890ff' }} />;
      case 'number':
        return <ClockCircleOutlined style={{ color: '#52c41a' }} />;
      case 'boolean':
        return <Switch disabled size="small" checked={data.value} />;
      case 'group':
      case 'category-group':
        return <FolderOpenOutlined style={{ color: data.color || '#1890ff' }} />;
      case 'category-node':
      case 'bc-node':
      case 'b-service-node':
        return <DatabaseOutlined style={{ color: data.color || '#1890ff' }} />;
      default:
        return <SettingOutlined style={{ color: '#666' }} />;
    }
  };

  const getCategoryIcon = (category: string) => {
    switch (category) {
      case 'ASR':
        return <span style={{ color: '#fa8c16' }}>🎤</span>;
      case 'TTS':
        return <span style={{ color: '#52c41a' }}>🔊</span>;
      case 'LLM':
        return <span style={{ color: '#1890ff' }}>🤖</span>;
      case 'VLLM':
        return <span style={{ color: '#722ed1' }}>👁️</span>;
      case 'server':
        return <span style={{ color: '#13c2c2' }}>🖥️</span>;
      case 'web':
        return <span style={{ color: '#eb2f96' }}>🌐</span>;
      case 'transport':
        return <span style={{ color: '#faad14' }}>📡</span>;
      case 'system':
        return <span style={{ color: '#f5222d' }}>⚙️</span>;
      case 'audio':
        return <span style={{ color: '#a0d911' }}>🎵</span>;
      case 'database':
        return <span style={{ color: '#2f54eb' }}>💾</span>;
      default:
        return <SettingOutlined style={{ color: '#666' }} />;
    }
  };

  const formatValue = (value: any): string => {
    if (typeof value === 'object') {
      return JSON.stringify(value, null, 2);
    }
    return String(value);
  };

  const handleSave = async () => {
    try {
      await configService.updateConfig(data.key, { value: editValue });
      data.value = editValue;
      setEditing(false);
      message.success('配置已更新');
      log.info('配置更新成功', { key: data.key, value: editValue }, 'config', 'ConfigView');
    } catch (error) {
      message.error('配置更新失败');
      log.error('配置更新失败', { key: data.key, error }, 'config', 'ConfigView');
    }
  };

  const handleDelete = async () => {
    try {
      await configService.deleteConfig(data.key);
      message.success('配置已删除');
      log.info('配置删除成功', { key: data.key }, 'config', 'ConfigView');
    } catch (error) {
      message.error('配置删除失败');
      log.error('配置删除失败', { key: data.key, error }, 'config', 'ConfigView');
    }
  };

  const renderValueEditor = () => {
    switch (data.dataType) {
      case 'boolean':
        return (
          <Switch
            checked={editValue}
            onChange={setEditValue}
          />
        );
      case 'number':
        return (
          <InputNumber
            value={editValue}
            onChange={setEditValue}
            style={{ width: '100%' }}
          />
        );
      case 'object':
      case 'array':
        return (
          <TextArea
            value={formatValue(editValue)}
            onChange={(e) => {
              try {
                setEditValue(JSON.parse(e.target.value));
              } catch {
                setEditValue(e.target.value);
              }
            }}
            rows={6}
            placeholder="JSON格式"
          />
        );
      default:
        return (
          <Input
            value={editValue}
            onChange={(e) => setEditValue(e.target.value)}
          />
        );
    }
  };

  const isGroupNode = data.dataType === 'group' || data.dataType === 'category-group';
  const groupConfigs = isGroupNode ? (data.value as any[]) : [];

  return (
    <Card
      size="small"
      className={`config-node ${selected ? 'selected' : ''} ${isGroupNode ? 'group-node' : ''}`}
      style={{
        width: isGroupNode ? 320 : 280,
        minWidth: isGroupNode ? 320 : 280,
        border: selected ? `2px solid ${data.color || '#1890ff'}` : '1px solid #d9d9d9',
        backgroundColor: isGroupNode ? `${data.color}08` : '#ffffff',
        borderRadius: '8px',
        boxShadow: selected ? `0 4px 12px ${data.color}33` : '0 2px 8px rgba(0, 0, 0, 0.1)',
      }}
      title={
        <Space size="small">
          {getCategoryIcon(data.category)}
          {getNodeIcon(data.dataType)}
          <Text strong style={{ fontSize: 12 }}>
            {data.label}
            {isGroupNode && <Badge count={data.configCount} size="small" style={{ marginLeft: 8 }} />}
          </Text>
          <Badge color={data.color || '#666'} />
        </Space>
      }
      extra={
        <Space size="small">
          {isGroupNode && (
            <Tooltip title={expanded ? "收起详情" : "展开详情"}>
              <Button
                type="text"
                size="small"
                icon={expanded ? <EyeInvisibleOutlined /> : <EyeOutlined />}
                onClick={() => setExpanded(!expanded)}
              />
            </Tooltip>
          )}
          {!isGroupNode && data.editable && (
            <Tooltip title="编辑">
              <Button
                type="text"
                size="small"
                icon={editing ? <SaveOutlined /> : <EditOutlined />}
                onClick={editing ? handleSave : () => setEditing(true)}
              />
            </Tooltip>
          )}
          {!isGroupNode && (
            <Tooltip title="删除">
              <Popconfirm
                title="确定要删除这个配置吗？"
                onConfirm={handleDelete}
                okText="确定"
                cancelText="取消"
              >
                <Button
                  type="text"
                  size="small"
                  icon={<DeleteOutlined />}
                  danger
                />
              </Popconfirm>
            </Tooltip>
          )}
        </Space>
      }
    >
      <Handle
        type="target"
        position={Position.Top}
        style={{ background: data.color || '#1890ff', width: 8, height: 8 }}
      />

      {data.description && (
        <Text type="secondary" style={{ fontSize: 10, display: 'block', marginBottom: 8 }}>
          {data.description}
        </Text>
      )}

      {isGroupNode ? (
        <div style={{ marginTop: 8 }}>
          {expanded && groupConfigs.length > 0 ? (
            <div style={{ maxHeight: 200, overflowY: 'auto' }}>
              {groupConfigs.map((config: any, index: number) => (
                <div key={index} style={{
                  marginBottom: 8,
                  padding: 6,
                  backgroundColor: '#f5f5f5',
                  borderRadius: 4,
                  borderLeft: `3px solid ${data.color}`
                }}>
                  <Text strong style={{ fontSize: 10, color: data.color }}>
                    {config.key}
                  </Text>
                  <div style={{ marginTop: 2 }}>
                    <Text code style={{ fontSize: 9, wordBreak: 'break-all' }}>
                      {formatValue(config.value)}
                    </Text>
                  </div>
                  {config.description && (
                    <Text type="secondary" style={{ fontSize: 8, display: 'block', marginTop: 2 }}>
                      {config.description}
                    </Text>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <Text type="secondary" style={{ fontSize: 10, fontStyle: 'italic' }}>
              包含 {groupConfigs.length} 个配置项，点击查看详情
            </Text>
          )}
        </div>
      ) : (
        <div style={{ marginTop: 8 }}>
          {editing ? (
            renderValueEditor()
          ) : (
            <Text code style={{ fontSize: 11, wordBreak: 'break-all' }}>
              {formatValue(data.value)}
            </Text>
          )}
        </div>
      )}

      <div style={{ marginTop: 8 }}>
        <Space size="small" wrap>
          {data.category && (
            <Tag size="small" color={data.color}>{data.category}</Tag>
          )}
          {data.subCategory && (
            <Tag size="small" style={{ backgroundColor: '#f0f0f0', border: '1px solid #d9d9d9' }}>
              {data.subCategory}
            </Tag>
          )}
          {isGroupNode && (
            <Tag size="small" style={{ backgroundColor: '#e6f7ff', border: '1px solid #91d5ff', color: '#1890ff' }}>
              分组
            </Tag>
          )}
        </Space>
      </div>

      <Handle
        type="source"
        position={Position.Bottom}
        style={{ background: data.color || '#1890ff', width: 8, height: 8 }}
      />
    </Card>
  );
};

// 配置画布组件（内部组件，用于ReactFlowProvider）
const ConfigCanvas: React.FC<{
  configs: ConfigRecord[];
  loading: boolean;
  onConfigsLoad: (configs: ConfigRecord[]) => void;
}> = ({ configs, loading, onConfigsLoad }) => {
  // ReactFlow 状态
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);

  // 记忆化节点类型定义
  const nodeTypes = useMemo(() => {
    const dynamicNodeTypes: Record<string, React.ComponentType<NodeProps>> = {};

    // 注册动态节点组件
    const allNodeDefinitions = nodeRegistry.getAllNodeDefinitions();
    allNodeDefinitions.forEach(definition => {
      if (definition.customComponent) {
        dynamicNodeTypes[definition.id] = definition.customComponent;
      } else {
        dynamicNodeTypes[definition.id] = DynamicNodeRenderer;
      }
    });

    return {
      config: ConfigNodeComponent,
      plugin: DynamicNodeRenderer,
      ...dynamicNodeTypes
    };
  }, []);

  // 加载配置数据
  useEffect(() => {
    if (configs.length > 0) {
      // 转换为画布节点
      const newNodes = configService.configsToNodes(configs);
      console.log('ConfigView: Generated nodes:', newNodes);
      console.log('ConfigView: Nodes length:', newNodes?.length || 0);
      setNodes(newNodes);
    }
  }, [configs, setNodes]);

  // 处理连接创建
  const onConnect = useCallback(
    (params: Edge | Connection) => setEdges((eds) => addEdge({ ...params, type: 'smoothstep', animated: true }, eds)),
    [setEdges]
  );

  // 处理拖拽创建节点
  const handleNodeCreate = useCallback((templateData: any) => {
    let newNode: ConfigNode;

    if (templateData.isDynamicNode && templateData.nodeDefinition) {
      // 创建动态节点
      const definition = templateData.nodeDefinition;
      newNode = dynamicNodeFactory.createNode(definition.id, definition, {
        position: {
          x: Math.random() * 400 + 100,
          y: Math.random() * 300 + 100
        },
        data: templateData.defaultValue
      });
    } else if (templateData.isDatabaseNode) {
      // 创建数据库节点
      newNode = {
        id: `config-${Date.now()}`,
        type: 'config',
        position: {
          x: Math.random() * 400 + 100,
          y: Math.random() * 300 + 100
        },
        data: {
          key: templateData.originalNode.key,
          label: templateData.label,
          description: templateData.description,
          category: templateData.category,
          value: templateData.defaultValue || templateData.originalNode.value,
          dataType: templateData.dataType,
          required: templateData.required,
          editable: templateData.editable,
          icon: templateData.icon,
          color: templateData.color
        }
      };
    } else {
      // 创建内置节点（原有逻辑）
      newNode = {
        id: `${templateData.id}-${Date.now()}`,
        type: 'config',
        position: {
          x: Math.random() * 400 + 100,
          y: Math.random() * 300 + 100
        },
        data: {
          key: templateData.id,
          label: templateData.label,
          description: templateData.description,
          category: templateData.category,
          value: templateData.defaultValue || {},
          dataType: templateData.dataType,
          required: false,
          editable: true,
          icon: templateData.icon,
          color: templateData.color
        }
      };
    }

    setNodes((nds) => [...nds, newNode]);
    message.success(`已添加 ${newNode.data.label} 节点`);
    log.info('拖拽创建节点成功', {
      nodeId: newNode.id,
      label: newNode.data.label,
      type: templateData.isDynamicNode ? 'dynamic' : templateData.isDatabaseNode ? 'database' : 'builtin'
    }, 'config', 'ConfigView');
  }, [setNodes]);

  // 处理清空画布
  const handleClearCanvas = useCallback(() => {
    setNodes([]);
    setEdges([]);
    message.success('画布已清空');
  }, [setNodes, setEdges]);

  // 处理保存配置
  const handleSaveConfig = useCallback(() => {
    // 实现保存配置逻辑
    message.success('配置已保存');
    log.info('配置保存成功', { nodesCount: nodes.length }, 'config', 'ConfigView');
  }, [nodes]);

  // 处理加载配置
  const handleLoadConfig = useCallback(async () => {
    try {
      const configData = await configService.getConfigs({});
      onConfigsLoad(configData);
      message.success('配置已加载');
    } catch (error) {
      message.error('加载配置失败');
      log.error('加载配置失败', error, 'config', 'ConfigView');
    }
  }, [onConfigsLoad]);

  if (loading && nodes.length === 0) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-gray-200 rounded-full animate-spin border-t-blue-500 border-r-blue-500 mx-auto mb-4"></div>
          <div className="text-lg text-gray-600">正在加载配置...</div>
        </div>
      </div>
    );
  }

  return (
    <DragHandler onNodeCreate={handleNodeCreate}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        nodeTypes={nodeTypes}
        connectionMode="loose"
        fitView
        attributionPosition="bottom-left"
        style={{ background: '#fafafa' }}
      >
        <Background variant={BackgroundVariant.Dots} gap={20} size={1} />
        <Controls className="bg-white border border-gray-200" />
        <MiniMap
          className="bg-white border border-gray-200"
          nodeColor={(node) => node.data.color || '#1890ff'}
          maskColor="rgba(255, 255, 255, 0.8)"
        />
      </ReactFlow>

      {/* 悬浮组件库面板 */}
      <FloatingComponentLibrary databaseNodes={configs} />
    </DragHandler>
  );
};

// 配置视图主组件
const ConfigView: React.FC<ConfigViewProps> = () => {
  const [configs, setConfigs] = useState<ConfigRecord[]>([]);
  const [loading, setLoading] = useState(true);

  // 加载配置数据
  const loadConfigs = async () => {
    try {
      setLoading(true);
      console.log('ConfigView: Loading configs');
      const configData = await configService.getConfigs({});
      console.log('ConfigView: Retrieved config data:', configData);
      console.log('ConfigView: Config data length:', configData?.length || 0);
      setConfigs(configData);
    } catch (error) {
      console.error('ConfigView: Error loading configs:', error);
      message.error('加载配置失败');
      log.error('加载配置失败', error, 'config', 'ConfigView');
    } finally {
      setLoading(false);
    }
  };

  // 初始化加载
  useEffect(() => {
    loadConfigs();
  }, []);

  return (
    <div className="flex w-full h-full">
      {/* 左侧配置侧边栏 */}
      <ConfigSidebar
        onClearCanvas={() => {
          // 清空画布逻辑将在ConfigCanvas中处理
          console.log('Clear canvas requested');
        }}
        onSaveConfig={() => {
          // 保存配置逻辑将在ConfigCanvas中处理
          console.log('Save config requested');
        }}
        onLoadConfig={loadConfigs}
      />

      {/* 主画布区域 */}
      <main className="flex-1 relative">
        <ReactFlowProvider>
          <ConfigCanvas
            configs={configs}
            loading={loading}
            onConfigsLoad={setConfigs}
          />
        </ReactFlowProvider>
      </main>
    </div>
  );
};

export default ConfigView;