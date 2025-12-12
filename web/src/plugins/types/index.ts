import type { NodeProps } from '@xyflow/react';
import type { ComponentType, ReactNode } from 'react';

// 基础插件接口
export interface IPlugin {
  id: string;
  name: string;
  version: string;
  description: string;
  author: string;
  type: 'frontend' | 'backend' | 'fullstack';
  runtime: 'javascript' | 'python' | 'go' | 'docker';

  // 后端配置
  backend?: BackendConfig;

  // 节点定义
  nodeDefinition: NodeDefinition;

  // 插件元数据
  metadata: {
    category: string;
    subCategory?: string;
    icon?: ReactNode;
    color: string;
    tags: string[];
    homepage?: string;
    repository?: string;
    license?: string;
  };

  // 生命周期钩子
  onLoad?(context: PluginContext): Promise<void>;
  onUnload?(): Promise<void>;
  onActivate?(): Promise<void>;
  onDeactivate?(): Promise<void>;
}

// 后端配置
export interface BackendConfig {
  entryPoint: string; // main.py, main.go等
  dependencies?: string[]; // requirements.txt, go.mod等
  port?: number; // 服务端口，0表示自动分配
  envVars?: Record<string, string>; // 环境变量
  workingDirectory?: string; // 工作目录
  startupTimeout?: number; // 启动超时时间（毫秒）
  healthCheck?: {
    path: string;
    interval: number;
    retries: number;
  };
}

// 节点定义
export interface NodeDefinition {
  id: string;
  type: string;
  displayName: string;
  description: string;
  category: string;
  subCategory?: string;
  icon?: ReactNode;
  color: string;
  tags: string[];

  // 动态参数配置
  parameters: ParameterDefinition[];

  // API端点配置
  endpoints: APIEndpoint[];

  // 输入输出端口
  inputs?: PortDefinition[];
  outputs?: PortDefinition[];

  // 验证规则
  validation?: ValidationRule[];

  // 渲染配置
  rendering?: RenderingConfig;

  // 前端组件（可选）
  customComponent?: ComponentType<NodeProps>;
}

// 参数定义
export interface ParameterDefinition {
  id: string;
  name: string;
  type: ParameterType;
  required: boolean;
  defaultValue: any;

  // 参数分组
  group?: string;

  // 后端参数映射
  backendMapping?: {
    field: string;
    transform?: (value: any) => any;
  };

  // 动态配置
  dynamic?: {
    visible?: (values: Record<string, any>) => boolean;
    options?: (
      values: Record<string, any>,
    ) => Promise<Array<{ label: string; value: any; description?: string }>>;
    validation?: (value: any, values: Record<string, any>) => ValidationResult;
    loadFromAPI?: {
      endpoint: string;
      method: 'GET' | 'POST';
      params?: Record<string, any>;
      mapping?: (response: any) => Array<{ label: string; value: any }>;
    };
  };

  // UI配置
  ui?: UIConfig;

  // 参数约束
  constraints?: {
    min?: number;
    max?: number;
    pattern?: string;
    minLength?: number;
    maxLength?: number;
    options?: Array<{ label: string; value: any }>;
  };
}

// 参数类型
export type ParameterType =
  | 'string'
  | 'number'
  | 'boolean'
  | 'object'
  | 'array'
  | 'select'
  | 'multiselect'
  | 'textarea'
  | 'code'
  | 'json'
  | 'secret'
  | 'api-key'
  | 'connection'
  | 'file'
  | 'directory';

// UI配置
export interface UIConfig {
  label?: string;
  placeholder?: string;
  helpText?: string;
  advanced?: boolean;
  hidden?: boolean;
  disabled?: boolean;
  component?: string; // 自定义组件名称
  width?: 'small' | 'medium' | 'large' | 'full';
  height?: 'small' | 'medium' | 'large';
}

// API端点定义
export interface APIEndpoint {
  id: string;
  name: string;
  description?: string;
  path: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  streaming?: boolean;
  timeout?: number;
  parameters?: {
    path?: Record<string, string>;
    query?: Record<string, string>;
    body?: Record<string, string>;
  };
  response?: {
    type: 'json' | 'text' | 'binary' | 'stream';
    schema?: any;
  };
}

// 端口定义
export interface PortDefinition {
  id: string;
  name: string;
  type: 'input' | 'output';
  dataType: string;
  required?: boolean;
  multiple?: boolean;
  validation?: (value: any) => boolean;
}

// 验证规则
export interface ValidationRule {
  field: string;
  rule: 'required' | 'pattern' | 'min' | 'max' | 'custom';
  value?: any;
  message: string;
  validator?: (value: any) => ValidationResult;
}

// 渲染配置
export interface RenderingConfig {
  width?: number;
  height?: number;
  minWidth?: number;
  minHeight?: number;
  resizable?: boolean;
  collapsible?: boolean;
  color?: string;
  borderStyle?: 'solid' | 'dashed' | 'dotted';
  backgroundColor?: string;
}

// 验证结果
export interface ValidationResult {
  valid: boolean;
  message?: string;
  errors?: string[];
}

// 插件上下文
export interface PluginContext {
  pluginId: string;
  workingDirectory: string;
  api: PluginAPI;
  storage: PluginStorage;
  logger: PluginLogger;
}

// 插件API
export interface PluginAPI {
  // 节点操作
  registerNode: (definition: NodeDefinition) => void;
  unregisterNode: (nodeId: string) => void;

  // 服务操作
  registerService: (serviceId: string, config: ServiceConfig) => void;
  unregisterService: (serviceId: string) => void;

  // 事件系统
  emit: (event: string, data: any) => void;
  on: (event: string, handler: (data: any) => void) => void;
  off: (event: string, handler: (data: any) => void) => void;

  // 配置管理
  getConfig: (key: string) => any;
  setConfig: (key: string, value: any) => void;

  // 通知系统
  showNotification: (
    message: string,
    type: 'info' | 'success' | 'warning' | 'error',
  ) => void;
}

// 服务配置
export interface ServiceConfig {
  name: string;
  version: string;
  description?: string;
  endpoints: APIEndpoint[];
  healthCheck?: {
    path: string;
    interval: number;
    timeout: number;
  };
}

// 插件存储
export interface PluginStorage {
  get: (key: string) => Promise<any>;
  set: (key: string, value: any) => Promise<void>;
  delete: (key: string) => Promise<void>;
  clear: () => Promise<void>;
  keys: () => Promise<string[]>;
}

// 插件日志
export interface PluginLogger {
  debug: (message: string, ...args: any[]) => void;
  info: (message: string, ...args: any[]) => void;
  warn: (message: string, ...args: any[]) => void;
  error: (message: string, ...args: any[]) => void;
}

// 插件源
export interface PluginSource {
  type: 'local' | 'url' | 'market' | 'registry';

  // 本地文件
  localPath?: string;

  // URL安装
  url?: string;

  // 市场安装
  marketId?: string;

  // 注册表安装
  registry?: {
    name: string;
    version?: string;
  };

  // 安装选项
  options?: {
    force?: boolean;
    autoStart?: boolean;
    overwrite?: boolean;
  };
}

// 插件信息
export interface PluginInfo {
  id: string;
  name: string;
  description: string;
  version: string;
  author: string;
  downloads: number;
  rating: number;
  tags: string[];
  lastUpdated: Date;
  homepage?: string;
  repository?: string;
  license?: string;
  size?: number;
  dependencies?: string[];
}

// 插件详情
export interface PluginDetails extends PluginInfo {
  readme?: string;
  changelog?: string;
  screenshots?: string[];
  maintainer?: string;
  keywords?: string[];
  versions?: PluginVersion[];
  reviews?: PluginReview[];
}

// 插件版本
export interface PluginVersion {
  version: string;
  releaseDate: Date;
  changelog?: string;
  deprecated?: boolean;
  securityIssues?: number;
}

// 插件评论
export interface PluginReview {
  id: string;
  userId: string;
  username: string;
  rating: number;
  comment: string;
  createdAt: Date;
  helpful: number;
}

// 服务信息
export interface ServiceInfo {
  id: string;
  pluginId: string;
  name: string;
  status: 'starting' | 'running' | 'stopping' | 'stopped' | 'error';
  port: number;
  host: string;
  baseUrl: string;
  pid?: number;
  startTime?: Date;
  memoryUsage?: number;
  cpuUsage?: number;
  lastHealthCheck?: Date;
  healthStatus?: 'healthy' | 'unhealthy' | 'unknown';
  error?: string;
}

// 插件状态
export interface PluginStatus {
  id: string;
  status:
    | 'loading'
    | 'loaded'
    | 'activating'
    | 'active'
    | 'deactivating'
    | 'inactive'
    | 'error';
  loadedAt?: Date;
  activatedAt?: Date;
  error?: string;
  services: ServiceInfo[];
}

// 插件配置
export interface PluginConfig {
  pluginId: string;
  enabled: boolean;
  autoStart: boolean;
  settings: Record<string, any>;
  permissions: string[];
}

// 插件市场
export interface PluginMarketplace {
  name: string;
  url: string;
  description?: string;
  official?: boolean;
  enabled: boolean;
}

// 运行时适配器接口
export interface RuntimeAdapter {
  type: string;
  name: string;
  version: string;

  start(config: BackendConfig): Promise<ServiceInfo>;
  stop(serviceInfo: ServiceInfo): Promise<void>;
  restart(serviceInfo: ServiceInfo): Promise<ServiceInfo>;
  getStatus(serviceInfo: ServiceInfo): Promise<ServiceInfo>;
  healthCheck(serviceInfo: ServiceInfo): Promise<boolean>;

  // 运行时特定方法
  validateEnvironment?(): Promise<boolean>;
  installDependencies?(dependencies: string[]): Promise<void>;
  cleanup?(serviceInfo: ServiceInfo): Promise<void>;
}

// 插件事件
export interface PluginEvent {
  type: string;
  pluginId: string;
  data?: any;
  timestamp: Date;
}

// 插件过滤器
export interface PluginFilter {
  type?: string;
  runtime?: string;
  category?: string;
  author?: string;
  tags?: string[];
  status?: string;
  enabled?: boolean;
  active?: boolean;
}

// 插件搜索查询
export interface PluginSearchQuery {
  q?: string;
  category?: string;
  runtime?: string;
  author?: string;
  tags?: string[];
  limit?: number;
  offset?: number;
  sortBy?: 'name' | 'created' | 'updated' | 'downloads' | 'rating';
  sortOrder?: 'asc' | 'desc';
}
