import type {
  IParticleSystem,
  Particle,
  ParticleConfig,
  ParticleSystemState,
  AIStatus,
  PerformanceStats,
  BackgroundParticleConfig,
  InteractiveParticleConfig,
  AIStatusParticleConfig,
} from '../../types/particle';

/**
 * 粒子系统核心类
 * 管理粒子的生成、更新、渲染和交互
 */
export class ParticleSystem implements IParticleSystem {
  private canvas: HTMLCanvasElement | null = null;
  private ctx: CanvasRenderingContext2D | null = null;
  private animationId: number | null = null;
  private lastTime = 0;
  private isInitialized = false;
  private isRunning = false;

  // 状态
  private state: ParticleSystemState = {
    particles: [],
    config: this.getDefaultConfig(),
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

  // 性能优化
  private frameCount = 0;
  private fpsUpdateTime = 0;
  private updateThrottleCounter = 0;

  /**
   * 获取默认粒子配置
   */
  private getDefaultConfig(): ParticleConfig {
    return {
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
        maxParticles: 200,
        targetFPS: 60,
        adaptiveQuality: true,
        reduceMotion: false,
        pixelRatio: 1,
        culling: true,
        updateThrottle: 16, // 约60fps
      },
    };
  }

  /**
   * 初始化粒子系统
   */
  public init(canvas: HTMLCanvasElement): void {
    this.canvas = canvas;
    const ctx = canvas.getContext('2d');
    if (!ctx) {
      throw new Error('Unable to get 2D context from canvas');
    }
    this.ctx = ctx;

    // 设置画布大小
    this.resizeCanvas();

    // 监听窗口大小变化
    window.addEventListener('resize', () => this.resizeCanvas());

    // 监听鼠标移动
    canvas.addEventListener('mousemove', (e) => {
      const rect = canvas.getBoundingClientRect();
      this.state.mousePosition = {
        x: e.clientX - rect.left,
        y: e.clientY - rect.top,
      };
    });

    // 监听鼠标点击
    canvas.addEventListener('click', (e) => {
      const rect = canvas.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;
      this.handleClick(x, y);
    });

    // 生成初始粒子
    this.generateBackgroundParticles();

    this.state.isInitialized = true;
    this.isInitialized = true;
  }

  /**
   * 调整画布大小
   */
  private resizeCanvas(): void {
    if (!this.canvas) return;

    const container = this.canvas.parentElement;
    if (!container) return;

    this.canvas.width = container.clientWidth;
    this.canvas.height = container.clientHeight;

    // 设置设备像素比
    const dpr = this.state.config.performance.pixelRatio;
    if (dpr > 1) {
      this.canvas.width = this.canvas.width * dpr;
      this.canvas.height = this.canvas.height * dpr;
      this.canvas.style.width = `${container.clientWidth}px`;
      this.canvas.style.height = `${container.clientHeight}px`;
    }

    if (this.ctx) {
      this.ctx.scale(dpr, dpr);
    }
  }

  /**
   * 生成背景粒子
   */
  private generateBackgroundParticles(): void {
    const config = this.state.config.background;
    const particlesToGenerate = Math.min(config.count, this.state.config.performance.maxParticles);

    for (let i = 0; i < particlesToGenerate; i++) {
      this.state.particles.push(this.createBackgroundParticle());
    }
  }

  /**
   * 创建背景粒子
   */
  private createBackgroundParticle(): Particle {
    const config = this.state.config.background;
    const canvas = this.canvas!;

    return {
      id: `bg-${Date.now()}-${Math.random()}`,
      x: Math.random() * canvas.clientWidth,
      y: Math.random() * canvas.clientHeight,
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
    };
  }

  /**
   * 创建AI状态粒子
   */
  private createAIStatusParticle(x: number, y: number): Particle | null {
    const config = this.state.config.aiStatus;
    if (!config.enabled) return null;

    const statusConfig = config.statusEffects[this.state.aiStatus];
    const canvas = this.canvas!;

    // 限制AI状态粒子数量
    const aiStatusParticles = this.state.particles.filter(p => p.type === 'ai-status');
    if (aiStatusParticles.length >= config.particleCount) {
      return null;
    }

    let vx = 0, vy = 0;

    // 根据模式设置速度方向
    switch (statusConfig.pattern) {
      case 'random':
        vx = (Math.random() - 0.5) * statusConfig.speed;
        vy = (Math.random() - 0.5) * statusConfig.speed;
        break;
      case 'wave':
        const angle = Math.random() * Math.PI * 2;
        vx = Math.cos(angle) * statusConfig.speed;
        vy = Math.sin(angle) * statusConfig.speed;
        break;
      case 'spiral':
        const spiralAngle = Math.random() * Math.PI * 2;
        vx = Math.cos(spiralAngle) * statusConfig.speed;
        vy = Math.sin(spiralAngle) * statusConfig.speed;
        break;
      case 'burst':
        const burstAngle = Math.random() * Math.PI * 2;
        vx = Math.cos(burstAngle) * statusConfig.speed * 2;
        vy = Math.sin(burstAngle) * statusConfig.speed * 2;
        break;
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
        status: this.state.aiStatus,
        animation: statusConfig.animation,
      },
    };
  }

  /**
   * 处理鼠标点击
   */
  private handleClick(x: number, y: number): void {
    const config = this.state.config.interactive;
    if (!config.clickEffect) return;

    // 在点击位置创建爆发效果
    for (let i = 0; i < 10; i++) {
      const angle = (Math.PI * 2 * i) / 10;
      const speed = config.clickForce * 2;
      const particle: Particle = {
        id: `click-${Date.now()}-${i}`,
        x,
        y,
        vx: Math.cos(angle) * speed,
        vy: Math.sin(angle) * speed,
        size: 3,
        color: config.connectionColor,
        opacity: 1,
        rotation: 0,
        rotationSpeed: 0.1,
        lifetime: 1000,
        maxLifetime: 1000,
        type: 'interactive',
      };
      this.state.particles.push(particle);
    }
  }

  /**
   * 开始粒子系统
   */
  public start(): void {
    if (!this.isInitialized) {
      throw new Error('ParticleSystem must be initialized before starting');
    }
    this.isRunning = true;
    this.state.isRunning = true;
    this.lastTime = performance.now();
    this.animate();
  }

  /**
   * 停止粒子系统
   */
  public stop(): void {
    this.isRunning = false;
    this.state.isRunning = false;
    if (this.animationId) {
      cancelAnimationFrame(this.animationId);
      this.animationId = null;
    }
  }

  /**
   * 动画循环
   */
  private animate(): void {
    if (!this.isRunning || !this.canvas || !this.ctx) return;

    const currentTime = performance.now();
    const deltaTime = currentTime - this.lastTime;
    this.lastTime = currentTime;

    // 性能节流
    this.updateThrottleCounter += deltaTime;
    if (this.updateThrottleCounter < this.state.config.performance.updateThrottle) {
      this.animationId = requestAnimationFrame(() => this.animate());
      return;
    }
    this.updateThrottleCounter = 0;

    // 更新FPS
    this.updateFPS(currentTime);

    // 清空画布
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

    // 更新粒子
    this.update(deltaTime);

    // 渲染粒子
    this.render();

    // 生成AI状态粒子
    this.emitAIStatusParticles();

    this.animationId = requestAnimationFrame(() => this.animate());
  }

  /**
   * 更新FPS计算
   */
  private updateFPS(currentTime: number): void {
    this.frameCount++;
    if (currentTime - this.fpsUpdateTime >= 1000) {
      this.state.fps = this.frameCount;
      this.state.performanceStats.fps = this.frameCount;
      this.frameCount = 0;
      this.fpsUpdateTime = currentTime;
    }
  }

  /**
   * 更新粒子
   */
  private update(deltaTime: number): void {
    const canvas = this.canvas!;
    const config = this.state.config;

    // 移除死亡的粒子
    this.state.particles = this.state.particles.filter(particle => {
      particle.lifetime -= deltaTime;
      return particle.lifetime > 0 && particle.maxLifetime !== particle.lifetime;
    });

    // 更新每个粒子
    this.state.particles.forEach(particle => {
      this.updateParticle(particle, deltaTime);
    });

    // 补充背景粒子
    const backgroundParticles = this.state.particles.filter(p => p.type === 'background');
    const neededBackgroundParticles = Math.max(0, config.background.count - backgroundParticles.length);
    for (let i = 0; i < neededBackgroundParticles; i++) {
      this.state.particles.push(this.createBackgroundParticle());
    }

    // 限制总粒子数
    if (this.state.particles.length > config.performance.maxParticles) {
      this.state.particles = this.state.particles.slice(0, config.performance.maxParticles);
    }

    // 更新性能统计
    this.state.performanceStats.particleCount = this.state.particles.length;
    this.state.performanceStats.frameTime = deltaTime;
  }

  /**
   * 更新单个粒子
   */
  private updateParticle(particle: Particle, deltaTime: number): void {
    const canvas = this.canvas!;
    const config = this.state.config.interactive;

    // 基础物理更新
    particle.x += particle.vx;
    particle.y += particle.vy;
    particle.rotation += particle.rotationSpeed;

    // 边界检测
    if (particle.x < 0 || particle.x > canvas.clientWidth) {
      particle.vx = -particle.vx;
      particle.x = Math.max(0, Math.min(canvas.clientWidth, particle.x));
    }
    if (particle.y < 0 || particle.y > canvas.clientHeight) {
      particle.vy = -particle.vy;
      particle.y = Math.max(0, Math.min(canvas.clientHeight, particle.y));
    }

    // 鼠标交互
    if (config.enabled) {
      const dx = this.state.mousePosition.x - particle.x;
      const dy = this.state.mousePosition.y - particle.y;
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
    }

    // 生命周期透明度
    if (particle.lifetime !== Infinity) {
      particle.opacity = Math.max(0, particle.lifetime / particle.maxLifetime);
    }
  }

  /**
   * 渲染粒子
   */
  private render(): void {
    if (!this.ctx || !this.canvas) return;

    const config = this.state.config;
    let drawCalls = 0;

    // 渲染背景粒子连接线
    if (config.background.connectDistance > 0) {
      this.renderConnections();
      drawCalls++;
    }

    // 渲染粒子
    this.state.particles.forEach(particle => {
      this.renderParticle(particle);
      drawCalls++;
    });

    this.state.performanceStats.drawCalls = drawCalls;
  }

  /**
   * 渲染粒子连接线
   */
  private renderConnections(): void {
    if (!this.ctx) return;

    const config = this.state.config.background;
    const backgroundParticles = this.state.particles.filter(p => p.type === 'background');

    for (let i = 0; i < backgroundParticles.length; i++) {
      for (let j = i + 1; j < backgroundParticles.length; j++) {
        const p1 = backgroundParticles[i];
        const p2 = backgroundParticles[j];
        const distance = Math.sqrt(
          Math.pow(p1.x - p2.x, 2) + Math.pow(p1.y - p2.y, 2)
        );

        if (distance < config.connectDistance) {
          const opacity = (1 - distance / config.connectDistance) * config.connectionOpacity;
          this.ctx.strokeStyle = `${config.color}${Math.floor(opacity * 255).toString(16).padStart(2, '0')}`;
          this.ctx.lineWidth = 0.5;
          this.ctx.beginPath();
          this.ctx.moveTo(p1.x, p1.y);
          this.ctx.lineTo(p2.x, p2.y);
          this.ctx.stroke();
        }
      }
    }
  }

  /**
   * 渲染单个粒子
   */
  private renderParticle(particle: Particle): void {
    if (!this.ctx) return;

    this.ctx.save();
    this.ctx.globalAlpha = particle.opacity;
    this.ctx.fillStyle = particle.color;
    this.ctx.translate(particle.x, particle.y);
    this.ctx.rotate(particle.rotation);

    const bgConfig = this.state.config.background;

    switch (bgConfig.shape) {
      case 'circle':
        this.ctx.beginPath();
        this.ctx.arc(0, 0, particle.size, 0, Math.PI * 2);
        this.ctx.fill();
        break;
      case 'square':
        this.ctx.fillRect(-particle.size, -particle.size, particle.size * 2, particle.size * 2);
        break;
      case 'triangle':
        this.ctx.beginPath();
        this.ctx.moveTo(0, -particle.size);
        this.ctx.lineTo(-particle.size, particle.size);
        this.ctx.lineTo(particle.size, particle.size);
        this.ctx.closePath();
        this.ctx.fill();
        break;
    }

    this.ctx.restore();
  }

  /**
   * 发射AI状态粒子
   */
  private emitAIStatusParticles(): void {
    const config = this.state.config.aiStatus;
    if (!config.enabled) return;

    // 随机从画布边缘发射粒子
    if (Math.random() < config.emitRate / 60) {
      const canvas = this.canvas!;
      let x = 0, y = 0;

      switch (Math.floor(Math.random() * 4)) {
        case 0: // 左边
          x = 0;
          y = Math.random() * canvas.clientHeight;
          break;
        case 1: // 右边
          x = canvas.clientWidth;
          y = Math.random() * canvas.clientHeight;
          break;
        case 2: // 顶部
          x = Math.random() * canvas.clientWidth;
          y = 0;
          break;
        case 3: // 底部
          x = Math.random() * canvas.clientWidth;
          y = canvas.clientHeight;
          break;
      }

      const particle = this.createAIStatusParticle(x, y);
      if (particle) {
        this.state.particles.push(particle);
      }
    }
  }

  /**
   * 设置AI状态
   */
  public setAIStatus(status: AIStatus): void {
    this.state.aiStatus = status;
  }

  /**
   * 更新配置
   */
  public updateConfig(config: Partial<ParticleConfig>): void {
    this.state.config = { ...this.state.config, ...config };
  }

  /**
   * 获取性能统计
   */
  public getStats(): PerformanceStats {
    return { ...this.state.performanceStats };
  }

  /**
   * 销毁粒子系统
   */
  public destroy(): void {
    this.stop();
    this.state.particles = [];
    this.state.isInitialized = false;

    if (this.canvas) {
      window.removeEventListener('resize', () => this.resizeCanvas());
      this.canvas.removeEventListener('mousemove', () => {});
      this.canvas.removeEventListener('click', () => {});
    }

    this.canvas = null;
    this.ctx = null;
  }
}