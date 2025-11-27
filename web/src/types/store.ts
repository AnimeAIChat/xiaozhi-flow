// Zustand Store 相关类型定义

import type {
  AuthState,
  User,
  Theme,
  Language,
  Notification,
  ModalProps,
  AIStatus
} from './index';
import type {
  SystemConfig,
  ProviderConfig,
  ProviderType,
  WorkflowConfig,
  ConnectionTestResult
} from './api';
import type { ParticleConfig, ParticleSystemState } from './particle';

// 应用状态接口
export interface AppState {
  // 认证状态
  auth: AuthState;

  // 配置状态
  config: ConfigState;

  // UI状态
  ui: UIState;

  // AI状态
  ai: AIState;

  // 粒子状态
  particles: ParticleSystemState;

  // 通知状态
  notifications: Notification[];
}

// 配置状态
export interface ConfigState {
  // 服务器配置
  server: {
    config: ServerConfig;
    connectionStatus: 'idle' | 'testing' | 'connected' | 'error';
    lastConnectionTest?: ConnectionTestResult;
  };

  // 系统配置
  system: SystemConfig | null;
  systemConfigLoading: boolean;
  systemConfigError: string | null;

  // 提供商配置
  providers: {
    [K in ProviderType]: {
      list: ProviderConfig[];
      selected: string | null;
      loading: boolean;
      error: string | null;
    };
  };

  // 工作流配置
  workflows: {
    list: WorkflowConfig[];
    selected: string | null;
    loading: boolean;
    error: string | null;
  };

  // 配置验证
  validation: {
    isValid: boolean;
    errors: string[];
    warnings: string[];
  };
}

// UI状态
export interface UIState {
  // 主题和外观
  theme: Theme;
  language: Language;

  // 布局状态
  sidebar: {
    collapsed: boolean;
    width: number;
  };

  // 页面状态
  page: {
    loading: boolean;
    title: string;
    breadcrumb: Array<{ title: string; path?: string }>;
  };

  // 模态框状态
  modals: {
    [key: string]: ModalProps;
  };

  // 抽屉状态
  drawers: {
    [key: string]: {
      visible: boolean;
      title?: string;
      content?: React.ReactNode;
    };
  };

  // 响应式状态
  responsive: {
    isMobile: boolean;
    isTablet: boolean;
    isDesktop: boolean;
    screenSize: {
      width: number;
      height: number;
    };
  };
}

// AI状态
export interface AIState {
  // AI服务状态
  status: AIStatus;

  // 对话状态
  conversation: {
    id: string | null;
    messages: AIMessage[];
    isTyping: boolean;
    currentResponse: string;
  };

  // 语音状态
  audio: {
    isRecording: boolean;
    isPlaying: boolean;
    volume: number;
    inputDevice: string | null;
    outputDevice: string | null;
  };

  // 服务状态
  services: {
    asr: {
      connected: boolean;
      provider: string | null;
      lastActivity: number;
    };
    tts: {
      connected: boolean;
      provider: string | null;
      lastActivity: number;
    };
    llm: {
      connected: boolean;
      provider: string | null;
      lastActivity: number;
    };
    vllm: {
      connected: boolean;
      provider: string | null;
      lastActivity: number;
    };
  };

  // 性能状态
  performance: {
    totalRequests: number;
    successfulRequests: number;
    failedRequests: number;
    averageResponseTime: number;
    lastError: string | null;
  };
}

// AI消息类型
export interface AIMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: number;
  metadata?: {
    model?: string;
    provider?: string;
    tokens?: number;
    duration?: number;
    confidence?: number;
  };
}

// Store Actions
export interface AppActions {
  // 认证相关
  auth: {
    login: (credentials: { username: string; password: string }) => Promise<void>;
    logout: () => void;
    refreshToken: () => Promise<void>;
    updateProfile: (profile: Partial<User>) => void;
  };

  // 配置相关
  config: {
    // 服务器配置
    updateServerConfig: (config: ServerConfig) => void;
    testConnection: (config: ServerConfig) => Promise<ConnectionTestResult>;

    // 系统配置
    loadSystemConfig: () => Promise<void>;
    updateSystemConfig: (config: SystemConfig) => Promise<void>;

    // 提供商配置
    loadProviders: (type?: ProviderType) => Promise<void>;
    updateProvider: (type: ProviderType, config: ProviderConfig) => Promise<void>;
    testProvider: (type: ProviderType, config: ProviderConfig) => Promise<void>;
    selectProvider: (type: ProviderType, providerId: string | null) => void;

    // 工作流配置
    loadWorkflows: () => Promise<void>;
    createWorkflow: (workflow: Omit<WorkflowConfig, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>;
    updateWorkflow: (id: string, workflow: Partial<WorkflowConfig>) => Promise<void>;
    deleteWorkflow: (id: string) => Promise<void>;

    // 配置验证
    validateConfig: (config: any) => Promise<boolean>;
  };

  // UI相关
  ui: {
    setTheme: (theme: Theme) => void;
    setLanguage: (language: Language) => void;
    toggleSidebar: () => void;
    setSidebarCollapsed: (collapsed: boolean) => void;

    // 页面管理
    setPageLoading: (loading: boolean) => void;
    setPageTitle: (title: string) => void;
    setBreadcrumb: (breadcrumb: Array<{ title: string; path?: string }>) => void;

    // 模态框管理
    openModal: (key: string, modal: Omit<ModalProps, 'visible'>) => void;
    closeModal: (key: string) => void;
    closeAllModals: () => void;

    // 抽屉管理
    openDrawer: (key: string, drawer: { title?: string; content?: React.ReactNode }) => void;
    closeDrawer: (key: string) => void;
    closeAllDrawers: () => void;

    // 响应式更新
    updateResponsive: (width: number, height: number) => void;
  };

  // AI相关
  ai: {
    setStatus: (status: AIStatus) => void;

    // 对话管理
    sendMessage: (message: string) => Promise<void>;
    clearConversation: () => void;

    // 语音控制
    startRecording: () => void;
    stopRecording: () => void;
    setVolume: (volume: number) => void;

    // 服务管理
    connectService: (type: 'asr' | 'tts' | 'llm' | 'vllm', providerId: string) => Promise<void>;
    disconnectService: (type: 'asr' | 'tts' | 'llm' | 'vllm') => Promise<void>;

    // 性能重置
    resetPerformanceStats: () => void;
  };

  // 粒子相关
  particles: {
    initialize: (canvas: HTMLCanvasElement) => void;
    start: () => void;
    stop: () => void;
    updateConfig: (config: Partial<ParticleConfig>) => void;
    setAIStatus: (status: AIStatus) => void;
    destroy: () => void;
  };

  // 通知相关
  notifications: {
    add: (notification: Omit<Notification, 'id' | 'timestamp' | 'read'>) => void;
    remove: (id: string) => void;
    markAsRead: (id: string) => void;
    markAllAsRead: () => void;
    clear: () => void;
  };
}

// 完整的Store类型
export type AppStore = AppState & AppActions;