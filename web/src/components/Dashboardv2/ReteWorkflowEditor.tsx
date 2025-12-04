import React, { useRef, useEffect, useState, useCallback } from 'react';
import { NodeEditor, Node } from 'rete';
import { AreaPlugin, AreaExtensions } from 'rete-area-plugin';
import { ConnectionPlugin, ClassicFlow } from 'rete-connection-plugin';
import { ReactPlugin, ReactRenderExtensions, Presets as ReactPresets } from 'rete-react-plugin';

import { BaseNode, renderNode } from './nodes';
import { DatabaseNode, ApiNode, AiNode, CloudNode, ConfigNode } from './nodes';

interface ReteWorkflowEditorProps {
  onNodesChange?: (nodes: Node[]) => void;
  onConnectionsChange?: (connections: any[]) => void;
  onExecute?: () => void;
}

export const ReteWorkflowEditor: React.FC<ReteWorkflowEditorProps> = ({
  onNodesChange,
  onConnectionsChange,
  onExecute
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const areaRef = useRef<AreaPlugin<any> | null>(null);
  const editorRef = useRef<NodeEditor | null>(null);
  const [isReady, setIsReady] = useState(false);
  const [selectedNodes, setSelectedNodes] = useState<string[]>([]);

  // 初始化编辑器
  useEffect(() => {
    if (!containerRef.current) return;

    const initializeEditor = async () => {
      try {
        // 创建编辑器实例
        const editor = new NodeEditor(containerRef.current);
        editorRef.current = editor;

        // 创建插件实例
        const areaPlugin = new AreaPlugin(containerRef.current);
        const connectionPlugin = new ConnectionPlugin();
        const reactPlugin = new ReactPlugin();

        // 使用插件
        editor.use(areaPlugin);
        editor.use(connectionPlugin);
        editor.use(reactPlugin);

        // 配置区域插件
        AreaExtensions.configure(areaPlugin, {
          snap: true,
          translateExtent: [['-200%', '-200%'], ['200%', '200%']],
          wheelZoom: {
            smooth: true
          }
        });

        // 配置连接插件
        connectionPlugin.addPreset(ClassicFlow.setup, {
          getRotation: () => 0,
          curvature: 0.4
        });

        // 配置 React 渲染插件
        reactPlugin.addPreset(ReactPresets.classic.setup, {
          customize: {
            render: (props) => {
              const { node } = props;
              const data = node.getData();

              // 包装 renderNode 以支持点击和选择功能
              const EnhancedNode = () => {
                const handleClick = (e: React.MouseEvent) => {
                  e.stopPropagation();
                  const nodeId = node.id;
                  setSelectedNodes(prev =>
                    prev.includes(nodeId)
                      ? prev.filter(id => id !== nodeId)
                      : [...prev, nodeId]
                  );
                };

                return (
                  <div
                    onClick={handleClick}
                    style={{
                      outline: selectedNodes.includes(node.id)
                        ? '2px solid #1890ff'
                        : 'none',
                      borderRadius: '8px'
                    }}
                  >
                    {renderNode({ data, emitter: null, node })}
                  </div>
                );
              };

              return <EnhancedNode />;
            }
          }
        });

        
        // 添加示例节点
        await addExampleNodes(editor);

        // 监听节点变化
        editor.addPipe((context) => {
          if (context.type === 'nodecreated' || context.type === 'noderemoved') {
            onNodesChange?.(editor.getNodes());
          }
          return context;
        });

        // 监听连接变化
        editor.addPipe((context) => {
          if (context.type === 'connectioncreated' || context.type === 'connectionremoved') {
            onConnectionsChange?.(editor.getConnections());
          }
          return context;
        });

        // 启动编辑器
        await editor.view.resize();
        AreaExtensions.zoomAt(editor, editor.getNodes());

        areaRef.current = areaPlugin;
        setIsReady(true);

        return editor;
      } catch (error) {
        console.error('Rete.js 编辑器初始化失败:', error);
        return null;
      }
    };

    const editorPromise = initializeEditor();

    return () => {
      editorPromise.then(editor => {
        if (editor) {
          editor.destroy();
        }
      });
    };
  }, []);

  // 添加示例节点
  const addExampleNodes = async (editor: NodeEditor) => {
    try {
      const databaseNode = new DatabaseNode('db-1', 'MySQL Database');
      const apiNode = new ApiNode('api-1', 'REST API');
      const aiNode = new AiNode('ai-1', 'AI Service');
      const cloudNode = new CloudNode('cloud-1', 'Cloud Storage');
      const configNode = new ConfigNode('config-1', 'Workflow Config');

      await editor.addNode(databaseNode);
      await editor.addNode(apiNode);
      await editor.addNode(aiNode);
      await editor.addNode(cloudNode);
      await editor.addNode(configNode);

      // 设置节点位置
      const positions = [
        { x: 100, y: 100 },   // database
        { x: 400, y: 100 },   // api
        { x: 700, y: 100 },   // ai
        { x: 100, y: 300 },   // cloud
        { x: 400, y: 300 }    // config
      ];

      const nodes = [databaseNode, apiNode, aiNode, cloudNode, configNode];
      nodes.forEach((node, index) => {
        if (areaRef.current) {
          areaRef.current.area.translate(node.id, positions[index]);
        }
      });

      // 添加一些示例连接
      await editor.addConnection(
        databaseNode.outputs.get('data')!,
        apiNode.inputs.get('input')!
      );
      await editor.addConnection(
        apiNode.outputs.get('response')!,
        aiNode.inputs.get('prompt')!
      );
      await editor.addConnection(
        configNode.outputs.get('settings')!,
        cloudNode.inputs.get('credentials')!
      );

    } catch (error) {
      console.error('添加示例节点失败:', error);
    }
  };

  // 节点操作函数
  const copyNode = async (node: any) => {
    if (!editorRef.current) return;

    const newNode = { ...node, id: `${node.id}-copy-${Date.now()}` };
    await editorRef.current.addNode(newNode);

    if (areaRef.current) {
      const position = areaRef.current.area.transform(node.id);
      areaRef.current.area.translate(newNode.id, { x: position.x + 50, y: position.y + 50 });
    }
  };

  const deleteNode = async (node: any) => {
    if (!editorRef.current) return;
    await editorRef.current.removeNode(node.id);
  };

  const editNode = (node: any) => {
    // 这里可以实现节点编辑对话框
    console.log('编辑节点:', node);
  };

  // 添加新节点
  const addNode = useCallback(async (nodeType: string) => {
    if (!editorRef.current || !areaRef.current) return;

    let newNode: BaseNode;
    const nodeId = `${nodeType}-${Date.now()}`;

    switch (nodeType) {
      case 'database':
        newNode = new DatabaseNode(nodeId, 'New Database');
        break;
      case 'api':
        newNode = new ApiNode(nodeId, 'New API');
        break;
      case 'ai':
        newNode = new AiNode(nodeId, 'New AI Service');
        break;
      case 'cloud':
        newNode = new CloudNode(nodeId, 'New Cloud Service');
        break;
      case 'config':
        newNode = new ConfigNode(nodeId, 'New Config');
        break;
      default:
        return;
    }

    await editorRef.current.addNode(newNode);

    // 将新节点放置在画布中心
    const centerPosition = { x: 400, y: 200 };
    areaRef.current.area.translate(nodeId, centerPosition);
  }, []);

  // 执行工作流
  const executeWorkflow = useCallback(() => {
    if (!editorRef.current) return;

    const nodes = editorRef.current.getNodes();
    const connections = editorRef.current.getConnections();

    console.log('执行工作流:', { nodes, connections });

    // 模拟执行过程
    nodes.forEach((node, index) => {
      setTimeout(() => {
        const data = node.getData();
        node.setData({ ...data, status: 'running' as const });

        setTimeout(() => {
          node.setData({ ...data, status: 'stopped' as const });
        }, 2000);
      }, index * 500);
    });

    onExecute?.();
  }, [onExecute]);

  // 清空画布
  const clearCanvas = useCallback(async () => {
    if (!editorRef.current) return;

    const nodes = editorRef.current.getNodes();
    for (const node of nodes) {
      await editorRef.current.removeNode(node.id);
    }
    setSelectedNodes([]);
  }, []);

  if (!isReady) {
    return (
      <div className="w-full h-full flex items-center justify-center">
        <div className="text-gray-500">正在加载工作流编辑器...</div>
      </div>
    );
  }

  return (
    <div className="w-full h-full relative">
      {/* 工具栏 */}
      <div className="absolute top-4 left-4 z-10 bg-white rounded-lg shadow-lg p-2">
        <div className="flex flex-col space-y-2">
          <div className="text-xs font-semibold text-gray-600 mb-1">添加节点</div>
          <button
            onClick={() => addNode('database')}
            className="px-3 py-2 text-xs bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
          >
            数据库
          </button>
          <button
            onClick={() => addNode('api')}
            className="px-3 py-2 text-xs bg-green-500 text-white rounded hover:bg-green-600 transition-colors"
          >
            API
          </button>
          <button
            onClick={() => addNode('ai')}
            className="px-3 py-2 text-xs bg-purple-500 text-white rounded hover:bg-purple-600 transition-colors"
          >
            AI
          </button>
          <button
            onClick={() => addNode('cloud')}
            className="px-3 py-2 text-xs bg-orange-500 text-white rounded hover:bg-orange-600 transition-colors"
          >
            云服务
          </button>
          <button
            onClick={() => addNode('config')}
            className="px-3 py-2 text-xs bg-pink-500 text-white rounded hover:bg-pink-600 transition-colors"
          >
            配置
          </button>
        </div>
      </div>

      {/* 控制面板 */}
      <div className="absolute top-4 right-4 z-10 bg-white rounded-lg shadow-lg p-4">
        <div className="flex flex-col space-y-3">
          <div className="text-xs font-semibold text-gray-600">工作流控制</div>
          <button
            onClick={executeWorkflow}
            className="px-4 py-2 text-sm bg-green-500 text-white rounded hover:bg-green-600 transition-colors flex items-center"
          >
            ▶ 执行
          </button>
          <button
            onClick={clearCanvas}
            className="px-4 py-2 text-sm bg-red-500 text-white rounded hover:bg-red-600 transition-colors"
          >
            清空
          </button>
        </div>
      </div>

      {/* 状态信息 */}
      <div className="absolute bottom-4 left-4 z-10 bg-white rounded-lg shadow-lg p-3">
        <div className="text-xs text-gray-600">
          <div>节点数量: {editorRef.current?.getNodes().length || 0}</div>
          <div>连接数量: {editorRef.current?.getConnections().length || 0}</div>
          <div>选中: {selectedNodes.length} 个节点</div>
        </div>
      </div>

      {/* 编辑器容器 */}
      <div
        ref={containerRef}
        className="w-full h-full"
        style={{ width: '100%', height: '100%' }}
      />
    </div>
  );
};