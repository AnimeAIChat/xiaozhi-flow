import React, {
  createContext,
  useContext,
  useReducer,
  useEffect,
  useCallback,
  ReactNode,
} from 'react';
import {
  AuthState,
  AuthContextType,
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  ApiResponse,
  ApiError,
  User,
  AUTH_KEYS,
  AUTH_ERRORS,
  AuthStatus,
  TOKEN_CONFIG,
} from '../types/auth';
import { apiService } from '../services/api';

// Auth action types
type AuthAction =
  | { type: 'AUTH_START' }
  | { type: 'AUTH_SUCCESS'; payload: { user: User; token: string; expiresAt: number } }
  | { type: 'AUTH_FAILURE'; payload: string }
  | { type: 'AUTH_LOGOUT' }
  | { type: 'AUTH_REFRESH'; payload: { token: string; expiresAt: number } }
  | { type: 'CLEAR_ERROR' }
  | { type: 'SET_LOADING'; payload: boolean };

// Initial state
const initialState: AuthState = {
  user: null,
  token: null,
  isAuthenticated: false,
  isLoading: true,
  error: null,
};

// Reducer function
const authReducer = (state: AuthState, action: AuthAction): AuthState => {
  switch (action.type) {
    case 'AUTH_START':
      return {
        ...state,
        isLoading: true,
        error: null,
      };

    case 'AUTH_SUCCESS':
      return {
        ...state,
        user: action.payload.user,
        token: action.payload.token,
        isAuthenticated: true,
        isLoading: false,
        error: null,
      };

    case 'AUTH_FAILURE':
      return {
        ...state,
        user: null,
        token: null,
        isAuthenticated: false,
        isLoading: false,
        error: action.payload,
      };

    case 'AUTH_LOGOUT':
      return {
        ...state,
        user: null,
        token: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      };

    case 'AUTH_REFRESH':
      return {
        ...state,
        token: action.payload.token,
        error: null,
      };

    case 'CLEAR_ERROR':
      return {
        ...state,
        error: null,
      };

    case 'SET_LOADING':
      return {
        ...state,
        isLoading: action.payload,
      };

    default:
      return state;
  }
};

// Local storage utilities
const storage = {
  getToken: (): string | null => {
    const token = localStorage.getItem(AUTH_KEYS.TOKEN);
    console.log('Storage: getToken called, result:', !!token);
    return token;
  },

  setToken: (token: string): void => {
    console.log('Storage: setToken called, token length:', token.length);
    localStorage.setItem(AUTH_KEYS.TOKEN, token);
    console.log('Storage: setToken completed');
  },

  removeToken: (): void => {
    console.log('Storage: removeToken called');
    localStorage.removeItem(AUTH_KEYS.TOKEN);
  },

  getUser: (): User | null => {
    const userStr = localStorage.getItem(AUTH_KEYS.USER);
    const user = userStr ? JSON.parse(userStr) : null;
    console.log('Storage: getUser called, result:', !!user);
    return user;
  },

  setUser: (user: User): void => {
    console.log('Storage: setUser called, username:', user.username);
    localStorage.setItem(AUTH_KEYS.USER, JSON.stringify(user));
    console.log('Storage: setUser completed');
  },

  removeUser: (): void => {
    console.log('Storage: removeUser called');
    localStorage.removeItem(AUTH_KEYS.USER);
  },

  getExpiresAt: (): number | null => {
    const expiresAt = localStorage.getItem(AUTH_KEYS.EXPIRES_AT);
    const result = expiresAt ? parseInt(expiresAt, 10) : null;
    console.log('Storage: getExpiresAt called, result:', result ? new Date(result).toISOString() : null);
    return result;
  },

  setExpiresAt: (expiresAt: number): void => {
    console.log('Storage: setExpiresAt called, expiresAt:', new Date(expiresAt).toISOString());
    localStorage.setItem(AUTH_KEYS.EXPIRES_AT, expiresAt.toString());
    console.log('Storage: setExpiresAt completed');
  },

  removeExpiresAt: (): void => {
    console.log('Storage: removeExpiresAt called');
    localStorage.removeItem(AUTH_KEYS.EXPIRES_AT);
  },

  clear: (): void => {
    console.log('Storage: clear called');
    storage.removeToken();
    storage.removeUser();
    storage.removeExpiresAt();
  },
};

// Create context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Provider component
interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [state, dispatch] = useReducer(authReducer, initialState);
  const checkAuthTimeoutRef = React.useRef<NodeJS.Timeout | null>(null);

  // Save auth data to localStorage
  const saveAuthData = useCallback((token: string, user: User, expiresAt: number) => {
    console.log('AuthContext: saveAuthData called', {
      tokenLength: token.length,
      username: user.username,
      expiresAt: new Date(expiresAt * 1000).toISOString()
    });

    storage.setToken(token);
    storage.setUser(user);
    storage.setExpiresAt(expiresAt);

    console.log('AuthContext: saveAuthData completed, verifying storage:', {
      token: !!storage.getToken(),
      user: !!storage.getUser(),
      expiresAt: !!storage.getExpiresAt()
    });
  }, []);

  // Clear auth data from localStorage
  const clearAuthData = useCallback(() => {
    storage.clear();
  }, []);

  // Handle API errors
  const handleApiError = useCallback((error: any): string => {
    if (error?.response?.data) {
      const apiError = error.response.data as ApiError;
      return apiError.message || 'Authentication failed';
    }
    return error?.message || 'An unexpected error occurred';
  }, []);

  // Login function
  const login = useCallback(async (credentials: LoginRequest): Promise<void> => {
    dispatch({ type: 'AUTH_START' });

    try {
      const authResponse = await apiService.login(credentials);

      // The apiService.login already handles token storage and returns AuthResponse
      const { token, user, expires_at } = authResponse;

      console.log('AuthContext: Login successful', { user: user.username, expires_at });

      // Ensure API service has the token set
      apiService.setAuthToken(token);

      // Ensure all auth data is stored properly
      console.log('AuthContext: Saving auth data', {
        tokenLength: token.length,
        username: user.username,
        expiresAt: new Date(expires_at * 1000).toISOString()
      });

      saveAuthData(token, user, expires_at);

      // Verify data was stored
      console.log('AuthContext: Verifying stored data', {
        storedToken: !!storage.getToken(),
        storedUser: !!storage.getUser(),
        storedExpiresAt: !!storage.getExpiresAt()
      });

      // Set authenticated state immediately after successful login
      dispatch({
        type: 'AUTH_SUCCESS',
        payload: { user, token, expiresAt: expires_at }
      });

      console.log('AuthContext: Login state updated, isAuthenticated should be true now');
    } catch (error) {
      const errorMessage = handleApiError(error);
      dispatch({ type: 'AUTH_FAILURE', payload: errorMessage });
      throw new Error(errorMessage);
    }
  }, [saveAuthData, handleApiError]);

  // Register function
  const register = useCallback(async (userData: RegisterRequest): Promise<void> => {
    dispatch({ type: 'AUTH_START' });

    try {
      const authResponse = await apiService.register(userData);

      // The apiService.register already handles token storage and returns AuthResponse
      const { token, user, expires_at } = authResponse;

      // Ensure API service has the token set
      apiService.setAuthToken(token);

      // Ensure all auth data is stored properly
      saveAuthData(token, user, expires_at);

      // Set authenticated state immediately after successful registration
      dispatch({
        type: 'AUTH_SUCCESS',
        payload: { user, token, expiresAt: expires_at }
      });
    } catch (error) {
      const errorMessage = handleApiError(error);
      dispatch({ type: 'AUTH_FAILURE', payload: errorMessage });
      throw new Error(errorMessage);
    }
  }, [saveAuthData, handleApiError]);

  // Logout function
  const logout = useCallback(async (): Promise<void> => {
    try {
      // Call logout endpoint if token is available
      if (state.token) {
        await apiService.delete('/auth/logout').catch(() => {
          // Ignore logout API errors and proceed with local logout
        });
      }
    } catch (error) {
      // Ignore logout API errors
    } finally {
      dispatch({ type: 'AUTH_LOGOUT' });
      clearAuthData();
      // Ensure API service token is also cleared
      apiService.clearAuthToken();
    }
  }, [state.token, clearAuthData]);

  // Refresh token function
  const refresh = useCallback(async (): Promise<void> => {
    if (!state.token) {
      return;
    }

    try {
      const refreshResponse = await apiService.refreshToken();

      // The apiService.refreshToken already handles token storage
      const { token, expires_at } = refreshResponse;

      // Ensure API service has the new token set
      apiService.setAuthToken(token);

      dispatch({
        type: 'AUTH_REFRESH',
        payload: { token, expiresAt: expires_at }
      });

      storage.setToken(token);
      storage.setExpiresAt(expires_at);
    } catch (error) {
      // Token refresh failed, logout user
      const errorMessage = handleApiError(error);
      dispatch({ type: 'AUTH_FAILURE', payload: errorMessage });
      clearAuthData();
      throw new Error(errorMessage);
    }
  }, [state.token, clearAuthData, handleApiError]);

  // Check authentication status (with debounce)
  const checkAuth = useCallback(async (skipDebounce = false): Promise<void> => {
    // 防抖处理，避免短时间内多次调用
    if (!skipDebounce && checkAuthTimeoutRef.current) {
      return; // 如果已经在等待执行，直接返回
    }

    if (!skipDebounce) {
      checkAuthTimeoutRef.current = setTimeout(async () => {
        checkAuthTimeoutRef.current = null;
        await performAuthCheck();
      }, 100);
      return;
    }

    await performAuthCheck();
  }, []);

  // 实际的认证检查逻辑
  const performAuthCheck = useCallback(async (): Promise<void> => {
    console.log('AuthContext: performAuthCheck called');
    dispatch({ type: 'SET_LOADING', payload: true });

    try {
      const token = storage.getToken();
      const expiresAt = storage.getExpiresAt();
      const user = storage.getUser();

      console.log('AuthContext: Stored auth data', {
        hasToken: !!token,
        hasUser: !!user,
        hasExpiresAt: !!expiresAt,
        expiresAt: expiresAt ? new Date(expiresAt * 1000).toISOString() : null,
        currentTime: new Date().toISOString()
      });

      // Check if we have stored auth data
      if (!token || !user || !expiresAt) {
        console.log('AuthContext: No stored auth data found');
        dispatch({ type: 'AUTH_FAILURE', payload: 'No authentication data found' });
        return;
      }

      // Check if token is expired
      // expiresAt from API is in seconds, Date.now() returns milliseconds
      const currentTimestampSeconds = Math.floor(Date.now() / 1000);
      if (currentTimestampSeconds >= expiresAt) {
        clearAuthData();
        dispatch({ type: 'AUTH_FAILURE', payload: 'Token expired' });
        return;
      }

      // Validate token with server (but don't block UI for too long)
      try {
        const currentUser = await Promise.race([
          apiService.getCurrentUser(),
          new Promise((_, reject) =>
            setTimeout(() => reject(new Error('Auth check timeout')), 5000)
          )
        ]) as User;

        // Update user info with latest server data
        const updatedUser = currentUser || user;
        storage.setUser(updatedUser);

        // Set authenticated state after successful validation
        console.log('AuthContext: Server validation successful, setting authenticated state');
        dispatch({
          type: 'AUTH_SUCCESS',
          payload: {
            user: updatedUser,
            token: token,
            expiresAt: expiresAt
          }
        });

        // Try to refresh token if it's about to expire
        if (shouldRefreshToken(expiresAt)) {
          refresh().catch((error) => {
            // Refresh failed, clear auth data
            console.warn('Token refresh failed:', error);
            clearAuthData();
            dispatch({ type: 'AUTH_FAILURE', payload: 'Session expired' });
          });
        }

      } catch (error) {
        // Server validation failed or timeout
        if (error?.response?.status === 401) {
          console.warn('Server validation failed, token is invalid:', error);
          clearAuthData();
          dispatch({ type: 'AUTH_FAILURE', payload: 'Invalid session' });
        } else {
          // Network error, timeout, or server error - use local cached data if token is not expired
          console.warn('Server validation failed, using cached auth data:', error.message);

          // Use cached data if token is not expired
          if (Date.now() < expiresAt) {
            console.log('AuthContext: Using cached auth data, setting authenticated state');
            dispatch({
              type: 'AUTH_SUCCESS',
              payload: {
                user: user,
                token: token,
                expiresAt: expiresAt
              }
            });

            // Still try to refresh if needed
            if (shouldRefreshToken(expiresAt)) {
              refresh().catch((refreshError) => {
                console.warn('Token refresh failed:', refreshError);
              });
            }
          } else {
            clearAuthData();
            dispatch({ type: 'AUTH_FAILURE', payload: 'Token expired' });
          }
        }
      }

    } catch (error) {
      dispatch({ type: 'AUTH_FAILURE', payload: 'Authentication check failed' });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, [clearAuthData, refresh]);

  // Clear error function
  const clearError = useCallback((): void => {
    dispatch({ type: 'CLEAR_ERROR' });
  }, []);

  // Initialize auth on mount only
  useEffect(() => {
    console.log('AuthContext: Initializing auth on mount');

    // Ensure API service has proper token headers before checking auth
    const token = storage.getToken();
    console.log('AuthContext: Found token in storage:', !!token);

    if (token) {
      apiService.setAuthToken(token);
    }

    checkAuth(true); // Skip debounce for initial mount
  }, []); // Empty dependency array - only run once on mount

  // Setup token refresh timer
  useEffect(() => {
    if (!state.token || !state.isAuthenticated) {
      return;
    }

    const expiresAt = storage.getExpiresAt();
    if (!expiresAt) {
      return;
    }

    const refreshTime = expiresAt - TOKEN_CONFIG.REFRESH_THRESHOLD;
    const timeUntilRefresh = refreshTime - Date.now();

    if (timeUntilRefresh > 0) {
      const refreshTimer = setTimeout(() => {
        refresh().catch(() => {
          // Refresh failed, user will be logged out
        });
      }, timeUntilRefresh);

      return () => clearTimeout(refreshTimer);
    }
  }, [state.token, state.isAuthenticated, refresh]);

  const value: AuthContextType = {
    ...state,
    login,
    register,
    logout,
    refresh,
    checkAuth,
    clearError,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

// Hook to use auth context
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

// Helper function to check if token should be refreshed
const shouldRefreshToken = (expiresAt: number): boolean => {
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

// Permission checking hooks
export const usePermissions = () => {
  const { user } = useAuth();

  const hasPermission = useCallback((permission: string): boolean => {
    if (!user) return false;

    // Admin has all permissions
    if (user.role === 'admin') {
      return true;
    }

    // Implement specific permission logic here based on user role
    switch (user.role) {
      case 'user':
        return permission.startsWith('user:') || permission.startsWith('read:') || permission.startsWith('write:');
      default:
        return false;
    }
  }, [user]);

  const isAdmin = useCallback((): boolean => {
    return user?.role === 'admin';
  }, [user]);

  return { hasPermission, isAdmin };
};