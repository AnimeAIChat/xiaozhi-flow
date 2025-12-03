import React, { useRef, useEffect, useState } from 'react';
import { ClassicPreset } from 'rete';
import { Editor } from 'rete';
import { ReactRenderPlugin, Presets as ReactPresets } from '@retejs/react-renderer';
import { ConnectionPlugin, Presets as ConnectionPresets } from '@retejs/connection-plugin';

import { WorkflowViewProps } from '../Dashboard/types';
import { convertReactFlowToRete } from '../../utils/reteDataConverter';
import { log } from '../../utils/logger';

interface ReteNodeData {
  label: string;
  type: 'database' | 'api' | 'ai' | 'cloud' | 'config';
  status: 'running' | 'stopped' | 'warning';
  description?: string;
  metrics?: Record<string, string | number>;
}

class ReteNode extends ClassicPreset.Node {
  width = 180;
  height = 100;

  constructor(data: ReteNodeData) {
    super(data.label);
    this.data = data;
  }

  data: ReteNodeData;
}

const ReteWorkflowView: React.FC<WorkflowViewProps> = ({
  nodes,
  edges,
  onNodesChange,
  onEdgesChange,
  onConnect,
  onDoubleClick,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [isReady, setIsReady] = useState(false);
  const editorRef = useRef<Editor | null>(null);

  useEffect(() => {
    if (!containerRef.current) return;

    const initializeEditor = async () => {
      try {
        // 创建编辑器实例
        const editor = new Editor(containerRef.current);

        // 添加插件
        editor.use(ReactRenderPlugin, {
          element: containerRef.current
        });

        editor.use(ConnectionPlugin, {
          connector: ConnectionPresets.curve
        });

        // 配置节点类型
        editor.register(ReteNode);

        editorRef.current = editor;
        setIsReady(true);

        log.info('Rete.js 编辑器初始化成功', null, 'system', 'ReteWorkflowView');
      } catch (error) {
        log.error('Rete.js 编辑器初始化失败', { error }, 'system', 'ReteWorkflowView');
      }
    };

    initializeEditor();

    return () => {
      if (editorRef.current) {
        editorRef.current.destroy();
      }
    };
  }, []);

  useEffect(() => {
    if (!editorRef.current || !isReady) return;

    const updateEditorData = async () => {
      try {
        const editor = editorRef.current!;

        // 清除现有节点
        editor.getNodes().forEach(node => {
          editor.removeNode(node.id);
        });

        // 转换并添加节点
        const { nodes: reteNodes, connections } = convertReactFlowToRete(nodes, edges);

        for (const nodeData of reteNodes) {
          const node = new ReteNode(nodeData.data);
          node.id = nodeData.id;
          node.position.x = nodeData.x;
          node.position.y = nodeData.y;

          editor.addNode(node);
        }

        // 添加连接
        for (const connection of connections) {
          try {
            await editor.addConnection(connection);
          } catch (error) {
            log.warn('连接创建失败', { connection, error }, 'ui', 'ReteWorkflowView');
          }
        }

        log.info('编辑器数据更新完成', {
          nodeCount: reteNodes.length,
          connectionCount: connections.length
        }, 'ui', 'ReteWorkflowView');
      } catch (error) {
        log.error('编辑器数据更新失败', { error }, 'system', 'ReteWorkflowView');
      }
    };

    updateEditorData();
  }, [nodes, edges, isReady]);

  const handleDoubleClick = useCallback((event: React.MouseEvent) => {
    if (event.target === containerRef.current) {
      onDoubleClick?.();
    }
  }, [onDoubleClick]);

  return (
    <div
      className="w-full h-full cursor-pointer relative"
      onDoubleClick={handleDoubleClick}
      title="双击进入配置编辑器"
    >
      <div
        ref={containerRef}
        className="w-full h-full"
        style={{
          background: '#f9fafb',
          backgroundImage: `
            radial-gradient(circle, #e5e7eb 1px, transparent 1px)
          `,
          backgroundSize: '20px 20px'
        }}
      />

      {!isReady && (
        <div className="absolute inset-0 flex items-center justify-center bg-white bg-opacity-75">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-2"></div>
            <p className="text-gray-600">正在初始化编辑器...</p>
          </div>
        </div>
      )}
    </div>
  );
};

export default ReteWorkflowView;