import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useSystemStatus } from '../hooks/useApi';
import { Spin } from 'antd';

interface SystemInitializerProps {
  children: React.ReactNode;
}

const SystemInitializer: React.FC<SystemInitializerProps> = ({ children }) => {
  const navigate = useNavigate();
  const { data: systemStatus, isLoading, error } = useSystemStatus();
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    if (!isLoading) {
      console.log('SystemInitializer Debug:', {
        systemStatus,
        error,
        currentPath: window.location.pathname
      });

      const currentPath = window.location.pathname;

      // 如果正在配置页面，允许访问并跳过路由检查
      if (currentPath === '/config') {
        console.log('SystemInitializer: User is on config page, skipping route checks');
        setInitialized(true);
        return;
      }

      // 如果从配置页面刚完成初始化跳转到 dashboard，给予特殊处理
      const isComingFromConfig = document.referrer.includes('/config') ||
                                sessionStorage.getItem('comingFromConfig') === 'true';

      if (isComingFromConfig && currentPath === '/dashboard') {
        console.log('SystemInitializer: Coming from config page to dashboard, allowing access');
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
            console.log('System already initialized, redirecting to dashboard...');
            setTimeout(() => {
              navigate('/dashboard', { replace: true });
            }, 100);
            return;
          } else if (currentPath === '/dashboard') {
            console.log('System initialized, allowing access to dashboard');
            setInitialized(true);
          } else {
            // 其他页面也允许访问
            console.log('System initialized, allowing access to:', currentPath);
            setInitialized(true);
          }
        } else {
          // 系统未初始化
          console.log('System not initialized, needs setup');
          // 只允许访问 setup、config 和根路径
          const allowedPaths = ['/', '/setup', '/config'];
          if (allowedPaths.includes(currentPath)) {
            console.log('Allowing access to setup page:', currentPath);
            setInitialized(true);
          } else {
            console.log('System not initialized, redirecting to setup...');
            setTimeout(() => {
              navigate('/setup', { replace: true });
            }, 100);
            return;
          }
        }
      } else {
        // API 调用出错，只在第一次出错时重定向到 setup
        if (!initialized && currentPath !== '/setup') {
          console.log('API error, redirecting to setup...', { error });
          setTimeout(() => {
            navigate('/setup', { replace: true });
          }, 100);
          return;
        } else if (currentPath === '/setup') {
          // 如果已经在 setup 页面，即使 API 出错也显示页面
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