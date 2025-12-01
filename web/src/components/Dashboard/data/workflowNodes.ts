import { Node } from 'reactflow';
import { WorkflowNodeData } from '../types';

export const workflowNodes: Node<WorkflowNodeData>[] = [
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