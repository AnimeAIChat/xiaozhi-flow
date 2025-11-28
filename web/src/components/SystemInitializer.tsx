import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useSystemStatus } from '../hooks/useApi';
import { Spin } from 'antd';
import { log } from '../utils/logger';

interface SystemInitializerProps {
  children: React.ReactNode;
}

const SystemInitializer: React.FC<SystemInitializerProps> = ({ children }) => {
  const navigate = useNavigate();
  const { data: systemStatus, isLoading, error } = useSystemStatus();
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    if (!isLoading) {
      log.debug('系统初始化器状态检查', {
        systemStatus,
        error,
        currentPath: window.location.pathname
      }, 'system', 'SystemInitializer');

      const currentPath = window.location.pathname;

      // 如果正在配置页面，允许访问并跳过路由检查
      if (currentPath === '/config') {
        log.info('用户在配置页面，跳过路由检查', { currentPath }, 'system', 'SystemInitializer');
        setInitialized(true);
        return;
      }

      // 如果从配置页面刚完成初始化跳转到 dashboard，给予特殊处理
      const isComingFromConfig = document.referrer.includes('/config') ||
                                sessionStorage.getItem('comingFromConfig') === 'true';

      if (isComingFromConfig && currentPath === '/dashboard') {
        log.info('从配置页面跳转到仪表板，允许访问', { currentPath, referrer: document.referrer }, 'system', 'SystemInitializer');
        sessionStorage.removeItem('comingFromConfig');
        setInitialized(true);
        return;
      }

      if (systemStatus && !error) {
        // 如果系统已初始化
        const isSystemInitialized = systemStatus.initialized === true && systemStatus.needs_setup !== true;

        if (isSystemInitialized) {
          // 如果当前在 setup 或根路径，重定向到 dashboard
          if (currentPath === '/' || currentPath === '/setup') {
            log.info('系统已初始化，重定向到仪表板', { currentPath, systemStatus }, 'system', 'SystemInitializer');
            setTimeout(() => {
              navigate('/dashboard', { replace: true });
            }, 100);
            return;
          } else if (currentPath === '/dashboard') {
            log.info('系统已初始化，允许访问仪表板', { currentPath, systemStatus }, 'system', 'SystemInitializer');
            setInitialized(true);
          } else {
            // 其他页面也允许访问
            log.info('系统已初始化，允许访问页面', { currentPath, systemStatus }, 'system', 'SystemInitializer');
            setInitialized(true);
          }
        } else {
          // 系统未初始化
          log.info('系统未初始化，需要设置', { currentPath, systemStatus }, 'system', 'SystemInitializer');
          // 只允许访问 setup、config 和根路径
          const allowedPaths = ['/', '/setup', '/config'];
          if (allowedPaths.includes(currentPath)) {
            log.info('允许访问设置页面', { currentPath, allowedPaths }, 'system', 'SystemInitializer');
            setInitialized(true);
          } else {
            log.warn('系统未初始化，重定向到设置页面', { currentPath, systemStatus }, 'system', 'SystemInitializer');
            setTimeout(() => {
              navigate('/setup', { replace: true });
            }, 100);
            return;
          }
        }
      } else {
        // API 调用出错，只在第一次出错时重定向到 setup
        if (!initialized && currentPath !== '/setup') {
          log.error('API调用错误，重定向到设置页面', { error, currentPath }, 'system', 'SystemInitializer', error instanceof Error ? error.stack : undefined);
          setTimeout(() => {
            navigate('/setup', { replace: true });
          }, 100);
          return;
        } else if (currentPath === '/setup') {
          // 如果已经在 setup 页面，即使 API 出错也显示页面
          log.info('已在设置页面，即使API出错也显示页面', { currentPath }, 'system', 'SystemInitializer');
          setInitialized(true);
        }
      }
    }
  }, [systemStatus, isLoading, error, navigate, initialized]);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-center">
          <Spin size="large" />
          <div className="mt-4 text-gray-600">正在检查系统状态...</div>
        </div>
      </div>
    );
  }

  if (!initialized) {
    return null; // 将会重定向到配置页面
  }

  return <>{children}</>;
};

export default SystemInitializer;