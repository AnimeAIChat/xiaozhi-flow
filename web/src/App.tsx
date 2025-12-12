import { App as AntApp, ConfigProvider } from 'antd';
import React from 'react';
import { Route, BrowserRouter as Router, Routes } from 'react-router-dom';
import './App.css';
import DevTools from './components/DevTools';
import { ErrorBoundary } from './components/ErrorBoundary';
import { QueryProvider } from './components/QueryProvider';
import { log } from './utils/logger';

// 页面组件
const Dashboard = React.lazy(() => import('./pages/Dashboard'));

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

const App: React.FC = () => {
  // 全局错误处理函数
  const handleGlobalError = (error: Error, errorInfo: any) => {
    log.error(
      '全局错误捕获',
      {
        error: error.message,
        stack: error.stack,
        componentStack: errorInfo.componentStack,
      },
      'global',
      'App',
      error.stack,
    );
  };

  return (
    <ErrorBoundary onError={handleGlobalError} componentName="App">
      <QueryProvider>
        <Router>
          <ConfigProvider>
            <AntApp>
              <React.Suspense fallback={<LoadingSpinner />}>
                <div className="App min-h-screen bg-white text-gray-900">
                  <Routes>
                    {/* 根路径和 dashboard 都指向主页 */}
                    <Route
                      path="/"
                      element={
                        <ErrorBoundary componentName="Dashboard">
                          <Dashboard />
                        </ErrorBoundary>
                      }
                    />
                    <Route
                      path="/dashboard"
                      element={
                        <ErrorBoundary componentName="Dashboard">
                          <Dashboard />
                        </ErrorBoundary>
                      }
                    />
                    <Route
                      path="/flow"
                      element={
                        <ErrorBoundary componentName="Dashboard">
                          <Dashboard />
                        </ErrorBoundary>
                      }
                    />
                  </Routes>
                  <DevTools />
                </div>
              </React.Suspense>
            </AntApp>
          </ConfigProvider>
        </Router>
      </QueryProvider>
    </ErrorBoundary>
  );
};

export default App;
