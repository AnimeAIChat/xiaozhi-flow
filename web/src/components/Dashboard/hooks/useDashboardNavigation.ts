import { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { message } from 'antd';
import { log } from '../../../utils/logger';
import { DashboardViewMode } from '../types';

export const useDashboardNavigation = () => {
  const navigate = useNavigate();

  const handleDoubleClick = useCallback((fromView?: DashboardViewMode) => {
    log.info('用户双击进入配置编辑器', { fromView }, 'ui', 'Dashboard');

    // 显示提示信息
    message.info('正在打开配置编辑器...', 1);

    // 延迟导航以显示消息
    setTimeout(() => {
      navigate('/config-editor');
    }, 500);
  }, [navigate]);

  return {
    handleDoubleClick,
  };
};