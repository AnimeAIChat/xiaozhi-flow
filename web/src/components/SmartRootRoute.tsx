import React, { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useSystemStatus } from '../hooks/useApi';
import { Spin } from 'antd';

interface SmartRootRouteProps {
  children: ReactNode;
}

/**
 * SmartRootRoute - 根据系统状态智能路由到合适的页面
 * - 如果系统未初始化，显示 setup 页面
 * - 如果系统已初始化，重定向到 dashboard
 */
export const SmartRootRoute: React.FC<SmartRootRouteProps> = ({ children }) => {
  const { data: systemStatus, isLoading: systemLoading } = useSystemStatus();

  // 显示加载状态
  if (systemLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen gap-8">
        <Spin size="large" />
        <div className="text-gray-500 text-sm">
          检查系统状态...
        </div>
      </div>
    );
  }

  // 检查系统是否已初始化
  if (systemStatus) {
    const isSystemInitialized = systemStatus.initialized === true && systemStatus.needs_setup !== true;

    // 如果系统未初始化，显示子组件（Setup 页面）
    if (!isSystemInitialized) {
      return <>{children}</>;
    }

    // 系统已初始化，直接重定向到dashboard
    return <Navigate to="/dashboard" replace />;
  }

  // 无法获取系统状态，显示子组件（Setup 页面）
  return <>{children}</>;
};

export default SmartRootRoute;