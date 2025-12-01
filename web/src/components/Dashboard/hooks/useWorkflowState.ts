import { useCallback } from 'react';
import { useNodesState, useEdgesState, addEdge, Node, Edge, Connection } from 'reactflow';
import { workflowNodes, workflowEdges } from '../data';

export const useWorkflowState = () => {
  const [nodes, setNodes, onNodesChange] = useNodesState(workflowNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(workflowEdges);

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge({ ...params, type: 'smoothstep' }, eds)),
    [setEdges]
  );

  return {
    nodes,
    edges,
    onNodesChange,
    onEdgesChange,
    onConnect,
  };
};