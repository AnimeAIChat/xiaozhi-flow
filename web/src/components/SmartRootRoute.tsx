import React, { useEffect, ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { Spin } from 'antd';

interface SmartRootRouteProps {
  children: ReactNode;
}

/**
 * SmartRootRoute - 根据认证状态智能路由到合适的页面
 * - 如果用户已认证，重定向到 dashboard
 * - 如果用户未认证，显示 setup 页面
 */
export const SmartRootRoute: React.FC<SmartRootRouteProps> = ({ children }) => {
  const { isAuthenticated, isLoading, checkAuth } = useAuth();

  // 在组件挂载时检查认证状态
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      checkAuth().catch(() => {
        // 错误处理在 AuthContext 中完成
      });
    }
  }, [isLoading, isAuthenticated, checkAuth]);

  // 显示加载状态
  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen gap-8">
        <Spin size="large" />
        <div className="text-gray-500 text-sm">检查身份状态...</div>
      </div>
    );
  }

  // 如果已认证，重定向到 dashboard
  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  // 未认证，显示子组件（Setup 页面）
  return <>{children}</>;
};

export default SmartRootRoute;