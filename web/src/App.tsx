import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { theme } from 'antd';
import './App.css';
import { QueryProvider } from './components/QueryProvider';
import SystemInitializer from './components/SystemInitializer';
import { ErrorBoundary } from './components/ErrorBoundary';
import DevTools from './components/DevTools';

// 页面组件
const Setup = React.lazy(() => import('./pages/Setup'));
const Dashboard = React.lazy(() => import('./pages/Dashboard'));
const Config = React.lazy(() => import('./pages/Config'));
const ConfigEditor = React.lazy(() => import('./pages/ConfigEditor'));
const Login = React.lazy(() => import('./pages/Login'));

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
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => (
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
        <Router>
          <React.Suspense fallback={<LoadingSpinner />}>
            <div className="App min-h-screen bg-white text-gray-900">
              <Routes>
                <Route path="/" element={<SystemInitializer><Setup /></SystemInitializer>} />
                <Route path="/setup" element={<SystemInitializer><Setup /></SystemInitializer>} />
                <Route
                  path="/dashboard"
                  element={
                    <SystemInitializer>
                      <ErrorBoundary componentName="Dashboard">
                        <Dashboard />
                      </ErrorBoundary>
                    </SystemInitializer>
                  }
                />
                <Route
                  path="/flow"
                  element={
                    <SystemInitializer>
                      <ErrorBoundary componentName="Dashboard">
                        <Dashboard />
                      </ErrorBoundary>
                    </SystemInitializer>
                  }
                />
                <Route
                  path="/config"
                  element={
                    <SystemInitializer>
                      <ErrorBoundary componentName="Config">
                        <Config />
                      </ErrorBoundary>
                    </SystemInitializer>
                  }
                />
                <Route
                  path="/config-editor"
                  element={
                    <SystemInitializer>
                      <ErrorBoundary componentName="ConfigEditor">
                        <ConfigEditor />
                      </ErrorBoundary>
                    </SystemInitializer>
                  }
                />
                <Route
                  path="/login"
                  element={
                    <SystemInitializer>
                      <ErrorBoundary componentName="Login">
                        <Login />
                      </ErrorBoundary>
                    </SystemInitializer>
                  }
                />
              </Routes>
              <DevTools />
            </div>
          </React.Suspense>
        </Router>
      </QueryProvider>
    </ErrorBoundary>
  );
};

export default App;
