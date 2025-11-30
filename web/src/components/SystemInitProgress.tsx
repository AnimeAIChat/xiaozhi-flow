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
  Tag,
  Result
} from 'antd';
import {
  RocketOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined
} from '@ant-design/icons';
import { useInitializeProject } from '../hooks/useApi';

const { Text, Title } = Typography;

interface InitStep {
  name: string;
  success: boolean;
  message: string;
}

interface SystemInitProgressProps {
  config: {
    database: any;
    admin: {
      username: string;
      password: string;
      email?: string;
    };
  };
  onInitComplete?: (success: boolean) => void;
  autoStart?: boolean;
}

const initSteps = [
  {
    key: 'validate_config',
    title: '验证配置参数',
    description: '验证数据库和管理员配置'
  },
  {
    key: 'init_database',
    title: '初始化数据库',
    description: '连接数据库并创建表结构'
  },
  {
    key: 'create_admin',
    title: '创建管理员用户',
    description: '创建默认管理员账户'
  },
  {
    key: 'load_config',
    title: '加载默认配置',
    description: '加载系统默认配置'
  },
  {
    key: 'start_services',
    title: '启动核心服务',
    description: '启动系统核心服务'
  },
  {
    key: 'verify_services',
    title: '验证服务连接',
    description: '验证所有服务正常运行'
  },
  {
    key: 'update_config',
    title: '更新配置文件',
    description: '标记系统为已初始化状态'
  }
];

const SystemInitProgress: React.FC<SystemInitProgressProps> = ({
  config,
  onInitComplete,
  autoStart = false
}) => {
  const [currentStep, setCurrentStep] = useState(-1);
  const [isRunning, setIsRunning] = useState(false);
  const [steps, setSteps] = useState<InitStep[]>([]);
  const [initCompleted, setInitCompleted] = useState(false);
  const [initSuccess, setInitSuccess] = useState(false);
  const initializeProject = useInitializeProject();

  // 初始化步骤
  useEffect(() => {
    setSteps(initSteps.map(step => ({
      name: step.title,
      success: false,
      message: step.description
    })));
  }, []);

  // 自动开始初始化
  useEffect(() => {
    if (autoStart && config && !isRunning && steps.length > 0 && currentStep === -1) {
      setTimeout(() => startInit(), 500);
    }
  }, [autoStart, config, steps, isRunning, currentStep]);

  const startInit = async (isRetry = false) => {
    if (!config || isRunning) return;

    setIsRunning(true);
    setCurrentStep(0);

    // 只有在非重试时才重置结果
    if (!isRetry) {
      setSteps(initSteps.map(step => ({
        name: step.title,
        success: false,
        message: step.description
      })));
      setInitCompleted(false);
      setInitSuccess(false);
    }

    try {
      const initConfig = {
        databaseConfig: config.database,
        adminConfig: config.admin
      };

      const result = await initializeProject.mutateAsync(initConfig);

      // API 返回的数据结构: {success: true, steps: [...], message: '...', configId: '...'}
      if (result.success && result.steps) {
        // 模拟实时进度更新
        for (let i = 0; i < result.steps.length; i++) {
          setCurrentStep(i);
          const step = result.steps[i];

          // 更新步骤状态
          setSteps(prev => prev.map((s, index) =>
            index === i
              ? { ...s, success: step.success, message: step.message }
              : s
          ));

          // 添加延迟以便用户看到进度
          await new Promise(resolve => setTimeout(resolve, 800));
        }

        setCurrentStep(result.steps.length);
        setInitCompleted(true);
        setInitSuccess(result.steps.every((step: any) => step.success));

        if (onInitComplete) {
          onInitComplete(result.steps.every((step: any) => step.success));
        }
      } else {
        // 处理错误情况
        setSteps(prev => prev.map(s => ({
          ...s,
          success: false,
          message: result.message || '初始化失败'
        })));
        setInitCompleted(true);
        setInitSuccess(false);

        if (onInitComplete) {
          onInitComplete(false);
        }
      }
    } catch (error) {
      console.error('系统初始化失败:', error);
      setSteps(prev => prev.map(s => ({
        ...s,
        success: false,
        message: `网络错误: ${error instanceof Error ? error.message : '未知错误'}`
      })));
      setInitCompleted(true);
      setInitSuccess(false);

      if (onInitComplete) {
        onInitComplete(false);
      }
    } finally {
      setIsRunning(false);
    }
  };

  const getStepStatus = (step: InitStep, index: number) => {
    if (index === currentStep && isRunning) {
      return { status: 'process', color: '#1890ff', text: '进行中' };
    }
    if (step.success) {
      return { status: 'finish', color: '#52c41a', text: '完成' };
    }
    if (initCompleted && !step.success) {
      return { status: 'error', color: '#ff4d4f', text: '失败' };
    }
    return { status: 'wait', color: '#d9d9d9', text: '等待' };
  };

  const getStepIcon = (step: InitStep, index: number) => {
    const status = getStepStatus(step, index);

    if (index === currentStep && isRunning) {
      return <Spin size="small" />;
    }
    if (step.success) {
      return <CheckCircleOutlined style={{ color: status.color }} />;
    }
    if (initCompleted && !step.success) {
      return <CloseCircleOutlined style={{ color: status.color }} />;
    }
    return <span style={{ color: status.color }}>○</span>;
  };

  const getOverallStatus = () => {
    if (isRunning) {
      return { status: 'running', color: '#1890ff', text: '系统初始化中...' };
    }
    if (initCompleted && initSuccess) {
      return { status: 'success', color: '#52c41a', text: '系统初始化成功' };
    }
    if (initCompleted && !initSuccess) {
      return { status: 'failed', color: '#ff4d4f', text: '系统初始化失败' };
    }
    return { status: 'pending', color: '#8c8c8c', text: '准备就绪' };
  };

  const successCount = steps.filter(s => s.success).length;
  const progress = steps.length > 0 ? (successCount / steps.length) * 100 : 0;
  const overallStatus = getOverallStatus();

  return (
    <div>
      <Card title="系统初始化进度" extra={
        <Space>
          {!initCompleted ? (
            <Button
              type="primary"
              onClick={() => startInit()}
              loading={isRunning}
              disabled={!config || isRunning}
              icon={<RocketOutlined />}
            >
              {isRunning ? '初始化中...' : '开始初始化'}
            </Button>
          ) : (
            <Button
              onClick={() => startInit(true)}
              disabled={!config || isRunning}
            >
              重新初始化
            </Button>
          )}
        </Space>
      }>
        <Space orientation="vertical" size="large" style={{ width: '100%' }}>
          {/* 总体进度 */}
          <div>
            <Space orientation="vertical" size="small" style={{ width: '100%' }}>
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
            {initSteps.map((step, index) => {
              const stepResult = steps[index] || { name: step.title, success: false, message: step.description };
              const stepStatus = getStepStatus(stepResult, index);

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
                    {getStepIcon(stepResult, index)}
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
                          color: stepStatus.status === 'error' ? '#ff4d4f' :
                                 stepStatus.status === 'finish' ? '#52c41a' : '#1890ff'
                        }}>
                          {stepResult.message}
                        </Text>
                      </div>
                    )}
                  </div>
                  <div style={{ marginLeft: '12px' }}>
                    <Tag color={stepStatus.color}>
                      {stepStatus.text}
                    </Tag>
                  </div>
                </div>
              );
            })}
          </div>
        </Space>
      </Card>
    </div>
  );
};

export default SystemInitProgress;