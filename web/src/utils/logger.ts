/**
 * 统一日志工具
 * 提供结构化日志记录、性能监控、错误追踪等功能
 */

import { envConfig } from './envConfig';

// 日志级别枚举
export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
}

// 日志级别名称映射
const LogLevelNames: Record<LogLevel, string> = {
  [LogLevel.DEBUG]: 'DEBUG',
  [LogLevel.INFO]: 'INFO',
  [LogLevel.WARN]: 'WARN',
  [LogLevel.ERROR]: 'ERROR',
};

// 日志级别颜色映射（控制台输出用）
const LogLevelColors: Record<LogLevel, string> = {
  [LogLevel.DEBUG]: '#95a5a6', // 灰色
  [LogLevel.INFO]: '#3498db',  // 蓝色
  [LogLevel.WARN]: '#f39c12',  // 橙色
  [LogLevel.ERROR]: '#e74c3c', // 红色
};

// 日志条目接口
export interface LogEntry {
  id: string;
  timestamp: string;
  level: LogLevel;
  message: string;
  data?: any;
  category?: string;
  source?: string;
  userId?: string;
  sessionId?: string;
  stack?: string;
  performance?: {
    duration?: number;
    memory?: number;
    type?: string;
  };
}

// 日志监听器接口
export interface LogListener {
  (entry: LogEntry): void;
}

// 性能监控接口
export interface PerformanceMetric {
  name: string;
  value: number;
  unit: 'ms' | 'bytes' | 'count' | 'percentage';
  category?: string;
  tags?: Record<string, string>;
}

/**
 * 统一日志管理器
 */
class Logger {
  private logs: LogEntry[] = [];
  private listeners: LogListener[] = [];
  private sessionId: string;
  private userId: string | null = null;
  private isEnabled: boolean = true;
  private currentLogLevel: LogLevel = LogLevel.DEBUG;
  private maxLogEntries: number = 1000;

  constructor() {
    // 初始化会话ID
    this.sessionId = `session-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

    // 根据环境配置初始化
    this.setupFromEnvConfig();

    // 设置全局错误监听
    this.setupGlobalErrorHandling();

    // 设置性能监控
    if (envConfig.enablePerformanceMonitoring) {
      this.setupPerformanceMonitoring();
    }

    // 清理旧日志
    this.setupLogCleanup();
  }

  // 根据环境配置设置日志器
  private setupFromEnvConfig() {
    this.isEnabled = envConfig.enableConsoleLog || envConfig.isDevelopment;
    this.maxLogEntries = envConfig.logMaxEntries;

    // 设置日志级别
    switch (envConfig.logLevel) {
      case 'debug':
        this.currentLogLevel = LogLevel.DEBUG;
        break;
      case 'info':
        this.currentLogLevel = LogLevel.INFO;
        break;
      case 'warn':
        this.currentLogLevel = LogLevel.WARN;
        break;
      case 'error':
        this.currentLogLevel = LogLevel.ERROR;
        break;
      default:
        this.currentLogLevel = LogLevel.DEBUG;
    }

    // 从本地存储恢复日志
    this.restoreLogsFromStorage();
  }

  // 设置全局错误处理
  private setupGlobalErrorHandling() {
    // 监听未捕获的 JavaScript 错误
    window.addEventListener('error', (event) => {
      // 过滤掉已知无害的错误
      const ignoredErrors = [
        'ResizeObserver loop completed with undelivered notifications',
        'React DevTools',
        'Download the React DevTools',
        'Immersion Translate ERROR: UnknownError: Model not available'
      ];

      if (ignoredErrors.some(ignored => event.message && event.message.includes(ignored))) {
        // 这些是已知的无害错误或开发工具信息，忽略它们
        return;
      }

      // 过滤 React Router 的未来版本警告
      if (event.message && event.message.includes('React Router Future Flag Warning')) {
        return;
      }

      this.error('未捕获的 JavaScript 错误', {
        message: event.message,
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
        stack: event.error?.stack,
      }, 'JavaScript');
    });

    // 监听 Promise 拒绝
    window.addEventListener('unhandledrejection', (event) => {
      this.error('未处理的 Promise 拒绝', {
        reason: event.reason,
        stack: event.reason?.stack,
      }, 'Promise');
    });

    // 重写 console 方法来过滤开发日志
    this.filterConsoleLogs();
  }

  // 过滤控制台日志
  private filterConsoleLogs() {
    const originalLog = console.log;
    const originalWarn = console.warn;
    const originalInfo = console.info;

    console.log = (...args: any[]) => {
      const message = args.join(' ');

      // 过滤掉 Storage 操作的详细日志
      if (message.includes('Storage:') &&
          (message.includes('called, result:') ||
           message.includes('completed') ||
           message.includes('called, token length:') ||
           message.includes('called, username:') ||
           message.includes('called, expiresAt:') ||
           message.includes('removeToken called') ||
           message.includes('removeUser called') ||
           message.includes('removeExpiresAt called') ||
           message.includes('clear called'))) {
        return;
      }

      // 过滤掉 AuthContext 的详细调试日志
      if (message.includes('AuthContext:') &&
          (message.includes('saveAuthData called') ||
           message.includes('saveAuthData completed') ||
           message.includes('Login successful') ||
           message.includes('Saving auth data') ||
           message.includes('Verifying stored data') ||
           message.includes('Login state updated') ||
           message.includes('Found token in storage') ||
           message.includes('checkAuth called') ||
           message.includes('Stored auth data') ||
           message.includes('Server validation successful') ||
           message.includes('Using cached auth data') ||
           message.includes('Initializing auth on mount'))) {
        return;
      }

      // 过滤掉 Login 组件的重复 useEffect 日志
      if (message.includes('Login useEffect:')) {
        return;
      }

      // 过滤掉 React 的开发工具提示
      if (message.includes('Download the React DevTools')) {
        return;
      }

      // 过滤掉 WebSocket 连接日志
      if (message.includes('[rsbuild] WebSocket')) {
        return;
      }

      // 调用原始方法
      return originalLog.apply(console, args);
    };

    console.info = (...args: any[]) => {
      const message = args.join(' ');

      // 过滤掉系统初始化器的重复日志
      if (message.includes('[system]') &&
          (message.includes('系统已初始化，允许访问仪表板') ||
           message.includes('系统已初始化，允许访问页面'))) {
        return;
      }

      // 过滤掉 API 调用的重复成功日志
      if (message.includes('[api]') &&
          (message.includes('✅ GET /auth/me') ||
           message.includes('✅ GET /admin/system/status') ||
           message.includes('✅ GET /admin/database/schema'))) {
        return;
      }

      // 调用原始方法
      return originalInfo.apply(console, args);
    };

    console.warn = (...args: any[]) => {
      const message = args.join(' ');

      // 过滤掉各种开发警告
      const ignoredWarnings = [
        'React Router Future Flag Warning',
        'React Flow',
        'It looks like you\'ve created a new nodeTypes or edgeTypes object'
      ];

      if (ignoredWarnings.some(warning => message.includes(warning))) {
        return;
      }

      // 调用原始方法
      return originalWarn.apply(console, args);
    };
  }

  // 设置性能监控
  private setupPerformanceMonitoring() {
    if (!envConfig.enablePerformanceApi) return;

    // 监听页面加载性能
    window.addEventListener('load', () => {
      setTimeout(() => {
        this.recordPageLoadMetrics();
      }, 0);
    });

    // 定期记录内存使用情况（如果支持）
    if ((performance as any).memory) {
      setInterval(() => {
        this.recordMemoryUsage();
      }, envConfig.performanceReportInterval);
    }
  }

  // 记录页面加载指标
  private recordPageLoadMetrics() {
    try {
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;

      if (navigation) {
        const metrics = {
          'page.load.domContentLoaded': navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
          'page.load.load': navigation.loadEventEnd - navigation.loadEventStart,
          'page.load.firstPaint': performance.getEntriesByType('paint')[0]?.startTime,
          'page.load.firstContentfulPaint': performance.getEntriesByType('paint')[1]?.startTime,
          'page.load.total': navigation.loadEventEnd - navigation.fetchStart,
        };

        Object.entries(metrics).forEach(([name, value]) => {
          if (value && value > 0) {
            this.performance(name, value, 'ms', 'performance');
          }
        });
      }
    } catch (error) {
      this.warn('页面性能指标记录失败', error);
    }
  }

  // 记录内存使用情况
  private recordMemoryUsage() {
    try {
      const memory = (performance as any).memory;
      if (memory) {
        this.performance('memory.used', memory.usedJSHeapSize, 'bytes', 'memory');
        this.performance('memory.total', memory.totalJSHeapSize, 'bytes', 'memory');
        this.performance('memory.limit', memory.jsHeapSizeLimit, 'bytes', 'memory');
        this.performance('memory.usage', (memory.usedJSHeapSize / memory.jsHeapSizeLimit) * 100, 'percentage', 'memory');
      }
    } catch (error) {
      this.warn('内存使用记录失败', error);
    }
  }

  // 设置日志清理
  private setupLogCleanup() {
    setInterval(() => {
      this.cleanupOldLogs();
    }, 5 * 60 * 1000); // 每5分钟清理一次
  }

  // 清理旧日志
  private cleanupOldLogs() {
    if (this.logs.length > this.maxLogEntries) {
      const excessCount = this.logs.length - this.maxLogEntries;
      this.logs.splice(0, excessCount);
      this.saveLogsToStorage();
    }
  }

  // 生成日志ID
  private generateLogId(): string {
    return `log-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  // 创建日志条目
  private createLogEntry(
    level: LogLevel,
    message: string,
    data?: any,
    category?: string,
    source?: string,
    stack?: string
  ): LogEntry {
    return {
      id: this.generateLogId(),
      timestamp: new Date().toISOString(),
      level,
      message,
      data,
      category,
      source,
      userId: this.userId,
      sessionId: this.sessionId,
      stack,
    };
  }

  // 格式化日志消息
  private formatLogEntry(entry: LogEntry): string {
    const parts: string[] = [];

    // 时间戳
    if (envConfig.logIncludeTimestamp) {
      parts.push(`[${new Date(entry.timestamp).toLocaleString()}]`);
    }

    // 日志级别
    parts.push(`[${LogLevelNames[entry.level]}]`);

    // 分类
    if (entry.category) {
      parts.push(`[${entry.category}]`);
    }

    // 消息
    parts.push(entry.message);

    // 数据
    if (entry.data && typeof entry.data === 'object') {
      try {
        parts.push('\n', JSON.stringify(entry.data, null, 2));
      } catch (e) {
        parts.push('\n[对象无法序列化]');
      }
    } else if (entry.data) {
      parts.push('\n', String(entry.data));
    }

    return parts.join(' ');
  }

  // 输出到控制台
  private outputToConsole(entry: LogEntry) {
    if (!envConfig.enableConsoleLog || !this.shouldLog(entry.level)) {
      return;
    }

    const message = this.formatLogEntry(entry);
    const style = `color: ${LogLevelColors[entry.level]}; font-weight: bold;`;

    switch (entry.level) {
      case LogLevel.DEBUG:
        console.debug(`%c${message}`, style);
        break;
      case LogLevel.INFO:
        console.info(`%c${message}`, style);
        break;
      case LogLevel.WARN:
        console.warn(`%c${message}`, style);
        if (entry.data) console.warn('数据:', entry.data);
        break;
      case LogLevel.ERROR:
        console.error(`%c${message}`, style);
        if (entry.data) console.error('数据:', entry.data);
        if (entry.stack) console.error('堆栈:', entry.stack);
        break;
    }
  }

  // 检查是否应该记录日志
  private shouldLog(level: LogLevel): boolean {
    return this.isEnabled && level >= this.currentLogLevel;
  }

  // 添加日志条目
  private addLogEntry(entry: LogEntry) {
    // 输出到控制台
    this.outputToConsole(entry);

    // 添加到内存
    this.logs.push(entry);

    // 通知监听器
    this.listeners.forEach(listener => {
      try {
        listener(entry);
      } catch (error) {
        console.error('日志监听器错误:', error);
      }
    });

    // 保存到本地存储
    if (envConfig.logPersistToStorage) {
      this.saveLogsToStorage();
    }

    // 上报错误日志
    if (entry.level === LogLevel.ERROR && envConfig.enableErrorReporting && envConfig.errorReportUrl) {
      this.reportError(entry);
    }
  }

  // 保存日志到本地存储
  private saveLogsToStorage() {
    try {
      if (!envConfig.logPersistToStorage) return;

      const recentLogs = this.logs.slice(-this.maxLogEntries);
      localStorage.setItem('xiaozhi-logs', JSON.stringify(recentLogs));
    } catch (error) {
      console.warn('保存日志到本地存储失败:', error);
    }
  }

  // 从本地存储恢复日志
  private restoreLogsFromStorage() {
    try {
      const storedLogs = localStorage.getItem('xiaozhi-logs');
      if (storedLogs) {
        const logs = JSON.parse(storedLogs) as LogEntry[];
        this.logs = logs.slice(-this.maxLogEntries); // 只保留最近的日志
      }
    } catch (error) {
      console.warn('从本地存储恢复日志失败:', error);
    }
  }

  // 上报错误
  private async reportError(entry: LogEntry) {
    try {
      if (!envConfig.errorReportUrl) return;

      const errorData = {
        id: entry.id,
        timestamp: entry.timestamp,
        message: entry.message,
        level: LogLevelNames[entry.level],
        data: entry.data,
        category: entry.category,
        source: entry.source,
        stack: entry.stack,
        userId: entry.userId,
        sessionId: entry.sessionId,
        userAgent: envConfig.errorIncludeUserAgent ? navigator.userAgent : undefined,
        url: window.location.href,
      };

      await fetch(envConfig.errorReportUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(errorData),
      });
    } catch (error) {
      console.warn('错误上报失败:', error);
    }
  }

  // === 公共 API ===

  // 美化的 API 响应日志
  apiResponse(method: string, url: string, status: number, duration: number, data?: any) {
    if (!this.shouldLog(LogLevel.INFO)) return;

    const statusIcon = status >= 200 && status < 300 ? '✅' : status >= 400 ? '❌' : '⚠️';

    // 美化响应数据
    let beautifiedData = '';
    if (data) {
      if (typeof data === 'object') {
        if (data.success !== undefined) {
          beautifiedData = `Status: ${data.success ? 'SUCCESS' : 'FAILED'}`;
          if (data.message) beautifiedData += ` | Message: ${data.message}`;
          if (data.code) beautifiedData += ` | Code: ${data.code}`;
        } else if (data.total !== undefined) {
          beautifiedData = `Total: ${data.total} items`;
        } else {
          beautifiedData = `${Object.keys(data).length} data fields`;
        }
      } else {
        beautifiedData = String(data);
      }
    }

    const message = `${statusIcon} ${method} ${url} - ${status} (${duration}ms)${beautifiedData ? ` | ${beautifiedData}` : ''}`;

    const entry = this.createLogEntry(LogLevel.INFO, message, {
      method,
      url,
      status,
      duration,
      data: this.sanitizeApiData(data)
    }, 'api', 'ApiService');

    this.addLogEntry(entry);
  }

  // 清理 API 数据以便日志记录
  private sanitizeApiData(data: any): any {
    if (!data || typeof data !== 'object') return data;

    try {
      // 移除敏感字段和过长的内容
      const sanitized = { ...data };

      // 移除可能的敏感字段
      const sensitiveFields = ['password', 'token', 'secret', 'key', 'auth'];
      sensitiveFields.forEach(field => {
        if (sanitized[field]) {
          sanitized[field] = '[REDACTED]';
        }
      });

      // 如果数据太大，只显示摘要
      const jsonStr = JSON.stringify(sanitized);
      if (jsonStr.length > 1000) {
        return '[Large data object - ' + jsonStr.length + ' characters]';
      }

      return sanitized;
    } catch {
      return '[Unserializable data]';
    }
  }

  debug(message: string, data?: any, category?: string, source?: string) {
    if (!this.shouldLog(LogLevel.DEBUG)) return;
    const entry = this.createLogEntry(LogLevel.DEBUG, message, data, category, source);
    this.addLogEntry(entry);
  }

  info(message: string, data?: any, category?: string, source?: string) {
    if (!this.shouldLog(LogLevel.INFO)) return;
    const entry = this.createLogEntry(LogLevel.INFO, message, data, category, source);
    this.addLogEntry(entry);
  }

  warn(message: string, data?: any, category?: string, source?: string) {
    if (!this.shouldLog(LogLevel.WARN)) return;
    const entry = this.createLogEntry(LogLevel.WARN, message, data, category, source);
    this.addLogEntry(entry);
  }

  error(message: string, data?: any, category?: string, source?: string, stack?: string) {
    if (!this.shouldLog(LogLevel.ERROR)) return;
    const entry = this.createLogEntry(LogLevel.ERROR, message, data, category, source, stack);
    this.addLogEntry(entry);
  }

  // 性能记录
  performance(name: string, value: number, unit: 'ms' | 'bytes' | 'count' | 'percentage' = 'ms', category?: string, tags?: Record<string, string>) {
    if (!envConfig.enablePerformanceMonitoring) return;

    const metric: PerformanceMetric = {
      name,
      value,
      unit,
      category,
      tags,
    };

    this.debug(`性能指标: ${name}`, metric, 'performance');

    // 如果启用了性能上报，可以在这里添加上报逻辑
  }

  // 时间测量
  time(label: string, category?: string) {
    if (!this.shouldLog(LogLevel.DEBUG)) return;
    console.time(label);
  }

  timeEnd(label: string, category?: string) {
    if (!this.shouldLog(LogLevel.DEBUG)) return;
    console.timeEnd(label);

    // 尝试获取测量结果并记录
    try {
      const measurements = performance.getEntriesByName(label, 'measure');
      if (measurements.length > 0) {
        const latest = measurements[measurements.length - 1];
        this.performance(`timer.${label}`, latest.duration, 'ms', category);
      }
    } catch (error) {
      // 忽略错误
    }
  }

  // 设置用户ID
  setUserId(userId: string) {
    this.userId = userId;
    this.info(`设置用户ID: ${userId}`, null, 'auth');
  }

  // 设置日志级别
  setLogLevel(level: LogLevel) {
    this.currentLogLevel = level;
    this.info(`设置日志级别: ${LogLevelNames[level]}`, null, 'logger');
  }

  // 启用/禁用日志
  setEnabled(enabled: boolean) {
    this.isEnabled = enabled;
    this.info(`日志${enabled ? '启用' : '禁用'}`, null, 'logger');
  }

  // 添加日志监听器
  addListener(listener: LogListener) {
    this.listeners.push(listener);
  }

  // 移除日志监听器
  removeListener(listener: LogListener) {
    const index = this.listeners.indexOf(listener);
    if (index > -1) {
      this.listeners.splice(index, 1);
    }
  }

  // 获取所有日志
  getLogs(): LogEntry[] {
    return [...this.logs];
  }

  // 获取按级别过滤的日志
  getLogsByLevel(level: LogLevel): LogEntry[] {
    return this.logs.filter(log => log.level === level);
  }

  // 获取按分类过滤的日志
  getLogsByCategory(category: string): LogEntry[] {
    return this.logs.filter(log => log.category === category);
  }

  // 清空日志
  clearLogs() {
    this.logs = [];
    localStorage.removeItem('xiaozhi-logs');
    this.info('日志已清空', null, 'logger');
  }

  // 导出日志
  exportLogs(): string {
    return JSON.stringify(this.logs, null, 2);
  }

  // 获取统计信息
  getStats() {
    const stats = {
      total: this.logs.length,
      debug: 0,
      info: 0,
      warn: 0,
      error: 0,
      categories: {} as Record<string, number>,
    };

    this.logs.forEach(log => {
      switch (log.level) {
        case LogLevel.DEBUG:
          stats.debug++;
          break;
        case LogLevel.INFO:
          stats.info++;
          break;
        case LogLevel.WARN:
          stats.warn++;
          break;
        case LogLevel.ERROR:
          stats.error++;
          break;
      }

      if (log.category) {
        stats.categories[log.category] = (stats.categories[log.category] || 0) + 1;
      }
    });

    return stats;
  }
}

// 创建全局日志实例
export const logger = new Logger();

// 导出便捷方法
export const log = {
  debug: (message: string, data?: any, category?: string, source?: string) => logger.debug(message, data, category, source),
  info: (message: string, data?: any, category?: string, source?: string) => logger.info(message, data, category, source),
  warn: (message: string, data?: any, category?: string, source?: string) => logger.warn(message, data, category, source),
  error: (message: string, data?: any, category?: string, source?: string, stack?: string) => logger.error(message, data, category, source, stack),
  performance: (name: string, value: number, unit?: 'ms' | 'bytes' | 'count' | 'percentage', category?: string, tags?: Record<string, string>) => logger.performance(name, value, unit, category, tags),
  time: (label: string, category?: string) => logger.time(label, category),
  timeEnd: (label: string, category?: string) => logger.timeEnd(label, category),
  apiResponse: (method: string, url: string, status: number, duration: number, data?: any) => logger.apiResponse(method, url, status, duration, data),
};

export default logger;