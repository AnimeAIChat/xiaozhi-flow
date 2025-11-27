// Zustand Store 入口文件

export { default as useAppStore } from './useAppStore';
export { default as useParticleStore } from './useParticleStore';

// 选择器hooks
export {
  useAuth,
  useConfig,
  useUI,
  useAI,
  useParticles,
  useNotifications,
  useServerConfig,
  useProviders,
  useSystemConfig,
  useTheme,
  useLanguage,
  useSidebar,
} from './useAppStore';

export {
  useParticles as useParticleParticles,
  useParticleConfig,
  useAIStatus as useParticleAIStatus,
  useMousePosition,
  useParticleStats,
  useParticleInitialized,
  useParticleRunning,
  useParticleCount,
} from './useParticleStore';