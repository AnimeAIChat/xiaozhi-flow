import { MinimapPlugin } from 'rete-minimap-plugin';
import { getTypeColor } from '../../../utils/nodeUtils';

/**
 * 创建启动流程专用的小地图插件
 */
export const createStartupMinimap = () => {
  return new MinimapPlugin({
    width: 200,
    height: 150,
    position: 'bottom-right',
    offset: { x: 10, y: 10 },
    // 节点颜色函数
    nodeColor: (node: any) => {
      try {
        const nodeType = node.data?.type || 'api';
        const nodeStatus = node.data?.status || 'stopped';

        // 根据状态调整颜色透明度
        const baseColor = getTypeColor(nodeType);
        const alpha =
          nodeStatus === 'running' ? 1 : nodeStatus === 'warning' ? 0.7 : 0.5;

        return hexToRgba(baseColor, alpha);
      } catch (error) {
        return '#d9d9d9'; // 默认灰色
      }
    },
    // 连接线颜色
    connectionColor: '#3b82f6',
    // 背景颜色
    backgroundColor: 'rgba(255, 255, 255, 0.9)',
    // 边框颜色
    borderColor: '#e5e7eb',
    // 显示连接线
    showConnections: true,
    // 缩放比例
    scale: 0.15,
    // 自定义样式
    className: 'startup-minimap',
  });
};

/**
 * 将十六进制颜色转换为 RGBA
 */
const hexToRgba = (hex: string, alpha: number): string => {
  try {
    const r = parseInt(hex.slice(1, 3), 16);
    const g = parseInt(hex.slice(3, 5), 16);
    const b = parseInt(hex.slice(5, 7), 16);
    return `rgba(${r}, ${g}, ${b}, ${alpha})`;
  } catch (error) {
    return `rgba(217, 217, 217, ${alpha})`; // 默认灰色
  }
};
