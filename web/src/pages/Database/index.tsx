import React, { useState, useEffect } from 'react';
import { Card, Spin, Alert, Button, Space, Typography, message } from 'antd';
import { ReloadOutlined, DatabaseOutlined, FullscreenOutlined } from '@ant-design/icons';
import { DatabaseTableNodes } from '../../components/DatabaseTableNodes';
import { apiService } from '../../services/api';
import { DatabaseSchema } from '../../types';
import { useNavigate } from 'react-router-dom';

const { Title, Text } = Typography;

const DatabasePage: React.FC = () => {
  const [schema, setSchema] = useState<DatabaseSchema | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  // 加载数据库schema
  const loadDatabaseSchema = async () => {
    setLoading(true);
    setError(null);

    try {
      const data = await apiService.getDatabaseSchema();

      // 转换后端数据格式到前端格式
      const transformedSchema: DatabaseSchema = {
        name: data.name,
        type: data.type,
        tables: data.tables.map((table: any) => ({
          id: table.name,
          name: table.name,
          type: 'table' as const,
          schema: data.name,
          rowCount: table.rowCount,
          size: table.size,
          columns: table.columns.map((col: any) => ({
            id: `${table.name}.${col.name}`,
            name: col.name,
            type: col.type,
            nullable: col.nullable,
            primaryKey: col.primaryKey,
            unique: col.unique,
            defaultValue: col.defaultValue,
            description: col.description,
            position: { x: 0, y: 0 },
          })),
          indexes: table.indexes?.map((idx: any) => ({
            id: `${table.name}.${idx.name}`,
            name: idx.name,
            columns: idx.columns,
            unique: idx.unique,
            type: idx.type,
          })) || [],
          foreignKeys: [],
          position: { x: 0, y: 0 },
        })),
        relationships: data.relationships?.map((rel: any) => ({
          id: rel.name || `${rel.sourceTable}_${rel.targetTable}`,
          source: rel.sourceTable,
          target: rel.targetTable,
          type: 'foreign_key' as const,
          label: `${rel.sourceColumn} → ${rel.targetColumn}`,
          style: {
            color: '#1890ff',
            width: 2,
            style: 'solid' as const,
            arrowType: 'arrow' as const,
          },
        })) || [],
      };

      setSchema(transformedSchema);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load database schema';
      setError(errorMessage);
      message.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  // 页面初始化时加载数据
  useEffect(() => {
    loadDatabaseSchema();
  }, []);

  // 处理表选择
  const handleTableSelect = (tableName: string) => {
    message.info(`选择了表: ${tableName}`);
    // 这里可以添加跳转到表详情页面的逻辑
    // navigate(`/database/tables/${tableName}`);
  };

  // 处理刷新
  const handleRefresh = () => {
    loadDatabaseSchema();
  };

  // 处理全屏
  const handleFullscreen = () => {
    if (document.documentElement.requestFullscreen) {
      document.documentElement.requestFullscreen();
    }
  };

  return (
    <div style={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* 页面头部 */}
      <div style={{
        padding: '16px 24px',
        borderBottom: '1px solid #f0f0f0',
        background: '#fff',
        zIndex: 100
      }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Space>
            <DatabaseOutlined style={{ fontSize: 24, color: '#1890ff' }} />
            <div>
              <Title level={3} style={{ margin: 0, fontWeight: 600 }}>
                数据库表关系图
              </Title>
              <Text type="secondary" style={{ fontSize: 14 }}>
                可视化展示数据库表结构和关系
              </Text>
            </div>
          </Space>

          <Space>
            <Button
              icon={<ReloadOutlined />}
              onClick={handleRefresh}
              loading={loading}
            >
              刷新
            </Button>
            <Button
              icon={<FullscreenOutlined />}
              onClick={handleFullscreen}
            >
              全屏
            </Button>
            <Button
              type="default"
              onClick={() => navigate('/dashboard')}
            >
              返回仪表盘
            </Button>
          </Space>
        </div>
      </div>

      {/* 内容区域 */}
      <div style={{ flex: 1, position: 'relative' }}>
        {loading && (
          <div style={{
            position: 'absolute',
            top: '50%',
            left: '50%',
            transform: 'translate(-50%, -50%)',
            zIndex: 1000,
            textAlign: 'center'
          }}>
            <Card style={{ textAlign: 'center', minWidth: 300 }}>
              <Spin size="large" />
              <div style={{ marginTop: 16 }}>
                <Text>正在加载数据库表结构...</Text>
              </div>
            </Card>
          </div>
        )}

        {error && (
          <div style={{
            position: 'absolute',
            top: '50%',
            left: '50%',
            transform: 'translate(-50%, -50%)',
            zIndex: 1000,
            width: '80%',
            maxWidth: 600
          }}>
            <Alert
              message="加载失败"
              description={error}
              type="error"
              showIcon
              action={
                <Space>
                  <Button size="small" onClick={handleRefresh}>
                    重试
                  </Button>
                  <Button size="small" onClick={() => navigate('/dashboard')}>
                    返回
                  </Button>
                </Space>
              }
            />
          </div>
        )}

        {schema && !loading && !error && (
          <DatabaseTableNodes
            schema={schema}
            onTableSelect={handleTableSelect}
          />
        )}
      </div>
    </div>
  );
};

export default DatabasePage;