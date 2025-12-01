import React from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { Tag } from 'antd';
import {
  DatabaseOutlined,
  ApiOutlined,
  RobotOutlined,
  CloudOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { WorkflowNodeProps } from '../../types';

const WorkflowNode: React.FC<NodeProps<WorkflowNodeProps['data']>> = ({ data }) => {
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
      {/* 输入Handle */}
      <Handle
        type="target"
        position={Position.Top}
        style={{ background: '#1890ff', width: 8, height: 8 }}
      />

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

      {/* 输出Handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        style={{ background: '#1890ff', width: 8, height: 8 }}
      />
    </div>
  );
};

export default WorkflowNode;