import { create } from 'zustand';
import { devtools, persist, subscribeWithSelector } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import type { AppState, AppActions, AppStore } from '../types/store';
import type { AIStatus, Theme, Language } from '../types/index';
import type { ServerConfig, SystemConfig, ProviderConfig, ProviderType } from '../types/api';
import type { ParticleConfig } from '../types/particle';

// 初始状态
const initialState: AppState = {
  // 认证状态
  auth: {
    isAuthenticated: false,
    user: null,
    token: null,
    permissions: [],
  },

  // 配置状态
  config: {
    server: {
      config: {
        host: 'localhost',
        port: 8080,
        protocol: 'http',
      },
      connectionStatus: 'idle',
    },
    system: null,
    systemConfigLoading: false,
    systemConfigError: null,
    providers: {
      asr: {
        list: [],
        selected: null,
        loading: false,
        error: null,
      },
      tts: {
        list: [],
        selected: null,
        loading: false,
        error: null,
      },
      llm: {
        list: [],
        selected: null,
        loading: false,
        error: null,
      },
      vllm: {
        list: [],
        selected: null,
        loading: false,
        error: null,
      },
    },
    workflows: {
      list: [],
      selected: null,
      loading: false,
      error: null,
    },
    validation: {
      isValid: false,
      errors: [],
      warnings: [],
    },
  },

  // UI状态
  ui: {
    theme: 'dark',
    language: 'zh-CN',
    sidebar: {
      collapsed: false,
      width: 256,
    },
    page: {
      loading: false,
      title: 'Xiaozhi-Flow',
      breadcrumb: [],
    },
    modals: {},
    drawers: {},
    responsive: {
      isMobile: false,
      isTablet: false,
      isDesktop: true,
      screenSize: {
        width: 1920,
        height: 1080,
      },
    },
  },

  // AI状态
  ai: {
    status: 'idle',
    conversation: {
      id: null,
      messages: [],
      isTyping: false,
      currentResponse: '',
    },
    audio: {
      isRecording: false,
      isPlaying: false,
      volume: 0.7,
      inputDevice: null,
      outputDevice: null,
    },
    services: {
      asr: {
        connected: false,
        provider: null,
        lastActivity: 0,
      },
      tts: {
        connected: false,
        provider: null,
        lastActivity: 0,
      },
      llm: {
        connected: false,
        provider: null,
        lastActivity: 0,
      },
      vllm: {
        connected: false,
        provider: null,
        lastActivity: 0,
      },
    },
    performance: {
      totalRequests: 0,
      successfulRequests: 0,
      failedRequests: 0,
      averageResponseTime: 0,
      lastError: null,
    },
  },

  // 粒子状态
  particles: {
    particles: [],
    config: {
      background: {
        count: 50,
        speed: 0.5,
        size: 2,
        opacity: 0.6,
        color: '#1890ff',
        shape: 'circle',
        connectDistance: 150,
        connectionOpacity: 0.2,
        pulse: true,
        pulseSpeed: 2,
        rotation: false,
        rotationSpeed: 1,
      },
      interactive: {
        enabled: true,
        mouseRadius: 100,
        mouseForce: 0.5,
        clickEffect: true,
        clickRadius: 200,
        clickForce: 2,
        hoverEffect: true,
        connections: true,
        connectionColor: '#1890ff',
        repulsion: true,
        attraction: false,
      },
      aiStatus: {
        enabled: true,
        statusEffects: {
          idle: {
            color: '#1890ff',
            particleColor: '#1890ff',
            speed: 0.3,
            size: 3,
            count: 20,
            shape: 'circle',
            pattern: 'random',
            animation: 'fade',
          },
          listening: {
            color: '#52c41a',
            particleColor: '#52c41a',
            speed: 0.8,
            size: 4,
            count: 30,
            shape: 'circle',
            pattern: 'wave',
            animation: 'pulse',
          },
          processing: {
            color: '#faad14',
            particleColor: '#faad14',
            speed: 1.2,
            size: 3,
            count: 40,
            shape: 'star',
            pattern: 'spiral',
            animation: 'rotate',
          },
          speaking: {
            color: '#722ed1',
            particleColor: '#722ed1',
            speed: 0.6,
            size: 5,
            count: 35,
            shape: 'heart',
            pattern: 'burst',
            animation: 'scale',
          },
          error: {
            color: '#ff4d4f',
            particleColor: '#ff4d4f',
            speed: 1.5,
            size: 4,
            count: 25,
            shape: 'triangle',
            pattern: 'random',
            animation: 'shake',
          },
        },
        transitionDuration: 500,
        particleCount: 30,
        emitRate: 2,
        particleLifetime: 3000,
      },
      performance: {
        maxParticles: 150,
        targetFPS: 60,
        adaptiveQuality: true,
        reduceMotion: false,
        pixelRatio: 1,
        culling: true,
        updateThrottle: 16,
      },
    },
    aiStatus: 'idle',
    mousePosition: { x: 0, y: 0 },
    isInitialized: false,
    isRunning: false,
    lastUpdateTime: 0,
    fps: 0,
    performanceStats: {
      fps: 60,
      frameTime: 0,
      particleCount: 0,
      drawCalls: 0,
      memoryUsage: 0,
      cpuUsage: 0,
    },
  },

  // 通知
  notifications: [],
};

// 动作实现
const actions: AppActions = {
  // 认证相关
  auth: {
    login: async (credentials) => {
      // 实际实现中这里会调用API
      console.log('Login action:', credentials);
    },
    logout: () => {
      // 清除认证信息
    },
    refreshToken: async () => {
      // 刷新token
    },
    updateProfile: (profile) => {
      // 更新用户资料
    },
  },

  // 配置相关
  config: {
    updateServerConfig: (config) => {
      // 更新服务器配置
    },
    testConnection: async (config) => {
      // 测试连接
      return { success: false, message: 'Not implemented' };
    },
    loadSystemConfig: async () => {
      // 加载系统配置
    },
    updateSystemConfig: async (config) => {
      // 更新系统配置
    },
    loadProviders: async (type) => {
      // 加载提供商列表
    },
    updateProvider: async (type, config) => {
      // 更新提供商配置
    },
    testProvider: async (type, config) => {
      // 测试提供商
    },
    selectProvider: (type, providerId) => {
      // 选择提供商
    },
    loadWorkflows: async () => {
      // 加载工作流
    },
    createWorkflow: async (workflow) => {
      // 创建工作流
    },
    updateWorkflow: async (id, workflow) => {
      // 更新工作流
    },
    deleteWorkflow: async (id) => {
      // 删除工作流
    },
    validateConfig: async (config) => {
      // 验证配置
      return true;
    },
  },

  // UI相关
  ui: {
    setTheme: (theme) => {
      // 设置主题
    },
    setLanguage: (language) => {
      // 设置语言
    },
    toggleSidebar: () => {
      // 切换侧边栏
    },
    setSidebarCollapsed: (collapsed) => {
      // 设置侧边栏折叠状态
    },
    setPageLoading: (loading) => {
      // 设置页面加载状态
    },
    setPageTitle: (title) => {
      // 设置页面标题
    },
    setBreadcrumb: (breadcrumb) => {
      // 设置面包屑
    },
    openModal: (key, modal) => {
      // 打开模态框
    },
    closeModal: (key) => {
      // 关闭模态框
    },
    closeAllModals: () => {
      // 关闭所有模态框
    },
    openDrawer: (key, drawer) => {
      // 打开抽屉
    },
    closeDrawer: (key) => {
      // 关闭抽屉
    },
    closeAllDrawers: () => {
      // 关闭所有抽屉
    },
    updateResponsive: (width, height) => {
      // 更新响应式状态
    },
  },

  // AI相关
  ai: {
    setStatus: (status) => {
      // 设置AI状态
    },
    sendMessage: async (message) => {
      // 发送消息
    },
    clearConversation: () => {
      // 清空对话
    },
    startRecording: () => {
      // 开始录音
    },
    stopRecording: () => {
      // 停止录音
    },
    setVolume: (volume) => {
      // 设置音量
    },
    connectService: async (type, providerId) => {
      // 连接服务
    },
    disconnectService: async (type) => {
      // 断开服务
    },
    resetPerformanceStats: () => {
      // 重置性能统计
    },
  },

  // 粒子相关
  particles: {
    initialize: (canvas) => {
      // 初始化粒子系统
    },
    start: () => {
      // 启动粒子系统
    },
    stop: () => {
      // 停止粒子系统
    },
    updateConfig: (config) => {
      // 更新粒子配置
    },
    setAIStatus: (status) => {
      // 设置AI状态
    },
    destroy: () => {
      // 销毁粒子系统
    },
  },

  // 通知相关
  notifications: {
    add: (notification) => {
      // 添加通知
    },
    remove: (id) => {
      // 移除通知
    },
    markAsRead: (id) => {
      // 标记为已读
    },
    markAllAsRead: () => {
      // 全部标记为已读
    },
    clear: () => {
      // 清空通知
    },
  },
};

// 创建store
export const useAppStore = create<AppStore>()(
  devtools(
    subscribeWithSelector(
      persist(
        immer((set, get) => ({
          ...initialState,
          ...actions,
          // 简化的状态更新方法
          setAuth: (auth) => set((state) => ({ auth })),
          setServerConfig: (config) =>
            set((state) => ({
              config: { ...state.config, server: { ...state.config.server, config } },
            })),
          setAIStatus: (status) => set((state) => ({ ai: { ...state.ai, status } })),
          setTheme: (theme: Theme) => set((state) => ({ ui: { ...state.ui, theme } })),
          setLanguage: (language: Language) =>
            set((state) => ({ ui: { ...state.ui, language } })),
          toggleSidebar: () =>
            set((state) => ({
              ui: {
                ...state.ui,
                sidebar: {
                  ...state.ui.sidebar,
                  collapsed: !state.ui.sidebar.collapsed,
                },
              },
            })),
        })),
        {
          name: 'xiaozhi-flow-app-store',
          partialize: (state) => ({
            // 只持久化部分状态
            ui: {
              theme: state.ui.theme,
              language: state.ui.language,
              sidebar: state.ui.sidebar,
            },
            config: {
              server: state.config.server,
            },
          }),
        }
      )
    ),
    {
      name: 'xiaozhi-flow-store',
    }
  )
);

// 选择器hooks
export const useAuth = () => useAppStore((state) => state.auth);
export const useConfig = () => useAppStore((state) => state.config);
export const useUI = () => useAppStore((state) => state.ui);
export const useAI = () => useAppStore((state) => state.ai);
export const useParticles = () => useAppStore((state) => state.particles);
export const useNotifications = () => useAppStore((state) => state.notifications);

// 特定状态的选择器
export const useServerConfig = () => useAppStore((state) => state.config.server);
export const useProviders = (type: ProviderType) =>
  useAppStore((state) => state.config.providers[type]);
export const useSystemConfig = () => useAppStore((state) => state.config.system);
export const useTheme = () => useAppStore((state) => state.ui.theme);
export const useLanguage = () => useAppStore((state) => state.ui.language);
export const useSidebar = () => useAppStore((state) => state.ui.sidebar);

export default useAppStore;