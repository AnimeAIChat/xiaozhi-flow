/**
 * 配置管理相关类型定义
 */

// 配置记录接口
export interface ConfigRecord {
  id: number;
  key: string;
  value: any; // JSON 类型
  description?: string;
  category?: string;
  version?: number;
  is_active?: boolean;
  created_at?: string;
  updated_at?: string;
}

// 配置分类
export interface ConfigCategory {
  id: string;
  name: string;
  description?: string;
  icon?: string;
  color?: string;
  count: number;
}

// 配置节点类型（用于画布显示）
export interface ConfigNode {
  id: string;
  type: 'config' | 'category' | 'action';
  position: { x: number; y: number };
  data: {
    key: string;
    label: string;
    description?: string;
    category?: string;
    value?: any;
    dataType?: string;
    required?: boolean;
    editable?: boolean;
    icon?: string;
    color?: string;
  };
}

// 配置连接线
export interface ConfigEdge {
  id: string;
  source: string;
  target: string;
  type?: 'dependency' | 'reference' | 'logic';
  label?: string;
  animated?: boolean;
  style?: {
    stroke?: string;
    strokeWidth?: number;
  };
}

// 配置画布状态
export interface ConfigCanvasState {
  nodes: ConfigNode[];
  edges: ConfigEdge[];
  selectedNode?: string;
  editingNode?: string;
  viewport: {
    x: number;
    y: number;
    zoom: number;
  };
  history: {
    past: ConfigCanvasState[];
    present: ConfigCanvasState;
    future: ConfigCanvasState[];
  };
}

// 配置编辑器模式
export type ConfigEditMode =
  | 'view'      // 查看模式
  | 'edit'      // 编辑模式
  | 'connect'   // 连接模式
  | 'debug';    // 调试模式

// 配置验证结果
export interface ConfigValidation {
  isValid: boolean;
  errors: ConfigValidationError[];
  warnings: ConfigValidationWarning[];
}

export interface ConfigValidationError {
  key: string;
  message: string;
  severity: 'error' | 'warning' | 'info';
  path?: string;
}

export interface ConfigValidationWarning extends ConfigValidationError {}

// 配置导出/导入格式
export interface ConfigExport {
  version: string;
  timestamp: string;
  categories: ConfigCategory[];
  configs: ConfigRecord[];
  relationships: ConfigEdge[];
  metadata: {
    exportedBy?: string;
    description?: string;
    tags?: string[];
  };
}

// 配置搜索过滤器
export interface ConfigFilter {
  category?: string;
  dataType?: string;
  isActive?: boolean;
  searchText?: string;
  tags?: string[];
}

// 配置更新操作
export interface ConfigUpdateOperation {
  type: 'create' | 'update' | 'delete' | 'move';
  key?: string;
  value?: any;
  oldValue?: any;
  newValue?: any;
  timestamp: string;
  user?: string;
}

// 配置画布布局算法
export type ConfigLayoutAlgorithm =
  | 'force'      // 力导向布局
  | 'hierarchical' // 层次布局
  | 'circular'   // 圆形布局
  | 'grid'       // 网格布局
  | 'tree';      // 树形布局

// 配置节点渲染器
export interface ConfigNodeRenderer {
  type: string;
  component: React.ComponentType<any>;
  defaultProps?: any;
  validator?: (value: any) => ConfigValidation;
}

// 配置上下文菜单项
export interface ConfigContextMenuItem {
  key: string;
  label: string;
  icon?: string;
  disabled?: boolean;
  children?: ConfigContextMenuItem[];
  action?: (node: ConfigNode) => void;
  separator?: boolean;
}

// 配置快照
export interface ConfigSnapshot {
  id: string;
  name: string;
  description?: string;
  version: string;
  data: ConfigCanvasState;
  created_at: string;
  created_by?: string;
  tags?: string[];
}

// 配置模板
export interface ConfigTemplate {
  id: string;
  name: string;
  description?: string;
  category: string;
  nodes: ConfigNode[];
  edges: ConfigEdge[];
  thumbnail?: string;
  created_at: string;
  usage_count?: number;
}

// 配置批处理操作
export interface ConfigBatchOperation {
  id: string;
  type: 'update' | 'delete' | 'export' | 'import';
  keys: string[];
  data?: any;
  status: 'pending' | 'running' | 'completed' | 'failed';
  progress?: number;
  result?: any;
  error?: string;
  created_at: string;
  completed_at?: string;
}