import { Spin } from 'antd';
import React, { type ReactNode, useEffect } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useSystemStatus } from '../hooks/useApi';
import { AuthLevel, AuthStatus, getAuthStatus } from '../types/auth';

interface ProtectedRouteProps {
  children: ReactNode;
  authLevel?: AuthLevel;
  redirectTo?: string;
  fallback?: ReactNode;
  roles?: string[];
  permissions?: string[];
}

/**
 * ProtectedRoute component that restricts access based on authentication status.
 *
 * @param children - The protected content to render
 * @param authLevel - Required authentication level (default: AUTHENTICATED)
 * @param redirectTo - Where to redirect if access is denied (default: '/login')
 * @param fallback - Component to show while loading
 * @param roles - Required user roles (admin role overrides all restrictions)
 * @param permissions - Required permissions
 */
export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  authLevel = AuthLevel.AUTHENTICATED,
  redirectTo = '/login',
  fallback,
  roles = [],
  permissions = [],
}) => {
  const { user, isAuthenticated, isLoading, checkAuth } = useAuth();
  const { data: systemStatus } = useSystemStatus();
  const location = useLocation();

  // Default loading fallback
  const defaultFallback = (
    <div className="flex flex-col items-center justify-center min-h-screen gap-8">
      <Spin size="large" />
      <div className="text-gray-500 text-sm">验证身份中...</div>
    </div>
  );

  // Handle authentication check on mount
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      checkAuth().catch(() => {
        // Error handling is done in AuthContext
      });
    }
  }, [isLoading, isAuthenticated, checkAuth]);

  // Check system status first
  if (systemStatus) {
    const isSystemInitialized =
      systemStatus.initialized === true && systemStatus.needs_setup !== true;

    // If system is not initialized, redirect to setup page (unless already there)
    if (!isSystemInitialized && location.pathname !== '/setup') {
      return <Navigate to="/setup" replace />;
    }

    // If already on setup page and system is not initialized, just show nothing
    if (!isSystemInitialized && location.pathname === '/setup') {
      return null;
    }
  }

  // Show loading state
  if (isLoading) {
    return <>{fallback || defaultFallback}</>;
  }

  // Handle different auth levels
  switch (authLevel) {
    case AuthLevel.PUBLIC:
      // Always render public routes
      return <>{children}</>;

    case AuthLevel.OPTIONAL:
      // Render regardless of auth status
      return <>{children}</>;

    case AuthLevel.AUTHENTICATED:
      // Must be authenticated
      if (!isAuthenticated || !user) {
        return <Navigate to={redirectTo} state={{ from: location }} replace />;
      }
      break;

    case AuthLevel.ADMIN:
      // Must be admin
      if (!isAuthenticated || !user || user.role !== 'admin') {
        const adminRedirect = user ? '/dashboard' : redirectTo;
        return (
          <Navigate to={adminRedirect} state={{ from: location }} replace />
        );
      }
      break;
  }

  // Check role-based access if roles are specified
  if (roles.length > 0 && user) {
    const hasRequiredRole = roles.includes(user.role) || user.role === 'admin';
    if (!hasRequiredRole) {
      return <Navigate to="/dashboard" state={{ from: location }} replace />;
    }
  }

  // Check permission-based access if permissions are specified
  if (permissions.length > 0 && user) {
    // Admin has all permissions
    if (user.role !== 'admin') {
      // Implement permission checking logic here
      // For now, we'll assume users have basic permissions
      const hasPermission = true; // Replace with actual permission logic
      if (!hasPermission) {
        return <Navigate to="/dashboard" state={{ from: location }} replace />;
      }
    }
  }

  // User has access, render children
  return <>{children}</>;
};

// Higher-order component for protecting components
export const withAuth = <P extends object>(
  Component: React.ComponentType<P>,
  options?: Omit<ProtectedRouteProps, 'children'>,
) => {
  const WrappedComponent = (props: P) => (
    <ProtectedRoute {...options}>
      <Component {...props} />
    </ProtectedRoute>
  );

  WrappedComponent.displayName = `withAuth(${Component.displayName || Component.name})`;
  return WrappedComponent;
};

// Hook-based protection for function components
export const useAuthProtection = (options?: Partial<ProtectedRouteProps>) => {
  const { user, isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  const {
    authLevel = AuthLevel.AUTHENTICATED,
    redirectTo = '/login',
    roles = [],
    permissions = [],
  } = options || {};

  const isAuthorized = React.useMemo(() => {
    if (isLoading) return false;

    // Handle different auth levels
    switch (authLevel) {
      case AuthLevel.PUBLIC:
        return true;

      case AuthLevel.OPTIONAL:
        return true;

      case AuthLevel.AUTHENTICATED:
        return isAuthenticated && !!user;

      case AuthLevel.ADMIN:
        return isAuthenticated && !!user && user.role === 'admin';

      default:
        return false;
    }
  }, [authLevel, isAuthenticated, user, isLoading]);

  const hasRequiredRole = React.useMemo(() => {
    if (roles.length === 0 || !user) return true;
    return roles.includes(user.role) || user.role === 'admin';
  }, [roles, user]);

  const hasRequiredPermissions = React.useMemo(() => {
    if (permissions.length === 0 || !user) return true;
    if (user.role === 'admin') return true;
    // Implement permission checking logic here
    return true; // Placeholder
  }, [permissions, user]);

  const canAccess = isAuthorized && hasRequiredRole && hasRequiredPermissions;
  const redirectPath = user ? '/dashboard' : redirectTo;

  return {
    canAccess,
    isLoading,
    redirectPath,
    user,
    isAuthenticated,
    location,
  };
};

// Utility components for common protection scenarios

/**
 * AdminOnly - Renders children only for admin users
 */
export const AdminOnly: React.FC<{
  children: ReactNode;
  fallback?: ReactNode;
}> = ({ children, fallback }) => {
  return (
    <ProtectedRoute
      authLevel={AuthLevel.ADMIN}
      fallback={fallback || <div>无权限访问</div>}
    >
      {children}
    </ProtectedRoute>
  );
};

/**
 * AuthOptional - Renders children regardless of auth status
 */
export const AuthOptional: React.FC<{ children: ReactNode }> = ({
  children,
}) => {
  return (
    <ProtectedRoute authLevel={AuthLevel.OPTIONAL}>{children}</ProtectedRoute>
  );
};

/**
 * RequireRole - Renders children only for users with specific roles
 */
export const RequireRole: React.FC<{
  children: ReactNode;
  roles: string[];
  fallback?: ReactNode;
}> = ({ children, roles, fallback }) => {
  return (
    <ProtectedRoute
      authLevel={AuthLevel.AUTHENTICATED}
      roles={roles}
      fallback={fallback || <div>权限不足</div>}
    >
      {children}
    </ProtectedRoute>
  );
};

export default ProtectedRoute;
