// 通用类型定义
export interface BaseEntity {
  id: string;
  created_at: string;
  updated_at: string;
}

// API响应类型
export interface ApiResponse<T = any> {
  success: boolean;
  data: T;
  message: string;
  code: number;
}

// 分页响应类型
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// 错误类型
export interface ApiError {
  code: number;
  message: string;
  details?: any;
}

// 状态类型
export type Status = 'idle' | 'loading' | 'success' | 'error';

// 主题类型
export type Theme = 'light' | 'dark';

// 语言类型
export type Language = 'zh-CN' | 'en-US';

// 排序类型
export interface SortOption {
  field: string;
  order: 'asc' | 'desc';
}

// 过滤选项
export interface FilterOption {
  field: string;
  operator: 'eq' | 'ne' | 'gt' | 'gte' | 'lt' | 'lte' | 'like' | 'in';
  value: any;
}

// 通用查询参数
export interface QueryParams {
  page?: number;
  pageSize?: number;
  sort?: SortOption[];
  filter?: FilterOption[];
  search?: string;
}

// 用户相关类型
export interface User extends BaseEntity {
  username: string;
  email: string;
  avatar?: string;
  role: UserRole;
  isActive: boolean;
  lastLoginAt?: string;
}

export type UserRole = 'admin' | 'user' | 'viewer';

// 认证相关类型
export interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  permissions: string[];
}

export interface LoginCredentials {
  username: string;
  password: string;
  remember?: boolean;
}

export interface LoginResult {
  user: User;
  token: string;
  refreshToken: string;
  expiresIn: number;
}

// 配置相关类型
export interface ConfigSection {
  key: string;
  title: string;
  description?: string;
  fields: ConfigField[];
}

export interface ConfigField {
  key: string;
  label: string;
  type:
    | 'string'
    | 'number'
    | 'boolean'
    | 'select'
    | 'multiselect'
    | 'textarea'
    | 'password';
  required?: boolean;
  default?: any;
  options?: Array<{ label: string; value: any }>;
  validation?: {
    min?: number;
    max?: number;
    pattern?: string;
    message?: string;
  };
  description?: string;
  placeholder?: string;
}

// 通知类型
export interface Notification {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message: string;
  duration?: number;
  timestamp: string;
  read: boolean;
}

// 模态框类型
export interface ModalProps {
  visible: boolean;
  title?: string;
  content?: React.ReactNode;
  footer?: React.ReactNode;
  width?: number;
  closable?: boolean;
  maskClosable?: boolean;
  onClose?: () => void;
  onOk?: () => void;
}

// 表格列定义
export interface TableColumn {
  key: string;
  title: string;
  dataIndex: string;
  width?: number;
  fixed?: 'left' | 'right';
  sortable?: boolean;
  filterable?: boolean;
  render?: (value: any, record: any, index: number) => React.ReactNode;
}

// 表单字段类型
export interface FormField {
  name: string;
  label: string;
  type: string;
  required?: boolean;
  rules?: any[];
  component?: React.ComponentType<any>;
  componentProps?: Record<string, any>;
}

// 菜单项类型
export interface MenuItem {
  key: string;
  label: string;
  icon?: React.ReactNode;
  path?: string;
  children?: MenuItem[];
  badge?: string | number;
  disabled?: boolean;
  hidden?: boolean;
}

// 面包屑类型
export interface BreadcrumbItem {
  title: string;
  path?: string;
  icon?: React.ReactNode;
}

// 文件上传类型
export interface UploadFile {
  id: string;
  name: string;
  size: number;
  type: string;
  url?: string;
  status: 'uploading' | 'done' | 'error';
  progress?: number;
}

// 统计数据类型
export interface StatisticData {
  title: string;
  value: number | string;
  prefix?: React.ReactNode;
  suffix?: React.ReactNode;
  precision?: number;
  formatter?: (value: number | string) => string;
  color?: string;
  trend?: {
    value: number;
    isPositive: boolean;
  };
}

// 图表数据类型
export interface ChartData {
  name: string;
  value: number;
  color?: string;
}

export interface TimeSeriesData {
  timestamp: string;
  value: number;
  label?: string;
}

// WebSocket消息类型
export interface WebSocketMessage {
  type: string;
  payload: any;
  timestamp: string;
  id: string;
}

// 系统状态类型
export interface SystemHealth {
  status: 'healthy' | 'warning' | 'error';
  services: ServiceHealth[];
  metrics: SystemMetrics;
  timestamp: string;
}

export interface ServiceHealth {
  name: string;
  status: 'running' | 'stopped' | 'error';
  uptime?: number;
  lastCheck: string;
  error?: string;
}

export interface SystemMetrics {
  cpu: number;
  memory: number;
  disk: number;
  network: {
    in: number;
    out: number;
  };
}

// 数据库表节点相关类型
export interface TableNode {
  id: string;
  name: string;
  type: 'table';
  schema: string;
  rowCount?: number;
  size?: number;
  columns: ColumnNode[];
  indexes?: IndexNode[];
  foreignKeys?: ForeignKeyNode[];
  position: Position;
  style?: NodeStyle;
}

export interface ColumnNode {
  id: string;
  name: string;
  type: string;
  nullable: boolean;
  primaryKey: boolean;
  unique: boolean;
  defaultValue?: any;
  description?: string;
  position: Position;
}

export interface IndexNode {
  id: string;
  name: string;
  columns: string[];
  unique: boolean;
  type: 'btree' | 'hash' | 'gist' | 'gin';
}

export interface ForeignKeyNode {
  id: string;
  name: string;
  sourceTable: string;
  sourceColumn: string;
  targetTable: string;
  targetColumn: string;
  onDelete: 'cascade' | 'restrict' | 'set null' | 'set default';
  onUpdate: 'cascade' | 'restrict' | 'set null' | 'set default';
}

export interface RelationshipEdge {
  id: string;
  source: string;
  target: string;
  type: 'foreign_key' | 'one_to_one' | 'one_to_many' | 'many_to_many';
  label?: string;
  style?: EdgeStyle;
}

export interface Position {
  x: number;
  y: number;
}

export interface NodeStyle {
  backgroundColor?: string;
  borderColor?: string;
  borderWidth?: number;
  textColor?: string;
  fontSize?: number;
  width?: number;
  height?: number;
}

export interface EdgeStyle {
  color?: string;
  width?: number;
  style?: 'solid' | 'dashed' | 'dotted';
  arrowType?: 'arrow' | 'arrowclosed';
}

export interface DatabaseSchema {
  name: string;
  type: 'sqlite' | 'mysql' | 'postgresql';
  tables: TableNode[];
  relationships: RelationshipEdge[];
}

export interface TableNodeViewConfig {
  showColumns: boolean;
  showIndexes: boolean;
  showForeignKeys: boolean;
  showRowCount: boolean;
  layoutType: 'force' | 'hierarchical' | 'circular' | 'grid';
  zoomLevel: number;
  filterText: string;
  selectedTables: string[];
}

// 导出所有类型
export * from './api';
export * from './particle';
export * from './store';
