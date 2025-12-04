import React, { useState, useCallback } from 'react';
import { FullscreenLayout } from '../layout';
import { useDatabaseSchema, useWorkflowState } from '../Dashboard/hooks';
import { DashboardViewMode } from '../Dashboard/types';
import { message } from 'antd';

// 导入子组件
import ViewSwitcher from '../Dashboard/components/ViewSwitcher';
import LoadingState from '../Dashboard/components/LoadingState';
import ErrorState from '../Dashboard/components/ErrorState';
import DatabaseView from '../Dashboard/components/DatabaseView/DatabaseView';
import ConfigView from '../Dashboard/components/ConfigView/ConfigView';
import { PluginManager } from '../PluginManager/PluginManager';

// 导入简化版工作流编辑器
import { SimpleReteEditor } from './SimpleReteEditor';

const Dashboardv2: React.FC = () => {
  const [viewMode, setViewMode] = useState<DashboardViewMode>('workflow');
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

  // 处理视图切换
  const handleViewChange = (newView: DashboardViewMode) => {
    setViewMode(newView);
  };

  // 处理节点变化
  const handleNodesChange = useCallback((nodes: any[]) => {
    console.log('节点变化:', nodes);
    // 这里可以添加状态同步逻辑
  }, []);

  // 处理连接变化
  const handleConnectionsChange = useCallback((connections: any[]) => {
    console.log('连接变化:', connections);
    // 这里可以添加状态同步逻辑
  }, []);

  // 处理工作流执行
  const handleExecute = useCallback(() => {
    message.success('工作流执行开始！');
    // 这里可以添加实际的工作流执行逻辑
  }, []);

  // 加载状态
  if (loading) {
    return <LoadingState message="加载数据库表结构..." />;
  }

  // 错误状态
  if (error || !schema) {
    return <ErrorState error={error || '未知错误'} />;
  }

  return (
    <FullscreenLayout>
      <div className="w-full h-full bg-gray-50 overflow-hidden relative">
        {/* 视图切换按钮 */}
        <ViewSwitcher
          currentView={viewMode}
          onViewChange={handleViewChange}
          onPluginManagerOpen={() => setPluginManagerVisible(true)}
        />

        {/* 内容区域 */}
        {viewMode === 'database' ? (
          <DatabaseView
            schema={schema}
            onTableSelect={onTableSelect}
            onDoubleClick={() => {}}
          />
        ) : viewMode === 'config' ? (
          <ConfigView />
        ) : (
          <div className="w-full h-full relative">
            <SimpleReteEditor
              workflowId="xiaozhi-flow-default-startup"
              autoConnect={true}
              onNodesChange={handleNodesChange}
              onConnectionsChange={handleConnectionsChange}
              onExecute={handleExecute}
            />
          </div>
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

export default Dashboardv2;