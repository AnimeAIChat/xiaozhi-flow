import { useCallback } from 'react';
import { message } from 'antd';
import { log } from '../../../utils/logger';
import { DashboardViewMode } from '../types';

// Dashboard视图管理Hook
export const useDashboardNavigation = () => {
  // 这里可以添加状态管理来控制视图切换
  // 目前先简单处理，返回一个空函数，因为双击功能在各个视图内部处理

  const handleDoubleClick = useCallback((fromView?: DashboardViewMode) => {
    log.info('用户双击操作', { fromView }, 'ui', 'Dashboard');

    // 根据不同的视图显示不同的提示
    if (fromView === 'workflow') {
      message.info('工作流节点双击功能待实现');
    } else if (fromView === 'database') {
      message.info('数据库表双击功能待实现');
    } else if (fromView === 'config') {
      message.info('配置节点双击功能待实现');
    } else {
      message.info('双击功能待实现');
    }
  }, []);

  return {
    handleDoubleClick,
  };
};