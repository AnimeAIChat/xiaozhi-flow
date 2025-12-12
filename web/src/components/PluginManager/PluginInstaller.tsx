import {
  CheckCircleOutlined,
  CloudUploadOutlined,
  FileTextOutlined,
  InboxOutlined,
  InfoCircleOutlined,
  LinkOutlined,
  LoadingOutlined,
  UploadOutlined,
  WarningOutlined,
} from '@ant-design/icons';
import {
  Alert,
  Button,
  Card,
  Checkbox,
  Col,
  Descriptions,
  Divider,
  Form,
  Input,
  List,
  Modal,
  message,
  Progress,
  Radio,
  Row,
  Select,
  Space,
  Steps,
  Tabs,
  Tag,
  Typography,
  Upload,
} from 'antd';
import type { UploadFile, UploadProps } from 'antd/es/upload/interface';
import type React from 'react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { pluginInstaller } from '../../plugins/core/PluginInstaller';
import type { PluginSource } from '../../plugins/types';

const { Title, Text, Paragraph } = Typography;
const { Step } = Steps;
const { TabPane } = Tabs;
const { Dragger } = Upload;

interface PluginInstallerProps {
  visible: boolean;
  onClose: () => void;
  onComplete?: () => void;
  initialSource?: PluginSource;
}

interface InstallationStep {
  title: string;
  description: string;
  status: 'wait' | 'process' | 'finish' | 'error';
  icon?: React.ReactNode;
  error?: string;
}

export const PluginInstaller: React.FC<PluginInstallerProps> = ({
  visible,
  onClose,
  onComplete,
  initialSource,
}) => {
  const [currentStep, setCurrentStep] = useState(0);
  const [sourceType, setSourceType] = useState<'local' | 'url' | 'market'>(
    initialSource?.type || 'local',
  );
  const [installationSteps, setInstallationSteps] = useState<
    InstallationStep[]
  >([]);
  const [progress, setProgress] = useState(0);
  const [loading, setLoading] = useState(false);
  const [installedPlugin, setInstalledPlugin] = useState<any>(null);
  const [validationErrors, setValidationErrors] = useState<string[]>([]);

  const [localPath, setLocalPath] = useState('');
  const [urlPath, setUrlPath] = useState('');
  const [marketplacePlugin, setMarketplacePlugin] = useState<any>(null);
  const [advancedOptions, setAdvancedOptions] = useState({
    force: false,
    autoStart: true,
    overwrite: false,
  });

  const formRef = useRef<any>(null);

  // 重置状态
  const resetState = useCallback(() => {
    setCurrentStep(0);
    setInstallationSteps([]);
    setProgress(0);
    setLoading(false);
    setInstalledPlugin(null);
    setValidationErrors([]);
    setLocalPath('');
    setUrlPath('');
    setMarketplacePlugin(null);
  }, []);

  // 验证输入
  const validateInput = useCallback(
    (type: typeof sourceType): boolean => {
      const errors: string[] = [];

      switch (type) {
        case 'local':
          if (!localPath.trim()) {
            errors.push('请选择插件文件夹');
          }
          break;
        case 'url':
          if (!urlPath.trim()) {
            errors.push('请输入插件URL');
          } else if (!isValidUrl(urlPath.trim())) {
            errors.push('请输入有效的URL');
          }
          break;
        case 'market':
          if (!marketplacePlugin) {
            errors.push('请选择要安装的插件');
          }
          break;
      }

      setValidationErrors(errors);
      return errors.length === 0;
    },
    [localPath, urlPath, marketplacePlugin],
  );

  // URL验证
  const isValidUrl = useCallback((url: string): boolean => {
    try {
      new URL(url);
      return true;
    } catch {
      return false;
    }
  }, []);

  // 创建安装步骤
  const createInstallationSteps = useCallback(
    (type: typeof sourceType): InstallationStep[] => {
      const steps: InstallationStep[] = [
        {
          title: '准备安装',
          description: '验证插件源',
          status: 'wait',
          icon: <FileTextOutlined />,
        },
      ];

      if (type === 'local') {
        steps.push({
          title: '读取本地文件',
          description: '检查插件配置文件',
          status: 'wait',
          icon: <CloudUploadOutlined />,
        });
      } else if (type === 'url') {
        steps.push({
          title: '下载插件',
          description: '从远程URL下载插件文件',
          status: 'wait',
          icon: <CloudUploadOutlined />,
        });
      } else if (type === 'market') {
        steps.push({
          title: '获取插件',
          description: '从市场获取插件文件',
          status: 'wait',
          icon: <CloudUploadOutlined />,
        });
      }

      steps.push(
        {
          title: '验证插件',
          description: '验证插件配置和依赖',
          status: 'wait',
          icon: <CheckCircleOutlined />,
        },
        {
          title: '安装插件',
          description: '复制文件并注册插件',
          status: 'wait',
          icon: <LoadingOutlined />,
        },
        {
          title: '完成安装',
          description: '插件安装成功',
          status: 'wait',
          icon: <CheckCircleOutlined />,
        },
      );

      return steps;
    },
    [],
  );

  // 更新步骤状态
  const updateStepStatus = useCallback(
    (stepIndex: number, status: InstallationStep['status'], error?: string) => {
      setInstallationSteps((prev) => {
        const newSteps = [...prev];
        if (stepIndex < newSteps.length) {
          newSteps[stepIndex] = { ...newSteps[stepIndex], status, error };
        }
        return newSteps;
      });
    },
    [],
  );

  // 开始安装
  const startInstallation = useCallback(async () => {
    if (!validateInput(sourceType)) {
      return;
    }

    setLoading(true);
    setCurrentStep(0);
    setProgress(0);
    setInstallationSteps(createInstallationSteps(sourceType));

    try {
      // 创建插件源
      let source: PluginSource;

      switch (sourceType) {
        case 'local':
          source = {
            type: 'local',
            localPath: localPath.trim(),
            options: advancedOptions,
          };
          break;
        case 'url':
          source = {
            type: 'url',
            url: urlPath.trim(),
            options: advancedOptions,
          };
          break;
        case 'market':
          source = {
            type: 'market',
            marketId: marketplacePlugin.id,
            options: advancedOptions,
          };
          break;
      }

      // 监听安装进度
      pluginInstaller.on('install-progress', (data) => {
        const stepIndex = Math.floor(data.progress.progress / 20); // 假设每个步骤占20%
        if (stepIndex < installationSteps.length) {
          updateStepStatus(stepIndex, 'process');
        }
        setProgress(data.progress.progress);
      });

      // 开始安装
      const plugin = await pluginInstaller.install(source);
      setInstalledPlugin(plugin);

      // 完成所有步骤
      for (let i = 0; i < installationSteps.length; i++) {
        updateStepStatus(i, 'finish');
      }

      setProgress(100);
      message.success(`插件 "${plugin.name}" 安装成功！`);
    } catch (error) {
      // 标记错误步骤
      updateStepStatus(
        currentStep,
        'error',
        error instanceof Error ? error.message : '安装失败',
      );
      message.error(`安装失败: ${error}`);
    } finally {
      setLoading(false);
    }
  }, [
    sourceType,
    validateInput,
    createInstallationSteps,
    installationSteps,
    updateStepStatus,
    localPath,
    urlPath,
    marketplacePlugin,
    advancedOptions,
  ]);

  // 本地文件夹上传配置
  const localUploadProps: UploadProps = {
    name: 'file',
    multiple: false,
    showUploadList: false,
    customRequest: ({ file, onSuccess }) => {
      // 这里只是模拟，实际应该处理文件夹上传
      setLocalPath(file.name);
      onSuccess?.({} as any, file as any);
    },
    onChange(info) {
      if (info.file.status === 'done') {
        setLocalPath(info.file.name);
      }
    },
  };

  // 插件信息展示
  const PluginInfo = useCallback(() => {
    if (!installedPlugin) return null;

    return (
      <Card title="插件信息" size="small" style={{ marginBottom: 16 }}>
        <Descriptions column={1} size="small">
          <Descriptions.Item label="名称">
            {installedPlugin.name}
          </Descriptions.Item>
          <Descriptions.Item label="版本">
            {installedPlugin.version}
          </Descriptions.Item>
          <Descriptions.Item label="作者">
            {installedPlugin.author}
          </Descriptions.Item>
          <Descriptions.Item label="类型">
            <Tag color={installedPlugin.type === 'backend' ? 'red' : 'blue'}>
              {installedPlugin.type}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="运行时">
            <Tag color="green">{installedPlugin.runtime}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="分类">
            <Tag color="orange">{installedPlugin.metadata?.category}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="描述">
            {installedPlugin.description}
          </Descriptions.Item>
        </Descriptions>
      </Card>
    );
  }, [installedPlugin]);

  // 渲染步骤内容
  const renderStepContent = useCallback(() => {
    switch (currentStep) {
      case 0:
        return (
          <div>
            <Title level={4}>选择安装方式</Title>
            <Radio.Group
              value={sourceType}
              onChange={(e) => {
                setSourceType(e.target.value);
                setValidationErrors([]);
              }}
            >
              <Space direction="vertical">
                <Radio value="local">
                  <Space>
                    <Text strong>本地文件夹</Text>
                    <Text type="secondary">从本地目录安装插件</Text>
                  </Space>
                </Radio>
                <Radio value="url">
                  <Space>
                    <Text strong>远程URL</Text>
                    <Text type="secondary">通过URL下载安装插件</Text>
                  </Space>
                </Radio>
                <Radio value="market">
                  <Space>
                    <Text strong>插件市场</Text>
                    <Text type="secondary">从插件市场选择安装</Text>
                  </Space>
                </Radio>
              </Space>
            </Radio.Group>

            {validationErrors.length > 0 && (
              <Alert
                message="请完成以下必填项"
                description={
                  <ul style={{ margin: 0, paddingLeft: 20 }}>
                    {validationErrors.map((error, index) => (
                      <li key={index}>{error}</li>
                    ))}
                  </ul>
                }
                type="error"
                style={{ marginTop: 16 }}
              />
            )}
          </div>
        );

      case 1:
        switch (sourceType) {
          case 'local':
            return (
              <div>
                <Title level={4}>选择插件文件夹</Title>
                <Dragger {...localUploadProps}>
                  <p className="ant-upload-drag-icon">
                    <InboxOutlined />
                  </p>
                  <p className="ant-upload-text">
                    点击或拖拽插件文件夹到此区域上传
                  </p>
                  <p className="ant-upload-hint">
                    请选择包含 plugin.json 配置文件的插件目录
                  </p>
                </Dragger>
                {localPath && (
                  <Alert
                    message={`已选择: ${localPath}`}
                    type="success"
                    style={{ marginTop: 16 }}
                  />
                )}
              </div>
            );

          case 'url':
            return (
              <div>
                <Title level={4}>输入插件URL</Title>
                <Input
                  placeholder="请输入插件的下载URL"
                  value={urlPath}
                  onChange={(e) => setUrlPath(e.target.value)}
                  prefix={<LinkOutlined />}
                  size="large"
                />
                <div style={{ marginTop: 16 }}>
                  <Alert
                    message="支持的URL格式"
                    description={
                      <ul>
                        <li>HTTP/HTTPS 直链下载</li>
                        <li>GitHub Releases</li>
                        <li>GitLab Releases</li>
                        <li>其他符合标准的下载链接</li>
                      </ul>
                    }
                    type="info"
                    showIcon
                  />
                </div>
              </div>
            );

          case 'market':
            return (
              <div>
                <Title level={4}>从市场选择插件</Title>
                <Alert
                  message="插件市场功能开发中"
                  description="目前您可以通过插件市场浏览页选择要安装的插件"
                  type="info"
                />
              </div>
            );
        }

      case 2:
      case 3:
      case 4:
        return (
          <div style={{ textAlign: 'center', padding: '20px' }}>
            {currentStep < 4 ? (
              <LoadingOutlined style={{ fontSize: 48, marginBottom: 16 }} />
            ) : (
              <CheckCircleOutlined
                style={{ fontSize: 48, marginBottom: 16, color: '#52c41a' }}
              />
            )}
            <div>
              <Title level={4}>{installationSteps[currentStep]?.title}</Title>
              <Text type="secondary">
                {installationSteps[currentStep]?.description}
              </Text>
            </div>
          </div>
        );

      default:
        return null;
    }
  }, [
    currentStep,
    sourceType,
    localUploadProps,
    localPath,
    urlPath,
    installationSteps,
    validationErrors,
  ]);

  return (
    <Modal
      title={
        <Space>
          <UploadOutlined />
          <span>安装插件</span>
        </Space>
      }
      open={visible}
      onCancel={onClose}
      width={600}
      footer={null}
      destroyOnHidden
      afterClose={resetState}
    >
      <Steps current={currentStep} items={installationSteps} size="small" />

      <div style={{ margin: '24px 0' }}>
        <Progress
          percent={progress}
          status={
            installationSteps[currentStep]?.status === 'error'
              ? 'exception'
              : 'active'
          }
          style={{ marginBottom: 16 }}
        />
        {renderStepContent()}
      </div>

      {installedPlugin && (
        <>
          <Divider />
          <PluginInfo />
          <Alert
            message="安装成功"
            description="插件已成功安装并可以使用。您可以继续安装其他插件或关闭此窗口。"
            type="success"
            action={
              <Button size="small" onClick={onComplete}>
                完成
              </Button>
            }
            showIcon
            style={{ marginTop: 16 }}
          />
        </>
      )}

      {!loading && currentStep === 0 && (
        <div style={{ marginTop: 16, textAlign: 'right' }}>
          <Space>
            <Button onClick={onClose}>取消</Button>
            <Button
              type="primary"
              onClick={startInstallation}
              disabled={validationErrors.length > 0}
            >
              开始安装
            </Button>
          </Space>
        </div>
      )}
    </Modal>
  );
};

export default PluginInstaller;
