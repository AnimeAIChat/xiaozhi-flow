import React, { useCallback, useMemo } from 'react';
import {
  ReactFlow,
  MiniMap,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  addEdge,
  Node,
  Edge,
  Connection,
  ConnectionMode,
  Panel,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { FullscreenLayout } from '../../components/layout';
import { Card, Typography, Space, Button, Tag } from 'antd';
import {
  DatabaseOutlined,
  ApiOutlined,
  RobotOutlined,
  CloudOutlined,
  SettingOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
} from '@ant-design/icons';

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

// 节点类型定义
const nodeTypes = {
  custom: CustomNode,
};

const Dashboard: React.FC = () => {
  // 初始节点数据 - 工作流节点
  const initialNodes: Node[] = [
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

  // 初始边数据 - 工作流连接
  const initialEdges: Edge[] = [
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

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  );

  const handleStartAll = () => {
    setNodes((nodes) =>
      nodes.map((node) => ({
        ...node,
        data: { ...node.data, status: 'running' },
      }))
    );
  };

  const handleStopAll = () => {
    setNodes((nodes) =>
      nodes.map((node) => ({
        ...node,
        data: { ...node.data, status: 'stopped' },
      }))
    );
  };

  const handleRefresh = () => {
    // 模拟刷新节点状态
    setNodes((nodes) =>
      nodes.map((node) => {
        const randomStatus = Math.random() > 0.3 ? 'running' : Math.random() > 0.5 ? 'warning' : 'stopped';
        return {
          ...node,
          data: { ...node.data, status: randomStatus },
        };
      })
    );
  };

  const runningCount = useMemo(() =>
    nodes.filter(node => node.data.status === 'running').length, [nodes]
  );

  const totalCount = nodes.length;
  const warningCount = nodes.filter(node => node.data.status === 'warning').length;
  const stoppedCount = nodes.filter(node => node.data.status === 'stopped').length;

  return (
    <FullscreenLayout>
      {/* React Flow 画布容器 - 占满整个屏幕 */}
      <div className="w-full h-full bg-white overflow-hidden">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            nodeTypes={nodeTypes}
            connectionMode={ConnectionMode.Loose}
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
            <Panel position="top-left" className="bg-white p-3 rounded-lg shadow-sm border border-gray-200">
              <div className="text-gray-700 text-sm space-y-1">
                <div>🎯 拖拽节点重新排列</div>
                <div>🔗 点击边缘创建连接</div>
                <div>🔍 滚轮缩放画布</div>
              </div>
            </Panel>
          </ReactFlow>
      </div>
    </FullscreenLayout>
  );
};

export default Dashboard;