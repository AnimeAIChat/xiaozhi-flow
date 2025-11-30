import React, { useState, useEffect } from 'react';
import {
  Typography,
  Card,
  Button,
  Alert,
  Steps,
  Space,
  Spin,
  App,
  Divider,
  Tag,
  Result,
  Progress
} from 'antd';
import {
  DatabaseOutlined,
  SaveOutlined,
  PlayCircleOutlined,
  RocketOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useSystemStatus } from '../../hooks/useApi';
import { FullscreenLayout } from '../../components/layout';
import DatabaseConfigForm from '../../components/DatabaseConfigForm';
import DatabaseTestProgress from '../../components/DatabaseTestProgress';
import SystemInitProgress from '../../components/SystemInitProgress';
import { envConfig } from '../../utils/envConfig';
import { apiService } from '../../services/api';

const { Title, Text } = Typography;
const { Step } = Steps;

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

interface AdminConfig {
  username: string;
  password: string;
  email?: string;
}

interface FullConfig {
  database: DatabaseConfig;
  admin: AdminConfig;
}

interface DatabaseTestResult {
  step: string;
  status: 'success' | 'failed' | 'running' | 'pending';
  message: string;
  latency?: number;
  details?: any;
}

const Config: React.FC = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const { data: systemStatus, isLoading: systemLoading } = useSystemStatus();

  const [currentStep, setCurrentStep] = useState(0);
  const [config, setConfig] = useState<FullConfig | null>(null);
  const [testResults, setTestResults] = useState<DatabaseTestResult[]>([]);
  const [initCompleted, setInitCompleted] = useState(false);
  const [loading, setLoading] = useState(false);

  const steps = [
    {
      title: '数据库配置',
      content: '配置数据库连接参数',
      icon: <DatabaseOutlined />
    },
    {
      title: '连接测试',
      content: '测试数据库连接和权限',
      icon: <PlayCircleOutlined />
    },
    {
      title: '保存配置',
      content: '保存配置到系统',
      icon: <SaveOutlined />
    },
    {
      title: '系统初始化',
      content: '初始化数据库和创建管理员',
      icon: <RocketOutlined />
    },
    {
      title: '完成',
      content: '配置和初始化完成',
      icon: <SaveOutlined />
    }
  ];

  // 检查系统是否已经初始化
  useEffect(() => {
    if (!systemLoading && systemStatus) {
      // 只有当系统已经完全初始化且用户刚开始配置流程时才跳转到 dashboard
      if (systemStatus.initialized === true && systemStatus.needs_setup !== true && currentStep === 0) {
        navigate('/dashboard', { replace: true });
      }
    }
  }, [systemStatus, systemLoading, navigate, currentStep]);

  const handleConfigChange = (newConfig: FullConfig) => {
    setConfig(newConfig);
  };

  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handlePrev = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleTestComplete = (results: DatabaseTestResult[]) => {
    setTestResults(results);
    const allSuccess = results.every(r => r.status === 'success');

    if (allSuccess) {
      message.success('数据库连接测试通过！');
      // 不再自动跳转到下一步，让用户手动确认
    } else {
      message.error('数据库连接测试失败，请检查配置');
    }
  };

  const handleSaveConfig = async () => {
    if (!config) {
      message.error('请先配置数据库连接');
      return;
    }

    try {
      setLoading(true);

      const configData = {
        ...config,
        initialized: false,
        version: '1.0.0',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      };

      const data = await apiService.saveDatabaseConfig(configData);

      if (data.success) {
        message.success('配置保存成功！');
        // 保存配置成功后跳转到初始化步骤
        setTimeout(() => handleNext(), 1000);
      } else {
        message.error(`保存配置失败: ${data.message}`);
      }
    } catch (error) {
      console.error('保存配置失败:', error);
      message.error(`保存配置失败: ${error instanceof Error ? error.message : '未知错误'}`);
    } finally {
      setLoading(false);
    }
  };

  
  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return (
          <DatabaseConfigForm
            onConfigChange={handleConfigChange}
            loading={loading}
          />
        );

      case 1:
        return (
          <DatabaseTestProgress
            config={config}
            onTestComplete={handleTestComplete}
            autoStart={true}
          />
        );

      case 2:
        return (
          <Card title="配置确认" extra={<Tag color="blue">待保存</Tag>}>
            <Space orientation="vertical" size="large" style={{ width: '100%' }}>
              <div>
                <Title level={5}>数据库配置</Title>
                <Space orientation="vertical" size="small" style={{ width: '100%' }}>
                  <div><Text strong>类型:</Text> {config?.database.type}</div>
                  {config?.database.type === 'sqlite' && (
                    <div><Text strong>路径:</Text> {config?.database.path}</div>
                  )}
                  {(config?.database.type === 'mysql' || config?.database.type === 'postgresql') && (
                    <>
                      <div><Text strong>主机:</Text> {config?.database.host}:{config?.database.port}</div>
                      <div><Text strong>数据库:</Text> {config?.database.database}</div>
                      <div><Text strong>用户名:</Text> {config?.database.username}</div>
                    </>
                  )}
                </Space>
              </div>

              <Divider />

              <div>
                <Title level={5}>管理员配置</Title>
                <Space orientation="vertical" size="small" style={{ width: '100%' }}>
                  <div><Text strong>用户名:</Text> {config?.admin.username}</div>
                  <div><Text strong>密码:</Text> {config?.admin.password || '未设置'}</div>
                  {config?.admin.email && (
                    <div><Text strong>邮箱:</Text> {config?.admin.email}</div>
                  )}
                </Space>
              </div>

              <Alert
                title="配置保存确认"
                description="点击保存按钮将配置写入系统，之后可以继续进行系统初始化。"
                type="info"
                showIcon
              />
            </Space>
          </Card>
        );

      case 3:
        return (
          <SystemInitProgress
            config={config!}
            onInitComplete={(success) => {
              if (success) {
                message.success('系统初始化成功！');
                setInitCompleted(true);
                setTimeout(() => {
                  setCurrentStep(4); // 跳转到完成步骤
                }, 1000);
              } else {
                message.error('系统初始化失败，请检查错误信息');
              }
            }}
            autoStart={true}
          />
        );

      case 4:
        return (
          <Card title="配置完成">
            <Space orientation="vertical" size="large" style={{ width: '100%', textAlign: 'center' }}>
              <Result
                status="success"
                title="数据库配置和系统初始化完成！"
                subTitle="系统已成功配置并初始化，现在可以开始使用了。"
                extra={
                  <Button type="primary" size="large" onClick={() => {
                    console.log('Config: Manual navigation to dashboard');
                    // 设置标志表示从配置页面跳转
                    sessionStorage.setItem('comingFromConfig', 'true');
                    // 添加延迟确保系统状态完全更新
                    setTimeout(() => {
                      navigate('/dashboard', { replace: true });
                    }, 100);
                  }}>
                    进入管理界面
                  </Button>
                }
              />
            </Space>
          </Card>
        );

      default:
        return null;
    }
  };

  const canGoNext = () => {
    switch (currentStep) {
      case 0:
        return config !== null;
      case 1:
        return testResults.length > 0 && testResults.every(r => r.status === 'success');
      case 2:
        return true; // 保存步骤总是可以继续
      case 3:
        return initCompleted;
      default:
        return false;
    }
  };

  if (loading && !config) {
    return (
      <FullscreenLayout>
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
          <Space orientation="vertical" align="center">
            <Spin size="large" />
            <Text>加载配置中...</Text>
          </Space>
        </div>
      </FullscreenLayout>
    );
  }

  return (
    <FullscreenLayout>
      <div style={{ maxWidth: 1200, margin: '0 auto', width: '100%' }}>
        <div style={{ marginBottom: 32, textAlign: 'center' }}>
          <Title level={2}>数据库配置</Title>
          <Text type="secondary">配置数据库连接并初始化系统</Text>
        </div>

        <Card>
          <Steps current={currentStep} items={steps} style={{ marginBottom: 32 }} />

          <div style={{ marginBottom: 32, minHeight: 400 }}>
            {renderStepContent()}
          </div>

          {currentStep < 3 && (
            <div style={{ textAlign: 'right' }}>
              <Space>
                {currentStep > 0 && (
                  <Button onClick={handlePrev}>
                    上一步
                  </Button>
                )}
                {currentStep < 2 && (
                  <Button
                    type="primary"
                    onClick={handleNext}
                    disabled={!canGoNext()}
                  >
                    下一步
                  </Button>
                )}
                {currentStep === 2 && (
                  <Button
                    type="primary"
                    onClick={handleSaveConfig}
                    loading={loading}
                    icon={<SaveOutlined />}
                  >
                    保存配置
                  </Button>
                )}
              </Space>
            </div>
          )}
        </Card>
      </div>
    </FullscreenLayout>
  );
};

export default Config;