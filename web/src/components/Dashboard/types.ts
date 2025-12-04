import { Node, Edge, OnNodesChange, OnEdgesChange, Connection } from '@xyflow/react';

export type DashboardViewMode = 'database' | 'workflow' | 'config';

export interface WorkflowNodeData {
  label: string;
  type: 'database' | 'api' | 'ai' | 'cloud' | 'config';
  status: 'running' | 'stopped' | 'warning';
  description?: string;
  metrics?: Record<string, string | number>;
}

export interface DatabaseSchema {
  name: string;
  type: string;
  tables: any[];
  relationships: any[];
}

export interface ViewSwitcherProps {
  currentView: DashboardViewMode;
  onViewChange: (view: DashboardViewMode) => void;
  onPluginManagerOpen?: () => void;
}

export interface LoadingStateProps {
  message?: string;
}

export interface ErrorStateProps {
  error: string;
}

export interface WorkflowViewProps {
  nodes: Node<WorkflowNodeData>[];
  edges: Edge[];
  onNodesChange: OnNodesChange;
  onEdgesChange: OnEdgesChange;
  onConnect: (params: Connection) => void;
  onDoubleClick: () => void;
}

export interface DatabaseViewProps {
  schema: DatabaseSchema;
  onTableSelect: (tableName: string) => void;
  onDoubleClick: () => void;
}

export interface ConfigViewProps {}

export interface WorkflowNodeProps {
  data: WorkflowNodeData;
}