import {
  BugOutlined,
  ExceptionOutlined,
  FileTextOutlined,
  HomeOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import {
  Button,
  Card,
  Collapse,
  Divider,
  Result,
  Space,
  Tag,
  Typography,
} from 'antd';
import type React from 'react';
import { Component, type ErrorInfo, type ReactNode } from 'react';
import { envConfig } from '../../utils/envConfig';

const { Text, Title, Paragraph } = Typography;
const { Panel } = Collapse;

// 错误边界状态接口
interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
  errorId: string | null;
  componentName: string | null;
  timestamp: string | null;
}

// 错误边界Props接口
interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
  componentName?: string;
  enableRetry?: boolean;
  enableDetails?: boolean;
  maxRetries?: number;
}

/**
 * React 错误边界组件
 * 捕获和处理 React 组件树中的错误，提供调试信息和恢复机制
 */
export class ErrorBoundary extends Component<
  ErrorBoundaryProps,
  ErrorBoundaryState
> {
  private retryCount: number = 0;
  private errorBoundaryId: string;

  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: null,
      componentName: props.componentName || null,
      timestamp: null,
    };

    // 为每个错误边界实例生成唯一ID
    this.errorBoundaryId = `eb-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    // 更新 state 使下一次渲染能够显示降级后的 UI
    return {
      hasError: true,
      error,
      timestamp: new Date().toISOString(),
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // 记录错误信息
    const errorId = `error-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

    this.setState({
      error,
      errorInfo,
      errorId,
    });

    // 调用外部错误处理函数
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }

    // 在开发环境输出详细错误信息
    if (envConfig.isDevelopment) {
      console.group('🚨 React Error Boundary - 组件错误捕获');
      console.error('错误ID:', errorId);
      console.error('组件名:', this.props.componentName || '未知组件');
      console.error('错误信息:', error);
      console.error('错误堆栈:', errorInfo.componentStack);
      console.groupEnd();
    }

    // 上报错误到服务器（如果在生产环境启用）
    if (envConfig.enableErrorReporting && envConfig.isProduction) {
      this.reportError(error, errorInfo, errorId);
    }
  }

  // 错误上报函数
  private reportError = async (
    error: Error,
    errorInfo: ErrorInfo,
    errorId: string,
  ) => {
    try {
      const errorData = {
        id: errorId,
        message: error.message,
        stack: error.stack,
        componentStack: errorInfo.componentStack,
        componentName: this.props.componentName,
        timestamp: new Date().toISOString(),
        userAgent: envConfig.errorIncludeUserAgent
          ? navigator.userAgent
          : undefined,
        url: window.location.href,
        errorBoundaryId: this.errorBoundaryId,
      };

      // 发送错误报告到服务器
      if (envConfig.errorReportUrl) {
        await fetch(envConfig.errorReportUrl, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(errorData),
        });
      }
    } catch (reportError) {
      console.warn('错误上报失败:', reportError);
    }
  };

  // 重试处理
  private handleRetry = () => {
    const { maxRetries = 3 } = this.props;

    if (this.retryCount < maxRetries) {
      this.retryCount++;
      this.setState({
        hasError: false,
        error: null,
        errorInfo: null,
        errorId: null,
        timestamp: null,
      });

      // 延迟重试以避免立即重复错误
      setTimeout(() => {
        this.forceUpdate();
      }, 100);
    } else {
      console.warn(`错误重试次数已达上限 (${maxRetries}次)`);
    }
  };

  // 刷新页面
  private handleRefresh = () => {
    window.location.reload();
  };

  // 返回首页
  private handleGoHome = () => {
    window.location.href = '/';
  };

  // 复制错误信息
  private handleCopyError = () => {
    const { error, errorInfo, errorId, componentName } = this.state;

    const errorText = [
      `错误ID: ${errorId}`,
      `组件: ${componentName || '未知组件'}`,
      `时间: ${this.state.timestamp}`,
      `错误信息: ${error?.message}`,
      `错误堆栈: ${error?.stack}`,
      `组件堆栈: ${errorInfo?.componentStack}`,
    ].join('\n\n');

    navigator.clipboard.writeText(errorText).then(() => {
      // 这里可以添加复制成功的提示
      if (envConfig.isDevelopment) {
        console.log('✅ 错误信息已复制到剪贴板');
      }
    });
  };

  render() {
    const { hasError, error, errorInfo, errorId, componentName, timestamp } =
      this.state;
    const {
      children,
      fallback,
      enableRetry = true,
      enableDetails = envConfig.isDevelopment,
    } = this.props;

    // 如果有自定义 fallback，优先使用
    if (hasError && fallback) {
      return fallback;
    }

    // 错误状态显示
    if (hasError && error) {
      return (
        <div
          style={{
            padding: '40px 20px',
            minHeight: '400px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: '#fafafa',
          }}
        >
          <Card
            style={{
              maxWidth: 800,
              width: '100%',
              boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
              borderRadius: '8px',
            }}
          >
            <Result
              status="error"
              icon={<BugOutlined style={{ color: '#ff4d4f' }} />}
              title={
                <Space orientation="vertical" size="small">
                  <Title level={3} style={{ color: '#ff4d4f', margin: 0 }}>
                    组件渲染错误
                  </Title>
                  <Space>
                    <Tag color="red">ID: {errorId}</Tag>
                    {componentName && <Tag color="blue">{componentName}</Tag>}
                    <Tag color="gray">{timestamp}</Tag>
                  </Space>
                </Space>
              }
              subTitle={
                <Paragraph>
                  <Text type="secondary">
                    {envConfig.isDevelopment
                      ? '组件发生了错误，请查看详细信息进行调试。'
                      : '抱歉，应用程序遇到了意外错误，请尝试刷新页面。'}
                  </Text>
                </Paragraph>
              }
              extra={
                <Space wrap>
                  {enableRetry && (
                    <Button
                      type="primary"
                      icon={<ReloadOutlined />}
                      onClick={this.handleRetry}
                    >
                      重试组件
                    </Button>
                  )}
                  <Button
                    icon={<ReloadOutlined />}
                    onClick={this.handleRefresh}
                  >
                    刷新页面
                  </Button>
                  <Button icon={<HomeOutlined />} onClick={this.handleGoHome}>
                    返回首页
                  </Button>
                  {envConfig.isDevelopment && (
                    <Button
                      icon={<FileTextOutlined />}
                      onClick={this.handleCopyError}
                      type="dashed"
                    >
                      复制错误信息
                    </Button>
                  )}
                </Space>
              }
            />

            {enableDetails && (
              <>
                <Divider />
                <Collapse ghost>
                  <Panel
                    header={
                      <Space>
                        <ExceptionOutlined />
                        <Text strong>错误详情</Text>
                        <Tag color="orange" size="small">
                          开发者模式
                        </Tag>
                      </Space>
                    }
                    key="error-details"
                  >
                    <Space orientation="vertical" style={{ width: '100%' }}>
                      {/* 错误信息 */}
                      <Card size="small" title={<Text strong>错误信息</Text>}>
                        <Text
                          code
                          style={{
                            whiteSpace: 'pre-wrap',
                            wordBreak: 'break-all',
                            fontSize: '12px',
                            color: '#ff4d4f',
                          }}
                        >
                          {error.message}
                        </Text>
                      </Card>

                      {/* 错误堆栈 */}
                      {error.stack && (
                        <Card size="small" title={<Text strong>错误堆栈</Text>}>
                          <Text
                            code
                            style={{
                              whiteSpace: 'pre-wrap',
                              fontSize: '11px',
                              fontFamily:
                                'Monaco, Menlo, "Ubuntu Mono", monospace',
                            }}
                          >
                            {error.stack}
                          </Text>
                        </Card>
                      )}

                      {/* 组件堆栈 */}
                      {errorInfo?.componentStack && (
                        <Card size="small" title={<Text strong>组件堆栈</Text>}>
                          <Text
                            code
                            style={{
                              whiteSpace: 'pre-wrap',
                              fontSize: '11px',
                              fontFamily:
                                'Monaco, Menlo, "Ubuntu Mono", monospace',
                            }}
                          >
                            {errorInfo.componentStack}
                          </Text>
                        </Card>
                      )}

                      {/* 环境信息 */}
                      <Card size="small" title={<Text strong>环境信息</Text>}>
                        <Space orientation="vertical" style={{ width: '100%' }}>
                          <Text>
                            <strong>用户代理:</strong> {navigator.userAgent}
                          </Text>
                          <Text>
                            <strong>当前URL:</strong> {window.location.href}
                          </Text>
                          <Text>
                            <strong>错误边界ID:</strong> {this.errorBoundaryId}
                          </Text>
                          <Text>
                            <strong>重试次数:</strong> {this.retryCount}
                          </Text>
                        </Space>
                      </Card>
                    </Space>
                  </Panel>
                </Collapse>
              </>
            )}
          </Card>
        </div>
      );
    }

    // 正常渲染子组件
    return children;
  }
}

// 默认错误边界组件
export const DefaultErrorBoundary: React.FC<{ children: ReactNode }> = ({
  children,
}) => <ErrorBoundary>{children}</ErrorBoundary>;

// 高阶组件：为组件添加错误边界
export const withErrorBoundary = <P extends object>(
  Component: React.ComponentType<P>,
  errorBoundaryProps?: Omit<ErrorBoundaryProps, 'children'>,
) => {
  const WrappedComponent = (props: P) => (
    <ErrorBoundary {...errorBoundaryProps}>
      <Component {...props} />
    </ErrorBoundary>
  );

  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`;

  return WrappedComponent;
};

export default ErrorBoundary;
