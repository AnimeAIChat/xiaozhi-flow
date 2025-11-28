// Authentication related TypeScript types

export interface User {
  id: number;
  username: string;
  email: string;
  nickname: string;
  role: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  expires_at: number;
  user: User;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

export interface AuthContextType extends AuthState {
  login: (credentials: LoginRequest) => Promise<void>;
  register: (userData: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
  checkAuth: () => Promise<void>;
  clearError: () => void;
}

// API Response types
export interface ApiResponse<T = any> {
  success: boolean;
  data: T;
  message: string;
  code: number;
}

export interface ApiError {
  success: false;
  data: { error: string };
  message: string;
  code: number;
}

// Local storage keys
export const AUTH_KEYS = {
  TOKEN: 'auth_token',
  USER: 'auth_user',
  EXPIRES_AT: 'auth_expires_at',
} as const;

// Authentication status
export enum AuthStatus {
  INITIALIZING = 'initializing',
  AUTHENTICATED = 'authenticated',
  UNAUTHENTICATED = 'unauthenticated',
  ERROR = 'error',
}

// Error types
export interface AuthError {
  code: string;
  message: string;
  details?: any;
}

export const AUTH_ERRORS = {
  INVALID_CREDENTIALS: 'INVALID_CREDENTIALS',
  USER_NOT_FOUND: 'USER_NOT_FOUND',
  TOKEN_EXPIRED: 'TOKEN_EXPIRED',
  TOKEN_INVALID: 'TOKEN_INVALID',
  NETWORK_ERROR: 'NETWORK_ERROR',
  REGISTRATION_FAILED: 'REGISTRATION_FAILED',
  USER_EXISTS: 'USER_EXISTS',
  EMAIL_EXISTS: 'EMAIL_EXISTS',
  SESSION_EXPIRED: 'SESSION_EXPIRED',
  SERVER_ERROR: 'SERVER_ERROR',
  UNKNOWN_ERROR: 'UNKNOWN_ERROR',
} as const;

export type AuthErrorCode = keyof typeof AUTH_ERRORS;

// Role-based access
export type UserRole = 'admin' | 'user';

export interface Permission {
  resource: string;
  action: string;
}

export const PERMISSIONS = {
  // Admin permissions
  ADMIN_DASHBOARD: 'admin:dashboard',
  ADMIN_USERS: 'admin:users',
  ADMIN_SYSTEM: 'admin:system',
  ADMIN_CONFIG: 'admin:config',

  // User permissions
  USER_DASHBOARD: 'user:dashboard',
  USER_CONFIG: 'user:config',
  USER_PROFILE: 'user:profile',

  // Common permissions
  READ_OWN_DATA: 'read:own',
  WRITE_OWN_DATA: 'write:own',
} as const;

export type PermissionKey = keyof typeof PERMISSIONS;

// Session management
export interface SessionInfo {
  token: string;
  user: User;
  expiresAt: number;
  refreshToken?: string;
  lastActivity: number;
}

// Token refresh configuration
export const TOKEN_CONFIG = {
  // Time before expiration to attempt refresh (5 minutes)
  REFRESH_THRESHOLD: 5 * 60 * 1000,
  // Maximum retry attempts for token refresh
  MAX_REFRESH_RETRIES: 3,
  // Delay between retry attempts (1 second)
  RETRY_DELAY: 1000,
} as const;

// Route protection levels
export enum AuthLevel {
  PUBLIC = 'public',           // No authentication required
  AUTHENTICATED = 'authenticated', // User must be logged in
  ADMIN = 'admin',             // User must have admin role
  OPTIONAL = 'optional',       // Authentication optional but preferred
}

export interface ProtectedRouteConfig {
  authLevel: AuthLevel;
  redirectTo?: string;
  fallback?: React.ComponentType;
  permissions?: PermissionKey[];
}

// Theme and preferences (can be extended)
export interface UserPreferences {
  theme: 'light' | 'dark' | 'auto';
  language: string;
  timezone: string;
  notifications: {
    email: boolean;
    push: boolean;
    inApp: boolean;
  };
}

export interface AuthUserProfile extends User {
  avatar?: string;
  phone?: string;
  preferences: UserPreferences;
  createdAt: string;
  updatedAt: string;
  lastLoginAt?: string;
}

// Device information for security
export interface DeviceInfo {
  id: string;
  name: string;
  type: 'web' | 'mobile' | 'desktop';
  userAgent: string;
  ip: string;
  lastActiveAt: string;
  isCurrent: boolean;
}

// Security settings
export interface SecuritySettings {
  twoFactorEnabled: boolean;
  emailVerified: boolean;
  passwordUpdatedAt: string;
  loginAttempts: number;
  lockedUntil?: string;
  devices: DeviceInfo[];
}

// Utility functions
export const isTokenExpired = (expiresAt: number): boolean => {
  return Date.now() >= expiresAt;
};

export const shouldRefreshToken = (expiresAt: number): boolean => {
  return Date.now() >= (expiresAt - TOKEN_CONFIG.REFRESH_THRESHOLD);
};

// Auth status helper
export const getAuthStatus = (state: AuthState): AuthStatus => {
  if (state.isLoading) {
    return AuthStatus.INITIALIZING;
  }
  if (state.isAuthenticated) {
    return AuthStatus.AUTHENTICATED;
  }
  if (state.error) {
    return AuthStatus.ERROR;
  }
  return AuthStatus.UNAUTHENTICATED;
};

export const hasPermission = (user: User, permission: PermissionKey): boolean => {
  // Admin has all permissions
  if (user.role === 'admin') {
    return true;
  }

  // User-specific permissions
  const userPermissions: PermissionKey[] = [
    PERMISSIONS.USER_DASHBOARD,
    PERMISSIONS.USER_CONFIG,
    PERMISSIONS.USER_PROFILE,
    PERMISSIONS.READ_OWN_DATA,
    PERMISSIONS.WRITE_OWN_DATA,
  ];

  return userPermissions.includes(permission);
};

export const hasRole = (user: User, role: UserRole): boolean => {
  return user.role === role;
};

export const isAdmin = (user: User): boolean => {
  return hasRole(user, 'admin');
};

// Default values
export const DEFAULT_USER_PREFERENCES: UserPreferences = {
  theme: 'auto',
  language: 'zh-CN',
  timezone: 'Asia/Shanghai',
  notifications: {
    email: true,
    push: false,
    inApp: true,
  },
} as const;