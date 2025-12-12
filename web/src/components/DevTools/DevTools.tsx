import {
  ApiOutlined,
  BugOutlined,
  ClearOutlined,
  DashboardOutlined,
  DownloadOutlined,
  InfoCircleOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import {
  Button,
  Card,
  Collapse,
  Divider,
  Drawer,
  Space,
  Statistic,
  Table,
  Tabs,
  Tag,
  Typography,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type React from 'react';
import { useEffect, useState } from 'react';
import { apiService } from '../../services/api';
import { envConfig } from '../../utils/envConfig';
import type { LogEntry } from '../../utils/logger';
import { log, logger } from '../../utils/logger';

const { Title, Text, Paragraph } = Typography;
const { TabPane } = Tabs;
const { Panel } = Collapse;

// 开发者工具组件
const DevTools: React.FC = () => {
  const [visible, setVisible] = useState(false);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [apiHistory, setApiHistory] = useState<any[]>([]);
  const [logStats, setLogStats] = useState<any>({});
  const [apiStats, setApiStats] = useState<any>({});

  // 快捷键切换
  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      // Ctrl+Shift+D 或 Cmd+Shift+D 切换开发者工具
      if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'D') {
        e.preventDefault();
        setVisible(!visible);
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [visible]);

  // 定期更新数据
  useEffect(() => {
    if (!visible) return;

    const updateData = () => {
      setLogs(logger.getLogs());
      setLogStats(logger.getStats());
      setApiHistory(apiService.getApiHistory());
      setApiStats(apiService.getPerformanceStats());
    };

    updateData();
    const interval = setInterval(updateData, 1000);

    return () => clearInterval(interval);
  }, [visible]);

  // 清空日志
  const handleClearLogs = () => {
    logger.clearLogs();
    setLogs([]);
  };

  // 清空API历史
  const handleClearApiHistory = () => {
    apiService.clearApiHistory();
    setApiHistory([]);
  };

  // 导出数据
  const handleExportLogs = () => {
    const data = logger.exportLogs();
    const blob = new Blob([data], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `logs-${new Date().toISOString().split('T')[0]}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleExportApiHistory = () => {
    const data = apiService.exportApiHistory();
    const blob = new Blob([data], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `api-history-${new Date().toISOString().split('T')[0]}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  // 日志表格列
  const logColumns: ColumnsType<LogEntry> = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 150,
      render: (timestamp: string) => new Date(timestamp).toLocaleTimeString(),
    },
    {
      title: '级别',
      dataIndex: 'level',
      key: 'level',
      width: 80,
      render: (level: number) => {
        const colors = ['default', 'blue', 'orange', 'red'];
        const labels = ['DEBUG', 'INFO', 'WARN', 'ERROR'];
        return <Tag color={colors[level]}>{labels[level]}</Tag>;
      },
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      width: 100,
      render: (category: string) => category && <Tag>{category}</Tag>,
    },
    {
      title: '消息',
      dataIndex: 'message',
      key: 'message',
      ellipsis: true,
    },
  ];

  // API历史表格列
  const apiColumns: ColumnsType<any> = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 150,
      render: (timestamp: string) => new Date(timestamp).toLocaleTimeString(),
    },
    {
      title: '方法',
      dataIndex: 'method',
      key: 'method',
      width: 80,
    },
    {
      title: 'URL',
      dataIndex: 'url',
      key: 'url',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: number) => {
        if (!status) return '-';
        const color =
          status >= 400 ? 'red' : status >= 300 ? 'orange' : 'green';
        return <Tag color={color}>{status}</Tag>;
      },
    },
    {
      title: '耗时',
      dataIndex: 'duration',
      key: 'duration',
      width: 100,
      render: (duration: number) => {
        if (!duration) return '-';
        const color =
          duration > 2000 ? 'red' : duration > 1000 ? 'orange' : 'green';
        return <Tag color={color}>{duration}ms</Tag>;
      },
    },
  ];

  // 开发环境浮动按钮
  if (!envConfig.enableDevtools || !envConfig.isDevelopment) {
    return null;
  }

  return (
    <>
      {/* 浮动按钮 */}
      <Button
        type="primary"
        shape="circle"
        size="large"
        icon={<BugOutlined />}
        style={{
          position: 'fixed',
          bottom: 20,
          right: 20,
          zIndex: 9999,
          boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
        }}
        onClick={() => setVisible(true)}
        title="开发者工具 (Ctrl+Shift+D)"
      />

      {/* 开发者工具面板 */}
      <Drawer
        title={
          <Space>
            <BugOutlined />
            <span>开发者工具</span>
            <Text type="secondary" style={{ fontSize: 12 }}>
              (Ctrl+Shift+D)
            </Text>
          </Space>
        }
        placement="right"
        width={800}
        open={visible}
        onClose={() => setVisible(false)}
        extra={
          <Space>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => window.location.reload()}
            >
              刷新页面
            </Button>
          </Space>
        }
      >
        <Tabs defaultActiveKey="overview">
          {/* 概览页 */}
          <TabPane
            tab={
              <span>
                <DashboardOutlined />
                概览
              </span>
            }
            key="overview"
          >
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              {/* 环境信息 */}
              <Card title="环境信息" size="small">
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Text>
                    <strong>环境:</strong> {envConfig.appEnv}
                  </Text>
                  <Text>
                    <strong>调试模式:</strong>{' '}
                    {envConfig.debug ? '启用' : '禁用'}
                  </Text>
                  <Text>
                    <strong>API地址:</strong> {envConfig.apiBaseUrl}
                  </Text>
                  <Text>
                    <strong>日志级别:</strong> {envConfig.logLevel}
                  </Text>
                  <Text>
                    <strong>SourceMap:</strong>{' '}
                    {envConfig.enableSourceMap ? '启用' : '禁用'}
                  </Text>
                </Space>
              </Card>

              {/* 日志统计 */}
              <Card
                title="日志统计"
                size="small"
                extra={
                  <Space>
                    <Button
                      icon={<ClearOutlined />}
                      size="small"
                      onClick={handleClearLogs}
                    >
                      清空
                    </Button>
                    <Button
                      icon={<DownloadOutlined />}
                      size="small"
                      onClick={handleExportLogs}
                    >
                      导出
                    </Button>
                  </Space>
                }
              >
                <Space wrap>
                  <Statistic title="总数" value={logStats.total || 0} />
                  <Statistic
                    title="信息"
                    value={logStats.info || 0}
                    valueStyle={{ color: '#1890ff' }}
                  />
                  <Statistic
                    title="警告"
                    value={logStats.warn || 0}
                    valueStyle={{ color: '#faad14' }}
                  />
                  <Statistic
                    title="错误"
                    value={logStats.error || 0}
                    valueStyle={{ color: '#ff4d4f' }}
                  />
                </Space>
              </Card>

              {/* API统计 */}
              <Card
                title="API统计"
                size="small"
                extra={
                  <Space>
                    <Button
                      icon={<ClearOutlined />}
                      size="small"
                      onClick={handleClearApiHistory}
                    >
                      清空
                    </Button>
                    <Button
                      icon={<DownloadOutlined />}
                      size="small"
                      onClick={handleExportApiHistory}
                    >
                      导出
                    </Button>
                  </Space>
                }
              >
                <Space wrap>
                  <Statistic title="总调用" value={apiStats.totalCalls || 0} />
                  <Statistic
                    title="成功"
                    value={apiStats.successCalls || 0}
                    valueStyle={{ color: '#52c41a' }}
                  />
                  <Statistic
                    title="失败"
                    value={apiStats.errorCalls || 0}
                    valueStyle={{ color: '#ff4d4f' }}
                  />
                  <Statistic
                    title="平均耗时"
                    value={
                      Math.round((apiStats.averageResponseTime || 0) * 100) /
                      100
                    }
                    suffix="ms"
                    valueStyle={{ color: '#722ed1' }}
                  />
                </Space>
              </Card>
            </Space>
          </TabPane>

          {/* 日志页 */}
          <TabPane
            tab={
              <span>
                <BugOutlined />
                日志
              </span>
            }
            key="logs"
          >
            <Space direction="vertical" style={{ width: '100%' }}>
              <Space>
                <Text strong>日志级别过滤:</Text>
                <Button size="small" onClick={() => setLogs(logger.getLogs())}>
                  全部
                </Button>
                <Button
                  size="small"
                  onClick={() => setLogs(logger.getLogsByLevel(3))}
                >
                  错误
                </Button>
                <Button
                  size="small"
                  onClick={() => setLogs(logger.getLogsByLevel(2))}
                >
                  警告
                </Button>
                <Button
                  size="small"
                  onClick={() => setLogs(logger.getLogsByLevel(1))}
                >
                  信息
                </Button>
                <Button
                  size="small"
                  onClick={() => setLogs(logger.getLogsByLevel(0))}
                >
                  调试
                </Button>
              </Space>

              <Table
                dataSource={logs.slice(0, 100)}
                columns={logColumns}
                rowKey="id"
                size="small"
                pagination={{ pageSize: 20, showSizeChanger: false }}
                scroll={{ y: 400 }}
              />
            </Space>
          </TabPane>

          {/* API页 */}
          <TabPane
            tab={
              <span>
                <ApiOutlined />
                API
              </span>
            }
            key="api"
          >
            <Space direction="vertical" style={{ width: '100%' }}>
              <Space>
                <Text strong>API分类过滤:</Text>
                <Button
                  size="small"
                  onClick={() => setApiHistory(apiService.getApiHistory())}
                >
                  全部
                </Button>
                {Object.keys(apiStats.categories || {}).map((category) => (
                  <Button
                    size="small"
                    key={category}
                    onClick={() =>
                      setApiHistory(
                        apiService.getApiHistoryByCategory(category),
                      )
                    }
                  >
                    {category}
                  </Button>
                ))}
                <Button
                  size="small"
                  onClick={() => setApiHistory(apiService.getErrorHistory())}
                >
                  仅错误
                </Button>
              </Space>

              <Table
                dataSource={apiHistory.slice(0, 100)}
                columns={apiColumns}
                rowKey="id"
                size="small"
                pagination={{ pageSize: 20, showSizeChanger: false }}
                scroll={{ y: 400 }}
                expandedRowRender={(record) => (
                  <div style={{ padding: '8px 0' }}>
                    <Collapse ghost size="small">
                      <Panel header="详细信息" key="details">
                        <Space direction="vertical" style={{ width: '100%' }}>
                          <Text>
                            <strong>ID:</strong> {record.id}
                          </Text>
                          <Text>
                            <strong>分类:</strong> {record.category}
                          </Text>
                          <Text>
                            <strong>参数:</strong>{' '}
                            {JSON.stringify(record.params, null, 2)}
                          </Text>
                          <Text>
                            <strong>请求体:</strong>{' '}
                            {JSON.stringify(record.data, null, 2)}
                          </Text>
                          <Text>
                            <strong>响应:</strong>{' '}
                            {JSON.stringify(record.response, null, 2)}
                          </Text>
                          {record.error && (
                            <Text>
                              <strong>错误:</strong>{' '}
                              {JSON.stringify(record.error, null, 2)}
                            </Text>
                          )}
                        </Space>
                      </Panel>
                    </Collapse>
                  </div>
                )}
              />
            </Space>
          </TabPane>
        </Tabs>
      </Drawer>
    </>
  );
};

export default DevTools;
