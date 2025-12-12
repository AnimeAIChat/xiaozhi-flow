import React from 'react';
import { ClassicPreset } from 'rete';

export interface NodeData {
  label: string;
  type: 'database' | 'api' | 'ai' | 'cloud' | 'config';
  status: 'running' | 'stopped' | 'warning';
  description?: string;
  metrics?: Record<string, string | number>;
}

export class BaseNode<
  T extends NodeData = NodeData,
> extends ClassicPreset.Node {
  width = 200;
  height = 120;

  constructor(id: string, label: string, type: T['type'], data?: Partial<T>) {
    super(id);

    const nodeData: T = {
      label,
      type,
      status: 'stopped',
      ...data,
    } as T;

    this.setData(nodeData);
    this.setupInputsAndOutputs(type);
  }

  private setupInputsAndOutputs(type: T['type']) {
    switch (type) {
      case 'database':
        this.addInput('config', new ClassicPreset.Input());
        this.addOutput('data', new ClassicPreset.Output());
        break;
      case 'api':
        this.addInput('input', new ClassicPreset.Input());
        this.addInput('config', new ClassicPreset.Input());
        this.addOutput('response', new ClassicPreset.Output());
        break;
      case 'ai':
        this.addInput('prompt', new ClassicPreset.Input());
        this.addInput('context', new ClassicPreset.Input());
        this.addOutput('result', new ClassicPreset.Output());
        break;
      case 'cloud':
        this.addInput('data', new ClassicPreset.Input());
        this.addInput('credentials', new ClassicPreset.Input());
        this.addOutput('output', new ClassicPreset.Output());
        break;
      case 'config':
        this.addOutput('settings', new ClassicPreset.Output());
        break;
    }
  }
}

// React 组件渲染函数
export const renderNode = (props: {
  data: NodeData;
  emitter: any;
  node: any;
}) => {
  const { data, emitter, node } = props;

  const getStatusColor = (status: NodeData['status']) => {
    switch (status) {
      case 'running':
        return '#52c41a';
      case 'warning':
        return '#faad14';
      case 'stopped':
        return '#ff4d4f';
      default:
        return '#d9d9d9';
    }
  };

  const getTypeIcon = (type: NodeData['type']) => {
    switch (type) {
      case 'database':
        return '🗄️';
      case 'api':
        return '🔌';
      case 'ai':
        return '🤖';
      case 'cloud':
        return '☁️';
      case 'config':
        return '⚙️';
      default:
        return '📦';
    }
  };

  return (
    <div
      className="bg-white rounded-lg shadow-lg border-2 border-gray-200 hover:border-blue-400 transition-colors duration-200"
      style={{
        width: '200px',
        minHeight: '120px',
        borderLeft: `4px solid ${getStatusColor(data.status)}`,
      }}
    >
      {/* 节点头部 */}
      <div className="flex items-center justify-between p-3 border-b border-gray-100">
        <div className="flex items-center space-x-2">
          <span className="text-lg">{getTypeIcon(data.type)}</span>
          <span className="font-semibold text-gray-800 text-sm">
            {data.label}
          </span>
        </div>
        <div
          className="w-3 h-3 rounded-full"
          style={{ backgroundColor: getStatusColor(data.status) }}
          title={`状态: ${data.status}`}
        />
      </div>

      {/* 节点内容 */}
      <div className="p-3">
        {data.description && (
          <p className="text-xs text-gray-600 mb-2">{data.description}</p>
        )}

        {data.metrics && Object.keys(data.metrics).length > 0 && (
          <div className="text-xs text-gray-500">
            {Object.entries(data.metrics).map(([key, value]) => (
              <div key={key} className="flex justify-between">
                <span>{key}:</span>
                <span className="font-mono">{String(value)}</span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 连接点提示 */}
      <div className="px-3 pb-2 text-xs text-gray-400">
        {node.inputs.size > 0 && <div>• 输入端口</div>}
        {node.outputs.size > 0 && <div>• 输出端口</div>}
      </div>
    </div>
  );
};
