import React, { ComponentType } from 'react';
import { NodeProps, Handle, Position } from '@xyflow/react';
import { Card, Button, Space } from 'antd';
import { ConfigNode } from '../../plugins/types';

interface DynamicNodeRendererProps extends NodeProps {
  data: ConfigNode['data'];
  selected?: boolean;
}

const DynamicNodeRenderer: ComponentType<DynamicNodeRendererProps> = ({ data, selected }) => {
  return (
    <div>
      <Handle type="target" position={Position.Top} />
      <Card
        size="small"
        style={{
          width: 280,
          borderColor: selected ? '#1890ff' : data.color,
          backgroundColor: '#fff'
        }}
        title={data.label}
      >
        <div style={{ fontSize: '12px', color: '#666' }}>
          {data.description || 'Plugin node'}
        </div>
      </Card>
      <Handle type="source" position={Position.Bottom} />
    </div>
  );
};

export default DynamicNodeRenderer;