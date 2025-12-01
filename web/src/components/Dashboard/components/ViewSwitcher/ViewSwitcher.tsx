import React from 'react';
import { Space, Button } from 'antd';
import { DatabaseOutlined, ApiOutlined, SettingOutlined, EditOutlined, AppstoreOutlined } from '@ant-design/icons';
import { ViewSwitcherProps } from '../../types';
import { log } from '../../../../utils/logger';

const ViewSwitcher: React.FC<ViewSwitcherProps> = ({
  currentView,
  onViewChange,
  onPluginManagerOpen
}) => {
  return (
    <div className="absolute top-4 right-4 z-10 bg-white rounded-lg shadow-sm border border-gray-200 p-2">
      <Space>
        <Button
          type={currentView === 'workflow' ? 'primary' : 'default'}
          size="small"
          icon={<ApiOutlined />}
          onClick={() => {
            log.info('用户切换到工作流节点视图', { from: currentView }, 'ui', 'Dashboard');
            onViewChange('workflow');
          }}
        >
          工作流节点
        </Button>
        <Button
          type={currentView === 'database' ? 'primary' : 'default'}
          size="small"
          icon={<DatabaseOutlined />}
          onClick={() => {
            log.info('用户切换到数据库表视图', { from: currentView }, 'ui', 'Dashboard');
            onViewChange('database');
          }}
        >
          数据库表
        </Button>
        <Button
          type={currentView === 'config' ? 'primary' : 'default'}
          size="small"
          icon={<SettingOutlined />}
          onClick={() => {
            log.info('用户切换到配置节点视图', { from: currentView }, 'ui', 'Dashboard');
            onViewChange('config');
          }}
        >
          配置节点
        </Button>
        <Button
          type="primary"
          size="small"
          icon={<AppstoreOutlined />}
          onClick={() => {
            log.info('用户打开插件管理器', { from: currentView }, 'ui', 'Dashboard');
            onPluginManagerOpen?.();
          }}
        >
          插件管理
        </Button>
      </Space>
    </div>
  );
};

export default ViewSwitcher;