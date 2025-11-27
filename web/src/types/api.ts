// 从api.ts重新导出类型，避免循环依赖
export {
  type ServerConfig,
  type ConnectionTestResult,
  type InitConfig,
  type InitResult,
  type ProviderType,
  type ProviderConfig,
  type ProviderTestResult,
  type SystemConfig,
  type ApiResponse,
} from '../services/api';

// 扩展API相关类型

// 提供商预设配置
export interface ProviderPreset {
  id: string;
  name: string;
  displayName: string;
  description: string;
  type: ProviderType;
  icon?: string;
  configFields: ProviderConfigField[];
  defaultConfig: Record<string, any>;
  examples?: ProviderExample[];
}

export interface ProviderConfigField {
  key: string;
  label: string;
  type: 'string' | 'number' | 'boolean' | 'select' | 'textarea' | 'password' | 'file';
  required: boolean;
  description?: string;
  placeholder?: string;
  validation?: {
    min?: number;
    max?: number;
    pattern?: string;
    message?: string;
  };
  options?: Array<{ label: string; value: any }>;
  defaultValue?: any;
}

export interface ProviderExample {
  name: string;
  description: string;
  config: Record<string, any>;
}

// ASR (语音识别) 配置
export interface ASRConfig extends ProviderConfig {
  language?: string;
  sampleRate?: number;
  channels?: number;
  encoding?: string;
  maxDuration?: number;
  vad?: boolean;
  punctuation?: boolean;
  confidence?: number;
}

// TTS (语音合成) 配置
export interface TTSConfig extends ProviderConfig {
  voice?: string;
  language?: string;
  speed?: number;
  pitch?: number;
  volume?: number;
  outputFormat?: string;
  sampleRate?: number;
}

// LLM (大语言模型) 配置
export interface LLMConfig extends ProviderConfig {
  model?: string;
  temperature?: number;
  maxTokens?: number;
  topP?: number;
  frequencyPenalty?: number;
  presencePenalty?: number;
  systemPrompt?: string;
}

// VLLM (视觉语言模型) 配置
export interface VLLMConfig extends ProviderConfig {
  model?: string;
  temperature?: number;
  maxTokens?: number;
  imageQuality?: 'high' | 'medium' | 'low';
  maxImageSize?: number;
  supportedFormats?: string[];
}

// 工作流配置
export interface WorkflowConfig {
  id: string;
  name: string;
  description?: string;
  steps: WorkflowStep[];
  variables: WorkflowVariable[];
  triggers: WorkflowTrigger[];
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface WorkflowStep {
  id: string;
  name: string;
  type: 'input' | 'process' | 'output' | 'condition' | 'loop';
  config: Record<string, any>;
  nextSteps?: string[];
  condition?: string;
}

export interface WorkflowVariable {
  name: string;
  type: 'string' | 'number' | 'boolean' | 'object' | 'array';
  defaultValue?: any;
  description?: string;
  required?: boolean;
}

export interface WorkflowTrigger {
  type: 'manual' | 'schedule' | 'webhook' | 'event';
  config: Record<string, any>;
  active: boolean;
}

// 日志类型
export interface LogEntry {
  id: string;
  timestamp: string;
  level: 'debug' | 'info' | 'warn' | 'error' | 'fatal';
  message: string;
  module: string;
  details?: any;
  stack?: string;
  userId?: string;
  sessionId?: string;
}

// 健康检查结果
export interface HealthCheck {
  status: 'healthy' | 'unhealthy' | 'degraded';
  checks: HealthCheckItem[];
  timestamp: string;
  uptime: number;
  version: string;
}

export interface HealthCheckItem {
  name: string;
  status: 'pass' | 'fail' | 'warn';
  duration?: number;
  message?: string;
  details?: any;
}

// 系统信息
export interface SystemInfo {
  version: string;
  buildTime: string;
  gitCommit: string;
  environment: 'development' | 'staging' | 'production';
  architecture: string;
  os: string;
  runtime: string;
}

// 性能指标
export interface PerformanceMetrics {
  timestamp: string;
  cpu: {
    usage: number;
    cores: number;
    loadAverage: number[];
  };
  memory: {
    total: number;
    used: number;
    free: number;
    usage: number;
  };
  disk: {
    total: number;
    used: number;
    free: number;
    usage: number;
  };
  network: {
    bytesIn: number;
    bytesOut: number;
    packetsIn: number;
    packetsOut: number;
  };
  processes: {
    total: number;
    running: number;
    sleeping: number;
  };
}

// API端点信息
export interface ApiEndpoint {
  method: string;
  path: string;
  description: string;
  parameters: ApiParameter[];
  requestBody?: any;
  responses: ApiResponseDefinition[];
}

export interface ApiParameter {
  name: string;
  in: 'path' | 'query' | 'header' | 'cookie';
  required: boolean;
  type: string;
  description?: string;
}

export interface ApiResponseDefinition {
  code: number;
  description: string;
  schema?: any;
}

// 配置验证结果
export interface ConfigValidationResult {
  valid: boolean;
  errors: ConfigValidationError[];
  warnings: ConfigValidationWarning[];
}

export interface ConfigValidationError {
  field: string;
  message: string;
  value?: any;
  code: string;
}

export interface ConfigValidationWarning {
  field: string;
  message: string;
  value?: any;
  code: string;
}