/**
 * 配置页面侧边栏组件
 * 默认展开状态，提供组件库访问和其他工具功能
 */

import React, { useEffect, useRef } from 'react';
import { Button, Space, Divider } from 'antd';
import {
  AppstoreOutlined,
  SettingOutlined,
  FileOutlined,
  SaveOutlined,
  ClearOutlined,
  BgColorsOutlined,
  ToolOutlined,
} from '@ant-design/icons';
import { useSidebarState } from '../hooks/useSidebarState';
import SidebarToggle from './SidebarToggle';
import './sidebar.css';

interface ConfigSidebarProps {
  className?: string;
  onClearCanvas?: () => void;
  onSaveConfig?: () => void;
  onLoadConfig?: () => void;
}

const ConfigSidebar: React.FC<ConfigSidebarProps> = ({
  className = '',
  onClearCanvas,
  onSaveConfig,
  onLoadConfig,
}) => {
  const {
    collapsed,
    width,
    toggleSidebar,
    showPanelAtCenter,
  } = useSidebarState();

  const sidebarRef = useRef<HTMLDivElement>(null);

  // 移动端点击外部关闭侧边栏
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        window.innerWidth <= 768 &&
        !collapsed &&
        sidebarRef.current &&
        !sidebarRef.current.contains(event.target as Node)
      ) {
        toggleSidebar();
      }
    };

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && !collapsed) {
        toggleSidebar();
      }
    };

    if (!collapsed) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleEscape);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
    };
  }, [collapsed, toggleSidebar]);

  // 移动端遮罩层
  const renderMobileOverlay = () => {
    if (window.innerWidth <= 768 && !collapsed) {
      return <div className="sidebar-overlay visible" onClick={toggleSidebar} />;
    }
    return null;
  };

  return (
    <>
      {renderMobileOverlay()}
      <aside
        ref={sidebarRef}
        className={`config-sidebar ${collapsed ? 'collapsed' : 'expanded'} ${className}`}
        style={{ width: collapsed ? 0 : width }}
      >
        {/* 侧边栏头部 */}
        <div className="sidebar-header">
          {!collapsed && <h3>组件工具</h3>}
          <SidebarToggle collapsed={collapsed} onToggle={toggleSidebar} />
        </div>

        {/* 侧边栏内容 */}
        {!collapsed && (
          <div className="sidebar-content">
            {/* 组件库按钮 */}
            <div className="sidebar-group">
              <h4 className="sidebar-group-title">组件库</h4>
              <Button
                type="primary"
                icon={<AppstoreOutlined />}
                className="sidebar-tool-button primary"
                onClick={showPanelAtCenter}
              >
                打开组件库
              </Button>
            </div>

            <Divider style={{ margin: '12px 0' }} />

            {/* 画布操作 */}
            <div className="sidebar-group">
              <h4 className="sidebar-group-title">画布操作</h4>
              <Space direction="vertical" style={{ width: '100%', display: 'flex' }}>
                <Button
                  icon={<SaveOutlined />}
                  className="sidebar-tool-button"
                  onClick={onSaveConfig}
                >
                  保存配置
                </Button>
                <Button
                  icon={<FileOutlined />}
                  className="sidebar-tool-button"
                  onClick={onLoadConfig}
                >
                  加载配置
                </Button>
                <Button
                  icon={<ClearOutlined />}
                  className="sidebar-tool-button"
                  onClick={onClearCanvas}
                >
                  清空画布
                </Button>
              </Space>
            </div>

            <Divider style={{ margin: '12px 0' }} />

            {/* 视图设置 */}
            <div className="sidebar-group">
              <h4 className="sidebar-group-title">视图设置</h4>
              <Space direction="vertical" style={{ width: '100%', display: 'flex' }}>
                <Button
                  icon={<BgColorsOutlined />}
                  className="sidebar-tool-button"
                  title="切换主题"
                >
                  主题设置
                </Button>
                <Button
                  icon={<SettingOutlined />}
                  className="sidebar-tool-button"
                  title="高级设置"
                >
                  高级设置
                </Button>
              </Space>
            </div>

            <Divider style={{ margin: '12px 0' }} />

            {/* 工具箱 */}
            <div className="sidebar-group">
              <h4 className="sidebar-group-title">工具箱</h4>
              <Space direction="vertical" style={{ width: '100%', display: 'flex' }}>
                <Button
                  icon={<ToolOutlined />}
                  className="sidebar-tool-button"
                  title="开发工具"
                >
                  开发工具
                </Button>
              </Space>
            </div>

            {/* 底部信息 */}
            <div style={{ marginTop: 'auto', paddingTop: '16px' }}>
              <div style={{
                fontSize: '12px',
                color: '#8c8c8c',
                textAlign: 'center',
                lineHeight: '1.4'
              }}>
                <div>小智工作流</div>
                <div>v1.0.0</div>
              </div>
            </div>
          </div>
        )}

        {/* 折叠状态的快速访问按钮 */}
        {collapsed && (
          <div className="sidebar-content">
            <Button
              type="primary"
              icon={<AppstoreOutlined />}
              className="sidebar-tool-button primary"
              onClick={showPanelAtCenter}
              title="打开组件库"
            />
          </div>
        )}
      </aside>
    </>
  );
};

export default ConfigSidebar;