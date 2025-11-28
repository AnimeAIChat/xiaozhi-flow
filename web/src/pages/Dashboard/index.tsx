import React, { useCallback, useMemo, useState, useEffect } from 'react';
import {
  ReactFlow,
  Node,
  Edge,
  addEdge,
  ConnectionLineType,
  Panel,
  useNodesState,
  useEdgesState,
  Controls,
  MiniMap,
  Background,
  BackgroundVariant,
  Handle,
  Position,
  NodeProps,
  ReactFlowProvider,
} from 'reactflow';
import '@xyflow/react/dist/style.css';
import { FullscreenLayout } from '../../components/layout';
import { Card, Typography, Space, Button, Tag, Switch } from 'antd';
import { useNavigate } from 'react-router-dom';
import {
  DatabaseOutlined,
  ApiOutlined,
  RobotOutlined,
  CloudOutlined,
  SettingOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
  TableOutlined,
} from '@ant-design/icons';
import { DatabaseTableNodes } from '../../components/DatabaseTableNodes';
import { apiService } from '../../services/api';
import { log } from '../../utils/logger';

const { Title } = Typography;

// 自定义节点类型 - 白色主题
const CustomNode = ({ data }: { data: any }) => {
  const getNodeIcon = (type: string) => {
    switch (type) {
      case 'database':
        return <DatabaseOutlined className="text-purple-500" />;
      case 'api':
        return <ApiOutlined className="text-blue-500" />;
      case 'ai':
        return <RobotOutlined className="text-green-500" />;
      case 'cloud':
        return <CloudOutlined className="text-cyan-500" />;
      case 'config':
        return <SettingOutlined className="text-orange-500" />;
      default:
        return <SettingOutlined className="text-gray-500" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return 'green';
      case 'stopped':
        return 'red';
      case 'warning':
        return 'orange';
      default:
        return 'default';
    }
  };

  return (
    <div className="px-4 py-3 shadow-sm rounded-lg bg-white border border-gray-200 hover:border-blue-400 hover:shadow-md transition-all">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <div className="text-lg">{getNodeIcon(data.type)}</div>
          <div>
            <div className="font-semibold text-gray-900">{data.label}</div>
            {data.description && (
              <div className="text-xs text-gray-500 mt-1">{data.description}</div>
            )}
          </div>
        </div>
        <Tag color={getStatusColor(data.status)}>
          {data.status}
        </Tag>
      </div>
      {data.metrics && (
        <div className="mt-3 pt-3 border-t border-gray-100 grid grid-cols-3 gap-2">
          {Object.entries(data.metrics).slice(0, 3).map(([key, value]) => (
            <div key={key} className="text-center">
              <div className="text-xs text-gray-500">{key}</div>
              <div className="text-sm font-medium text-gray-900">{String(value)}</div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};


// Dashboard 组件 - 支持切换显示数据库表结构或工作流节点
const Dashboard: React.FC = () => {
  const [schema, setSchema] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'database' | 'workflow'>('workflow'); // 默认显示工作流节点

  // 工作流节点示例数据 - 保留以备将来使用
  const workflowNodes: Node[] = [
    {
      id: '1',
      type: 'custom',
      position: { x: 250, y: 50 },
      data: {
        label: 'Web Server',
        type: 'api',
        status: 'running',
        description: 'Nginx Web服务器',
        metrics: {
          'CPU': '15%',
          'Memory': '256MB',
          'Connections': '124',
        },
      },
    },
    {
      id: '2',
      type: 'custom',
      position: { x: 50, y: 200 },
      data: {
        label: 'PostgreSQL',
        type: 'database',
        status: 'running',
        description: '主数据库',
        metrics: {
          'Connections': '45',
          'Size': '2.5GB',
          'Query/s': '1.2K',
        },
      },
    },
    {
      id: '3',
      type: 'custom',
      position: { x: 450, y: 200 },
      data: {
        label: 'Redis Cache',
        type: 'database',
        status: 'running',
        description: '缓存服务器',
        metrics: {
          'Memory': '512MB',
          'Hit Rate': '94%',
          'Keys': '12.5K',
        },
      },
    },
    {
      id: '4',
      type: 'custom',
      position: { x: 250, y: 350 },
      data: {
        label: 'AI Service',
        type: 'ai',
        status: 'running',
        description: 'OpenAI GPT-4',
        metrics: {
          'Tokens/s': '850',
          'Queue': '23',
          'Latency': '125ms',
        },
      },
    },
    {
      id: '5',
      type: 'custom',
      position: { x: 650, y: 125 },
      data: {
        label: 'Object Storage',
        type: 'cloud',
        status: 'running',
        description: 'AWS S3存储',
        metrics: {
          'Storage': '125GB',
          'Objects': '45K',
          'Bandwidth': '2.1MB/s',
        },
      },
    },
    {
      id: '6',
      type: 'custom',
      position: { x: 450, y: 350 },
      data: {
        label: 'Config Service',
        type: 'config',
        status: 'warning',
        description: '配置管理服务',
        metrics: {
          'Configs': '156',
          'Version': 'v1.2.3',
          'Sync': '延迟',
        },
      },
    },
    {
      id: '7',
      type: 'custom',
      position: { x: 50, y: 500 },
      data: {
        label: '监控系统',
        type: 'config',
        status: 'stopped',
        description: 'Prometheus监控',
        metrics: {
          'Metrics': '2.3K',
          'Targets': '8/12',
          'Status': '部分离线',
        },
      },
    },
  ];

  // 工作流边示例数据 - 保留以备将来使用
  const workflowEdges: Edge[] = [
    {
      id: 'e1-2',
      source: '1',
      target: '2',
      label: '数据查询',
      type: 'smoothstep',
      animated: true,
      style: { stroke: '#1890ff' },
    },
    {
      id: 'e1-3',
      source: '1',
      target: '3',
      label: '缓存读写',
      type: 'smoothstep',
      animated: true,
      style: { stroke: '#52c41a' },
    },
    {
      id: 'e1-4',
      source: '1',
      target: '4',
      label: 'AI请求',
      type: 'smoothstep',
      animated: true,
      style: { stroke: '#722ed1' },
    },
    {
      id: 'e1-5',
      source: '1',
      target: '5',
      label: '文件存储',
      type: 'smoothstep',
      animated: true,
      style: { stroke: '#13c2c2' },
    },
    {
      id: 'e2-4',
      source: '2',
      target: '4',
      label: '训练数据',
      type: 'smoothstep',
      style: { stroke: '#fa8c16' },
    },
    {
      id: 'e3-4',
      source: '3',
      target: '4',
      label: '模型缓存',
      type: 'smoothstep',
      style: { stroke: '#eb2f96' },
    },
    {
      id: 'e4-6',
      source: '4',
      target: '6',
      label: '配置同步',
      type: 'smoothstep',
      style: { stroke: '#faad14' },
    },
    {
      id: 'e6-7',
      source: '6',
      target: '7',
      label: '监控配置',
      type: 'smoothstep',
      style: { stroke: '#f5222d' },
    },
  ];

  useEffect(() => {
    const loadDatabaseSchema = async () => {
      try {
        setLoading(true);
        const data = await apiService.getDatabaseSchema();

        // 转换后端数据格式到前端格式
        const transformedSchema = {
          name: data.name,
          type: data.type,
          tables: data.tables.map((table: any) => ({
            id: table.name,
            name: table.name,
            type: 'table' as const,
            schema: data.name,
            rowCount: table.rowCount,
            size: table.size,
            columns: table.columns.map((col: any) => ({
              id: `${table.name}.${col.name}`,
              name: col.name,
              type: col.type,
              nullable: col.nullable,
              primaryKey: col.primaryKey,
              unique: col.unique,
              defaultValue: col.defaultValue,
              description: col.description,
              position: { x: 0, y: 0 },
            })),
            indexes: table.indexes?.map((idx: any) => ({
              id: `${table.name}.${idx.name}`,
              name: idx.name,
              columns: idx.columns,
              unique: idx.unique,
              type: idx.type,
            })) || [],
            foreignKeys: [],
            position: { x: 0, y: 0 },
          })),
          relationships: data.relationships?.map((rel: any) => ({
            id: rel.name || `${rel.sourceTable}_${rel.targetTable}`,
            source: rel.sourceTable,
            target: rel.targetTable,
            type: 'foreign_key' as const,
            label: `${rel.sourceColumn} → ${rel.targetColumn}`,
            style: {
              color: '#1890ff',
              width: 2,
              style: 'solid' as const,
              arrowType: 'arrow' as const,
            },
          })) || [],
        };

        setSchema(transformedSchema);
        setError(null);
      } catch (err) {
        log.error('数据库表结构加载失败', err, 'database', 'Dashboard', err instanceof Error ? err.stack : undefined);
        setError(err instanceof Error ? err.message : '加载数据库表结构失败');
      } finally {
        setLoading(false);
      }
    };

    loadDatabaseSchema();
  }, []);

  const handleTableSelect = (tableName: string) => {
    log.info(`用户选择数据库表: ${tableName}`, { tableName }, 'database', 'Dashboard');
    // 可以在这里添加表详情处理逻辑
  };

  // 工作流视图的状态 - 必须在所有条件渲染之前
  const [workflowNodesState, setWorkflowNodesState, onWorkflowNodesChange] = useNodesState(workflowNodes);
  const [workflowEdgesState, setWorkflowEdgesState, onWorkflowEdgesChange] = useEdgesState(workflowEdges);

  const onWorkflowConnect = useCallback(
    (params: any) => setWorkflowEdgesState((eds) => addEdge({ ...params, type: 'smoothstep' }, eds)),
    [setWorkflowEdgesState]
  );

  if (loading) {
    return (
      <FullscreenLayout>
        <div className="flex items-center justify-center min-h-screen bg-white">
          <div className="text-center">
            <div className="w-12 h-12 border-4 border-gray-200 rounded-full animate-spin border-t-blue-500 border-r-blue-500 mx-auto mb-4"></div>
            <div className="text-lg text-gray-600">加载数据库表结构...</div>
          </div>
        </div>
      </FullscreenLayout>
    );
  }

  if (error || !schema) {
    return (
      <FullscreenLayout>
        <div className="flex items-center justify-center min-h-screen bg-white">
          <div className="text-center p-8 bg-red-50 rounded-lg border border-red-200">
            <DatabaseOutlined className="text-4xl text-red-500 mb-4" />
            <div className="text-lg text-red-600 mb-2">数据库表结构加载失败</div>
            <div className="text-sm text-red-500">{error || '未知错误'}</div>
          </div>
        </div>
      </FullscreenLayout>
    );
  }

  return (
    <FullscreenLayout>
      <div className="w-full h-full bg-gray-50 overflow-hidden relative">
        {/* 视图切换按钮 - 移到右上角 */}
        <div className="absolute top-4 right-4 z-10 bg-white rounded-lg shadow-sm border border-gray-200 p-2">
          <Space>
            <Button
              type={viewMode === 'workflow' ? 'primary' : 'default'}
              size="small"
              icon={<ApiOutlined />}
              onClick={() => {
                log.info('用户切换到工作流节点视图', { from: viewMode }, 'ui', 'Dashboard');
                setViewMode('workflow');
              }}
            >
              工作流节点
            </Button>
            <Button
              type={viewMode === 'database' ? 'primary' : 'default'}
              size="small"
              icon={<DatabaseOutlined />}
              onClick={() => {
                log.info('用户切换到数据库表视图', { from: viewMode }, 'ui', 'Dashboard');
                setViewMode('database');
              }}
            >
              数据库表
            </Button>
          </Space>
        </div>

        {/* 内容区域 */}
        {viewMode === 'database' ? (
          <DatabaseTableNodes
            schema={schema}
            onTableSelect={handleTableSelect}
          />
        ) : (
          <ReactFlowProvider>
            <ReactFlow
              nodes={workflowNodesState}
              edges={workflowEdgesState}
              onNodesChange={onWorkflowNodesChange}
              onEdgesChange={onWorkflowEdgesChange}
              onConnect={onWorkflowConnect}
              nodeTypes={{ custom: CustomNode }}
              connectionMode="loose"
              fitView
              style={{ width: '100%', height: '100%' }}
              className="bg-gray-50"
            >
              <Background color="#e5e7eb" gap={20} />
              <Controls
                className="bg-white border border-gray-200 shadow-sm"
                showInteractive={false}
              />
              <MiniMap
                className="bg-white border border-gray-200 shadow-sm"
                nodeColor={(node) => {
                  switch (node.data?.status) {
                    case 'running':
                      return '#52c41a';
                    case 'warning':
                      return '#faad14';
                    case 'stopped':
                      return '#ff4d4f';
                    default:
                      return '#d9d9d9';
                  }
                }}
                maskColor="rgba(255, 255, 255, 0.8)"
              />
            </ReactFlow>
          </ReactFlowProvider>
        )}
      </div>
    </FullscreenLayout>
  );
};

export default Dashboard;