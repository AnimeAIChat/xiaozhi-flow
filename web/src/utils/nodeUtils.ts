// 节点工具函数

export type NodeType = 'database' | 'api' | 'ai' | 'cloud' | 'config';
export type NodeStatus = 'running' | 'stopped' | 'warning';

/**
 * 根据节点类型获取图标
 */
export const getIconByNodeType = (type: NodeType): string => {
  switch (type) {
    case 'database':
      return '🗄️';
    case 'api':
      return '🔌';
    case 'ai':
      return '🤖';
    case 'cloud':
      return '☁️';
    case 'config':
      return '⚙️';
    default:
      return '📦';
  }
};

/**
 * 根据节点状态获取颜色
 */
export const getStatusColor = (status: NodeStatus): string => {
  switch (status) {
    case 'running':
      return '#52c41a'; // green
    case 'warning':
      return '#faad14'; // orange
    case 'stopped':
      return '#ff4d4f'; // red
    default:
      return '#d9d9d9'; // gray
  }
};

/**
 * 根据节点类型获取颜色
 */
export const getTypeColor = (type: NodeType): string => {
  switch (type) {
    case 'database':
      return '#1890ff'; // blue
    case 'api':
      return '#722ed1'; // purple
    case 'ai':
      return '#13c2c2'; // cyan
    case 'cloud':
      return '#fa8c16'; // orange
    case 'config':
      return '#52c41a'; // green
    default:
      return '#d9d9d9'; // gray
  }
};

/**
 * 获取节点的默认尺寸
 */
export const getNodeSize = (
  type: NodeType,
): { width: number; height: number } => {
  switch (type) {
    case 'database':
      return { width: 180, height: 100 };
    case 'api':
      return { width: 160, height: 90 };
    case 'ai':
      return { width: 200, height: 110 };
    case 'cloud':
      return { width: 170, height: 95 };
    case 'config':
      return { width: 190, height: 100 };
    default:
      return { width: 180, height: 100 };
  }
};

/**
 * 获取节点类型的显示名称
 */
export const getNodeTypeName = (type: NodeType): string => {
  switch (type) {
    case 'database':
      return '数据库';
    case 'api':
      return 'API服务';
    case 'ai':
      return 'AI模型';
    case 'cloud':
      return '云服务';
    case 'config':
      return '配置';
    default:
      return '未知类型';
  }
};

/**
 * 获取状态的显示名称
 */
export const getStatusName = (status: NodeStatus): string => {
  switch (status) {
    case 'running':
      return '运行中';
    case 'warning':
      return '警告';
    case 'stopped':
      return '已停止';
    default:
      return '未知状态';
  }
};
