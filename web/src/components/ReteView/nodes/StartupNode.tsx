import React from 'react';
import { NodeProps } from 'rete-react-plugin';
import { ClassicPreset } from 'rete';
import { getIconByNodeType, getStatusColor } from '../../../utils/nodeUtils';

// 节点数据类型
export interface StartupNodeData {
  label: string;
  type: 'database' | 'api' | 'ai' | 'cloud' | 'config';
  status: 'running' | 'stopped' | 'warning';
  description?: string;
  metrics?: Record<string, string | number>;
}

export const StartupNode: React.FC<NodeProps<StartupNodeComponent>> = ({
  data,
  selected,
  id
}) => {
  const nodeData = data as StartupNodeData;

  return (
    <div
      className={`
        startup-node
        startup-node--${nodeData.type}
        ${selected ? 'startup-node--selected' : ''}
        startup-node--status-${nodeData.status}
      `}
      style={{
        borderColor: getStatusColor(nodeData.status),
        background: 'white',
        border: `2px solid ${getStatusColor(nodeData.status)}`,
        borderRadius: '8px',
        padding: '12px',
        minWidth: '180px',
        maxWidth: '240px',
        boxShadow: selected ? '0 4px 12px rgba(0, 0, 0, 0.15)' : '0 2px 8px rgba(0, 0, 0, 0.1)',
        cursor: 'pointer',
        transition: 'all 0.2s ease'
      }}
    >
      {/* 节点头部 */}
      <div className="startup-node__header" style={{ display: 'flex', alignItems: 'center', marginBottom: '8px' }}>
        <div
          className="startup-node__icon"
          style={{
            width: '24px',
            height: '24px',
            marginRight: '8px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: getStatusColor(nodeData.type),
            fontSize: '16px'
          }}
        >
          {getIconByNodeType(nodeData.type)}
        </div>
        <span
          className="startup-node__label"
          style={{
            fontWeight: '600',
            fontSize: '14px',
            color: '#1f2937',
            flex: 1,
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap'
          }}
        >
          {nodeData.label}
        </span>

        {/* 状态指示器 */}
        <div
          className="startup-node__status"
          style={{
            width: '8px',
            height: '8px',
            borderRadius: '50%',
            backgroundColor: getStatusColor(nodeData.status),
            marginLeft: '8px',
            animation: nodeData.status === 'running' ? 'pulse 2s infinite' : 'none'
          }}
        />
      </div>

      {/* 节点描述 */}
      {nodeData.description && (
        <div
          className="startup-node__description"
          style={{
            fontSize: '12px',
            color: '#6b7280',
            marginBottom: '8px',
            lineHeight: '1.4'
          }}
        >
          {nodeData.description}
        </div>
      )}

      {/* 节点指标 */}
      {nodeData.metrics && Object.keys(nodeData.metrics).length > 0 && (
        <div
          className="startup-node__metrics"
          style={{
            fontSize: '11px',
            color: '#9ca3af',
            marginBottom: '8px'
          }}
        >
          {Object.entries(nodeData.metrics).slice(0, 2).map(([key, value]) => (
            <div key={key} style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span>{key}:</span>
              <span style={{ fontWeight: '500' }}>{value}</span>
            </div>
          ))}
        </div>
      )}

      {/* 连接点 */}
      <div className="startup-node__connections" style={{ display: 'flex', justifyContent: 'space-between' }}>
        {/* 输入连接点 */}
        {nodeData.type !== 'database' && (
          <div
            className="startup-node__input"
            style={{
              width: '12px',
              height: '12px',
              borderRadius: '50%',
              backgroundColor: '#3b82f6',
              border: '2px solid white',
              position: 'absolute',
              left: '-6px',
              top: '50%',
              transform: 'translateY(-50%)',
              cursor: 'crosshair'
            }}
            title="输入连接点"
          />
        )}

        {/* 输出连接点 */}
        {nodeData.type !== 'cloud' && (
          <div
            className="startup-node__output"
            style={{
              width: '12px',
              height: '12px',
              borderRadius: '50%',
              backgroundColor: '#3b82f6',
              border: '2px solid white',
              position: 'absolute',
              right: '-6px',
              top: '50%',
              transform: 'translateY(-50%)',
              cursor: 'crosshair'
            }}
            title="输出连接点"
          />
        )}
      </div>

      {/* 添加脉冲动画的样式 */}
      <style jsx>{`
        @keyframes pulse {
          0% {
            opacity: 1;
            transform: scale(1);
          }
          50% {
            opacity: 0.7;
            transform: scale(1.1);
          }
          100% {
            opacity: 1;
            transform: scale(1);
          }
        }
      `}</style>
    </div>
  );
};

// 节点组件类（用于 Rete.js）
export class StartupNodeComponent extends ClassicPreset.Node {
  width = 200;
  height = 120;

  constructor(data: StartupNodeData) {
    super(data.label);
    this.data = data;
  }

  data: StartupNodeData;
}