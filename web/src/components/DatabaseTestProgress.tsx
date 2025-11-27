import React, { useState, useEffect } from 'react';
import {
  Card,
  Button,
  Alert,
  Space,
  Spin,
  Progress,
  Divider,
  Typography,
  Tag
} from 'antd';

const { Text, Title } = Typography;

interface DatabaseConfig {
  type: string;
  path?: string;
  host?: string;
  port?: number;
  database?: string;
  username?: string;
  password?: string;
  ssl_mode?: string;
  charset?: string;
  connection_pool: {
    max_open_conns: number;
    max_idle_conns: number;
    conn_max_lifetime: number;
  };
}

interface DatabaseTestResult {
  step: string;
  status: 'success' | 'failed' | 'running' | 'pending';
  message: string;
  latency?: number;
  details?: any;
}

interface DatabaseTestProgressProps {
  config: {
    database: DatabaseConfig;
    admin: {
      username: string;
      password: string;
      email?: string;
    };
  };
  onTestComplete?: (results: DatabaseTestResult[]) => void;
  autoStart?: boolean;
  showConfig?: boolean;
}

const testSteps = [
  {
    key: 'network_check',
    title: '网络连通性检查',
    description: '检查数据库服务器网络连接'
  },
  {
    key: 'database_connect',
    title: '数据库连接测试',
    description: '建立与数据库的实际连接'
  },
  {
    key: 'permission_check',
    title: '权限验证',
    description: '验证数据库读写权限'
  },
  {
    key: 'table_creation',
    title: '表结构创建测试',
    description: '测试必要的数据表创建'
  }
];

const DatabaseTestProgress: React.FC<DatabaseTestProgressProps> = ({
  config,
  onTestComplete,
  autoStart = false,
  showConfig = true
}) => {
  const [currentStep, setCurrentStep] = useState(0);
  const [isRunning, setIsRunning] = useState(false);
  const [results, setResults] = useState<DatabaseTestResult[]>([]);
  const [currentStepResult, setCurrentStepResult] = useState<DatabaseTestResult | null>(null);

  // 初始化结果
  useEffect(() => {
    setResults(testSteps.map(step => ({
      step: step.key,
      status: 'pending',
      message: step.description
    })));
  }, []);

  // 自动开始测试
  useEffect(() => {
    if (autoStart && config && !isRunning && results.length > 0 && results[0].status === 'pending') {
      setTimeout(() => startTest(), 500);
    }
  }, [autoStart, config, results, isRunning]);

  const startTest = async (isRetry = false) => {
    if (!config || isRunning) return;

    setIsRunning(true);
    setCurrentStep(0);

    // 只有在非重试时才重置结果，避免覆盖之前的测试结果
    if (!isRetry) {
      setResults(testSteps.map(step => ({
        step: step.key,
        status: 'pending',
        message: step.description
      })));
    }

    // 依次执行测试步骤
    const newResults = [];
    for (let i = 0; i < testSteps.length; i++) {
      setCurrentStep(i);
      const result = await executeTestStep(testSteps[i].key, config.database);
      newResults.push(result);

      // 添加延迟以便用户看到进度
      await new Promise(resolve => setTimeout(resolve, 800));
    }

    setCurrentStep(testSteps.length);
    setIsRunning(false);

    if (onTestComplete) {
      onTestComplete(newResults);
    }
  };

  const executeTestStep = async (step: string, dbConfig: DatabaseConfig): Promise<DatabaseTestResult> => {
    try {
      // 更新当前步骤状态为运行中
      setCurrentStepResult({
        step,
        status: 'running',
        message: '正在执行测试...'
      });

      const response = await fetch(`/api/admin/system/test-database-step?step=${step}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(dbConfig)
      });

      const data = await response.json();

      if (data.success) {
        const result = data.data;

        // 更新结果列表
        setResults(prev => prev.map(r =>
          r.step === step ? result : r
        ));

        setCurrentStepResult(result);
        return result;
      } else {
        const errorResult: DatabaseTestResult = {
          step,
          status: 'failed',
          message: data.message || '测试失败'
        };

        setResults(prev => prev.map(r =>
          r.step === step ? errorResult : r
        ));

        setCurrentStepResult(errorResult);
        return errorResult;
      }
    } catch (error) {
      const errorResult: DatabaseTestResult = {
        step,
        status: 'failed',
        message: `网络错误: ${error instanceof Error ? error.message : '未知错误'}`
      };

      setResults(prev => prev.map(r =>
        r.step === step ? errorResult : r
      ));

      setCurrentStepResult(errorResult);
      return errorResult;
    }
  };

  const getStepStatus = (result: DatabaseTestResult) => {
    switch (result.status) {
      case 'success':
        return 'finish';
      case 'failed':
        return 'error';
      case 'running':
        return 'process';
      default:
        return 'wait';
    }
  };

  const getStepIcon = (result: DatabaseTestResult) => {
    switch (result.status) {
      case 'success':
        return <span style={{ color: '#52c41a' }}>✓</span>;
      case 'failed':
        return <span style={{ color: '#ff4d4f' }}>✗</span>;
      case 'running':
        return <Spin size="small" />;
      default:
        return <span style={{ color: '#8c8c8c' }}>○</span>;
    }
  };

  const getOverallStatus = () => {
    const successCount = results.filter(r => r.status === 'success').length;
    const failedCount = results.filter(r => r.status === 'failed').length;
    const totalCount = results.length;

    if (isRunning) {
      return { status: 'running', color: '#1890ff', text: '测试进行中...' };
    } else if (failedCount === 0 && successCount === totalCount && successCount > 0) {
      return { status: 'success', color: '#52c41a', text: '所有测试通过' };
    } else if (failedCount > 0) {
      return { status: 'failed', color: '#ff4d4f', text: `测试失败 (${failedCount}/${totalCount})` };
    } else {
      return { status: 'pending', color: '#8c8c8c', text: '准备就绪' };
    }
  };

  const isTestCompleted = () => {
    const successCount = results.filter(r => r.status === 'success').length;
    const failedCount = results.filter(r => r.status === 'failed').length;
    const totalCount = results.length;
    return (successCount + failedCount) === totalCount && totalCount > 0;
  };

  const overallStatus = getOverallStatus();
  const progress = (results.filter(r => r.status === 'success').length / results.length) * 100;

  return (
    <div>
      {showConfig && (
        <Card title="测试配置" style={{ marginBottom: 16 }}>
          <Space direction="vertical" size="small" style={{ width: '100%' }}>
            <div>
              <Text strong>数据库类型:</Text> {config.database.type}
            </div>
            {config.database.type === 'sqlite' && (
              <div>
                <Text strong>数据库路径:</Text> {config.database.path}
              </div>
            )}
            {(config.database.type === 'mysql' || config.database.type === 'postgresql') && (
              <>
                <div><Text strong>主机:</Text> {config.database.host}:{config.database.port}</div>
                <div><Text strong>数据库:</Text> {config.database.database}</div>
                <div><Text strong>用户名:</Text> {config.database.username}</div>
              </>
            )}
          </Space>
        </Card>
      )}

      <Card title="连接测试进度" extra={
        <Space>
          {!isTestCompleted() ? (
            <Button
              type="primary"
              onClick={startTest}
              loading={isRunning}
              disabled={!config || isRunning}
            >
              {isRunning ? '测试中...' : '开始测试'}
            </Button>
          ) : (
            <Button
              onClick={() => startTest(true)}
              disabled={!config || isRunning}
            >
              重新测试
            </Button>
          )}
        </Space>
      }>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          {/* 总体进度 */}
          <div>
            <Space direction="vertical" size="small" style={{ width: '100%' }}>
              <Space>
                <Title level={5} style={{ margin: 0 }}>总体状态:</Title>
                <Tag color={overallStatus.color}>{overallStatus.text}</Tag>
              </Space>
              <Progress
                percent={progress}
                status={overallStatus.status === 'failed' ? 'exception' : overallStatus.status}
                strokeColor={overallStatus.color}
              />
            </Space>
          </div>

          <Divider />

          {/* 步骤进度展示 */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            {testSteps.map((step, index) => {
              const result = results.find(r => r.step === step.key);
              const stepResult = result || { step: step.key, status: 'pending', message: step.description };

              return (
                <div
                  key={step.key}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    padding: '12px',
                    border: `1px solid ${index === currentStep ? '#1890ff' : '#f0f0f0'}`,
                    borderRadius: '6px',
                    backgroundColor: index === currentStep ? '#f6ffed' : 'white'
                  }}
                >
                  <div style={{ marginRight: '12px' }}>
                    {getStepIcon(stepResult)}
                  </div>
                  <div style={{ flex: 1 }}>
                    <div style={{ fontWeight: 'bold', marginBottom: '4px' }}>
                      {step.title}
                    </div>
                    <div style={{ color: '#666', fontSize: '12px' }}>
                      {step.description}
                    </div>
                    {stepResult.message && (
                      <div style={{ marginTop: '4px', fontSize: '12px' }}>
                        <Text style={{
                          color: stepResult.status === 'failed' ? '#ff4d4f' :
                                 stepResult.status === 'success' ? '#52c41a' : '#1890ff'
                        }}>
                          {stepResult.message}
                        </Text>
                        {stepResult.latency && (
                          <Tag style={{ marginLeft: 8 }}>
                            {stepResult.latency}ms
                          </Tag>
                        )}
                      </div>
                    )}
                  </div>
                  <div style={{ marginLeft: '12px' }}>
                    <Tag color={
                      stepResult.status === 'success' ? 'success' :
                      stepResult.status === 'failed' ? 'error' : 'default'
                    }>
                      {stepResult.status}
                    </Tag>
                  </div>
                </div>
              );
            })}
          </div>

          {/* 错误汇总 */}
          {results.some(r => r.status === 'failed') && (
            <>
              <Divider />
              <Alert
                message="测试失败"
                description={
                  <div>
                    <Text>以下步骤失败：</Text>
                    <ul style={{ marginTop: 8, marginBottom: 0 }}>
                      {results
                        .filter(r => r.status === 'failed')
                        .map(r => (
                          <li key={r.step}>
                            {testSteps.find(s => s.key === r.step)?.title}: {r.message}
                          </li>
                        ))}
                    </ul>
                  </div>
                }
                type="error"
                showIcon
              />
            </>
          )}

          {/* 成功消息 */}
          {!isRunning && results.length > 0 && results.every(r => r.status === 'success') && (
            <>
              <Divider />
              <Alert
                message="连接测试成功"
                description="数据库连接测试全部通过，可以继续进行系统初始化。"
                type="success"
                showIcon
              />
            </>
          )}
        </Space>
      </Card>
    </div>
  );
};

export default DatabaseTestProgress;