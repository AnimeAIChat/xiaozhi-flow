import type React from 'react';
import { useCallback, useEffect, useRef, useState } from 'react';
import type {
  AIStatus,
  ParticleConfig,
  PerformanceStats,
} from '../../types/particle';
import { ParticleSystem } from './ParticleSystem';

export interface ParticleSystemComponentProps {
  className?: string;
  type?: 'background' | 'interactive' | 'ai-status' | 'all';
  aiStatus?: AIStatus;
  config?: Partial<ParticleConfig>;
  performance?: boolean;
  onMouseMove?: (position: { x: number; y: number }) => void;
  onClick?: (position: { x: number; y: number }) => void;
}

/**
 * 粒子系统React组件
 * 提供React接口来使用粒子系统
 */
export const ParticleSystemComponent: React.FC<
  ParticleSystemComponentProps
> = ({
  className = '',
  type = 'all',
  aiStatus = 'idle',
  config = {},
  performance = false,
  onMouseMove,
  onClick,
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const particleSystemRef = useRef<ParticleSystem | null>(null);
  const [stats, setStats] = useState<PerformanceStats | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);

  // 初始化粒子系统
  const initializeParticleSystem = useCallback(() => {
    if (!canvasRef.current || particleSystemRef.current) return;

    try {
      const system = new ParticleSystem();
      system.init(canvasRef.current);
      system.start();

      // 应用配置
      if (Object.keys(config).length > 0) {
        system.updateConfig(config);
      }

      // 设置AI状态
      if (type === 'ai-status' || type === 'all') {
        system.setAIStatus(aiStatus);
      }

      particleSystemRef.current = system;
      setIsInitialized(true);

      console.log('[ParticleSystem] Initialized successfully');
    } catch (error) {
      console.error('[ParticleSystem] Failed to initialize:', error);
    }
  }, [config, aiStatus, type]);

  // 清理粒子系统
  const cleanupParticleSystem = useCallback(() => {
    if (particleSystemRef.current) {
      particleSystemRef.current.destroy();
      particleSystemRef.current = null;
      setIsInitialized(false);
    }
  }, []);

  // 更新AI状态
  useEffect(() => {
    if (particleSystemRef.current && (type === 'ai-status' || type === 'all')) {
      particleSystemRef.current.setAIStatus(aiStatus);
    }
  }, [aiStatus, type]);

  // 更新配置
  useEffect(() => {
    if (particleSystemRef.current && Object.keys(config).length > 0) {
      particleSystemRef.current.updateConfig(config);
    }
  }, [config]);

  // 性能监控
  useEffect(() => {
    if (!performance) return;

    const interval = setInterval(() => {
      if (particleSystemRef.current) {
        const currentStats = particleSystemRef.current.getStats();
        setStats(currentStats);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [performance]);

  // 组件挂载时初始化
  useEffect(() => {
    initializeParticleSystem();

    return () => {
      cleanupParticleSystem();
    };
  }, [initializeParticleSystem, cleanupParticleSystem]);

  // 处理鼠标移动
  const handleMouseMove = useCallback(
    (event: React.MouseEvent<HTMLCanvasElement>) => {
      const rect = event.currentTarget.getBoundingClientRect();
      const x = event.clientX - rect.left;
      const y = event.clientY - rect.top;
      onMouseMove?.({ x, y });
    },
    [onMouseMove],
  );

  // 处理点击
  const handleClick = useCallback(
    (event: React.MouseEvent<HTMLCanvasElement>) => {
      const rect = event.currentTarget.getBoundingClientRect();
      const x = event.clientX - rect.left;
      const y = event.clientY - rect.top;
      onClick?.({ x, y });
    },
    [onClick],
  );

  return (
    <div className={`particle-system ${className}`}>
      <canvas
        ref={canvasRef}
        className={`w-full h-full ${isInitialized ? 'block' : 'hidden'}`}
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          pointerEvents: type === 'background' ? 'none' : 'auto',
        }}
        onMouseMove={type !== 'background' ? handleMouseMove : undefined}
        onClick={type !== 'background' ? handleClick : undefined}
      />

      {!isInitialized && (
        <div className="flex items-center justify-center w-full h-full">
          <div className="text-gray-500">正在初始化粒子系统...</div>
        </div>
      )}

      {performance && stats && (
        <div className="absolute top-2 right-2 bg-black bg-opacity-50 text-white text-xs p-2 rounded">
          <div>FPS: {stats.fps}</div>
          <div>粒子数: {stats.particleCount}</div>
          <div>帧时间: {stats.frameTime.toFixed(2)}ms</div>
        </div>
      )}
    </div>
  );
};

export default ParticleSystemComponent;
