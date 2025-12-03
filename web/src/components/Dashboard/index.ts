// 主要组件导出
export * from './Dashboard';

// 子组件导出（供需要时单独使用）
export { default as ViewSwitcher } from './components/ViewSwitcher';
export { default as LoadingState } from './components/LoadingState';
export { default as ErrorState } from './components/ErrorState';
export { default as WorkflowView } from './components/WorkflowView';
export { default as DatabaseView } from './components/DatabaseView';
export { default as ConfigView } from './components/ConfigView';

// Hooks导出
export { useDatabaseSchema, useWorkflowState, useDashboardNavigation } from './hooks';

// 类型导出
export type {
  DashboardViewMode,
  WorkflowNodeData,
  DatabaseSchema,
  ViewSwitcherProps,
  LoadingStateProps,
  ErrorStateProps,
  WorkflowViewProps,
  DatabaseViewProps,
  ConfigViewProps,
  WorkflowNodeProps,
} from './types';

// 数据导出
export { workflowNodes, workflowEdges } from './data';