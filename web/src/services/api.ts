import axios, { AxiosInstance, AxiosResponse } from 'axios';

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
 * API服务类 - 封装所有后端API调用
 */
export class ApiService {
  private client: AxiosInstance;
  private baseURL: string;

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
        // 在这里可以添加认证token
        const token = localStorage.getItem('auth_token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        console.log(`[API Request] ${config.method?.toUpperCase()} ${config.url}`);
        return config;
      },
      (error) => {
        console.error('[API Request Error]', error);
        return Promise.reject(error);
      }
    );

    // 响应拦截器
    this.client.interceptors.response.use(
      (response: AxiosResponse<ApiResponse>) => {
        console.log(`[API Response] ${response.config.method?.toUpperCase()} ${response.config.url}`, response.data);
        return response;
      },
      (error) => {
        console.error('[API Response Error]', error);

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