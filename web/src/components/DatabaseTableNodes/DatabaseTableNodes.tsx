import React, { useCallback, useMemo, useState, useEffect } from 'react';
import ReactFlow, {
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
import 'reactflow/dist/style.css';
import { TableNode, ColumnNode, DatabaseSchema, RelationshipEdge } from '../../types';
import { Card, Badge, Button, Space, Tooltip, Typography, Tag, Switch, Input, Select } from 'antd';
import { TableOutlined, KeyOutlined, NumberOutlined, FieldStringOutlined } from '@ant-design/icons';

const { Text, Title } = Typography;

// 表节点组件
const TableNodeComponent: React.FC<NodeProps<TableNode>> = ({ data, selected }) => {
  const [expanded, setExpanded] = useState(true);

  const primaryKeys = data.columns.filter(col => col.primaryKey);
  const regularColumns = data.columns.filter(col => !col.primaryKey);

  const getColumnIcon = (type: string) => {
    const lowerType = type.toLowerCase();
    if (lowerType.includes('int') || lowerType.includes('number')) {
      return <NumberOutlined style={{ color: '#1890ff' }} />;
    }
    if (lowerType.includes('text') || lowerType.includes('char') || lowerType.includes('varchar')) {
      return <FieldStringOutlined style={{ color: '#52c41a' }} />;
    }
    return <KeyOutlined style={{ color: '#faad14' }} />;
  };

  const formatDataType = (type: string) => {
    return type.toUpperCase();
  };

  return (
    <Card
      size="small"
      className={`table-node ${selected ? 'selected' : ''}`}
      title={
        <Space>
          <TableOutlined style={{ color: '#1890ff' }} />
          <Text strong>{data.name}</Text>
          <Tag color="blue">{(data.rowCount ?? 0).toLocaleString()} rows</Tag>
        </Space>
      }
      extra={
        <Button
          type="text"
          size="small"
          onClick={() => setExpanded(!expanded)}
        >
          {expanded ? '▼' : '▶'}
        </Button>
      }
      style={{
        width: 280,
        minWidth: 280,
        border: selected ? '2px solid #1890ff' : '1px solid #d9d9d9',
        backgroundColor: '#ffffff',
        borderRadius: '8px',
        boxShadow: selected ? '0 4px 12px rgba(24, 144, 255, 0.15)' : '0 2px 8px rgba(0, 0, 0, 0.1)',
      }}
    >
      <Handle
        type="target"
        position={Position.Top}
        style={{ background: '#1890ff' }}
      />

      {expanded && (
        <div style={{ marginTop: 8 }}>
          {primaryKeys.length > 0 && (
            <div style={{ marginBottom: 8 }}>
              <Text type="secondary" style={{ fontSize: 11, fontWeight: 'bold' }}>PRIMARY KEYS</Text>
              {primaryKeys.map((column: ColumnNode) => (
                <div
                  key={column.id}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    padding: '4px 8px',
                    margin: '2px 0',
                    backgroundColor: '#fff1f0',
                    border: '1px solid #ffccc7',
                    borderRadius: '4px',
                  }}
                >
                  <Space size={4}>
                    <KeyOutlined style={{ color: '#ff4d4f', fontSize: 12 }} />
                    <Text style={{ fontSize: 12, fontWeight: 500 }}>{column.name}</Text>
                  </Space>
                  <Space size={4}>
                    <Tag color="volcano" style={{ fontSize: 10, margin: 0, padding: '0 4px' }}>
                      {formatDataType(column.type)}
                    </Tag>
                    {column.nullable && (
                      <Tag color="default" style={{ fontSize: 10, margin: 0, padding: '0 4px' }}>
                        NULL
                      </Tag>
                    )}
                  </Space>
                </div>
              ))}
            </div>
          )}

          {regularColumns.length > 0 && (
            <div>
              <Text type="secondary" style={{ fontSize: 11, fontWeight: 'bold' }}>COLUMNS</Text>
              {regularColumns.map((column: ColumnNode) => (
                <Tooltip
                  key={column.id}
                  title={
                    <div>
                      <div>Type: {formatDataType(column.type)}</div>
                      <div>Nullable: {column.nullable ? 'Yes' : 'No'}</div>
                      {column.defaultValue && <div>Default: {column.defaultValue}</div>}
                    </div>
                  }
                >
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      padding: '3px 8px',
                      margin: '1px 0',
                      borderRadius: '4px',
                      cursor: 'pointer',
                      transition: 'background-color 0.2s',
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.backgroundColor = '#f5f5f5';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.backgroundColor = 'transparent';
                    }}
                  >
                    <Space size={4}>
                      {getColumnIcon(column.type)}
                      <Text style={{ fontSize: 12 }}>{column.name}</Text>
                    </Space>
                    <Space size={4}>
                      <Text type="secondary" style={{ fontSize: 10 }}>
                        {formatDataType(column.type)}
                      </Text>
                      {column.nullable && (
                        <Text type="secondary" style={{ fontSize: 10, opacity: 0.7 }}>NULL</Text>
                      )}
                    </Space>
                  </div>
                </Tooltip>
              ))}
            </div>
          )}
        </div>
      )}

      <Handle
        type="source"
        position={Position.Bottom}
        style={{ background: '#1890ff' }}
      />
    </Card>
  );
};

// 节点类型定义
const nodeTypes = {
  table: TableNodeComponent,
};

// 组件Props类型
interface DatabaseTableNodesProps {
  schema: DatabaseSchema;
  onTableSelect?: (tableName: string) => void;
}

// 数据库表节点组件
export const DatabaseTableNodes: React.FC<DatabaseTableNodesProps> = ({
  schema,
  onTableSelect,
}) => {

  // 将数据库schema转换为ReactFlow的节点和边
  const { initialNodes, initialEdges } = useMemo(() => {
    const nodes: Node[] = [];
    const edges: Edge[] = [];

    // 创建节点 - 使用层次布局
    const cols = Math.ceil(Math.sqrt((schema.tables || []).length));
    (schema.tables || []).forEach((table, index) => {
      const row = Math.floor(index / cols);
      const col = index % cols;

      nodes.push({
        id: table.name,
        type: 'table',
        position: {
          x: col * 350 + 50,
          y: row * 200 + 50,
        },
        data: table,
      });
    });

    // 创建边 - 基于外键关系
    schema.relationships?.forEach((relationship, index) => {
      const sourceNode = nodes.find(n => n.id === relationship.sourceTable);
      const targetNode = nodes.find(n => n.id === relationship.targetTable);

      if (sourceNode && targetNode) {
        edges.push({
          id: `edge_${relationship.sourceTable}_${relationship.targetTable}_${index}`,
          source: relationship.sourceTable,
          target: relationship.targetTable,
          type: 'smoothstep',
          animated: true,
          style: {
            stroke: '#1890ff',
            strokeWidth: 2,
          },
          label: `${relationship.sourceColumn} → ${relationship.targetColumn}`,
          labelStyle: {
            fontSize: 10,
            fontWeight: 'bold',
            backgroundColor: 'rgba(255, 255, 255, 0.8)',
          },
        });
      }
    });

    return { initialNodes: nodes, initialEdges: edges };
  }, [schema]);

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  const onConnect = useCallback(
    (params: any) => setEdges((eds) => addEdge({ ...params, type: 'smoothstep' }, eds)),
    [setEdges]
  );

  // 更新节点和边当数据变化时
  useEffect(() => {
    setNodes(initialNodes);
    setEdges(initialEdges);
  }, [initialNodes, initialEdges, setNodes, setEdges]);

  // 节点点击处理
  const onNodeClick = useCallback((event: React.MouseEvent, node: Node) => {
    if (onTableSelect) {
      onTableSelect(node.id);
    }
  }, [onTableSelect]);

  return (
    <ReactFlowProvider>
      <div style={{ width: '100%', height: '100vh', position: 'relative' }}>

        {/* ReactFlow */}
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onNodeClick={onNodeClick}
          nodeTypes={nodeTypes}
          connectionLineType={ConnectionLineType.SmoothStep}
          fitView
          defaultViewport={{ x: 0, y: 0, zoom: 0.8 }}
        >
          <Controls />
          <MiniMap
            style={{
              backgroundColor: 'rgba(255, 255, 255, 0.8)',
              border: '1px solid #d9d9d9',
            }}
            nodeColor={(node) => '#1890ff'}
            maskColor="rgba(255, 255, 255, 0.2)"
          />
          <Background variant={BackgroundVariant.Dots} gap={20} size={1} />
        </ReactFlow>
      </div>
    </ReactFlowProvider>
  );
};

export default DatabaseTableNodes;