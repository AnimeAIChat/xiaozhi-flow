import {
  ApiOutlined,
  AppstoreOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  CloudOutlined,
  DatabaseOutlined,
  DeleteOutlined,
  DownloadOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  ExportOutlined,
  EyeOutlined,
  FilterOutlined,
  GlobalOutlined,
  ImportOutlined,
  MoreOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  SearchOutlined,
  SettingOutlined,
  StopOutlined,
  ToolOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import {
  Avatar,
  Badge,
  Button,
  Card,
  Col,
  Divider,
  Dropdown,
  Empty,
  Input,
  Menu,
  Modal,
  message,
  Progress,
  Row,
  Select,
  Space,
  Statistic,
  Switch,
  Table,
  Tag,
  Tooltip,
  Typography,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type React from 'react';
import { useCallback, useMemo, useState } from 'react';
import { backendServiceManager } from '../../plugins/core/BackendServiceManager';
import type { IPlugin, PluginStatus, ServiceInfo } from '../../plugins/types';
import {
  usePluginList,
  usePluginServices,
  usePluginStatuses,
} from '../../stores/useAppStore';

const { Title, Text, Paragraph } = Typography;
const { Search } = Input;
const { Option } = Select;

interface PluginListProps {
  onInstall?: () => void;
  onToggle?: (plugin: IPlugin, enabled: boolean) => void;
  onUninstall?: (plugin: IPlugin) => void;
  onView?: (plugin: IPlugin) => void;
}

export const PluginList: React.FC<PluginListProps> = ({
  onInstall,
  onToggle,
  onUninstall,
  onView,
}) => {
  const installedPlugins = usePluginList();
  const pluginStatuses = usePluginStatuses();
  const pluginServices = usePluginServices();

  const [searchText, setSearchText] = useState('');
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [filterType, setFilterType] = useState<string>('all');
  const [filterCategory, setFilterCategory] = useState<string>('all');

  // 获取插件状态
  const getPluginStatus = useCallback(
    (pluginId: string): PluginStatus | undefined => {
      return pluginStatuses[pluginId];
    },
    [pluginStatuses],
  );

  // 获取插件服务
  const getPluginServices = useCallback(
    (pluginId: string): ServiceInfo[] => {
      return Object.values(pluginServices).filter(
        (service) => service.pluginId === pluginId,
      );
    },
    [pluginServices],
  );

  // 过滤插件
  const filteredPlugins = useMemo(() => {
    return Object.values(installedPlugins).filter((plugin) => {
      // 文本搜索
      if (searchText) {
        const searchLower = searchText.toLowerCase();
        const matchesSearch =
          plugin.name.toLowerCase().includes(searchLower) ||
          plugin.description.toLowerCase().includes(searchLower) ||
          plugin.author.toLowerCase().includes(searchLower) ||
          plugin.metadata.tags.some((tag) =>
            tag.toLowerCase().includes(searchLower),
          );

        if (!matchesSearch) return false;
      }

      // 状态过滤
      if (filterStatus !== 'all') {
        const status = getPluginStatus(plugin.id);
        if (
          filterStatus === 'active' &&
          (!status || status.status !== 'active')
        )
          return false;
        if (
          filterStatus === 'inactive' &&
          (!status || status.status !== 'inactive')
        )
          return false;
        if (filterStatus === 'error' && (!status || status.status !== 'error'))
          return false;
      }

      // 类型过滤
      if (filterType !== 'all' && plugin.type !== filterType) {
        return false;
      }

      // 分类过滤
      if (
        filterCategory !== 'all' &&
        plugin.metadata.category !== filterCategory
      ) {
        return false;
      }

      return true;
    });
  }, [
    installedPlugins,
    searchText,
    filterStatus,
    filterType,
    filterCategory,
    getPluginStatus,
  ]);

  // 获取插件图标
  const getPluginIcon = useCallback((plugin: IPlugin) => {
    const iconMap: Record<string, React.ReactNode> = {
      AI: <ApiOutlined style={{ color: '#1890ff' }} />,
      LLM: <ApiOutlined style={{ color: '#722ed1' }} />,
      ASR: <GlobalOutlined style={{ color: '#fa8c16' }} />,
      TTS: <GlobalOutlined style={{ color: '#52c41a' }} />,
      Database: <DatabaseOutlined style={{ color: '#13c2c2' }} />,
      Tool: <ToolOutlined style={{ color: '#faad14' }} />,
      Service: <CloudOutlined style={{ color: '#1890ff' }} />,
    };

    return (
      iconMap[plugin.metadata.category] ||
      plugin.metadata.icon || <AppstoreOutlined />
    );
  }, []);

  // 获取状态标签
  const getStatusTag = useCallback(
    (plugin: IPlugin) => {
      const status = getPluginStatus(plugin.id);
      if (!status) {
        return <Tag color="default">未知</Tag>;
      }

      switch (status.status) {
        case 'active':
          return (
            <Tag color="success" icon={<CheckCircleOutlined />}>
              运行中
            </Tag>
          );
        case 'inactive':
          return (
            <Tag color="default" icon={<CloseCircleOutlined />}>
              已停止
            </Tag>
          );
        case 'loading':
        case 'activating':
          return (
            <Tag color="processing" icon={<ReloadOutlined />}>
              启动中
            </Tag>
          );
        case 'deactivating':
          return (
            <Tag color="processing" icon={<ReloadOutlined />}>
              停止中
            </Tag>
          );
        case 'error':
          return (
            <Tag color="error" icon={<ExclamationCircleOutlined />}>
              错误
            </Tag>
          );
        default:
          return <Tag color="default">{status.status}</Tag>;
      }
    },
    [getPluginStatus],
  );

  // 获取服务状态
  const getServiceStatus = useCallback(
    (plugin: IPlugin) => {
      const services = getPluginServices(plugin.id);
      if (!plugin.backend || services.length === 0) {
        return null;
      }

      const runningServices = services.filter(
        (service) => service.status === 'running',
      );
      const totalServices = services.length;

      return (
        <Space size="small">
          <Tag color="blue">
            {runningServices.length}/{totalServices} 服务
          </Tag>
          {runningServices.length === totalServices &&
            runningServices.length > 0 && (
              <Tag color="green" icon={<CheckCircleOutlined />}>
                全部运行
              </Tag>
            )}
          {runningServices.length === 0 && totalServices > 0 && (
            <Tag color="red" icon={<CloseCircleOutlined />}>
              全部停止
            </Tag>
          )}
          {runningServices.length > 0 &&
            runningServices.length < totalServices && (
              <Tag color="orange" icon={<ExclamationCircleOutlined />}>
                部分运行
              </Tag>
            )}
        </Space>
      );
    },
    [getPluginServices],
  );

  // 操作菜单
  const ActionMenu = useCallback(
    ({ plugin }: { plugin: IPlugin }) => (
      <Menu>
        <Menu.Item
          key="view"
          icon={<EyeOutlined />}
          onClick={() => onView?.(plugin)}
        >
          查看详情
        </Menu.Item>
        <Menu.Item
          key="restart"
          icon={<ReloadOutlined />}
          disabled={!plugin.backend}
        >
          重启服务
        </Menu.Item>
        <Menu.Item key="export" icon={<ExportOutlined />}>
          导出配置
        </Menu.Item>
        <Menu.Divider />
        <Menu.Item
          key="uninstall"
          icon={<DeleteOutlined />}
          danger
          onClick={() => onUninstall?.(plugin)}
        >
          卸载插件
        </Menu.Item>
      </Menu>
    ),
    [onView, onUninstall],
  );

  // 表格列定义
  const columns: ColumnsType<IPlugin> = [
    {
      title: '插件',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      render: (name: string, record: IPlugin) => (
        <Space>
          <Avatar
            size="small"
            icon={getPluginIcon(record)}
            style={{ backgroundColor: record.metadata.color }}
          />
          <div>
            <div style={{ fontWeight: 500 }}>{name}</div>
            <Text type="secondary" style={{ fontSize: '12px' }}>
              v{record.version} · {record.author}
            </Text>
          </div>
        </Space>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 80,
      render: (type: string, record: IPlugin) => (
        <Tag color={type === 'backend' ? 'red' : 'blue'}>{type}</Tag>
      ),
    },
    {
      title: '运行时',
      dataIndex: 'runtime',
      key: 'runtime',
      width: 80,
      render: (runtime: string) => (
        <Tag
          color={
            runtime === 'python'
              ? 'green'
              : runtime === 'go'
                ? 'orange'
                : 'blue'
          }
        >
          {runtime}
        </Tag>
      ),
    },
    {
      title: '分类',
      dataIndex: ['metadata', 'category'],
      key: 'category',
      width: 100,
      render: (category: string, record: IPlugin) => (
        <Space direction="vertical" size="small">
          <Tag color={record.metadata.color}>{category}</Tag>
          {record.nodeDefinition && (
            <Tag color="blue">{record.nodeDefinition.displayName}</Tag>
          )}
        </Space>
      ),
    },
    {
      title: '状态',
      key: 'status',
      width: 120,
      render: (_, record: IPlugin) => (
        <Space direction="vertical" size="small">
          {getStatusTag(record)}
          {getServiceStatus(record)}
        </Space>
      ),
    },
    {
      title: '启用',
      key: 'enabled',
      width: 80,
      render: (_, record: IPlugin) => {
        const status = getPluginStatus(record.id);
        const isEnabled = status?.status === 'active';

        return (
          <Switch
            checked={isEnabled}
            onChange={(checked) => onToggle?.(record, checked)}
            loading={
              status?.status === 'loading' || status?.status === 'activating'
            }
          />
        );
      },
    },
    {
      title: '操作',
      key: 'actions',
      width: 100,
      render: (_, record: IPlugin) => (
        <Dropdown
          menu={{
            items: [
              {
                key: 'view',
                icon: <EyeOutlined />,
                label: '查看详情',
                onClick: () => onView?.(record),
              },
              {
                key: 'restart',
                icon: <ReloadOutlined />,
                label: '重启服务',
                disabled: !record.backend,
              },
              {
                key: 'export',
                icon: <ExportOutlined />,
                label: '导出配置',
              },
              {
                type: 'divider',
              },
              {
                key: 'uninstall',
                icon: <DeleteOutlined />,
                label: '卸载插件',
                danger: true,
                onClick: () => onUninstall?.(record),
              },
            ],
          }}
          trigger={['click']}
        >
          <Button size="small" icon={<MoreOutlined />} />
        </Dropdown>
      ),
    },
  ];

  // 统计信息
  const PluginStats = useCallback(() => {
    const stats = {
      total: filteredPlugins.length,
      active: filteredPlugins.filter((plugin) => {
        const status = getPluginStatus(plugin.id);
        return status?.status === 'active';
      }).length,
      withBackend: filteredPlugins.filter((plugin) => plugin.backend).length,
      errors: filteredPlugins.filter((plugin) => {
        const status = getPluginStatus(plugin.id);
        return status?.status === 'error';
      }).length,
    };

    return (
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card size="small">
            <Statistic
              title="总数"
              value={stats.total}
              valueStyle={{ fontSize: '16px' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <Statistic
              title="运行中"
              value={stats.active}
              valueStyle={{ fontSize: '16px', color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <Statistic
              title="后端服务"
              value={stats.withBackend}
              valueStyle={{ fontSize: '16px', color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <Statistic
              title="错误"
              value={stats.errors}
              valueStyle={{ fontSize: '16px', color: '#ff4d4f' }}
            />
          </Card>
        </Col>
      </Row>
    );
  }, [filteredPlugins, getPluginStatus]);

  // 筛选控件
  const FilterControls = useCallback(
    () => (
      <Space wrap>
        <Search
          placeholder="搜索插件名称、描述或作者"
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          style={{ width: 250 }}
          allowClear
        />
        <Select
          placeholder="状态"
          value={filterStatus}
          onChange={setFilterStatus}
          style={{ width: 120 }}
        >
          <Option value="all">全部状态</Option>
          <Option value="active">运行中</Option>
          <Option value="inactive">已停止</Option>
          <Option value="error">错误</Option>
        </Select>
        <Select
          placeholder="类型"
          value={filterType}
          onChange={setFilterType}
          style={{ width: 120 }}
        >
          <Option value="all">全部类型</Option>
          <Option value="frontend">前端</Option>
          <Option value="backend">后端</Option>
          <Option value="fullstack">全栈</Option>
        </Select>
        <Select
          placeholder="分类"
          value={filterCategory}
          onChange={setFilterCategory}
          style={{ width: 140 }}
        >
          <Option value="all">全部分类</Option>
          <Option value="AI">AI</Option>
          <Option value="LLM">大语言模型</Option>
          <Option value="ASR">语音识别</Option>
          <Option value="TTS">语音合成</Option>
          <Option value="Database">数据库</Option>
          <Option value="Tool">工具</Option>
        </Select>
        <Button icon={<FilterOutlined />}>高级筛选</Button>
      </Space>
    ),
    [searchText, filterStatus, filterType, filterCategory],
  );

  return (
    <div>
      {/* 统计信息 */}
      <PluginStats />

      {/* 筛选控件 */}
      <div style={{ marginBottom: 16 }}>
        <FilterControls />
      </div>

      {/* 插件列表 */}
      {filteredPlugins.length > 0 ? (
        <Table
          columns={columns}
          dataSource={filteredPlugins}
          rowKey="id"
          size="small"
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 个插件`,
            showQuickJumper: true,
          }}
          scroll={{ x: 1000 }}
        />
      ) : (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="没有找到匹配的插件"
          style={{ padding: '40px 0' }}
        >
          <Button type="primary" onClick={onInstall}>
            安装第一个插件
          </Button>
        </Empty>
      )}
    </div>
  );
};

export default PluginList;
