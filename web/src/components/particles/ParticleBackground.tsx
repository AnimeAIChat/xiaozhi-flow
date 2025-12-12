import type React from 'react';
import type { AIStatus } from '../../types/particle';
import type { ParticleSystemComponentProps } from './ParticleSystemComponent';
import { ParticleSystemComponent } from './ParticleSystemComponent';
import { getRecommendedPreset } from './presets';

export interface ParticleBackgroundProps
  extends Omit<ParticleSystemComponentProps, 'type'> {
  children?: React.ReactNode;
  showControls?: boolean;
  onPresetChange?: (preset: string) => void;
}

/**
 * 粒子背景组件
 * 提供带有粒子效果的背景容器
 */
export const ParticleBackground: React.FC<ParticleBackgroundProps> = ({
  children,
  showControls = false,
  onPresetChange,
  aiStatus = 'idle',
  config,
  performance = false,
  className = '',
  onMouseMove,
  onClick,
}) => {
  // 使用推荐预设或传入的配置
  const presetConfig = config || getRecommendedPreset().config;

  return (
    <div
      className={`particle-background relative w-full h-full overflow-hidden ${className}`}
    >
      {/* 粒子系统 */}
      <ParticleSystemComponent
        type="background"
        aiStatus={aiStatus}
        config={presetConfig}
        performance={performance}
        onMouseMove={onMouseMove}
        onClick={onClick}
        className="absolute inset-0 z-0"
      />

      {/* 内容 */}
      {children && <div className="relative z-10">{children}</div>}

      {/* 控制面板 */}
      {showControls && (
        <div className="absolute top-4 left-4 z-20 bg-black bg-opacity-50 text-white p-4 rounded-lg">
          <h3 className="text-sm font-bold mb-2">粒子效果控制</h3>
          <div className="text-xs space-y-1">
            <div>状态: {aiStatus}</div>
            <div>模式: 背景</div>
            {performance && (
              <div className="text-green-400">性能监控已启用</div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default ParticleBackground;
