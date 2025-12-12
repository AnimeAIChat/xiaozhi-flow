import {
  AppstoreOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  CloudOutlined,
  DeleteOutlined,
  DownloadOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  ExportOutlined,
  EyeOutlined,
  FilterOutlined,
  ImportOutlined,
  MoreOutlined,
  PlayCircleOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
  SettingOutlined,
  StopOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import {
  Alert,
  Avatar,
  Badge,
  Button,
  Card,
  Col,
  Divider,
  Dropdown,
  Form,
  Input,
  List,
  Menu,
  Modal,
  message,
  Progress,
  Row,
  Space,
  Statistic,
  Switch,
  Table,
  Tabs,
  Tag,
  Tooltip,
  Typography,
  Upload,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type React from 'react';
import { useCallback, useEffect, useState } from 'react';
import { backendServiceManager } from '../../plugins/core/BackendServiceManager';
import { pluginInstaller } from '../../plugins/core/PluginInstaller';
import { pluginManager } from '../../plugins/core/PluginManager';
import {
  type IPlugin,
  PluginConfig,
  PluginStatus,
  ServiceInfo,
} from '../../plugins/types';
import {
  useInstallationProgress,
  useInstalledPlugins,
  usePluginManagerState,
} from '../../stores/useAppStore';
import { PluginInstaller } from './PluginInstaller';
import { PluginList } from './PluginList';
import { PluginMarket } from './PluginMarket';

const { Title, Text, Paragraph } = Typography;

interface PluginManagerProps {
  visible: boolean;
  onClose: () => void;
}

export const PluginManager: React.FC<PluginManagerProps> = ({
  visible,
  onClose,
}) => {
  const plugins = useInstalledPlugins();
  const managerState = usePluginManagerState();
  const installationProgress = useInstallationProgress();

  const [activeTab, setActiveTab] = useState('installed');
  const [selectedPlugin, setSelectedPlugin] = useState<IPlugin | null>(null);
  const [installerVisible, setInstallerVisible] = useState(false);
  const [marketVisible, setMarketVisible] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [filterCategory, setFilterCategory] = useState<string>('');
  const [loading, setLoading] = useState(false);

  // 刷新插件列表
  const refreshPlugins = useCallback(async () => {
    setLoading(true);
    try {
      // 这里可以调用插件管理器加载插件
      console.log('Refreshing plugins...');
    } catch (error) {
      message.error('刷新插件列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  // 安装插件
  const handleInstallPlugin = useCallback((source: any) => {
    setSelectedPlugin(null);
    setInstallerVisible(true);
  }, []);

  // 卸载插件
  const handleUninstallPlugin = useCallback(
    async (plugin: IPlugin) => {
      Modal.confirm({
        title: '卸载插件',
        content: `确定要卸载插件 "${plugin.name}" 吗？这将删除所有相关文件和配置。`,
        okText: '卸载',
        okType: 'danger',
        cancelText: '取消',
        onOk: async () => {
          try {
            await pluginInstaller.uninstall(plugin.id);
            message.success(`插件 ${plugin.name} 卸载成功`);
            await refreshPlugins();
          } catch (error) {
            message.error(`卸载失败: ${error}`);
          }
        },
      });
    },
    [refreshPlugins],
  );

  // 启用/禁用插件
  const handleTogglePlugin = useCallback(
    async (plugin: IPlugin, enabled: boolean) => {
      try {
        if (enabled) {
          await pluginManager.activatePlugin(plugin.id);
          message.success(`插件 ${plugin.name} 已启用`);
        } else {
          await pluginManager.deactivatePlugin(plugin.id);
          message.success(`插件 ${plugin.name} 已禁用`);
        }
        await refreshPlugins();
      } catch (error) {
        message.error(`操作失败: ${error}`);
      }
    },
    [refreshPlugins],
  );

  // 重启插件服务
  const handleRestartPlugin = useCallback(async (plugin: IPlugin) => {
    try {
      if (plugin.backend) {
        const services = backendServiceManager.getPluginServices(plugin.id);
        for (const service of services) {
          await backendServiceManager.restartService(plugin.id, service.id);
        }
        message.success(`插件 ${plugin.name} 服务已重启`);
      }
    } catch (error) {
      message.error(`重启失败: ${error}`);
    }
  }, []);

  // 查看插件详情
  const handleViewPlugin = useCallback((plugin: IPlugin) => {
    setSelectedPlugin(plugin);
  }, []);

  // 安装进度显示
  const InstallationProgress = useCallback(() => {
    if (!installationProgress.inProgress) return null;

    return (
      <Card title="安装进度" size="small" style={{ marginBottom: 16 }}>
        <div style={{ marginBottom: 8 }}>
          <Text strong>
            {installationProgress.progress?.stage || '安装中...'}
          </Text>
        </div>
        <Progress
          percent={installationProgress.progress?.progress || 0}
          status="active"
          size="small"
        />
        <div style={{ marginTop: 8, fontSize: '12px', color: '#666' }}>
          {installationProgress.progress?.message}
        </div>
      </Card>
    );
  }, [installationProgress]);

  // 安装插件统计
  const PluginStats = useCallback(() => {
    const stats = {
      total: Object.keys(plugins).length,
      active: Object.values(plugins).filter((p) => p.enabled).length,
      running: 0, // 需要从后端服务管理器获取
      withBackend: Object.values(plugins).filter((p) => p.backend).length,
    };

    return (
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总插件数"
              value={stats.total}
              prefix={<AppstoreOutlined />}
              styles={{ content: { color: '#1890ff' } }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃插件"
              value={stats.active}
              prefix={<CheckCircleOutlined />}
              styles={{ content: { color: '#52c41a' } }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="运行中"
              value={stats.running}
              prefix={<PlayCircleOutlined />}
              styles={{ content: { color: '#faad14' } }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="后端服务"
              value={stats.withBackend}
              prefix={<CloudOutlined />}
              styles={{ content: { color: '#722ed1' } }}
            />
          </Card>
        </Col>
      </Row>
    );
  }, [plugins]);

  // 操作菜单
  const ActionMenu = useCallback(
    ({ plugin }: { plugin: IPlugin }) => (
      <Menu>
        <Menu.Item
          key="view"
          icon={<EyeOutlined />}
          onClick={() => handleViewPlugin(plugin)}
        >
          查看详情
        </Menu.Item>
        <Menu.Item
          key="restart"
          icon={<ReloadOutlined />}
          onClick={() => handleRestartPlugin(plugin)}
          disabled={!plugin.backend}
        >
          重启服务
        </Menu.Item>
        <Menu.Divider />
        <Menu.Item
          key="uninstall"
          icon={<DeleteOutlined />}
          danger
          onClick={() => handleUninstallPlugin(plugin)}
        >
          卸载插件
        </Menu.Item>
      </Menu>
    ),
    [handleViewPlugin, handleRestartPlugin, handleUninstallPlugin],
  );

  return (
    <Modal
      title={
        <Space>
          <AppstoreOutlined />
          <span>插件管理器</span>
          {installationProgress.inProgress && (
            <Badge status="processing" text="安装中" />
          )}
        </Space>
      }
      open={visible}
      onCancel={onClose}
      width={1200}
      footer={null}
      destroyOnHidden
    >
      {/* 安装进度 */}
      <InstallationProgress />

      {/* 统计信息 */}
      <PluginStats />

      {/* 主要内容 */}
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        tabBarExtraContent={
          <Space>
            <Button icon={<SearchOutlined />} size="small">
              搜索
            </Button>
            <Button icon={<FilterOutlined />} size="small">
              筛选
            </Button>
            <Button
              icon={<PlusOutlined />}
              type="primary"
              onClick={() => setInstallerVisible(true)}
            >
              安装插件
            </Button>
            <Button
              icon={<CloudOutlined />}
              onClick={() => setMarketVisible(true)}
            >
              插件市场
            </Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={refreshPlugins}
              loading={loading}
            >
              刷新
            </Button>
          </Space>
        }
        items={[
          {
            key: 'installed',
            label: '已安装插件',
            children: (
              <PluginList
                onInstall={handleInstallPlugin}
                onToggle={handleTogglePlugin}
                onUninstall={handleUninstallPlugin}
                onView={handleViewPlugin}
              />
            ),
          },
          {
            key: 'market',
            label: '插件市场',
            children: (
              <PluginMarket
                onInstall={(source) => {
                  handleInstallPlugin(source);
                  setActiveTab('installed');
                }}
              />
            ),
          },
          {
            key: 'config',
            label: '插件配置',
            children: (
              <div style={{ textAlign: 'center', padding: '40px' }}>
                <Text type="secondary">插件配置功能开发中...</Text>
              </div>
            ),
          },
          {
            key: 'services',
            label: '服务状态',
            children: (
              <div style={{ textAlign: 'center', padding: '40px' }}>
                <Text type="secondary">服务状态监控功能开发中...</Text>
              </div>
            ),
          },
        ]}
      />

      {/* 插件安装器 */}
      <PluginInstaller
        visible={installerVisible}
        onClose={() => setInstallerVisible(false)}
        onComplete={() => {
          setInstallerVisible(false);
          setActiveTab('installed');
          refreshPlugins();
        }}
      />

      {/* 插件市场 */}
      <PluginMarket
        visible={marketVisible}
        onClose={() => setMarketVisible(false)}
      />

      {/* 插件详情弹窗 */}
      <Modal
        title="插件详情"
        open={!!selectedPlugin}
        onCancel={() => setSelectedPlugin(null)}
        footer={[
          <Button key="close" onClick={() => setSelectedPlugin(null)}>
            关闭
          </Button>,
        ]}
        width={600}
      >
        {selectedPlugin && (
          <div>
            <Row gutter={16}>
              <Col span={8}>
                <div style={{ textAlign: 'center' }}>
                  <div
                    style={{
                      width: 80,
                      height: 80,
                      backgroundColor:
                        selectedPlugin.metadata.color || '#1890ff',
                      borderRadius: 8,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      margin: '0 auto 16px',
                    }}
                  >
                    {selectedPlugin.metadata.icon || (
                      <AppstoreOutlined
                        style={{ fontSize: 32, color: 'white' }}
                      />
                    )}
                  </div>
                  <Tag
                    color={selectedPlugin.type === 'backend' ? 'red' : 'blue'}
                  >
                    {selectedPlugin.type}
                  </Tag>
                </div>
              </Col>
              <Col span={16}>
                <Title level={4}>{selectedPlugin.name}</Title>
                <Paragraph type="secondary">
                  {selectedPlugin.description}
                </Paragraph>
                <Space wrap>
                  <Tag>版本 {selectedPlugin.version}</Tag>
                  <Tag
                    color={
                      selectedPlugin.runtime === 'python' ? 'green' : 'orange'
                    }
                  >
                    {selectedPlugin.runtime}
                  </Tag>
                  <Tag>作者: {selectedPlugin.author}</Tag>
                </Space>
              </Col>
            </Row>

            <Divider />

            <Row gutter={16}>
              <Col span={12}>
                <Text strong>节点定义:</Text>
                <div style={{ marginTop: 8 }}>
                  <Tag color="blue">
                    {selectedPlugin.nodeDefinition.displayName}
                  </Tag>
                  <Text type="secondary" style={{ marginLeft: 8 }}>
                    {selectedPlugin.nodeDefinition.description}
                  </Text>
                </div>
              </Col>
              <Col span={12}>
                <Text strong>参数数量:</Text>
                <div style={{ marginTop: 8 }}>
                  <Badge
                    count={selectedPlugin.nodeDefinition.parameters.length}
                  />
                  <Text type="secondary" style={{ marginLeft: 8 }}>
                    个参数
                  </Text>
                </div>
              </Col>
            </Row>

            {selectedPlugin.backend && (
              <>
                <Divider />
                <Text strong>后端配置:</Text>
                <div style={{ marginTop: 8 }}>
                  <Space direction="vertical" size="small">
                    <div>
                      <Text type="secondary">入口点: </Text>
                      <Text code>{selectedPlugin.backend.entryPoint}</Text>
                    </div>
                    <div>
                      <Text type="secondary">端口: </Text>
                      <Text code>
                        {selectedPlugin.backend.port || '自动分配'}
                      </Text>
                    </div>
                    {selectedPlugin.backend.dependencies && (
                      <div>
                        <Text type="secondary">依赖: </Text>
                        <Space wrap size="small">
                          {selectedPlugin.backend.dependencies.map(
                            (dep, index) => (
                              <Tag key={index}>{dep}</Tag>
                            ),
                          )}
                        </Space>
                      </div>
                    )}
                  </Space>
                </div>
              </>
            )}
          </div>
        )}
      </Modal>
    </Modal>
  );
};

export default PluginManager;
