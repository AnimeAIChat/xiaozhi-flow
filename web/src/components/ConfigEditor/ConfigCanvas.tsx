/**
 * 配置编辑器画布组件
 * 提供UE5风格的配置编辑界面
 */

import React, { useCallback, useEffect, useState, useRef, useMemo } from 'react';
import ReactFlow, {
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
  Panel,
  ReactFlowProvider,
  Position,
  Handle,
  NodeProps,
} from 'reactflow';
import 'reactflow/dist/style.css';
import {
  Card,
  Button,
  Space,
  Input,
  Select,
  Switch,
  Modal,
  Form,
  InputNumber,
  Tooltip,
  Badge,
  Typography,
  Divider,
  Dropdown,
  MenuProps,
  message,
  Drawer,
  List,
  Tag,
  Popconfirm,
} from 'antd';
import {
  SettingOutlined,
  SaveOutlined,
  UndoOutlined,
  RedoOutlined,
  PlusOutlined,
  DeleteOutlined,
  EditOutlined,
  CopyOutlined,
  DownloadOutlined,
  UploadOutlined,
  SearchOutlined,
  FilterOutlined,
  BugOutlined,
  EyeOutlined,
  EyeInvisibleOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
  CameraOutlined,
  HistoryOutlined,
  FolderOpenOutlined,
  CodeOutlined,
  DatabaseOutlined,
  ApiOutlined,
  CloudOutlined,
  SafetyOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import { configService } from '../../services/configService';
import { log } from '../../utils/logger';
import type { ConfigNode, ConfigEdge, ConfigEditMode, ConfigRecord } from '../../types/config';

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

// 自定义配置节点组件
const ConfigNodeComponent: React.FC<NodeProps<ConfigNode>> = ({ data, selected, id }) => {
  const [editing, setEditing] = useState(false);
  const [editValue, setEditValue] = useState<any>(data.value);

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
      log.info('配置更新成功', { key: data.key, value: editValue }, 'config', 'ConfigCanvas');
    } catch (error) {
      message.error('配置更新失败');
      log.error('配置更新失败', { key: data.key, error }, 'config', 'ConfigCanvas');
    }
  };

  const handleDelete = async () => {
    try {
      await configService.deleteConfig(data.key);
      message.success('配置已删除');
      log.info('配置删除成功', { key: data.key }, 'config', 'ConfigCanvas');
    } catch (error) {
      message.error('配置删除失败');
      log.error('配置删除失败', { key: data.key, error }, 'config', 'ConfigCanvas');
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

  return (
    <Card
      size="small"
      className={`config-node ${selected ? 'selected' : ''}`}
      style={{
        width: 280,
        minWidth: 280,
        border: selected ? `2px solid ${data.color || '#1890ff'}` : '1px solid #d9d9d9',
        backgroundColor: '#ffffff',
        borderRadius: '8px',
        boxShadow: selected ? `0 4px 12px ${data.color}33` : '0 2px 8px rgba(0, 0, 0, 0.1)',
      }}
      title={
        <Space size="small">
          {getNodeIcon(data.dataType)}
          <Text strong style={{ fontSize: 12 }}>{data.label}</Text>
          <Badge color={data.color || '#666'} />
        </Space>
      }
      extra={
        <Space size="small">
          {data.editable && (
            <Tooltip title="编辑">
              <Button
                type="text"
                size="small"
                icon={editing ? <SaveOutlined /> : <EditOutlined />}
                onClick={editing ? handleSave : () => setEditing(true)}
              />
            </Tooltip>
          )}
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

      {editing ? (
        <div style={{ marginTop: 8 }}>
          {renderValueEditor()}
        </div>
      ) : (
        <div style={{ marginTop: 8 }}>
          <Text code style={{ fontSize: 11, wordBreak: 'break-all' }}>
            {formatValue(data.value)}
          </Text>
        </div>
      )}

      {data.category && (
        <div style={{ marginTop: 8 }}>
          <Tag size="small" color={data.color}>{data.category}</Tag>
        </div>
      )}

      <Handle
        type="source"
        position={Position.Bottom}
        style={{ background: data.color || '#1890ff', width: 8, height: 8 }}
      />
    </Card>
  );
};

// 主组件 Props
interface ConfigCanvasProps {
  initialMode?: ConfigEditMode;
  onClose?: () => void;
}

// 配置画布组件
export const ConfigCanvas: React.FC<ConfigCanvasProps> = ({ initialMode = 'view', onClose }) => {
  const [mode, setMode] = useState<ConfigEditMode>(initialMode);
  const [configs, setConfigs] = useState<ConfigRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchText, setSearchText] = useState('');
  const [filterCategory, setFilterCategory] = useState('');
  const [showAddModal, setShowAddModal] = useState(false);
  const [showSearchDrawer, setShowSearchDrawer] = useState(false);

  // ReactFlow 状态
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [selectedNode, setSelectedNode] = useState<string | null>(null);

  // ReactFlow 实例引用
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const [reactFlowInstance, setReactFlowInstance] = useState<any>(null);

  // 记忆化节点类型定义以避免 React Flow 警告
  const nodeTypes = useMemo(() => ({
    config: ConfigNodeComponent,
  }), []);

  // 加载配置数据
  useEffect(() => {
    loadConfigs();
  }, [filterCategory, searchText]);

  const loadConfigs = async () => {
    try {
      setLoading(true);
      const filter = {
        category: filterCategory || undefined,
        searchText: searchText || undefined,
      };
      const configData = await configService.getConfigs(filter);
      setConfigs(configData);

      // 转换为画布节点
      const newNodes = configService.configsToNodes(configData);
      setNodes(newNodes);
    } catch (error) {
      message.error('加载配置失败');
      log.error('加载配置失败', error, 'config', 'ConfigCanvas');
    } finally {
      setLoading(false);
    }
  };

  // 处理连接创建
  const onConnect = useCallback(
    (params: Edge | Connection) => setEdges((eds) => addEdge({ ...params, type: 'smoothstep', animated: true }, eds)),
    []
  );

  // 保存画布状态
  const saveCanvasState = async () => {
    try {
      const canvasState = {
        nodes,
        edges,
        viewport: reactFlowInstance?.getViewport() || { x: 0, y: 0, zoom: 1 },
      };
      await configService.saveCanvasState(canvasState);
      message.success('画布状态已保存');
    } catch (error) {
      message.error('保存画布状态失败');
    }
  };

  // 创建新配置
  const handleCreateConfig = async (values: any) => {
    try {
      const newConfig = {
        key: values.key,
        value: values.value,
        description: values.description,
        category: values.category,
        is_active: true,
      };

      await configService.createConfig(newConfig);
      setShowAddModal(false);
      loadConfigs();
      message.success('配置创建成功');
      log.info('新配置创建成功', { key: values.key }, 'config', 'ConfigCanvas');
    } catch (error) {
      message.error('配置创建失败');
      log.error('配置创建失败', error, 'config', 'ConfigCanvas');
    }
  };

  // 工具栏按钮菜单
  const toolbarMenuItems: MenuProps['items'] = [
    {
      key: 'save',
      label: '保存画布',
      icon: <SaveOutlined />,
      onClick: saveCanvasState,
    },
    {
      key: 'export',
      label: '导出配置',
      icon: <DownloadOutlined />,
      onClick: handleExport,
    },
    {
      key: 'import',
      label: '导入配置',
      icon: <UploadOutlined />,
      onClick: handleImport,
    },
    {
      type: 'divider',
    },
    {
      key: 'snapshot',
      label: '创建快照',
      icon: <CameraOutlined />,
      onClick: handleCreateSnapshot,
    },
    {
      key: 'history',
      label: '查看历史',
      icon: <HistoryOutlined />,
      onClick: handleShowHistory,
    },
    {
      type: 'divider',
    },
    {
      key: 'validate',
      label: '验证配置',
      icon: <SafetyOutlined />,
      onClick: handleValidate,
    },
    {
      key: 'debug',
      label: '调试模式',
      icon: <BugOutlined />,
      onClick: () => setMode(mode === 'debug' ? 'view' : 'debug'),
    },
  ];

  const handleExport = async () => {
    try {
      const exportData = await configService.exportConfig({
        category: filterCategory || undefined,
      });

      // 下载文件
      const blob = new Blob([JSON.stringify(exportData, null, 2)], {
        type: 'application/json',
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `config-export-${new Date().toISOString().split('T')[0]}.json`;
      a.click();
      URL.revokeObjectURL(url);

      message.success('配置导出成功');
    } catch (error) {
      message.error('配置导出失败');
    }
  };

  const handleImport = () => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (file) {
        try {
          const text = await file.text();
          const exportData = JSON.parse(text);
          await configService.importConfig(exportData);
          loadConfigs();
          message.success('配置导入成功');
        } catch (error) {
          message.error('配置导入失败');
        }
      }
    };
    input.click();
  };

  const handleCreateSnapshot = async () => {
    const name = prompt('请输入快照名称');
    if (name) {
      try {
        await configService.createSnapshot(name);
        message.success('快照创建成功');
      } catch (error) {
        message.error('快照创建失败');
      }
    }
  };

  const handleShowHistory = () => {
    setShowSearchDrawer(true);
  };

  const handleValidate = async () => {
    try {
      const validation = await configService.validateConfigs(configs);
      if (validation.isValid) {
        message.success('配置验证通过');
      } else {
        message.error(`发现 ${validation.errors.length} 个配置错误`);
      }
    } catch (error) {
      message.error('配置验证失败');
    }
  };

  return (
    <div className="w-full h-full bg-gray-50 flex flex-col">
      {/* 顶部工具栏 */}
      <div className="bg-white border-b border-gray-200 p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Title level={4} style={{ margin: 0 }}>
              <DatabaseOutlined /> 配置编辑器
            </Title>
            <Badge count={configs.length} showZero>
              <Text type="secondary">配置项</Text>
            </Badge>
          </div>

          <div className="flex items-center space-x-2">
            {/* 搜索和过滤 */}
            <Space>
              <Input
                placeholder="搜索配置..."
                prefix={<SearchOutlined />}
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                style={{ width: 200 }}
                allowClear
              />
              <Select
                placeholder="分类"
                value={filterCategory}
                onChange={setFilterCategory}
                style={{ width: 120 }}
                allowClear
              >
                <Select.Option value="system">系统配置</Select.Option>
                <Select.Option value="user">用户配置</Select.Option>
                <Select.Option value="device">设备配置</Select.Option>
                <Select.Option value="network">网络配置</Select.Option>
              </Select>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => setShowAddModal(true)}
              >
                新建配置
              </Button>
              <Dropdown menu={{ items: toolbarMenuItems }} trigger={['click']}>
                <Button icon={<SettingOutlined />} />
              </Dropdown>
              {onClose && (
                <Button onClick={onClose}>关闭</Button>
              )}
            </Space>
          </div>
        </div>

        {/* 模式切换 */}
        <div className="flex items-center space-x-4 mt-4">
          <Space>
            <Text type="secondary">编辑模式:</Text>
            <Button.Group>
              <Button
                type={mode === 'view' ? 'primary' : 'default'}
                icon={<EyeOutlined />}
                onClick={() => setMode('view')}
                size="small"
              >
                查看
              </Button>
              <Button
                type={mode === 'edit' ? 'primary' : 'default'}
                icon={<EditOutlined />}
                onClick={() => setMode('edit')}
                size="small"
              >
                编辑
              </Button>
              <Button
                type={mode === 'connect' ? 'primary' : 'default'}
                icon={<ApiOutlined />}
                onClick={() => setMode('connect')}
                size="small"
              >
                连接
              </Button>
              <Button
                type={mode === 'debug' ? 'primary' : 'default'}
                icon={<BugOutlined />}
                onClick={() => setMode('debug')}
                size="small"
              >
                调试
              </Button>
            </Button.Group>
          </Space>
        </div>
      </div>

      {/* 画布区域 */}
      <div className="flex-1 relative" ref={reactFlowWrapper}>
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

          {/* 面板 */}
          <Panel position="top-left" className="bg-white rounded-lg shadow-sm p-2">
            <Space direction="vertical" size="small">
              <Text strong>配置画布</Text>
              <Text type="secondary" style={{ fontSize: 12 }}>
                节点: {nodes.length} | 连接: {edges.length}
              </Text>
              <Text type="secondary" style={{ fontSize: 12 }}>
                模式: {mode}
              </Text>
            </Space>
          </Panel>
        </ReactFlow>
      </div>

      {/* 新建配置弹窗 */}
      <Modal
        title="新建配置"
        open={showAddModal}
        onCancel={() => setShowAddModal(false)}
        footer={null}
        width={600}
      >
        <Form onFinish={handleCreateConfig} layout="vertical">
          <Form.Item name="key" label="配置键" rules={[{ required: true, message: '请输入配置键' }]}>
            <Input placeholder="例如: app.version" />
          </Form.Item>
          <Form.Item name="value" label="配置值" rules={[{ required: true, message: '请输入配置值' }]}>
            <TextArea rows={4} placeholder="JSON格式或字符串" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input placeholder="配置描述" />
          </Form.Item>
          <Form.Item name="category" label="分类" rules={[{ required: true, message: '请选择分类' }]}>
            <Select placeholder="选择分类">
              <Select.Option value="system">系统配置</Select.Option>
              <Select.Option value="user">用户配置</Select.Option>
              <Select.Option value="device">设备配置</Select.Option>
              <Select.Option value="network">网络配置</Select.Option>
              <Select.Option value="media">媒体配置</Select.Option>
              <Select.Option value="security">安全配置</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">创建</Button>
              <Button onClick={() => setShowAddModal(false)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 搜索抽屉 */}
      <Drawer
        title="配置搜索"
        placement="right"
        onClose={() => setShowSearchDrawer(false)}
        open={showSearchDrawer}
        width={400}
      >
        <List
          dataSource={configs}
          renderItem={(config) => (
            <List.Item
              key={config.id}
              actions={[
                <Button
                  type="link"
                  icon={<EditOutlined />}
                  onClick={() => {
                    // 聚焦到对应节点
                    const node = nodes.find(n => n.id === config.id.toString());
                    if (node && reactFlowInstance) {
                      reactFlowInstance.fitView({ nodes: [node], duration: 800 });
                    }
                  }}
                >
                  定位
                </Button>
              ]}
            >
              <List.Item.Meta
                title={config.key}
                description={config.description}
              />
            </List.Item>
          )}
        />
      </Drawer>
    </div>
  );
};

// 包装器组件
const ConfigCanvasWrapper: React.FC<ConfigCanvasProps> = (props) => (
  <ReactFlowProvider>
    <ConfigCanvas {...props} />
  </ReactFlowProvider>
);

export default ConfigCanvasWrapper;