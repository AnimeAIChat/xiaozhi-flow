import axios, { AxiosInstance, AxiosResponse } from 'axios';
import { log } from '../utils/logger';
import { envConfig } from '../utils/envConfig';

// 基础API响应类型
export interface ApiResponse<T = any> {
  success: boolean;
  data: T;
  message: string;
  code: number;
}

// 服务器配置类型
export interface ServerConfig {
  host: string;
  port: number;
  protocol?: 'http' | 'https';
}

// 连接测试结果
export interface ConnectionTestResult {
  success: boolean;
  message: string;
  latency?: number;
  version?: string;
}

// 项目初始化配置
export interface InitConfig {
  serverConfig: ServerConfig;
  providers: {
    asr?: any;
    tts?: any;
    llm?: any;
    vllm?: any;
  };
  systemConfig?: any;
}

// 初始化结果
export interface InitResult {
  success: boolean;
  message: string;
  configId?: string;
  steps?: Array<{
    name: string;
    status: 'pending' | 'running' | 'completed' | 'failed';
    message?: string;
  }>;
}

// 提供商类型
export type ProviderType = 'asr' | 'tts' | 'llm' | 'vllm';

// 提供商配置
export interface ProviderConfig {
  id: string;
  name: string;
  type: ProviderType;
  enabled: boolean;
  config: Record<string, any>;
}

// 提供商测试结果
export interface ProviderTestResult {
  success: boolean;
  message: string;
  latency?: number;
  details?: any;
}

// 系统配置
export interface SystemConfig {
  server: {
    ip: string;
    port: number;
    token?: string;
    auth: boolean;
  };
  audio: {
    input_sample_rate: number;
    output_sample_rate: number;
    channels: number;
  };
  transport: {
    websocket: boolean;
    mqtt: boolean;
    udp: boolean;
  };
  [key: string]: any;
}

/**
 * API 调试信息接口
 */
interface ApiCallInfo {
  id: string;
  method: string;
  url: string;
  headers: any;
  params?: any;
  data?: any;
  timestamp: string;
  duration?: number;
  status?: number;
  response?: any;
  error?: any;
  category?: string;
}

/**
 * API服务类 - 封装所有后端API调用
 */
export class ApiService {
  private client: AxiosInstance;
  private baseURL: string;
  private apiCallHistory: ApiCallInfo[] = [];
  private maxHistorySize: number = 100;

  constructor(baseURL: string = 'http://localhost:8080/api') {
    this.baseURL = baseURL;
    this.client = axios.create({
      baseURL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // 请求拦截器
    this.client.interceptors.request.use(
      (config) => {
        const startTime = Date.now();

        // 添加请求开始时间到配置中
        (config as any).metadata = { startTime };

        // 在这里可以添加认证token
        const token = localStorage.getItem('auth_token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }

        // 创建API调用记录
        const callInfo: ApiCallInfo = {
          id: this.generateApiCallId(),
          method: config.method?.toUpperCase() || 'UNKNOWN',
          url: config.url || '',
          headers: config.headers,
          params: config.params,
          data: config.data,
          timestamp: new Date().toISOString(),
          category: this.getCategoryFromUrl(config.url || ''),
        };

        // 存储调用信息到配置中，以便在响应拦截器中使用
        (config as any).callInfo = callInfo;

        if (envConfig.enableApiDebugging) {
          log.debug(`API 请求: ${callInfo.method} ${callInfo.url}`, {
            id: callInfo.id,
            headers: this.sanitizeHeaders(config.headers),
            params: config.params,
            data: this.sanitizeData(config.data),
          }, 'api', 'ApiService');
        }

        return config;
      },
      (error) => {
        log.error('API 请求错误', error, 'api', 'ApiService', error.stack);
        return Promise.reject(error);
      }
    );

    // 响应拦截器
    this.client.interceptors.response.use(
      (response: AxiosResponse<ApiResponse>) => {
        const config = response.config as any;
        const callInfo = config.callInfo as ApiCallInfo;
        const endTime = Date.now();

        if (callInfo) {
          callInfo.duration = endTime - (config.metadata?.startTime || endTime);
          callInfo.status = response.status;
          callInfo.response = this.sanitizeResponse(response.data);

          // 添加到历史记录
          this.addToHistory(callInfo);
        }

        if (envConfig.enableApiDebugging) {
          log.info(`API 响应: ${response.config.method?.toUpperCase()} ${response.config.url}`, {
            id: callInfo?.id,
            status: response.status,
            duration: callInfo?.duration,
            data: this.sanitizeResponse(response.data),
          }, 'api', 'ApiService');

          // 记录性能指标
          if (callInfo?.duration) {
            log.performance(`api.${this.getCategoryFromUrl(response.config.url || '')}.response_time`, callInfo.duration, 'ms', 'api');
          }
        }

        return response;
      },
      (error) => {
        const config = error.config as any;
        const callInfo = config?.callInfo as ApiCallInfo;
        const endTime = Date.now();

        if (callInfo) {
          callInfo.duration = endTime - (config.metadata?.startTime || endTime);
          callInfo.status = error.response?.status;
          callInfo.error = this.sanitizeError(error);

          // 添加到历史记录
          this.addToHistory(callInfo);
        }

        log.error('API 响应错误', {
          id: callInfo?.id,
          method: callInfo?.method,
          url: callInfo?.url,
          status: error.response?.status,
          statusText: error.response?.statusText,
          message: error.message,
          duration: callInfo?.duration,
        }, 'api', 'ApiService', error.stack);

        // 记录错误性能指标
        if (callInfo?.duration) {
          log.performance(`api.error.response_time`, callInfo.duration, 'ms', 'api');
        }

        // 统一错误处理
        if (error.response) {
          // 服务器返回了错误状态码
          const { status, data } = error.response;
          throw new Error(data.message || `HTTP ${status} Error`);
        } else if (error.request) {
          // 请求发出但没有收到响应
          throw new Error('Network error - unable to connect to server');
        } else {
          // 请求配置错误
          throw new Error(error.message || 'Request configuration error');
        }
      }
    );
  }

  // === API 调试方法 ===

  // 生成API调用ID
  private generateApiCallId(): string {
    return `api-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  // 从URL获取分类
  private getCategoryFromUrl(url: string): string {
    const path = url.replace(/^\/api\//, '');
    const parts = path.split('/');
    return parts[0] || 'unknown';
  }

  // 清理敏感的头部信息
  private sanitizeHeaders(headers: any): any {
    if (!headers) return {};

    const sanitized = { ...headers };
    const sensitiveHeaders = ['authorization', 'token', 'cookie'];

    sensitiveHeaders.forEach(key => {
      if (sanitized[key]) {
        sanitized[key] = '[REDACTED]';
      }
    });

    return sanitized;
  }

  // 清理敏感的请求数据
  private sanitizeData(data: any): any {
    if (!data) return data;

    // 如果是对象，移除敏感字段
    if (typeof data === 'object') {
      const sanitized = { ...data };
      const sensitiveFields = ['password', 'token', 'secret', 'key'];

      sensitiveFields.forEach(field => {
        if (sanitized[field]) {
          sanitized[field] = '[REDACTED]';
        }
      });

      return sanitized;
    }

    return data;
  }

  // 清理响应数据
  private sanitizeResponse(data: any): any {
    if (!data) return data;

    // 只记录响应的关键信息，避免记录大量数据
    if (typeof data === 'object' && data.data) {
      return {
        success: data.success,
        message: data.message,
        code: data.code,
        hasData: !!data.data,
        dataKeys: Array.isArray(data.data) ? `Array[${data.data.length}]` : typeof data.data,
      };
    }

    return data;
  }

  // 清理错误信息
  private sanitizeError(error: any): any {
    return {
      message: error.message,
      status: error.response?.status,
      statusText: error.response?.statusText,
      code: error.code,
    };
  }

  // 添加到历史记录
  private addToHistory(callInfo: ApiCallInfo) {
    this.apiCallHistory.unshift(callInfo);

    // 保持历史记录大小限制
    if (this.apiCallHistory.length > this.maxHistorySize) {
      this.apiCallHistory = this.apiCallHistory.slice(0, this.maxHistorySize);
    }
  }

  // 获取API调用历史
  getApiHistory(): ApiCallInfo[] {
    return [...this.apiCallHistory];
  }

  // 获取按分类过滤的API历史
  getApiHistoryByCategory(category: string): ApiCallInfo[] {
    return this.apiCallHistory.filter(call => call.category === category);
  }

  // 获取错误调用历史
  getErrorHistory(): ApiCallInfo[] {
    return this.apiCallHistory.filter(call => call.error || (call.status && call.status >= 400));
  }

  // 获取性能统计
  getPerformanceStats() {
    const stats = {
      totalCalls: this.apiCallHistory.length,
      successCalls: 0,
      errorCalls: 0,
      averageResponseTime: 0,
      categories: {} as Record<string, { count: number; totalTime: number; errors: number }>,
      slowCalls: this.apiCallHistory.filter(call => call.duration && call.duration > 2000),
    };

    let totalResponseTime = 0;

    this.apiCallHistory.forEach(call => {
      const isError = call.error || (call.status && call.status >= 400);

      if (isError) {
        stats.errorCalls++;
      } else {
        stats.successCalls++;
      }

      if (call.duration) {
        totalResponseTime += call.duration;
      }

      // 分类统计
      if (call.category) {
        if (!stats.categories[call.category]) {
          stats.categories[call.category] = { count: 0, totalTime: 0, errors: 0 };
        }
        stats.categories[call.category].count++;
        if (call.duration) {
          stats.categories[call.category].totalTime += call.duration;
        }
        if (isError) {
          stats.categories[call.category].errors++;
        }
      }
    });

    stats.averageResponseTime = stats.totalCalls > 0 ? totalResponseTime / stats.totalCalls : 0;

    // 计算每个分类的平均响应时间
    Object.keys(stats.categories).forEach(category => {
      const cat = stats.categories[category];
      cat.totalTime = cat.count > 0 ? cat.totalTime / cat.count : 0;
    });

    return stats;
  }

  // 清除API历史
  clearApiHistory() {
    this.apiCallHistory = [];
    log.info('API调用历史已清除', null, 'api', 'ApiService');
  }

  // 导出API历史
  exportApiHistory(): string {
    return JSON.stringify({
      exportTime: new Date().toISOString(),
      totalCalls: this.apiCallHistory.length,
      calls: this.apiCallHistory,
      stats: this.getPerformanceStats(),
    }, null, 2);
  }

  // 重放API调用（仅开发环境）
  async replayApiCall(callId: string): Promise<any> {
    if (!envConfig.isDevelopment) {
      throw new Error('API重放功能仅在开发环境可用');
    }

    const originalCall = this.apiCallHistory.find(call => call.id === callId);
    if (!originalCall) {
      throw new Error(`找不到API调用记录: ${callId}`);
    }

    log.info(`重放API调用: ${originalCall.method} ${originalCall.url}`, { callId }, 'api', 'ApiService');

    try {
      const response = await this.client.request({
        method: originalCall.method.toLowerCase() as any,
        url: originalCall.url,
        params: originalCall.params,
        data: originalCall.data,
        headers: originalCall.headers,
      });

      return response.data;
    } catch (error) {
      log.error('API重放失败', { callId, error }, 'api', 'ApiService');
      throw error;
    }
  }

  /**
   * 测试服务器连接
   */
  async testConnection(config: ServerConfig): Promise<ConnectionTestResult> {
    try {
      const startTime = Date.now();
      const response = await this.client.post('/admin/system/test-connection', config);
      const latency = Date.now() - startTime;

      return {
        success: response.data.success,
        message: response.data.message,
        latency,
        version: response.data.data?.version,
      };
    } catch (error) {
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Connection test failed',
      };
    }
  }

  /**
   * 初始化项目
   */
  async initializeProject(config: InitConfig): Promise<InitResult> {
    try {
      const response = await this.client.post('/admin/system/init', config);
      return response.data.data;
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'Project initialization failed');
    }
  }

  /**
   * 获取系统状态
   */
  async getSystemStatus(): Promise<any> {
    try {
      const response = await this.client.get('/admin/system/status');
      return response.data.data;
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'Failed to get system status');
    }
  }

  /**
   * 获取系统日志
   */
  async getSystemLogs(level?: string): Promise<any> {
    try {
      const params = level ? { level } : {};
      const response = await this.client.get('/admin/system/logs', { params });
      return response.data.data;
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'Failed to get system logs');
    }
  }

  /**
   * 获取提供商列表
   */
  async getProviders(type?: ProviderType): Promise<ProviderConfig[]> {
    try {
      const params = type ? { type } : {};
      const response = await this.client.get('/admin/config/providers', { params });
      return response.data.data;
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'Failed to get providers');
    }
  }

  /**
   * 测试提供商连接
   */
  async testProvider(type: ProviderType, config: ProviderConfig): Promise<ProviderTestResult> {
    try {
      const startTime = Date.now();
      const response = await this.client.post('/admin/config/providers/test', {
        type,
        config,
      });
      const latency = Date.now() - startTime;

      return {
        success: response.data.success,
        message: response.data.message,
        latency,
        details: response.data.data,
      };
    } catch (error) {
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Provider test failed',
      };
    }
  }

  /**
   * 更新提供商配置
   */
  async updateProvider(type: ProviderType, config: ProviderConfig): Promise<void> {
    try {
      await this.client.put('/admin/config/providers', { type, config });
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'Failed to update provider');
    }
  }

  /**
   * 获取系统配置
   */
  async getSystemConfig(): Promise<SystemConfig> {
    try {
      const response = await this.client.get('/admin/config/system');
      return response.data.data;
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'Failed to get system config');
    }
  }

  /**
   * 更新系统配置
   */
  async updateSystemConfig(config: SystemConfig): Promise<void> {
    try {
      await this.client.put('/admin/config/system', config);
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'Failed to update system config');
    }
  }

  /**
   * 验证配置
   */
  async validateConfig(config: any): Promise<any> {
    try {
      const response = await this.client.post('/admin/config/validate', config);
      return response.data.data;
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'Configuration validation failed');
    }
  }

  /**
   * 获取数据库模式信息
   */
  async getDatabaseSchema(): Promise<any> {
    try {
      const response = await this.client.get('/admin/database/schema');
      return response.data.data;
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'Failed to get database schema');
    }
  }

  /**
   * 获取数据库表列表
   */
  async getDatabaseTables(): Promise<any> {
    try {
      const response = await this.client.get('/admin/database/tables');
      return response.data.data;
    } catch (error) {
      throw new Error(error instanceof Error ? error.message : 'Failed to get database tables');
    }
  }

  /**
   * 获取基础URL
   */
  getBaseURL(): string {
    return this.baseURL;
  }

  /**
   * 更新基础URL
   */
  updateBaseURL(baseURL: string): void {
    this.baseURL = baseURL;
    this.client.defaults.baseURL = baseURL;
  }
}

// 创建默认的API服务实例
export const apiService = new ApiService();

export default apiService;