import React, { useState } from 'react';
import { FullscreenLayout } from '../layout';
import { useDatabaseSchema, useWorkflowState, useDashboardNavigation } from './hooks';
import { DashboardViewMode } from './types';

// 导入子组件
import ViewSwitcher from './components/ViewSwitcher';
import LoadingState from './components/LoadingState';
import ErrorState from './components/ErrorState';
import WorkflowView from './components/WorkflowView';
import DatabaseView from './components/DatabaseView';
import ConfigView from './components/ConfigView';
import StartupControls from './components/StartupControls';
import { PluginManager } from '../PluginManager/PluginManager';

const Dashboard: React.FC = () => {
  const [viewMode, setViewMode] = useState<DashboardViewMode>('workflow'); // 默认显示工作流节点
  const [pluginManagerVisible, setPluginManagerVisible] = useState(false);

  // 自定义Hooks
  const { schema, loading, error, onTableSelect } = useDatabaseSchema();
  const {
    nodes,
    edges,
    onNodesChange,
    onEdgesChange,
    onConnect,
    isLoading: workflowLoading,
    error: workflowError,
    execution,
    executionId,
    executeWorkflow,
    getExecutionStatus
  } = useWorkflowState({
    autoConnect: true,
    workflowId: 'xiaozhi-flow-default-startup'
  });
  const { handleDoubleClick } = useDashboardNavigation();

  // 处理视图切换
  const handleViewChange = (newView: DashboardViewMode) => {
    setViewMode(newView);
  };

  // 加载状态
  if (loading) {
    return <LoadingState message="加载数据库表结构..." />;
  }

  // 错误状态
  if (error || !schema) {
    return <ErrorState error={error || '未知错误'} />;
  }

  // 处理执行开始
  const handleExecutionStart = (newExecutionId: string) => {
    console.log('启动流程执行开始:', newExecutionId);
  };

  // 处理执行完成
  const handleExecutionComplete = (completedExecution: any) => {
    console.log('启动流程执行完成:', completedExecution);
  };

  return (
    <FullscreenLayout>
      <div className="w-full h-full bg-gray-50 overflow-hidden relative">
        {/* 视图切换按钮 */}
        <ViewSwitcher
          currentView={viewMode}
          onViewChange={handleViewChange}
          onPluginManagerOpen={() => setPluginManagerVisible(true)}
        />

        {/* 启动流程控制面板 */}
        {viewMode === 'workflow' && (
          <div className="absolute top-4 left-4 z-10" style={{ width: 320 }}>
            <StartupControls
              onExecutionStart={handleExecutionStart}
              onExecutionComplete={handleExecutionComplete}
            />
          </div>
        )}

        {/* 内容区域 */}
        {viewMode === 'database' ? (
          <DatabaseView
            schema={schema}
            onTableSelect={onTableSelect}
            onDoubleClick={handleDoubleClick}
          />
        ) : viewMode === 'config' ? (
          <ConfigView />
        ) : (
          <WorkflowView
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onDoubleClick={handleDoubleClick}
          />
        )}
      </div>

      {/* 插件管理器 */}
      <PluginManager
        visible={pluginManagerVisible}
        onClose={() => setPluginManagerVisible(false)}
      />
    </FullscreenLayout>
  );
};

export default Dashboard;