import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntApp } from 'antd';
import './App.css';
import { QueryProvider } from './components/QueryProvider';
import SystemInitializer from './components/SystemInitializer';
import { ErrorBoundary } from './components/ErrorBoundary';
import DevTools from './components/DevTools';
import { AuthProvider } from './contexts/AuthContext';
import { ProtectedRoute } from './components/ProtectedRoute';
import SmartRootRoute from './components/SmartRootRoute';
import { log } from './utils/logger';

// 页面组件
const Setup = React.lazy(() => import('./pages/Setup'));
const Dashboard = React.lazy(() => import('./pages/Dashboard'));
const Config = React.lazy(() => import('./pages/Config'));
const ConfigEditor = React.lazy(() => import('./pages/ConfigEditor'));
const Login = React.lazy(() => import('./pages/Login'));
const Register = React.lazy(() => import('./pages/Register'));

// 加载组件
const LoadingSpinner: React.FC = () => (
  <div className="flex items-center justify-center min-h-screen bg-white">
    <div className="relative">
      {/* 外圈旋转动画 */}
      <div className="w-12 h-12 border-2 border-gray-200 rounded-full animate-spin border-t-black border-r-black"></div>
      {/* 内圈脉冲效果 */}
      <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
        <div className="w-2 h-2 bg-black rounded-full animate-ping"></div>
      </div>
    </div>
  </div>
);

// 需要系统初始化检查的组件包装器
const SystemRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <SystemInitializer>
    {children}
  </SystemInitializer>
);

const App: React.FC = () => {
  // 全局错误处理函数
  const handleGlobalError = (error: Error, errorInfo: any) => {
    log.error('全局错误捕获', {
      error: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack,
    }, 'global', 'App', error.stack);
  };

  return (
    <ErrorBoundary onError={handleGlobalError} componentName="App">
      <QueryProvider>
        <AuthProvider>
          <Router>
            <ConfigProvider>
              <AntApp>
                <React.Suspense fallback={<LoadingSpinner />}>
                  <div className="App min-h-screen bg-white text-gray-900">
                <Routes>
                  {/* 智能根路由 - 根据认证状态决定重定向 */}
                  <Route
                    path="/"
                    element={
                      <SystemRoute>
                        <ErrorBoundary componentName="SmartRootRoute">
                          <SmartRootRoute>
                            <ErrorBoundary componentName="Setup">
                              <Setup />
                            </ErrorBoundary>
                          </SmartRootRoute>
                        </ErrorBoundary>
                      </SystemRoute>
                    }
                  />

                  {/* Setup 页面路由 - 明确访问 setup 路径 */}
                  <Route
                    path="/setup"
                    element={
                      <SystemRoute>
                        <ErrorBoundary componentName="Setup">
                          <Setup />
                        </ErrorBoundary>
                      </SystemRoute>
                    }
                  />

                  {/* 认证路由 - 也需要系统初始化检查 */}
                  <Route
                    path="/login"
                    element={
                      <SystemRoute>
                        <ErrorBoundary componentName="Login">
                          <Login />
                        </ErrorBoundary>
                      </SystemRoute>
                    }
                  />
                  <Route
                    path="/register"
                    element={
                      <SystemRoute>
                        <ErrorBoundary componentName="Register">
                          <Register />
                        </ErrorBoundary>
                      </SystemRoute>
                    }
                  />

                  {/* 需要认证和系统初始化的路由 */}
                  <Route
                    path="/dashboard"
                    element={
                      <SystemRoute>
                        <ProtectedRoute>
                          <ErrorBoundary componentName="Dashboard">
                            <Dashboard />
                          </ErrorBoundary>
                        </ProtectedRoute>
                      </SystemRoute>
                    }
                  />
                  <Route
                    path="/flow"
                    element={
                      <SystemRoute>
                        <ProtectedRoute>
                          <ErrorBoundary componentName="Dashboard">
                            <Dashboard />
                          </ErrorBoundary>
                        </ProtectedRoute>
                      </SystemRoute>
                    }
                  />
                  <Route
                    path="/config"
                    element={
                      <SystemRoute>
                        <ErrorBoundary componentName="Config">
                          <Config />
                        </ErrorBoundary>
                      </SystemRoute>
                    }
                  />
                  <Route
                    path="/config-editor"
                    element={
                      <SystemRoute>
                        <ProtectedRoute>
                          <ErrorBoundary componentName="ConfigEditor">
                            <ConfigEditor />
                          </ErrorBoundary>
                        </ProtectedRoute>
                      </SystemRoute>
                    }
                  />

                  {/* 默认重定向 */}
                  <Route path="*" element={<Navigate to="/" replace />} />
                </Routes>
                <DevTools />
              </div>
                </React.Suspense>
              </AntApp>
            </ConfigProvider>
          </Router>
        </AuthProvider>
      </QueryProvider>
    </ErrorBoundary>
  );
};

export default App;
