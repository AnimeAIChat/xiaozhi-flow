/**
 * 环境配置工具
 * 统一管理所有环境变量和调试配置
 */

// 获取环境变量值的辅助函数
const getEnvVar = (key: string, defaultValue: string = ''): string => {
  if (typeof import.meta !== 'undefined' && import.meta.env) {
    return import.meta.env[key] || defaultValue;
  }
  return process.env?.[key] || defaultValue;
};

// 获取布尔型环境变量
const getEnvBoolean = (key: string, defaultValue: boolean = false): boolean => {
  const value = getEnvVar(key).toLowerCase();
  if (value === 'true' || value === '1') return true;
  if (value === 'false' || value === '0') return false;
  return defaultValue;
};

// 获取数字型环境变量
const getEnvNumber = (key: string, defaultValue: number = 0): number => {
  const value = getEnvVar(key);
  const num = parseInt(value, 10);
  return isNaN(num) ? defaultValue : num;
};

// 环境配置接口
interface EnvConfig {
  // 基础环境配置
  nodeEnv: string;
  appEnv: string;
  isDevelopment: boolean;
  isProduction: boolean;
  isTest: boolean;
  debug: boolean;

  // API 配置
  apiBaseUrl: string;
  apiTimeout: number;

  // 调试配置
  enableSourceMap: boolean;
  enableConsoleLog: boolean;
  enablePerformanceMonitoring: boolean;
  enableErrorBoundary: boolean;
  enableApiDebugging: boolean;

  // 日志配置
  logLevel: 'debug' | 'info' | 'warn' | 'error';
  logMaxEntries: number;
  logPersistToStorage: boolean;
  logIncludeTimestamp: boolean;

  // 错误报告配置
  enableErrorReporting: boolean;
  errorReportUrl: string;
  errorIncludeUserAgent: boolean;
  errorIncludeUserData: boolean;

  // 性能监控配置
  enablePerformanceApi: boolean;
  performanceSampleRate: number;
  performanceReportInterval: number;

  // 开发工具配置
  enableDevtools: boolean;
  enableComponentInspector: boolean;
  enableStateInspector: boolean;
  enableReactDevtools: boolean;

  // 特性开关
  enableMockData: boolean;
  enableExperimentalFeatures: boolean;

  // 开发服务器配置
  devServerPort: number;
  devServerHost: string;
  devServerHttps: boolean;
  devServerOpen: boolean;

  // 构建配置
  buildSourceMap: boolean;
  buildMinify: boolean;
  buildTarget: string;
  buildCssCodeSplit: boolean;

  // 安全配置
  enableCors: boolean;
  corsOrigin: string;
  contentSecurityPolicy: boolean;
}

// 创建环境配置对象
export const envConfig: EnvConfig = {
  // 基础环境配置
  nodeEnv: getEnvVar('NODE_ENV', 'development'),
  appEnv: getEnvVar('VITE_APP_ENV', 'development'),
  isDevelopment: getEnvVar('NODE_ENV') === 'development',
  isProduction: getEnvVar('NODE_ENV') === 'production',
  isTest: getEnvVar('NODE_ENV') === 'test',
  debug: getEnvBoolean('VITE_APP_DEBUG', true),

  // API 配置
  apiBaseUrl: getEnvVar('VITE_API_BASE_URL', 'http://localhost:8080/api'),
  apiTimeout: getEnvNumber('VITE_API_TIMEOUT', 10000),

  // 调试配置
  enableSourceMap: getEnvBoolean('VITE_ENABLE_SOURCE_MAP', true),
  enableConsoleLog: getEnvBoolean('VITE_ENABLE_CONSOLE_LOG', false), // 减少控制台日志
  enablePerformanceMonitoring: getEnvBoolean('VITE_ENABLE_PERFORMANCE_MONITORING', false), // 关闭性能监控
  enableErrorBoundary: getEnvBoolean('VITE_ENABLE_ERROR_BOUNDARY', true),
  enableApiDebugging: getEnvBoolean('VITE_ENABLE_API_DEBUGGING', false), // 关闭API调试

  // 日志配置
  logLevel: (getEnvVar('VITE_LOG_LEVEL', 'error') as 'debug' | 'info' | 'warn' | 'error'), // 提高日志级别
  logMaxEntries: getEnvNumber('VITE_LOG_MAX_ENTRIES', 500), // 减少日志条目
  logPersistToStorage: getEnvBoolean('VITE_LOG_PERSIST_TO_STORAGE', false), // 关闭持久化
  logIncludeTimestamp: getEnvBoolean('VITE_LOG_INCLUDE_TIMESTAMP', false), // 关闭时间戳

  // 错误报告配置
  enableErrorReporting: getEnvBoolean('VITE_ENABLE_ERROR_REPORTING', false),
  errorReportUrl: getEnvVar('VITE_ERROR_REPORT_URL', ''),
  errorIncludeUserAgent: getEnvBoolean('VITE_ERROR_INCLUDE_USER_AGENT', true),
  errorIncludeUserData: getEnvBoolean('VITE_ERROR_INCLUDE_USER_DATA', false),

  // 性能监控配置
  enablePerformanceApi: getEnvBoolean('VITE_ENABLE_PERFORMANCE_API', true),
  performanceSampleRate: getEnvNumber('VITE_PERFORMANCE_SAMPLE_RATE', 100) / 100,
  performanceReportInterval: getEnvNumber('VITE_PERFORMANCE_REPORT_INTERVAL', 30000),

  // 开发工具配置
  enableDevtools: getEnvBoolean('VITE_ENABLE_DEVTOOLS', true),
  enableComponentInspector: getEnvBoolean('VITE_ENABLE_COMPONENT_INSPECTOR', true),
  enableStateInspector: getEnvBoolean('VITE_ENABLE_STATE_INSPECTOR', true),
  enableReactDevtools: getEnvBoolean('VITE_ENABLE_REACT_DEVTOOLS', true),

  // 特性开关
  enableMockData: getEnvBoolean('VITE_ENABLE_MOCK_DATA', false),
  enableExperimentalFeatures: getEnvBoolean('VITE_ENABLE_EXPERIMENTAL_FEATURES', false),

  // 开发服务器配置
  devServerPort: getEnvNumber('VITE_DEV_SERVER_PORT', 3000),
  devServerHost: getEnvVar('VITE_DEV_SERVER_HOST', 'localhost'),
  devServerHttps: getEnvBoolean('VITE_DEV_SERVER_HTTPS', false),
  devServerOpen: getEnvBoolean('VITE_DEV_SERVER_OPEN', true),

  // 构建配置
  buildSourceMap: getEnvBoolean('VITE_BUILD_SOURCEMAP', true),
  buildMinify: getEnvBoolean('VITE_BUILD_MINIFY', false),
  buildTarget: getEnvVar('VITE_BUILD_TARGET', 'esnext'),
  buildCssCodeSplit: getEnvBoolean('VITE_BUILD_CSS_CODE_SPLIT', true),

  // 安全配置
  enableCors: getEnvBoolean('VITE_ENABLE_CORS', true),
  corsOrigin: getEnvVar('VITE_CORS_ORIGIN', 'http://localhost:3000'),
  contentSecurityPolicy: getEnvBoolean('VITE_CONTENT_SECURITY_POLICY', false),
};

// 验证环境配置的函数
export const validateEnvConfig = (): boolean => {
  try {
    // 验证 API 配置
    if (!envConfig.apiBaseUrl) {
      console.warn('⚠️ API Base URL 未配置');
      return false;
    }

    // 验证端口配置
    if (envConfig.devServerPort < 1 || envConfig.devServerPort > 65535) {
      console.warn('⚠️ 开发服务器端口配置无效');
      return false;
    }

    // 验证日志级别
    const validLogLevels = ['debug', 'info', 'warn', 'error'];
    if (!validLogLevels.includes(envConfig.logLevel)) {
      console.warn('⚠️ 日志级别配置无效');
      return false;
    }

    // 验证性能采样率
    if (envConfig.performanceSampleRate < 0 || envConfig.performanceSampleRate > 1) {
      console.warn('⚠️ 性能采样率配置无效');
      return false;
    }

    console.log('✅ 环境配置验证通过');
    return true;
  } catch (error) {
    console.error('❌ 环境配置验证失败:', error);
    return false;
  }
};

// 获取当前环境描述
export const getEnvironmentDescription = (): string => {
  const { appEnv, debug, apiBaseUrl } = envConfig;

  let description = `当前环境: ${appEnv}`;
  if (debug) description += ' (调试模式)';
  description += `\nAPI 地址: ${apiBaseUrl}`;

  return description;
};

// 检查是否应该启用某个调试功能
export const shouldEnable = (feature: keyof EnvConfig): boolean => {
  const value = envConfig[feature];
  return typeof value === 'boolean' ? value : !!value;
};

// 默认导出环境配置
export default envConfig;

// 类型守卫函数
export const isDebugMode = (): boolean => envConfig.debug && envConfig.isDevelopment;
export const isProductionMode = (): boolean => envConfig.isProduction;
export const shouldLogToConsole = (): boolean => envConfig.enableConsoleLog && envConfig.debug;
export const shouldReportErrors = (): boolean => envConfig.enableErrorReporting;