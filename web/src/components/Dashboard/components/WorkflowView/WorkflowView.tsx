import React from 'react';
import {
  ReactFlow,
  Controls,
  MiniMap,
  Background,
  BackgroundVariant,
  ReactFlowProvider,
} from '@xyflow/react';
import { WorkflowViewProps } from '../../types';
import { workflowNodeTypes } from './workflowNodeTypes';
import { log } from '../../../../utils/logger';

const WorkflowCanvas: React.FC<WorkflowViewProps> = ({
  nodes,
  edges,
  onNodesChange,
  onEdgesChange,
  onConnect,
  onDoubleClick,
}) => {
  return (
    <div
      className="w-full h-full cursor-pointer"
      onDoubleClick={onDoubleClick}
      title="双击进入配置编辑器"
    >
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        nodeTypes={workflowNodeTypes}
        fitView
        style={{ width: '100%', height: '100%', cursor: 'pointer' }}
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
    </div>
  );
};

// 包装器组件 - 负责提供ReactFlow上下文
const WorkflowView: React.FC<WorkflowViewProps> = (props) => (
  <ReactFlowProvider>
    <WorkflowCanvas {...props} />
  </ReactFlowProvider>
);

export default WorkflowView;