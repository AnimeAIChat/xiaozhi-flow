import React, { useState, useEffect } from 'react';
import { Form, Input, Select, InputNumber, Button, Card, Divider, Space, Alert, Spin } from 'antd';

const { Option } = Select;
const { TextArea } = Input;

interface DatabaseConnection {
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
    conn_max_lifetime: number; // 分钟
  };
}

interface AdminConfig {
  username: string;
  password: string;
  email?: string;
}

interface DatabaseConfigProps {
  onConfigChange: (config: {
    database: DatabaseConnection;
    admin: AdminConfig;
  }) => void;
  initialConfig?: {
    database: DatabaseConnection;
    admin: AdminConfig;
  };
  loading?: boolean;
}

const DatabaseConfigForm: React.FC<DatabaseConfigProps> = ({
  onConfigChange,
  initialConfig,
  loading = false
}) => {
  const [form] = Form.useForm();
  const [dbType, setDbType] = useState<string>('sqlite');
  const [config, setConfig] = useState({
    database: initialConfig?.database || {
      type: 'sqlite',
      path: './data/xiaozhi.db',
      connection_pool: {
        max_open_conns: 25,
        max_idle_conns: 5,
        conn_max_lifetime: 5
      }
    },
    admin: initialConfig?.admin || {
      username: 'admin',
      password: '123456',
      email: 'admin@xiaozhi.local'
    }
  });

  useEffect(() => {
    if (initialConfig) {
      const updatedConfig = {
        ...initialConfig,
        database: {
          ...initialConfig.database,
          connection_pool: {
            ...initialConfig.database.connection_pool,
            conn_max_lifetime: initialConfig.database.connection_pool.conn_max_lifetime / 60
          }
        }
      };

      form.setFieldsValue({
        ...updatedConfig,
        database: {
          ...updatedConfig.database,
          connection_pool: {
            ...updatedConfig.database.connection_pool,
            conn_max_lifetime: updatedConfig.database.connection_pool.conn_max_lifetime
          }
        }
      });
      setConfig(updatedConfig);
      setDbType(initialConfig.database.type);
      onConfigChange(updatedConfig);
    } else {
      // 如果没有初始配置，使用默认配置并通知父组件
      // 只在组件第一次挂载时执行
      onConfigChange(config);
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleDbTypeChange = (type: string) => {
    setDbType(type);

    // 根据数据库类型设置默认配置
    let defaultConfig = { ...config.database };

    switch (type) {
      case 'sqlite':
        defaultConfig = {
          type: 'sqlite',
          path: './data/xiaozhi.db',
          connection_pool: {
            max_open_conns: 25,
            max_idle_conns: 5,
            conn_max_lifetime: 5
          }
        };
        break;
      case 'mysql':
        defaultConfig = {
          type: 'mysql',
          host: 'localhost',
          port: 3306,
          database: 'xiaozhi',
          username: 'root',
          password: '',
          charset: 'utf8mb4',
          connection_pool: {
            max_open_conns: 25,
            max_idle_conns: 5,
            conn_max_lifetime: 5
          }
        };
        break;
      case 'postgresql':
        defaultConfig = {
          type: 'postgresql',
          host: 'localhost',
          port: 5432,
          database: 'xiaozhi',
          username: 'postgres',
          password: '',
          ssl_mode: 'prefer',
          connection_pool: {
            max_open_conns: 25,
            max_idle_conns: 5,
            conn_max_lifetime: 5
          }
        };
        break;
    }

    setConfig({ ...config, database: defaultConfig });
    form.setFieldsValue({ database: defaultConfig });
  };

  const handleValuesChange = (changedValues: any, allValues: any) => {
    const newConfig = {
      database: {
        ...allValues.database,
        connection_pool: {
          ...allValues.database.connection_pool,
          conn_max_lifetime: allValues.database.connection_pool.conn_max_lifetime * 60 // 转换为秒
        }
      },
      admin: allValues.admin
    };

    setConfig(newConfig);
    onConfigChange(newConfig);
  };

  const renderDatabaseConfig = () => {
    switch (dbType) {
      case 'sqlite':
        return (
          <>
            <Form.Item
              label="数据库文件路径"
              name={['database', 'path']}
              rules={[
                { required: true, message: '请输入数据库文件路径' },
                { pattern: '^\\./.*|^[a-zA-Z]:.*|^/.*', message: '请输入有效的文件路径' }
              ]}
            >
              <Input
                placeholder="./data/xiaozhi.db"
              />
            </Form.Item>
          </>
        );

      case 'mysql':
        return (
          <>
            <Form.Item
              label="主机地址"
              name={['database', 'host']}
              rules={[{ required: true, message: '请输入数据库主机地址' }]}
            >
              <Input placeholder="localhost" />
            </Form.Item>

            <Form.Item
              label="端口"
              name={['database', 'port']}
              rules={[{ required: true, message: '请输入数据库端口' }]}
            >
              <InputNumber min={1} max={65535} placeholder="3306" style={{ width: '100%' }} />
            </Form.Item>

            <Form.Item
              label="数据库名"
              name={['database', 'database']}
              rules={[{ required: true, message: '请输入数据库名称' }]}
            >
              <Input placeholder="xiaozhi" />
            </Form.Item>

            <Form.Item
              label="用户名"
              name={['database', 'username']}
              rules={[{ required: true, message: '请输入数据库用户名' }]}
            >
              <Input placeholder="root" />
            </Form.Item>

            <Form.Item
              label="密码"
              name={['database', 'password']}
              rules={[{ required: true, message: '请输入数据库密码' }]}
            >
              <Input.Password placeholder="请输入密码" />
            </Form.Item>

            <Form.Item
              label="字符集"
              name={['database', 'charset']}
            >
              <Select placeholder="utf8mb4" defaultValue="utf8mb4">
                <Option value="utf8mb4">utf8mb4</Option>
                <Option value="utf8">utf8</Option>
                <Option value="latin1">latin1</Option>
              </Select>
            </Form.Item>
          </>
        );

      case 'postgresql':
        return (
          <>
            <Form.Item
              label="主机地址"
              name={['database', 'host']}
              rules={[{ required: true, message: '请输入数据库主机地址' }]}
            >
              <Input placeholder="localhost" />
            </Form.Item>

            <Form.Item
              label="端口"
              name={['database', 'port']}
              rules={[{ required: true, message: '请输入数据库端口' }]}
            >
              <InputNumber min={1} max={65535} placeholder="5432" style={{ width: '100%' }} />
            </Form.Item>

            <Form.Item
              label="数据库名"
              name={['database', 'database']}
              rules={[{ required: true, message: '请输入数据库名称' }]}
            >
              <Input placeholder="xiaozhi" />
            </Form.Item>

            <Form.Item
              label="用户名"
              name={['database', 'username']}
              rules={[{ required: true, message: '请输入数据库用户名' }]}
            >
              <Input placeholder="postgres" />
            </Form.Item>

            <Form.Item
              label="密码"
              name={['database', 'password']}
              rules={[{ required: true, message: '请输入数据库密码' }]}
            >
              <Input.Password placeholder="请输入密码" />
            </Form.Item>

            <Form.Item
              label="SSL模式"
              name={['database', 'ssl_mode']}
            >
              <Select placeholder="prefer" defaultValue="prefer">
                <Option value="disable">disable</Option>
                <Option value="allow">allow</Option>
                <Option value="prefer">prefer</Option>
                <Option value="require">require</Option>
                <Option value="verify-ca">verify-ca</Option>
                <Option value="verify-full">verify-full</Option>
              </Select>
            </Form.Item>
          </>
        );

      default:
        return null;
    }
  };

  if (loading) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '50px' }}>
          <Spin size="large" />
          <div style={{ marginTop: 16 }}>加载配置中...</div>
        </div>
      </Card>
    );
  }

  return (
    <Spin spinning={loading}>
      <Card title="数据库配置">
        <Form
          form={form}
          layout="vertical"
          initialValues={config}
          onValuesChange={handleValuesChange}
        >
          <Form.Item
            label="数据库类型"
            name={['database', 'type']}
            rules={[{ required: true, message: '请选择数据库类型' }]}
          >
            <Select placeholder="请选择数据库类型" onChange={handleDbTypeChange}>
              <Option value="sqlite">SQLite</Option>
              <Option value="mysql">MySQL</Option>
              <Option value="postgresql">PostgreSQL</Option>
            </Select>
          </Form.Item>

          {dbType === 'sqlite' && (
            <Alert
              title="SQLite 说明"
              description="SQLite 是一个轻量级的文件数据库，适合单机部署和开发环境使用。"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
          )}

          {dbType === 'mysql' && (
            <Alert
              title="MySQL 说明"
              description="MySQL 是一个流行的关系型数据库，适合生产环境和多用户场景。"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
          )}

          {dbType === 'postgresql' && (
            <Alert
              title="PostgreSQL 说明"
              description="PostgreSQL 是一个功能强大的开源数据库，支持高级特性和扩展。"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
          )}

          {renderDatabaseConfig()}

          <Divider>连接池配置</Divider>

          <Form.Item
            label="最大打开连接数"
            name={['database', 'connection_pool', 'max_open_conns']}
            rules={[{ required: true, message: '请输入最大打开连接数' }]}
          >
            <InputNumber min={1} max={1000} placeholder="25" style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            label="最大空闲连接数"
            name={['database', 'connection_pool', 'max_idle_conns']}
            rules={[{ required: true, message: '请输入最大空闲连接数' }]}
          >
            <InputNumber min={1} max={1000} placeholder="5" style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            label="连接最大生存时间（分钟）"
            name={['database', 'connection_pool', 'conn_max_lifetime']}
            rules={[{ required: true, message: '请输入连接最大生存时间' }]}
          >
            <InputNumber min={1} max={1440} placeholder="5" style={{ width: '100%' }} />
          </Form.Item>

          <Divider>管理员账户配置</Divider>

          <Form.Item
            label="管理员用户名"
            name={['admin', 'username']}
            rules={[
              { required: true, message: '请输入管理员用户名' },
              { min: 3, message: '用户名至少3个字符' }
            ]}
          >
            <Input placeholder="admin" />
          </Form.Item>

          <Form.Item
            label="管理员密码"
            name={['admin', 'password']}
            rules={[
              { required: true, message: '请输入管理员密码' },
              { min: 6, message: '密码至少6个字符' }
            ]}
          >
            <Input placeholder="123456" />
          </Form.Item>

          <Form.Item
            label="邮箱（可选）"
            name={['admin', 'email']}
            rules={[{ type: 'email', message: '请输入有效的邮箱地址' }]}
          >
            <Input placeholder="admin@xiaozhi.local" />
          </Form.Item>
        </Form>
      </Card>
    </Spin>
  );
};

export default DatabaseConfigForm;