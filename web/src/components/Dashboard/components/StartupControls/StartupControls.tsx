import React, { useState, useCallback } from 'react';
import { Button, Card, Space, Tag, Progress, Tooltip, Modal, message, Select, Input } from 'antd';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  StopOutlined,
  ReloadOutlined,
  SettingOutlined,
  DeleteOutlined,
  InfoCircleOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  CloseCircleOutlined
} from '@ant-design/icons';
import { useWorkflowState } from '../../hooks';
import { generateStartupStats, formatNodeStatus } from '../../../../utils/startupDataConverter';
import { log } from '../../../../utils/logger';

const { Option } = Select;

interface StartupControlsProps {
  className?: string;
  onExecutionStart?: (executionId: string) => void;
  onExecutionComplete?: (execution: any) => void;
}

const StartupControls: React.FC<StartupControlsProps> = ({
  className,
  onExecutionStart,
  onExecutionComplete
}) => {
  const [showSettingsModal, setShowSettingsModal] = useState(false);
  const [showStatsModal, setShowStatsModal] = useState(false);
  const [executionInputs, setExecutionInputs] = useState<Record<string, any>>({});
  const [selectedWorkflowId, setSelectedWorkflowId] = useState('xiaozhi-flow-default-startup');

  const {
    workflow,
    execution,
    isLoading,
    error,
    isConnected,
    workflowId,
    executionId,
    executeWorkflow,
    cancelExecution,
    pauseExecution,
    resumeExecution,
    refresh,
    switchWorkflow
  } = useWorkflowState({
    autoConnect: true,
    workflowId: selectedWorkflowId
  });

  // 处理执行工作流
  const handleExecuteWorkflow = useCallback(async () => {
    try {
      log.info('用户请求执行启动工作流', { workflow_id: workflowId }, 'ui', 'StartupControls');

      const newExecutionId = await executeWorkflow(executionInputs);

      message.success('启动工作流执行已开始');
      onExecutionStart?.(newExecutionId);

      log.info('启动工作流执行成功', {
        workflow_id: workflowId,
        execution_id: newExecutionId
      }, 'ui', 'StartupControls');

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '执行失败';
      message.error(`执行失败: ${errorMessage}`);
      log.error('启动工作流执行失败', {
        workflow_id: workflowId,
        error: errorMessage
      }, 'ui', 'StartupControls');
    }
  }, [workflowId, executionInputs, executeWorkflow, onExecutionStart]);

  // 处理取消执行
  const handleCancelExecution = useCallback(async () => {
    Modal.confirm({
      title: '确认取消执行',
      content: '取消执行将停止所有正在运行的节点，此操作不可恢复。',
      okText: '确认取消',
      cancelText: '继续执行',
      okType: 'danger',
      onOk: async () => {
        try {
          await cancelExecution();
          message.success('执行已取消');
          log.info('用户取消执行', { execution_id: executionId }, 'ui', 'StartupControls');
        } catch (err) {
          const errorMessage = err instanceof Error ? err.message : '取消失败';
          message.error(`取消失败: ${errorMessage}`);
          log.error('取消执行失败', {
            execution_id: executionId,
            error: errorMessage
          }, 'ui', 'StartupControls');
        }
      }
    });
  }, [executionId, cancelExecution]);

  // 处理暂停/恢复执行
  const handleToggleExecution = useCallback(async () => {
    if (!execution) return;

    try {
      if (execution.status === 'paused') {
        await resumeExecution();
        message.success('执行已恢复');
        log.info('用户恢复执行', { execution_id: executionId }, 'ui', 'StartupControls');
      } else if (execution.status === 'running') {
        await pauseExecution();
        message.success('执行已暂停');
        log.info('用户暂停执行', { execution_id: executionId }, 'ui', 'StartupControls');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '操作失败';
      message.error(`操作失败: ${errorMessage}`);
      log.error('切换执行状态失败', {
        execution_id: executionId,
        error: errorMessage
      }, 'ui', 'StartupControls');
    }
  }, [execution, executionId, pauseExecution, resumeExecution]);

  // 处理切换工作流
  const handleSwitchWorkflow = useCallback(async (newWorkflowId: string) => {
    if (execution && (execution.status === 'running' || execution.status === 'paused')) {
      message.warning('请先停止当前执行');
      return;
    }

    try {
      setSelectedWorkflowId(newWorkflowId);
      await switchWorkflow(newWorkflowId);
      message.success('已切换工作流');
      log.info('用户切换工作流', {
        from_workflow_id: workflowId,
        to_workflow_id: newWorkflowId
      }, 'ui', 'StartupControls');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '切换失败';
      message.error(`切换失败: ${errorMessage}`);
      log.error('切换工作流失败', {
        from_workflow_id: workflowId,
        to_workflow_id: newWorkflowId,
        error: errorMessage
      }, 'ui', 'StartupControls');
    }
  }, [workflowId, execution, switchWorkflow]);

  // 处理刷新
  const handleRefresh = useCallback(() => {
    refresh();
    message.info('已刷新数据');
    log.info('用户刷新工作流数据', { workflow_id: workflowId }, 'ui', 'StartupControls');
  }, [workflowId, refresh]);

  // 获取执行状态标签
  const getExecutionStatusTag = useCallback(() => {
    if (!execution) return null;

    const statusConfig = {
      pending: { color: 'default', icon: <ClockCircleOutlined />, text: '等待中' },
      running: { color: 'processing', icon: <ThunderboltOutlined />, text: '运行中' },
      paused: { color: 'warning', icon: <PauseCircleOutlined />, text: '已暂停' },
      completed: { color: 'success', icon: <CheckCircleOutlined />, text: '已完成' },
      failed: { color: 'error', icon: <ExclamationCircleOutlined />, text: '执行失败' },
      cancelled: { color: 'default', icon: <CloseCircleOutlined />, text: '已取消' }
    };

    const config = statusConfig[execution.status as keyof typeof statusConfig];
    if (!config) return null;

    return (
      <Tag color={config.color} icon={config.icon}>
        {config.text}
      </Tag>
    );
  }, [execution]);

  // 生成统计信息
  const stats = execution ? generateStartupStats(execution) : null;

  return (
    <div className={className}>
      <Card size="small" title="启动流程控制">
        <Space orientation="vertical" style={{ width: '100%' }}>
          {/* 工作流选择和状态 */}
          <Space wrap>
            <Select
              value={selectedWorkflowId}
              onChange={handleSwitchWorkflow}
              style={{ width: 200 }}
              placeholder="选择工作流"
              disabled={execution && (execution.status === 'running' || execution.status === 'paused')}
            >
              <Option value="xiaozhi-flow-default-startup">默认启动工作流</Option>
              <Option value="xiaozhi-flow-parallel-startup">并行启动工作流</Option>
              <Option value="xiaozhi-flow-minimal-startup">最小启动工作流</Option>
            </Select>

            {getExecutionStatusTag()}

            <Tooltip title={isConnected ? 'WebSocket已连接' : 'WebSocket未连接'}>
              <Tag color={isConnected ? 'success' : 'error'}>
                {isConnected ? '已连接' : '未连接'}
              </Tag>
            </Tooltip>

            {workflow && (
              <Tag color="blue">
                {workflow.nodes?.length || 0} 个节点
              </Tag>
            )}
          </Space>

          {/* 执行进度 */}
          {execution && (
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                <span>执行进度</span>
                <span>{Math.round(execution.progress * 100)}%</span>
              </div>
              <Progress
                percent={Math.round(execution.progress * 100)}
                status={execution.status === 'failed' ? 'exception' : execution.status === 'completed' ? 'success' : 'active'}
                size="small"
              />
            </div>
          )}

          {/* 控制按钮 */}
          <Space wrap>
            {/* 执行按钮 */}
            {!execution ? (
              <Button
                type="primary"
                icon={<PlayCircleOutlined />}
                onClick={handleExecuteWorkflow}
                loading={isLoading}
                disabled={!isConnected}
              >
                执行启动流程
              </Button>
            ) : (
              <>
                {/* 暂停/恢复按钮 */}
                {(execution.status === 'running' || execution.status === 'paused') && (
                  <Button
                    icon={execution.status === 'paused' ? <PlayCircleOutlined /> : <PauseCircleOutlined />}
                    onClick={handleToggleExecution}
                    disabled={!isConnected}
                  >
                    {execution.status === 'paused' ? '恢复' : '暂停'}
                  </Button>
                )}

                {/* 取消按钮 */}
                {(execution.status === 'running' || execution.status === 'paused') && (
                  <Button
                    danger
                    icon={<StopOutlined />}
                    onClick={handleCancelExecution}
                    disabled={!isConnected}
                  >
                    取消执行
                  </Button>
                )}

                {/* 重新执行按钮 */}
                {execution.status === 'completed' || execution.status === 'failed' || execution.status === 'cancelled' ? (
                  <Button
                    icon={<ReloadOutlined />}
                    onClick={handleExecuteWorkflow}
                    disabled={!isConnected}
                  >
                    重新执行
                  </Button>
                ) : null}
              </>
            )}

            {/* 功能按钮 */}
            <Button
              icon={<ReloadOutlined />}
              onClick={handleRefresh}
              loading={isLoading}
            >
              刷新
            </Button>

            <Button
              icon={<SettingOutlined />}
              onClick={() => setShowSettingsModal(true)}
            >
              执行配置
            </Button>

            <Button
              icon={<InfoCircleOutlined />}
              onClick={() => setShowStatsModal(true)}
              disabled={!stats}
            >
              执行统计
            </Button>
          </Space>

          {/* 错误信息 */}
          {error && (
            <div style={{ padding: '8px 12px', backgroundColor: '#fff2f0', border: '1px solid #ffccc7', borderRadius: 6 }}>
              <span style={{ color: '#ff4d4f', fontSize: 12 }}>
                错误: {error}
              </span>
            </div>
          )}
        </Space>
      </Card>

      {/* 执行配置模态框 */}
      <Modal
        title="执行配置"
        open={showSettingsModal}
        onOk={() => setShowSettingsModal(false)}
        onCancel={() => setShowSettingsModal(false)}
        footer={[
          <Button key="cancel" onClick={() => setShowSettingsModal(false)}>
            取消
          </Button>,
          <Button key="ok" type="primary" onClick={() => setShowSettingsModal(false)}>
            确定
          </Button>
        ]}
      >
        <Space orientation="vertical" style={{ width: '100%' }}>
          <div>
            <strong>工作流ID:</strong> {workflowId}
          </div>
          {workflow && (
            <div>
              <strong>工作流名称:</strong> {workflow.name}
            </div>
          )}
          <div>
            <strong>自定义输入参数:</strong>
          </div>
          <Input.TextArea
            placeholder="输入JSON格式的参数，例如: {&quot;env&quot;: &quot;development&quot;}"
            value={JSON.stringify(executionInputs, null, 2)}
            onChange={(e) => {
              try {
                const value = e.target.value;
                if (value.trim() === '') {
                  setExecutionInputs({});
                } else {
                  setExecutionInputs(JSON.parse(value));
                }
              } catch (err) {
                // 忽略JSON解析错误
              }
            }}
            rows={6}
          />
        </Space>
      </Modal>

      {/* 执行统计模态框 */}
      <Modal
        title="执行统计"
        open={showStatsModal}
        onOk={() => setShowStatsModal(false)}
        onCancel={() => setShowStatsModal(false)}
        footer={[
          <Button key="ok" type="primary" onClick={() => setShowStatsModal(false)}>
            确定
          </Button>
        ]}
      >
        {stats && (
          <Space orientation="vertical" style={{ width: '100%' }}>
            <div><strong>总节点数:</strong> {stats.total}</div>
            <div><strong>已完成:</strong> {stats.completed}</div>
            <div><strong>失败:</strong> {stats.failed}</div>
            <div><strong>运行中:</strong> {stats.running}</div>
            <div><strong>等待中:</strong> {stats.pending}</div>
            <div><strong>进度:</strong> {stats.progress}%</div>
            <div><strong>耗时:</strong> {stats.duration}秒</div>
            <div><strong>开始时间:</strong> {stats.startTime}</div>
            {stats.endTime && (
              <div><strong>结束时间:</strong> {stats.endTime}</div>
            )}
            {stats.criticalFailed > 0 && (
              <div style={{ color: '#ff4d4f' }}>
                <strong>关键节点失败:</strong> {stats.criticalFailed}
              </div>
            )}
            {stats.optionalSkipped > 0 && (
              <div style={{ color: '#fa8c16' }}>
                <strong>跳过可选节点:</strong> {stats.optionalSkipped}
              </div>
            )}
          </Space>
        )}
      </Modal>
    </div>
  );
};

export default StartupControls;