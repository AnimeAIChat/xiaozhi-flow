// Zustand Store 相关类型定义

import type {
  IPlugin,
  PluginConfig,
  PluginFilter,
  PluginMarketplace,
  PluginSource,
  PluginStatus,
  ServiceInfo,
} from '../plugins/types';
import type {
  ConnectionTestResult,
  ProviderConfig,
  ProviderType,
  SystemConfig,
  WorkflowConfig,
} from './api';
import type {
  AIStatus,
  AuthState,
  Language,
  ModalProps,
  Notification,
  Theme,
  User,
} from './index';
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

  // 插件状态
  plugins: PluginState;

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

  // 配置页面侧边栏
  configSidebar: {
    collapsed: boolean;
    width: number;
    defaultWidth: number;
    minWidth: number;
    maxWidth: number;
  };

  // 组件库悬浮面板
  componentLibraryPanel: {
    visible: boolean;
    position: { x: number; y: number };
    pinned: boolean;
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

// 插件状态
export interface PluginState {
  // 插件管理
  plugins: {
    list: IPlugin[];
    loading: boolean;
    error: string | null;
  };

  // 插件状态
  statuses: {
    [pluginId: string]: PluginStatus;
  };

  // 已安装插件
  installed: {
    [pluginId: string]: IPlugin;
  };

  // 插件配置
  configs: {
    [pluginId: string]: PluginConfig;
  };

  // 后端服务
  services: {
    [serviceId: string]: ServiceInfo;
  };

  // 插件市场
  marketplace: {
    available: IPlugin[];
    categories: string[];
    searchResults: IPlugin[];
    loading: boolean;
    error: string | null;
  };

  // 插件管理器状态
  manager: {
    loading: boolean;
    error: string | null;
    operationInProgress: boolean;
    currentOperation: {
      type:
        | 'install'
        | 'uninstall'
        | 'update'
        | 'activate'
        | 'deactivate'
        | null;
      pluginId: string | null;
      progress: number;
    };
  };

  // 安装状态
  installation: {
    inProgress: boolean;
    progress: {
      stage: string;
      progress: number;
      message: string;
    } | null;
  };
}

// Store Actions
export interface AppActions {
  // 认证相关
  auth: {
    login: (credentials: {
      username: string;
      password: string;
    }) => Promise<void>;
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
    updateProvider: (
      type: ProviderType,
      config: ProviderConfig,
    ) => Promise<void>;
    testProvider: (type: ProviderType, config: ProviderConfig) => Promise<void>;
    selectProvider: (type: ProviderType, providerId: string | null) => void;

    // 工作流配置
    loadWorkflows: () => Promise<void>;
    createWorkflow: (
      workflow: Omit<WorkflowConfig, 'id' | 'createdAt' | 'updatedAt'>,
    ) => Promise<void>;
    updateWorkflow: (
      id: string,
      workflow: Partial<WorkflowConfig>,
    ) => Promise<void>;
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

    // 配置页面侧边栏
    toggleConfigSidebar: () => void;
    setConfigSidebarCollapsed: (collapsed: boolean) => void;
    setConfigSidebarWidth: (width: number) => void;

    // 组件库悬浮面板
    toggleComponentLibraryPanel: () => void;
    showComponentLibraryPanel: (position?: { x: number; y: number }) => void;
    hideComponentLibraryPanel: () => void;
    setComponentLibraryPanelPosition: (position: {
      x: number;
      y: number;
    }) => void;
    toggleComponentLibraryPanelPin: () => void;
    setComponentLibraryPanelPin: (pinned: boolean) => void;

    // 页面管理
    setPageLoading: (loading: boolean) => void;
    setPageTitle: (title: string) => void;
    setBreadcrumb: (
      breadcrumb: Array<{ title: string; path?: string }>,
    ) => void;

    // 模态框管理
    openModal: (key: string, modal: Omit<ModalProps, 'visible'>) => void;
    closeModal: (key: string) => void;
    closeAllModals: () => void;

    // 抽屉管理
    openDrawer: (
      key: string,
      drawer: { title?: string; content?: React.ReactNode },
    ) => void;
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
    connectService: (
      type: 'asr' | 'tts' | 'llm' | 'vllm',
      providerId: string,
    ) => Promise<void>;
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
    add: (
      notification: Omit<Notification, 'id' | 'timestamp' | 'read'>,
    ) => void;
    remove: (id: string) => void;
    markAsRead: (id: string) => void;
    markAllAsRead: () => void;
    clear: () => void;
  };

  // 插件相关
  plugins: {
    // 插件管理
    loadPlugins: (filter?: PluginFilter) => Promise<void>;
    installPlugin: (source: PluginSource) => Promise<void>;
    uninstallPlugin: (pluginId: string) => Promise<void>;
    updatePlugin: (pluginId: string, source?: PluginSource) => Promise<void>;
    activatePlugin: (pluginId: string) => Promise<void>;
    deactivatePlugin: (pluginId: string) => Promise<void>;
    reloadPlugin: (pluginId: string) => Promise<void>;

    // 插件配置
    getPluginConfig: (pluginId: string) => PluginConfig | undefined;
    setPluginConfig: (pluginId: string, config: Partial<PluginConfig>) => void;
    resetPluginConfig: (pluginId: string) => void;

    // 插件状态
    getPluginStatus: (pluginId: string) => PluginStatus | undefined;
    updatePluginStatus: (
      pluginId: string,
      status: Partial<PluginStatus>,
    ) => void;

    // 服务管理
    startService: (pluginId: string, serviceId: string) => Promise<void>;
    stopService: (pluginId: string, serviceId: string) => Promise<void>;
    restartService: (pluginId: string, serviceId: string) => Promise<void>;
    getService: (serviceId: string) => ServiceInfo | undefined;

    // 市场相关
    searchMarketplace: (query: string) => Promise<void>;
    loadMarketplacePlugins: (category?: string) => Promise<void>;
    getPluginCategories: () => Promise<string[]>;

    // 安装进度
    setInstallationProgress: (progress: {
      stage: string;
      progress: number;
      message: string;
    }) => void;
    clearInstallationProgress: () => void;

    // 批量操作
    installMultiplePlugins: (sources: PluginSource[]) => Promise<void>;
    updateAllPlugins: () => Promise<void>;
    deactivateAllPlugins: () => Promise<void>;
  };
}

// 完整的Store类型
export type AppStore = AppState & AppActions;
