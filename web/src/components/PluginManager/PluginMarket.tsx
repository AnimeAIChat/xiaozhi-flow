import {
  ApiOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloudOutlined,
  CloudUploadOutlined,
  DatabaseOutlined,
  DownloadOutlined,
  EyeOutlined,
  FilterOutlined,
  GlobalOutlined,
  HeartOutlined,
  MessageOutlined,
  RobotOutlined,
  SearchOutlined,
  ShareAltOutlined,
  SortAscendingOutlined,
  StarOutlined,
  ToolOutlined,
} from '@ant-design/icons';
import {
  Avatar,
  Badge,
  Button,
  Card,
  Col,
  Divider,
  Empty,
  Input,
  List,
  Modal,
  message,
  Progress,
  Rate,
  Row,
  Select,
  Space,
  Statistic,
  Tabs,
  Tag,
  Tooltip,
  Typography,
} from 'antd';
import type React from 'react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import type { IPlugin, PluginSource } from '../../plugins/types';

const { Title, Text, Paragraph } = Typography;
const { Search } = Input;
const { Option } = Select;
const { TabPane } = Tabs;

interface PluginMarketProps {
  visible: boolean;
  onClose: () => void;
  onInstall?: (source: PluginSource) => void;
}

interface MarketplacePlugin {
  id: string;
  name: string;
  description: string;
  version: string;
  author: string;
  downloads: number;
  rating: number;
  reviews: number;
  tags: string[];
  category: string;
  runtime: string;
  icon: React.ReactNode;
  screenshots: string[];
  lastUpdated: Date;
  homepage?: string;
  repository?: string;
  license: string;
  fileUrl: string;
  size: string;
  dependencies: string[];
}

// 模拟市场插件数据
const mockMarketplacePlugins: MarketplacePlugin[] = [
  {
    id: 'openai-connector',
    name: 'OpenAI Connector',
    description: '官方OpenAI API连接器，支持GPT-3.5、GPT-4等模型',
    version: '2.1.0',
    author: 'Xiaozhi Team',
    downloads: 15420,
    rating: 4.8,
    reviews: 328,
    tags: ['AI', 'LLM', 'OpenAI', 'GPT'],
    category: 'LLM',
    runtime: 'python',
    icon: <RobotOutlined />,
    screenshots: [],
    lastUpdated: new Date('2024-01-15'),
    homepage: 'https://openai.com',
    repository: 'https://github.com/xiaozhi-flow/openai-connector',
    license: 'MIT',
    fileUrl: 'https://example.com/plugins/openai-connector.zip',
    size: '2.3MB',
    dependencies: ['openai>=1.0.0', 'requests>=2.0.0'],
  },
  {
    id: 'claude-api',
    name: 'Claude API',
    description: 'Anthropic Claude API连接器，支持Claude-2、Claude-3模型',
    version: '1.5.0',
    author: 'AI Solutions',
    downloads: 8760,
    rating: 4.7,
    reviews: 156,
    tags: ['AI', 'LLM', 'Anthropic', 'Claude'],
    category: 'LLM',
    runtime: 'python',
    icon: <ApiOutlined />,
    screenshots: [],
    lastUpdated: new Date('2024-01-20'),
    repository: 'https://github.com/ai-solutions/claude-api',
    license: 'Apache-2.0',
    fileUrl: 'https://example.com/plugins/claude-api.zip',
    size: '1.8MB',
    dependencies: ['anthropic>=0.8.0', 'asyncio>=3.0.0'],
  },
  {
    id: 'speech-recognition',
    name: 'Speech Recognition',
    description: '多语言语音识别插件，支持中文、英文等多种语言',
    version: '3.0.0',
    author: 'VoiceTech',
    downloads: 12350,
    rating: 4.6,
    reviews: 89,
    tags: ['ASR', '语音识别', '多语言', 'Whisper'],
    category: 'ASR',
    runtime: 'python',
    icon: <GlobalOutlined />,
    screenshots: [],
    lastUpdated: new Date('2024-01-18'),
    repository: 'https://github.com/voicetech/speech-recognition',
    license: 'MIT',
    fileUrl: 'https://example.com/plugins/speech-recognition.zip',
    size: '15.2MB',
    dependencies: ['whisper>=1.0.0', 'torch>=2.0.0', 'numpy>=1.24.0'],
  },
  {
    id: 'text-to-speech',
    name: 'Text to Speech',
    description: '高质量文本转语音插件，支持多种声音和语言',
    version: '2.2.0',
    author: 'TTS Pro',
    downloads: 9230,
    rating: 4.5,
    reviews: 67,
    tags: ['TTS', '语音合成', '多声音', 'Edge-TTS'],
    category: 'TTS',
    runtime: 'python',
    icon: <MessageOutlined />,
    screenshots: [],
    lastUpdated: new Date('2024-01-12'),
    repository: 'https://github.com/ttspro/text-to-speech',
    license: 'MIT',
    fileUrl: 'https://example.com/plugins/text-to-speech.zip',
    size: '4.1MB',
    dependencies: ['edge-tts>=6.0.0', 'pydub>=0.25.0'],
  },
  {
    id: 'redis-connector',
    name: 'Redis Connector',
    description: 'Redis数据库连接器，支持缓存、队列、发布订阅等功能',
    version: '1.8.0',
    author: 'CacheMaster',
    downloads: 6780,
    rating: 4.4,
    reviews: 45,
    tags: ['数据库', 'Redis', '缓存', '队列'],
    category: 'Database',
    runtime: 'python',
    icon: <DatabaseOutlined />,
    screenshots: [],
    lastUpdated: new Date('2024-01-10'),
    repository: 'https://github.com/cachemaster/redis-connector',
    license: 'BSD-3-Clause',
    fileUrl: 'https://example.com/plugins/redis-connector.zip',
    size: '1.2MB',
    dependencies: ['redis>=4.5.0', 'hiredis>=2.0.0'],
  },
  {
    id: 'custom-llm-server',
    name: 'Custom LLM Server',
    description: '自托管大语言模型服务端，支持Hugging Face模型',
    version: '1.3.0',
    author: 'LLM Foundation',
    downloads: 4530,
    rating: 4.3,
    reviews: 28,
    tags: ['LLM', '自托管', 'Hugging Face', 'Transformers'],
    category: 'LLM',
    runtime: 'python',
    icon: <CloudUploadOutlined />,
    screenshots: [],
    lastUpdated: new Date('2024-01-08'),
    repository: 'https://github.com/llm-foundation/custom-llm-server',
    license: 'Apache-2.0',
    fileUrl: 'https://example.com/plugins/custom-llm-server.zip',
    size: '8.7MB',
    dependencies: [
      'transformers>=4.30.0',
      'fastapi>=0.100.0',
      'uvicorn>=0.20.0',
    ],
  },
  {
    id: 'data-processor',
    name: 'Data Processor',
    description: '数据处理工具集，包含数据清洗、转换、验证等功能',
    version: '2.0.0',
    author: 'DataTools',
    downloads: 3150,
    rating: 4.2,
    reviews: 19,
    tags: ['工具', '数据处理', 'ETL', '验证'],
    category: 'Tool',
    runtime: 'python',
    icon: <ToolOutlined />,
    screenshots: [],
    lastUpdated: new Date('2024-01-05'),
    repository: 'https://github.com/datatools/data-processor',
    license: 'MIT',
    fileUrl: 'https://example.com/plugins/data-processor.zip',
    size: '2.8MB',
    dependencies: ['pandas>=2.0.0', 'numpy>=1.24.0', 'pydantic>=2.0.0'],
  },
];

export const PluginMarket: React.FC<PluginMarketProps> = ({
  visible,
  onClose,
  onInstall,
}) => {
  const [plugins, setPlugins] = useState<MarketplacePlugin[]>(
    mockMarketplacePlugins,
  );
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [sortBy, setSortBy] = useState<'popularity' | 'rating' | 'updated'>(
    'popularity',
  );
  const [selectedPlugin, setSelectedPlugin] =
    useState<MarketplacePlugin | null>(null);
  const [installingPlugin, setInstallingPlugin] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('browse');

  // 分类列表
  const categories = useMemo(() => {
    const cats = Array.from(new Set(plugins.map((p) => p.category)));
    return ['all', ...cats];
  }, [plugins]);

  // 排序插件
  const sortedPlugins = useMemo(() => {
    const filtered = plugins.filter((plugin) => {
      // 搜索过滤
      if (searchText) {
        const searchLower = searchText.toLowerCase();
        return (
          plugin.name.toLowerCase().includes(searchLower) ||
          plugin.description.toLowerCase().includes(searchLower) ||
          plugin.tags.some((tag) => tag.toLowerCase().includes(searchLower))
        );
      }

      // 分类过滤
      if (selectedCategory !== 'all' && plugin.category !== selectedCategory) {
        return false;
      }

      return true;
    });

    // 排序
    return [...filtered].sort((a, b) => {
      switch (sortBy) {
        case 'popularity':
          return b.downloads - a.downloads;
        case 'rating':
          return b.rating - a.rating;
        case 'updated':
          return b.lastUpdated.getTime() - a.lastUpdated.getTime();
        default:
          return 0;
      }
    });
  }, [plugins, searchText, selectedCategory, sortBy]);

  // 获取分类插件
  const getCategoryPlugins = useCallback(
    (category: string) => {
      return category === 'all'
        ? plugins
        : plugins.filter((p) => p.category === category);
    },
    [plugins],
  );

  // 获取推荐插件
  const getFeaturedPlugins = useCallback(() => {
    return [...plugins]
      .sort((a, b) => b.rating * b.reviews - a.rating * a.reviews)
      .slice(0, 6);
  }, [plugins]);

  // 安装插件
  const handleInstall = useCallback(
    async (plugin: MarketplacePlugin) => {
      setInstallingPlugin(plugin.id);

      try {
        // 模拟下载过程
        const source: PluginSource = {
          type: 'url',
          url: plugin.fileUrl,
          options: {
            autoStart: true,
            overwrite: true,
          },
        };

        // 这里可以显示下载进度
        message.loading(`正在下载 ${plugin.name}...`);

        // 模拟下载延迟
        await new Promise((resolve) => setTimeout(resolve, 2000));

        message.destroy();
        message.success(`${plugin.name} 安装成功！`);

        onInstall?.(source);
        onClose?.();
      } catch (error) {
        message.error(`安装失败: ${error}`);
      } finally {
        setInstallingPlugin(null);
      }
    },
    [onInstall, onClose],
  );

  // 渲染插件卡片
  const PluginCard = useCallback(
    ({ plugin }: { plugin: MarketplacePlugin }) => (
      <Card
        hoverable
        style={{ height: '100%' }}
        cover={
          <div
            style={{
              height: 120,
              backgroundColor: '#f5f5f5',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              border: '1px solid #d9d9d9',
            }}
          >
            <Avatar
              size={64}
              icon={plugin.icon}
              style={{ backgroundColor: '#1890ff' }}
            />
          </div>
        }
        actions={[
          <Button
            type="primary"
            icon={<DownloadOutlined />}
            loading={installingPlugin === plugin.id}
            onClick={() => handleInstall(plugin)}
            block
          >
            安装
          </Button>,
          <Button
            icon={<EyeOutlined />}
            onClick={() => setSelectedPlugin(plugin)}
          >
            详情
          </Button>,
        ]}
      >
        <Card.Meta
          title={
            <Space>
              <span>{plugin.name}</span>
              <Tag size="small">{plugin.version}</Tag>
            </Space>
          }
          description={
            <Space direction="vertical" size="small" style={{ width: '100%' }}>
              <Text ellipsis={{ rows: 2 }} type="secondary">
                {plugin.description}
              </Text>
              <Space wrap>
                <Rate
                  disabled
                  defaultValue={plugin.rating}
                  size="small"
                  style={{ fontSize: 12 }}
                />
                <Text type="secondary" style={{ fontSize: 12 }}>
                  ({plugin.reviews})
                </Text>
                <Tag color="blue">{plugin.category}</Tag>
                <Tag color="green">{plugin.runtime}</Tag>
              </Space>
              <Space>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  <DownloadOutlined /> {plugin.downloads.toLocaleString()}
                </Text>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  <ClockCircleOutlined />{' '}
                  {plugin.lastUpdated.toLocaleDateString()}
                </Text>
              </Space>
              <Space wrap>
                {plugin.tags.slice(0, 3).map((tag) => (
                  <Tag key={tag} size="small">
                    {tag}
                  </Tag>
                ))}
                {plugin.tags.length > 3 && (
                  <Tag size="small">+{plugin.tags.length - 3}</Tag>
                )}
              </Space>
            </Space>
          }
        />
      </Card>
    ),
    [installingPlugin, handleInstall],
  );

  return (
    <Modal
      title={
        <Space>
          <CloudOutlined />
          <span>插件市场</span>
        </Space>
      }
      open={visible}
      onCancel={onClose}
      width={1000}
      footer={null}
      destroyOnClose
    >
      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="浏览插件" key="browse">
          {/* 搜索和筛选 */}
          <div style={{ marginBottom: 16 }}>
            <Row gutter={16}>
              <Col span={12}>
                <Search
                  placeholder="搜索插件名称、描述或标签..."
                  value={searchText}
                  onChange={(e) => setSearchText(e.target.value)}
                  allowClear
                  size="large"
                />
              </Col>
              <Col span={8}>
                <Select
                  placeholder="分类"
                  value={selectedCategory}
                  onChange={setSelectedCategory}
                  style={{ width: '100%' }}
                  size="large"
                >
                  {categories.map((cat) => (
                    <Option key={cat} value={cat}>
                      {cat === 'all' ? '全部分类' : cat}
                    </Option>
                  ))}
                </Select>
              </Col>
              <Col span={4}>
                <Select
                  placeholder="排序"
                  value={sortBy}
                  onChange={setSortBy}
                  style={{ width: '100%' }}
                  size="large"
                >
                  <Option value="popularity">
                    <SortAscendingOutlined /> 热门程度
                  </Option>
                  <Option value="rating">
                    <StarOutlined /> 评分
                  </Option>
                  <Option value="updated">
                    <ClockCircleOutlined /> 最近更新
                  </Option>
                </Select>
              </Col>
            </Row>
          </div>

          {/* 插件列表 */}
          <Row gutter={[16, 16]}>
            {sortedPlugins.map((plugin) => (
              <Col xs={24} sm={12} md={8} lg={6} xl={6} key={plugin.id}>
                <PluginCard plugin={plugin} />
              </Col>
            ))}
          </Row>

          {sortedPlugins.length === 0 && (
            <Empty
              description="没有找到匹配的插件"
              style={{ padding: '60px 0' }}
            />
          )}
        </TabPane>

        <TabPane tab="分类浏览" key="categories">
          {categories.map((category) => (
            <div key={category} style={{ marginBottom: 24 }}>
              <Title level={4}>
                {category === 'all' ? '全部分类' : category}
                <Badge
                  count={getCategoryPlugins(category).length}
                  style={{ marginLeft: 8 }}
                />
              </Title>
              <Row gutter={[16, 16]}>
                {getCategoryPlugins(category)
                  .slice(0, category === 'all' ? 6 : 4)
                  .map((plugin) => (
                    <Col xs={24} sm={12} md={8} lg={6} key={plugin.id}>
                      <PluginCard plugin={plugin} />
                    </Col>
                  ))}
              </Row>
              {category !== 'all' &&
                getCategoryPlugins(category).length > 4 && (
                  <div style={{ textAlign: 'center', marginTop: 16 }}>
                    <Button>
                      查看更多 {getCategoryPlugins(category).length - 4} 个插件
                    </Button>
                  </div>
                )}
            </div>
          ))}
        </TabPane>

        <TabPane tab="推荐插件" key="featured">
          <Title level={3}>精选推荐</Title>
          <Paragraph type="secondary">
            根据评分和下载量为您推荐的热门插件
          </Paragraph>
          <Row gutter={[16, 16]}>
            {getFeaturedPlugins().map((plugin) => (
              <Col xs={24} sm={12} md={8} key={plugin.id}>
                <PluginCard plugin={plugin} />
              </Col>
            ))}
          </Row>
        </TabPane>
      </Tabs>

      {/* 插件详情弹窗 */}
      <Modal
        title={
          <Space>
            {selectedPlugin?.icon}
            <span>{selectedPlugin?.name}</span>
          </Space>
        }
        open={!!selectedPlugin}
        onCancel={() => setSelectedPlugin(null)}
        footer={[
          <Button key="close" onClick={() => setSelectedPlugin(null)}>
            关闭
          </Button>,
          <Button
            key="install"
            type="primary"
            icon={<DownloadOutlined />}
            loading={installingPlugin === selectedPlugin?.id}
            onClick={() => selectedPlugin && handleInstall(selectedPlugin)}
          >
            安装插件
          </Button>,
        ]}
        width={600}
      >
        {selectedPlugin && (
          <div>
            <Paragraph>{selectedPlugin.description}</Paragraph>

            <Row gutter={16} style={{ marginBottom: 16 }}>
              <Col span={8}>
                <Statistic
                  title="评分"
                  value={selectedPlugin.rating}
                  precision={1}
                  suffix={`/ 5.0`}
                  valueStyle={{ color: '#faad14' }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="下载量"
                  value={selectedPlugin.downloads}
                  valueStyle={{ color: '#1890ff' }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="评价数"
                  value={selectedPlugin.reviews}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Col>
            </Row>

            <Divider />

            <Space direction="vertical" size="middle" style={{ width: '100%' }}>
              <div>
                <Text strong>版本: </Text>
                <Tag color="blue">{selectedPlugin.version}</Tag>
                <Text type="secondary">({selectedPlugin.size})</Text>
              </div>

              <div>
                <Text strong>作者: </Text>
                <Text>{selectedPlugin.author}</Text>
              </div>

              <div>
                <Text strong>分类: </Text>
                <Tag color="green">{selectedPlugin.category}</Tag>
                <Tag color="orange">{selectedPlugin.runtime}</Tag>
              </div>

              <div>
                <Text strong>标签: </Text>
                <Space wrap>
                  {selectedPlugin.tags.map((tag) => (
                    <Tag key={tag}>{tag}</Tag>
                  ))}
                </Space>
              </div>

              {selectedPlugin.dependencies.length > 0 && (
                <div>
                  <Text strong>依赖: </Text>
                  <Space wrap>
                    {selectedPlugin.dependencies.map((dep) => (
                      <Tag key={dep} color="cyan">
                        {dep}
                      </Tag>
                    ))}
                  </Space>
                </div>
              )}

              <div>
                <Text strong>更新时间: </Text>
                <Text>{selectedPlugin.lastUpdated.toLocaleDateString()}</Text>
              </div>

              <div>
                <Text strong>许可证: </Text>
                <Text>{selectedPlugin.license}</Text>
              </div>

              {(selectedPlugin.homepage || selectedPlugin.repository) && (
                <Space>
                  {selectedPlugin.homepage && (
                    <Button
                      icon={<GlobalOutlined />}
                      href={selectedPlugin.homepage}
                      target="_blank"
                    >
                      主页
                    </Button>
                  )}
                  {selectedPlugin.repository && (
                    <Button
                      icon={<CloudUploadOutlined />}
                      href={selectedPlugin.repository}
                      target="_blank"
                    >
                      源码
                    </Button>
                  )}
                </Space>
              )}
            </Space>
          </div>
        )}
      </Modal>
    </Modal>
  );
};

export default PluginMarket;
