import React, { useState } from 'react';
import { Layout, Menu, Button, Typography, Avatar, Space } from 'antd';
import {
  DashboardOutlined,
  SettingOutlined,
  LoginOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  RobotOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';

const { Header, Sider, Content } = Layout;
const { Title, Text } = Typography;

interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  const menuItems = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: '仪表板',
    },
    {
      key: '/config',
      icon: <SettingOutlined />,
      label: '配置管理',
    },
    {
      key: 'divider1',
      type: 'divider',
    },
    {
      key: '/login',
      icon: <LoginOutlined />,
      label: '登录',
    },
  ];

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  return (
    <Layout className="min-h-screen bg-white">
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        className="bg-white border-r border-gray-200"
        theme="light"
        width={256}
        collapsedWidth={64}
      >
        <div className="p-4 border-b border-gray-200">
          <div className="flex items-center">
            <div className="w-8 h-8 bg-black rounded-lg flex items-center justify-center">
              <span className="text-white font-bold text-sm">X</span>
            </div>
            {!collapsed && (
              <span className="ml-3 text-lg font-medium text-gray-900">Xiaozhi</span>
            )}
          </div>
        </div>

        <Menu
          theme="light"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
          className="border-none bg-white"
          style={{
            paddingTop: '8px',
          }}
        />
      </Sider>

      <Layout>
        <Header className="bg-white border-b border-gray-200 px-6 flex items-center">
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            className="text-gray-600 hover:bg-gray-100 border-none"
          />

          <div className="ml-6">
            <h1 className="text-lg font-medium text-gray-900">Configuration</h1>
          </div>
        </Header>

        <Content className="bg-gray-50">
          <div className="p-8">
            {children}
          </div>
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;