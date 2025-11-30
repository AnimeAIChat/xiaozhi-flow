import React, { useEffect, ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useSystemStatus } from '../hooks/useApi';
import { Spin } from 'antd';

interface SmartRootRouteProps {
  children: ReactNode;
}

/**
 * SmartRootRoute - 根据系统状态和认证状态智能路由到合适的页面
 * - 如果系统未初始化，显示 setup 页面（无论认证状态如何）
 * - 如果系统已初始化且用户已认证，重定向到 dashboard
 * - 如果系统已初始化但用户未认证，显示 setup 页面
 */
export const SmartRootRoute: React.FC<SmartRootRouteProps> = ({ children }) => {
  const { isAuthenticated, isLoading, checkAuth } = useAuth();
  const { data: systemStatus, isLoading: systemLoading } = useSystemStatus();

  // 在组件挂载时检查认证状态和系统状态
  useEffect(() => {
    // 只有当系统已初始化时才检查认证状态
    if (!systemLoading && systemStatus && systemStatus.initialized && !isLoading && !isAuthenticated) {
      checkAuth().catch(() => {
        // 错误处理在 AuthContext 中完成
      });
    }
  }, [isLoading, isAuthenticated, checkAuth, systemLoading, systemStatus]);

  // 显示加载状态
  if (isLoading || systemLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen gap-8">
        <Spin size="large" />
        <div className="text-gray-500 text-sm">
          {systemLoading ? '检查系统状态...' : '检查身份状态...'}
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

    // 系统已初始化，根据认证状态决定
    if (isAuthenticated) {
      return <Navigate to="/dashboard" replace />;
    }

    // 系统已初始化但用户未认证，重定向到登录页面
    return <Navigate to="/login" replace />;
  }

  // 无法获取系统状态，显示子组件（Setup 页面）
  return <>{children}</>;
};

export default SmartRootRoute;