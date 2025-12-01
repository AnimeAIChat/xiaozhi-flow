/**
 * 侧边栏折叠切换按钮组件
 */

import React from 'react';
import { Button } from 'antd';
import { MenuFoldOutlined, MenuUnfoldOutlined } from '@ant-design/icons';
import './sidebar.css';

interface SidebarToggleProps {
  collapsed: boolean;
  onToggle: () => void;
  className?: string;
}

const SidebarToggle: React.FC<SidebarToggleProps> = ({
  collapsed,
  onToggle,
  className = '',
}) => {
  return (
    <Button
      type="text"
      icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
      onClick={onToggle}
      className={`sidebar-toggle ${className}`}
      title={collapsed ? '展开侧边栏' : '折叠侧边栏'}
    />
  );
};

export default SidebarToggle;