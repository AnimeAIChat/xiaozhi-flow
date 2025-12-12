// 粒子效果相关类型定义

// 粒子效果类型
export type ParticleType = 'background' | 'interactive' | 'ai-status';

// 粒子配置
export interface ParticleConfig {
  background: BackgroundParticleConfig;
  interactive: InteractiveParticleConfig;
  aiStatus: AIStatusParticleConfig;
  performance: PerformanceConfig;
}

// 背景粒子配置
export interface BackgroundParticleConfig {
  count: number; // 粒子数量
  speed: number; // 移动速度
  size: number; // 粒子大小
  opacity: number; // 透明度
  color: string; // 颜色
  shape: 'circle' | 'square' | 'triangle'; // 形状
  connectDistance: number; // 连接距离
  connectionOpacity: number; // 连接线透明度
  pulse: boolean; // 是否脉冲效果
  pulseSpeed: number; // 脉冲速度
  rotation: boolean; // 是否旋转
  rotationSpeed: number; // 旋转速度
}

// 交互粒子配置
export interface InteractiveParticleConfig {
  enabled: boolean; // 是否启用
  mouseRadius: number; // 鼠标影响半径
  mouseForce: number; // 鼠标影响力度
  clickEffect: boolean; // 点击效果
  clickRadius: number; // 点击影响半径
  clickForce: number; // 点击力度
  hoverEffect: boolean; // 悬停效果
  connections: boolean; // 粒子间连接
  connectionColor: string; // 连接线颜色
  repulsion: boolean; // 排斥效果
  attraction: boolean; // 吸引效果
}

// AI状态粒子配置
export interface AIStatusParticleConfig {
  enabled: boolean; // 是否启用
  statusEffects: {
    idle: AIStatusEffect;
    listening: AIStatusEffect;
    processing: AIStatusEffect;
    speaking: AIStatusEffect;
    error: AIStatusEffect;
  };
  transitionDuration: number; // 状态转换动画时间
  particleCount: number; // 状态粒子数量
  emitRate: number; // 发射速率
  particleLifetime: number; // 粒子生命周期
}

// AI状态效果
export interface AIStatusEffect {
  color: string; // 主色调
  particleColor: string; // 粒子颜色
  speed: number; // 粒子速度
  size: number; // 粒子大小
  count: number; // 粒子数量
  shape: 'circle' | 'star' | 'heart' | 'wave'; // 粒子形状
  pattern: 'random' | 'spiral' | 'wave' | 'burst'; // 粒子模式
  animation: 'fade' | 'pulse' | 'rotate' | 'scale'; // 动画效果
}

// 性能配置
export interface PerformanceConfig {
  maxParticles: number; // 最大粒子数
  targetFPS: number; // 目标帧率
  adaptiveQuality: boolean; // 自适应质量
  reduceMotion: boolean; // 减少动效
  pixelRatio: number; // 像素比例
  culling: boolean; // 视锥剔除
  updateThrottle: number; // 更新节流
}

// 粒子对象
export interface Particle {
  id: string;
  x: number;
  y: number;
  vx: number; // x方向速度
  vy: number; // y方向速度
  size: number;
  color: string;
  opacity: number;
  rotation: number;
  rotationSpeed: number;
  lifetime: number; // 生命周期
  maxLifetime: number; // 最大生命周期
  type: 'background' | 'interactive' | 'ai-status';
  metadata?: Record<string, any>; // 额外数据
}

// AI状态类型
export type AIStatus =
  | 'idle'
  | 'listening'
  | 'processing'
  | 'speaking'
  | 'error';

// 粒子系统状态
export interface ParticleSystemState {
  particles: Particle[];
  config: ParticleConfig;
  aiStatus: AIStatus;
  mousePosition: { x: number; y: number };
  isInitialized: boolean;
  isRunning: boolean;
  lastUpdateTime: number;
  fps: number;
  performanceStats: PerformanceStats;
}

// 性能统计
export interface PerformanceStats {
  fps: number;
  frameTime: number;
  particleCount: number;
  drawCalls: number;
  memoryUsage: number;
  cpuUsage: number;
}

// 粒子事件
export interface ParticleEvent {
  type:
    | 'mouse-move'
    | 'mouse-click'
    | 'status-change'
    | 'particle-created'
    | 'particle-destroyed';
  timestamp: number;
  data: any;
}

// 粒子效果预设
export interface ParticlePreset {
  id: string;
  name: string;
  description: string;
  config: ParticleConfig;
  thumbnail?: string;
  category: 'minimal' | 'moderate' | 'intensive' | 'custom';
}

// 粒子管理器接口
export interface IParticleSystem {
  init(canvas: HTMLCanvasElement): void;
  start(): void;
  stop(): void;
  update(deltaTime: number): void;
  render(): void;
  setAIStatus(status: AIStatus): void;
  updateConfig(config: Partial<ParticleConfig>): void;
  getStats(): PerformanceStats;
  destroy(): void;
}

// 粒子生成器
export interface ParticleGenerator {
  generate(config: ParticleConfig, count: number): Particle[];
  generateFromPattern(
    pattern: string,
    x: number,
    y: number,
    count: number,
  ): Particle[];
  generateStatusEffect(status: AIStatus, x: number, y: number): Particle[];
}

// 粒子动画器
export interface ParticleAnimator {
  update(particle: Particle, deltaTime: number, config: ParticleConfig): void;
  applyPhysics(particle: Particle, deltaTime: number): void;
  applyMouseInteraction(
    particle: Particle,
    mousePosition: { x: number; y: number },
    config: InteractiveParticleConfig,
  ): void;
  applyStatusEffect(
    particle: Particle,
    status: AIStatus,
    config: AIStatusParticleConfig,
  ): void;
}

// 粒子渲染器
export interface ParticleRenderer {
  render(
    ctx: CanvasRenderingContext2D,
    particles: Particle[],
    config: ParticleConfig,
  ): void;
  renderParticle(ctx: CanvasRenderingContext2D, particle: Particle): void;
  renderConnections(
    ctx: CanvasRenderingContext2D,
    particles: Particle[],
    config: BackgroundParticleConfig,
  ): void;
  clear(ctx: CanvasRenderingContext2D): void;
}
