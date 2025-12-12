import { create } from 'zustand';
import { devtools, subscribeWithSelector } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import type {
  AIStatus,
  Particle,
  ParticleConfig,
  ParticleSystemState,
} from '../types/particle';

interface ParticleStore extends ParticleSystemState {
  // Actions
  initialize: (canvas: HTMLCanvasElement) => void;
  start: () => void;
  stop: () => void;
  updateConfig: (config: Partial<ParticleConfig>) => void;
  setAIStatus: (status: AIStatus) => void;
  updateMousePosition: (x: number, y: number) => void;
  addParticle: (particle: Particle) => void;
  removeParticle: (id: string) => void;
  clearParticles: () => void;
  updatePerformanceStats: (
    stats: Partial<ParticleSystemState['performanceStats']>,
  ) => void;
  destroy: () => void;
}

// 初始粒子状态
const initialParticleState: ParticleSystemState = {
  particles: [],
  config: {
    background: {
      count: 30,
      speed: 0.3,
      size: 1,
      opacity: 0.4,
      color: '#ffffff',
      shape: 'circle',
      connectDistance: 120,
      connectionOpacity: 0.1,
      pulse: false,
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
      connectionColor: '#ffffff',
      repulsion: true,
      attraction: false,
    },
    aiStatus: {
      enabled: true,
      statusEffects: {
        idle: {
          color: '#ffffff',
          particleColor: '#ffffff',
          speed: 0.2,
          size: 2,
          count: 15,
          shape: 'circle',
          pattern: 'random',
          animation: 'fade',
        },
        listening: {
          color: '#ffffff',
          particleColor: '#ffffff',
          speed: 0.6,
          size: 3,
          count: 25,
          shape: 'circle',
          pattern: 'wave',
          animation: 'pulse',
        },
        processing: {
          color: '#ffffff',
          particleColor: '#ffffff',
          speed: 0.8,
          size: 2,
          count: 30,
          shape: 'circle',
          pattern: 'spiral',
          animation: 'rotate',
        },
        speaking: {
          color: '#ffffff',
          particleColor: '#ffffff',
          speed: 0.5,
          size: 4,
          count: 20,
          shape: 'circle',
          pattern: 'burst',
          animation: 'scale',
        },
        error: {
          color: '#ffffff',
          particleColor: '#ffffff',
          speed: 1.0,
          size: 3,
          count: 20,
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
};

// 粒子系统相关工具函数
const particleUtils = {
  // 创建背景粒子
  createBackgroundParticle: (
    config: ParticleConfig['background'],
    canvas: HTMLCanvasElement,
  ): Particle => ({
    id: `bg-${Date.now()}-${Math.random()}`,
    x: Math.random() * canvas.width,
    y: Math.random() * canvas.height,
    vx: (Math.random() - 0.5) * config.speed,
    vy: (Math.random() - 0.5) * config.speed,
    size: config.size + Math.random() * 2,
    color: config.color,
    opacity: config.opacity + Math.random() * 0.3,
    rotation: 0,
    rotationSpeed: (Math.random() - 0.5) * config.rotationSpeed * 0.01,
    lifetime: Infinity,
    maxLifetime: Infinity,
    type: 'background',
  }),

  // 创建AI状态粒子
  createAIStatusParticle: (
    x: number,
    y: number,
    status: AIStatus,
    config: ParticleConfig['aiStatus'],
  ): Particle | null => {
    if (!config.enabled) return null;

    const statusConfig = config.statusEffects[status];
    let vx = 0,
      vy = 0;

    switch (statusConfig.pattern) {
      case 'wave': {
        const angle = Math.random() * Math.PI * 2;
        vx = Math.cos(angle) * statusConfig.speed;
        vy = Math.sin(angle) * statusConfig.speed;
        break;
      }
      case 'burst': {
        const burstAngle = Math.random() * Math.PI * 2;
        vx = Math.cos(burstAngle) * statusConfig.speed * 2;
        vy = Math.sin(burstAngle) * statusConfig.speed * 2;
        break;
      }
      default:
        vx = (Math.random() - 0.5) * statusConfig.speed;
        vy = (Math.random() - 0.5) * statusConfig.speed;
    }

    return {
      id: `ai-${Date.now()}-${Math.random()}`,
      x,
      y,
      vx,
      vy,
      size: statusConfig.size + Math.random() * 2,
      color: statusConfig.particleColor,
      opacity: 0.8,
      rotation: 0,
      rotationSpeed: 0.02,
      lifetime: config.particleLifetime,
      maxLifetime: config.particleLifetime,
      type: 'ai-status',
      metadata: {
        status,
        animation: statusConfig.animation,
      },
    };
  },

  // 更新粒子位置
  updateParticle: (
    particle: Particle,
    deltaTime: number,
    canvas: HTMLCanvasElement,
  ): void => {
    particle.x += particle.vx;
    particle.y += particle.vy;
    particle.rotation += particle.rotationSpeed;

    // 边界检测
    if (particle.x < 0 || particle.x > canvas.width) {
      particle.vx = -particle.vx;
      particle.x = Math.max(0, Math.min(canvas.width, particle.x));
    }
    if (particle.y < 0 || particle.y > canvas.height) {
      particle.vy = -particle.vy;
      particle.y = Math.max(0, Math.min(canvas.height, particle.y));
    }

    // 生命周期透明度
    if (particle.lifetime !== Infinity) {
      particle.opacity = Math.max(0, particle.lifetime / particle.maxLifetime);
    }
  },

  // 应用鼠标交互
  applyMouseInteraction: (
    particle: Particle,
    mousePosition: { x: number; y: number },
    config: ParticleConfig['interactive'],
  ): void => {
    if (!config.enabled) return;

    const dx = mousePosition.x - particle.x;
    const dy = mousePosition.y - particle.y;
    const distance = Math.sqrt(dx * dx + dy * dy);

    if (distance < config.mouseRadius) {
      const force = (1 - distance / config.mouseRadius) * config.mouseForce;
      const angle = Math.atan2(dy, dx);

      if (config.repulsion) {
        particle.vx -= Math.cos(angle) * force;
        particle.vy -= Math.sin(angle) * force;
      }
      if (config.attraction) {
        particle.vx += Math.cos(angle) * force;
        particle.vy += Math.sin(angle) * force;
      }
    }
  },
};

// 创建粒子store
export const useParticleStore = create<ParticleStore>()(
  devtools(
    subscribeWithSelector(
      immer((set, get) => ({
        ...initialParticleState,

        // 初始化粒子系统
        initialize: (canvas: HTMLCanvasElement) =>
          set((state) => {
            state.isInitialized = true;
            // 生成初始背景粒子
            for (let i = 0; i < state.config.background.count; i++) {
              state.particles.push(
                particleUtils.createBackgroundParticle(
                  state.config.background,
                  canvas,
                ),
              );
            }
          }),

        // 启动粒子系统
        start: () =>
          set((state) => {
            state.isRunning = true;
            state.lastUpdateTime = performance.now();
          }),

        // 停止粒子系统
        stop: () =>
          set((state) => {
            state.isRunning = false;
          }),

        // 更新配置
        updateConfig: (newConfig: Partial<ParticleConfig>) =>
          set((state) => {
            Object.assign(state.config, newConfig);
          }),

        // 设置AI状态
        setAIStatus: (status: AIStatus) =>
          set((state) => {
            state.aiStatus = status;
          }),

        // 更新鼠标位置
        updateMousePosition: (x: number, y: number) =>
          set((state) => {
            state.mousePosition = { x, y };
          }),

        // 添加粒子
        addParticle: (particle: Particle) =>
          set((state) => {
            // 检查粒子数量限制
            if (
              state.particles.length < state.config.performance.maxParticles
            ) {
              state.particles.push(particle);
            }
          }),

        // 移除粒子
        removeParticle: (id: string) =>
          set((state) => {
            state.particles = state.particles.filter((p) => p.id !== id);
          }),

        // 清空所有粒子
        clearParticles: () =>
          set((state) => {
            state.particles = [];
          }),

        // 更新性能统计
        updatePerformanceStats: (
          stats: Partial<ParticleSystemState['performanceStats']>,
        ) =>
          set((state) => {
            Object.assign(state.performanceStats, stats);
          }),

        // 销毁粒子系统
        destroy: () =>
          set((state) => {
            state.particles = [];
            state.isInitialized = false;
            state.isRunning = false;
          }),
      })),
    ),
    {
      name: 'particle-store',
    },
  ),
);

// 选择器hooks
export const useParticles = () => useParticleStore((state) => state.particles);
export const useParticleConfig = () =>
  useParticleStore((state) => state.config);
export const useAIStatus = () => useParticleStore((state) => state.aiStatus);
export const useMousePosition = () =>
  useParticleStore((state) => state.mousePosition);
export const useParticleStats = () =>
  useParticleStore((state) => state.performanceStats);
export const useParticleInitialized = () =>
  useParticleStore((state) => state.isInitialized);
export const useParticleRunning = () =>
  useParticleStore((state) => state.isRunning);

// 复合选择器
export const useParticleCount = () => {
  const particles = useParticles();
  const config = useParticleConfig();
  return {
    total: particles.length,
    background: particles.filter((p) => p.type === 'background').length,
    aiStatus: particles.filter((p) => p.type === 'ai-status').length,
    interactive: particles.filter((p) => p.type === 'interactive').length,
    max: config.performance.maxParticles,
  };
};

export default useParticleStore;
