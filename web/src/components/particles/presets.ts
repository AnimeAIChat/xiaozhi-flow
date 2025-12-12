import type { ParticleConfig, ParticlePreset } from '../../types/particle';

/**
 * 粒子效果预设配置
 * 提供不同风格和性能级别的粒子效果
 */

// 轻量级预设 - 适合低端设备
export const minimalPreset: ParticlePreset = {
  id: 'minimal',
  name: '轻量级',
  description: '最少的粒子数量，适合性能要求高的场景',
  category: 'minimal',
  config: {
    background: {
      count: 20,
      speed: 0.3,
      size: 1.5,
      opacity: 0.4,
      color: '#1890ff',
      shape: 'circle',
      connectDistance: 100,
      connectionOpacity: 0.1,
      pulse: false,
      pulseSpeed: 1,
      rotation: false,
      rotationSpeed: 0.5,
    },
    interactive: {
      enabled: true,
      mouseRadius: 80,
      mouseForce: 0.3,
      clickEffect: true,
      clickRadius: 150,
      clickForce: 1,
      hoverEffect: true,
      connections: false,
      connectionColor: '#1890ff',
      repulsion: false,
      attraction: false,
    },
    aiStatus: {
      enabled: true,
      statusEffects: {
        idle: {
          color: '#1890ff',
          particleColor: '#1890ff',
          speed: 0.2,
          size: 2,
          count: 10,
          shape: 'circle',
          pattern: 'random',
          animation: 'fade',
        },
        listening: {
          color: '#52c41a',
          particleColor: '#52c41a',
          speed: 0.4,
          size: 3,
          count: 15,
          shape: 'circle',
          pattern: 'wave',
          animation: 'pulse',
        },
        processing: {
          color: '#faad14',
          particleColor: '#faad14',
          speed: 0.6,
          size: 2.5,
          count: 20,
          shape: 'star',
          pattern: 'spiral',
          animation: 'rotate',
        },
        speaking: {
          color: '#722ed1',
          particleColor: '#722ed1',
          speed: 0.3,
          size: 3,
          count: 18,
          shape: 'heart',
          pattern: 'burst',
          animation: 'scale',
        },
        error: {
          color: '#ff4d4f',
          particleColor: '#ff4d4f',
          speed: 0.8,
          size: 2.5,
          count: 12,
          shape: 'triangle',
          pattern: 'random',
          animation: 'shake',
        },
      },
      transitionDuration: 300,
      particleCount: 15,
      emitRate: 1,
      particleLifetime: 2000,
    },
    performance: {
      maxParticles: 50,
      targetFPS: 60,
      adaptiveQuality: true,
      reduceMotion: true,
      pixelRatio: 1,
      culling: true,
      updateThrottle: 32, // 约30fps
    },
  },
};

// 标准预设 - 平衡视觉效果和性能
export const standardPreset: ParticlePreset = {
  id: 'standard',
  name: '标准',
  description: '平衡的视觉效果和性能表现',
  category: 'moderate',
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
      updateThrottle: 16, // 约60fps
    },
  },
};

// 强烈效果预设 - 适合高端设备
export const intensivePreset: ParticlePreset = {
  id: 'intensive',
  name: '强烈效果',
  description: '丰富的视觉效果，适合高端设备',
  category: 'intensive',
  config: {
    background: {
      count: 100,
      speed: 0.8,
      size: 3,
      opacity: 0.8,
      color: '#1890ff',
      shape: 'circle',
      connectDistance: 200,
      connectionOpacity: 0.3,
      pulse: true,
      pulseSpeed: 3,
      rotation: true,
      rotationSpeed: 2,
    },
    interactive: {
      enabled: true,
      mouseRadius: 150,
      mouseForce: 0.8,
      clickEffect: true,
      clickRadius: 300,
      clickForce: 3,
      hoverEffect: true,
      connections: true,
      connectionColor: '#1890ff',
      repulsion: true,
      attraction: true,
    },
    aiStatus: {
      enabled: true,
      statusEffects: {
        idle: {
          color: '#1890ff',
          particleColor: '#1890ff',
          speed: 0.5,
          size: 4,
          count: 40,
          shape: 'circle',
          pattern: 'random',
          animation: 'fade',
        },
        listening: {
          color: '#52c41a',
          particleColor: '#52c41a',
          speed: 1.2,
          size: 6,
          count: 50,
          shape: 'circle',
          pattern: 'wave',
          animation: 'pulse',
        },
        processing: {
          color: '#faad14',
          particleColor: '#faad14',
          speed: 1.8,
          size: 5,
          count: 60,
          shape: 'star',
          pattern: 'spiral',
          animation: 'rotate',
        },
        speaking: {
          color: '#722ed1',
          particleColor: '#722ed1',
          speed: 1.0,
          size: 7,
          count: 55,
          shape: 'heart',
          pattern: 'burst',
          animation: 'scale',
        },
        error: {
          color: '#ff4d4f',
          particleColor: '#ff4d4f',
          speed: 2.0,
          size: 5,
          count: 40,
          shape: 'triangle',
          pattern: 'random',
          animation: 'shake',
        },
      },
      transitionDuration: 700,
      particleCount: 50,
      emitRate: 3,
      particleLifetime: 4000,
    },
    performance: {
      maxParticles: 300,
      targetFPS: 60,
      adaptiveQuality: true,
      reduceMotion: false,
      pixelRatio: 2,
      culling: true,
      updateThrottle: 16, // 约60fps
    },
  },
};

// 科技感预设 - 蓝色调的科技风格
export const techPreset: ParticlePreset = {
  id: 'tech',
  name: '科技感',
  description: '蓝色调的科技风格效果',
  category: 'moderate',
  config: {
    background: {
      count: 60,
      speed: 0.4,
      size: 1.5,
      opacity: 0.7,
      color: '#00d4ff',
      shape: 'circle',
      connectDistance: 180,
      connectionOpacity: 0.4,
      pulse: true,
      pulseSpeed: 1.5,
      rotation: true,
      rotationSpeed: 0.8,
    },
    interactive: {
      enabled: true,
      mouseRadius: 120,
      mouseForce: 0.6,
      clickEffect: true,
      clickRadius: 250,
      clickForce: 2.5,
      hoverEffect: true,
      connections: true,
      connectionColor: '#00d4ff',
      repulsion: true,
      attraction: false,
    },
    aiStatus: {
      enabled: true,
      statusEffects: {
        idle: {
          color: '#00d4ff',
          particleColor: '#00d4ff',
          speed: 0.4,
          size: 2.5,
          count: 25,
          shape: 'circle',
          pattern: 'random',
          animation: 'fade',
        },
        listening: {
          color: '#00ff88',
          particleColor: '#00ff88',
          speed: 0.9,
          size: 4,
          count: 35,
          shape: 'circle',
          pattern: 'wave',
          animation: 'pulse',
        },
        processing: {
          color: '#ff9500',
          particleColor: '#ff9500',
          speed: 1.5,
          size: 3.5,
          count: 45,
          shape: 'star',
          pattern: 'spiral',
          animation: 'rotate',
        },
        speaking: {
          color: '#b300ff',
          particleColor: '#b300ff',
          speed: 0.7,
          size: 5,
          count: 40,
          shape: 'heart',
          pattern: 'burst',
          animation: 'scale',
        },
        error: {
          color: '#ff0040',
          particleColor: '#ff0040',
          speed: 1.8,
          size: 3.5,
          count: 30,
          shape: 'triangle',
          pattern: 'random',
          animation: 'shake',
        },
      },
      transitionDuration: 600,
      particleCount: 35,
      emitRate: 2.5,
      particleLifetime: 3500,
    },
    performance: {
      maxParticles: 180,
      targetFPS: 60,
      adaptiveQuality: true,
      reduceMotion: false,
      pixelRatio: 1,
      culling: true,
      updateThrottle: 16,
    },
  },
};

// 自然风格预设 - 温暖的自然色调
export const naturePreset: ParticlePreset = {
  id: 'nature',
  name: '自然风格',
  description: '温暖的自然色调效果',
  category: 'moderate',
  config: {
    background: {
      count: 45,
      speed: 0.3,
      size: 2.5,
      opacity: 0.5,
      color: '#4caf50',
      shape: 'circle',
      connectDistance: 120,
      connectionOpacity: 0.15,
      pulse: true,
      pulseSpeed: 1,
      rotation: false,
      rotationSpeed: 0.5,
    },
    interactive: {
      enabled: true,
      mouseRadius: 90,
      mouseForce: 0.4,
      clickEffect: true,
      clickRadius: 180,
      clickForce: 1.5,
      hoverEffect: true,
      connections: true,
      connectionColor: '#4caf50',
      repulsion: false,
      attraction: true,
    },
    aiStatus: {
      enabled: true,
      statusEffects: {
        idle: {
          color: '#4caf50',
          particleColor: '#4caf50',
          speed: 0.25,
          size: 3,
          count: 18,
          shape: 'circle',
          pattern: 'random',
          animation: 'fade',
        },
        listening: {
          color: '#8bc34a',
          particleColor: '#8bc34a',
          speed: 0.6,
          size: 4.5,
          count: 28,
          shape: 'circle',
          pattern: 'wave',
          animation: 'pulse',
        },
        processing: {
          color: '#ffc107',
          particleColor: '#ffc107',
          speed: 1.0,
          size: 3.5,
          count: 38,
          shape: 'star',
          pattern: 'spiral',
          animation: 'rotate',
        },
        speaking: {
          color: '#ff5722',
          particleColor: '#ff5722',
          speed: 0.5,
          size: 5,
          count: 33,
          shape: 'heart',
          pattern: 'burst',
          animation: 'scale',
        },
        error: {
          color: '#f44336',
          particleColor: '#f44336',
          speed: 1.2,
          size: 3.5,
          count: 23,
          shape: 'triangle',
          pattern: 'random',
          animation: 'shake',
        },
      },
      transitionDuration: 400,
      particleCount: 25,
      emitRate: 1.8,
      particleLifetime: 2500,
    },
    performance: {
      maxParticles: 120,
      targetFPS: 60,
      adaptiveQuality: true,
      reduceMotion: false,
      pixelRatio: 1,
      culling: true,
      updateThrottle: 20,
    },
  },
};

// 所有预设
export const particlePresets: ParticlePreset[] = [
  minimalPreset,
  standardPreset,
  intensivePreset,
  techPreset,
  naturePreset,
];

// 根据设备性能推荐预设
export function getRecommendedPreset(): ParticlePreset {
  // 简单的性能检测
  const canvas = document.createElement('canvas');
  const gl =
    canvas.getContext('webgl') || canvas.getContext('experimental-webgl');

  if (!gl) {
    // 不支持WebGL，使用轻量级预设
    return minimalPreset;
  }

  // 检测GPU信息
  const debugInfo = gl.getExtension('WEBGL_debug_renderer_info');
  if (debugInfo) {
    const renderer = gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL);

    // 根据GPU信息判断性能
    if (renderer.includes('Intel') || renderer.includes('Mali')) {
      return minimalPreset; // 集成显卡，使用轻量级
    }
    if (
      renderer.includes('NVIDIA') ||
      renderer.includes('AMD') ||
      renderer.includes('Radeon')
    ) {
      return intensivePreset; // 独立显卡，可以承受更多粒子
    }
  }

  // 检测设备内存
  if ('memory' in navigator && navigator.memory) {
    const memoryGB =
      (navigator.memory as any).totalJSHeapSize / (1024 * 1024 * 1024);
    if (memoryGB < 4) {
      return minimalPreset;
    }
    if (memoryGB > 8) {
      return intensivePreset;
    }
  }

  // 默认返回标准预设
  return standardPreset;
}

// 根据用户偏好获取预设
export function getPresetById(id: string): ParticlePreset | undefined {
  return particlePresets.find((preset) => preset.id === id);
}

export default {
  minimalPreset,
  standardPreset,
  intensivePreset,
  techPreset,
  naturePreset,
  particlePresets,
  getRecommendedPreset,
  getPresetById,
};
