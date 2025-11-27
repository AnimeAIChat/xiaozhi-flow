import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { message } from 'antd';
import {
  apiService,
  type ServerConfig,
  type ConnectionTestResult,
  type InitConfig,
  type InitResult,
  type ProviderType,
  type ProviderConfig,
  type ProviderTestResult,
  type SystemConfig,
} from '../services/api';

// 查询键
export const queryKeys = {
  systemStatus: ['system', 'status'],
  systemLogs: ['system', 'logs'],
  providers: (type?: ProviderType) => ['providers', type].filter(Boolean),
  systemConfig: ['system', 'config'],
} as const;

/**
 * 测试服务器连接的Hook
 */
export const useTestConnection = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (config: ServerConfig): Promise<ConnectionTestResult> =>
      apiService.testConnection(config),
    onSuccess: (result) => {
      if (result.success) {
        message.success(`连接成功！延迟: ${result.latency}ms`);
      } else {
        message.error(`连接失败: ${result.message}`);
      }
    },
    onError: (error) => {
      message.error(`连接测试出错: ${error.message}`);
    },
  });
};

/**
 * 初始化项目的Hook
 */
export const useInitializeProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (config: InitConfig): Promise<InitResult> =>
      apiService.initializeProject(config),
    onSuccess: (result) => {
      if (result.success) {
        message.success('项目初始化成功！');
        // 刷新相关查询
        queryClient.invalidateQueries(queryKeys.systemStatus);
        queryClient.invalidateQueries(queryKeys.systemConfig);
      } else {
        message.error(`项目初始化失败: ${result.message}`);
      }
    },
    onError: (error) => {
      message.error(`项目初始化出错: ${error.message}`);
    },
  });
};

/**
 * 获取系统状态的Hook
 */
export const useSystemStatus = () => {
  return useQuery({
    queryKey: queryKeys.systemStatus,
    queryFn: () => apiService.getSystemStatus(),
    refetchInterval: 30000, // 每30秒刷新一次
    retry: 3,
    retryDelay: 1000,
    staleTime: 10000, // 10秒内的数据被认为是新的
  });
};

/**
 * 获取系统日志的Hook
 */
export const useSystemLogs = (level?: string) => {
  return useQuery({
    queryKey: [...queryKeys.systemLogs, level],
    queryFn: () => apiService.getSystemLogs(level),
    refetchInterval: 60000, // 每分钟刷新一次
    retry: 2,
    retryDelay: 2000,
  });
};

/**
 * 获取提供商列表的Hook
 */
export const useProviders = (type?: ProviderType) => {
  return useQuery({
    queryKey: queryKeys.providers(type),
    queryFn: () => apiService.getProviders(type),
    staleTime: 300000, // 5分钟内的数据被认为是新的
    retry: 2,
    retryDelay: 1000,
  });
};

/**
 * 测试提供商连接的Hook
 */
export const useTestProvider = () => {
  return useMutation({
    mutationFn: ({ type, config }: { type: ProviderType; config: ProviderConfig }): Promise<ProviderTestResult> =>
      apiService.testProvider(type, config),
    onSuccess: (result, variables) => {
      if (result.success) {
        message.success(`${variables.config.name} 连接测试成功！延迟: ${result.latency}ms`);
      } else {
        message.error(`${variables.config.name} 连接失败: ${result.message}`);
      }
    },
    onError: (error) => {
      message.error(`提供商测试出错: ${error.message}`);
    },
  });
};

/**
 * 更新提供商配置的Hook
 */
export const useUpdateProvider = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ type, config }: { type: ProviderType; config: ProviderConfig }) =>
      apiService.updateProvider(type, config),
    onSuccess: (_, variables) => {
      message.success(`${variables.config.name} 配置更新成功！`);
      // 刷新提供商列表
      queryClient.invalidateQueries(queryKeys.providers());
      queryClient.invalidateQueries(queryKeys.providers(variables.type));
    },
    onError: (error, variables) => {
      message.error(`${variables.config.name} 配置更新失败: ${error.message}`);
    },
  });
};

/**
 * 获取系统配置的Hook
 */
export const useSystemConfig = () => {
  return useQuery({
    queryKey: queryKeys.systemConfig,
    queryFn: () => apiService.getSystemConfig(),
    staleTime: 300000, // 5分钟内的数据被认为是新的
    retry: 3,
    retryDelay: 1000,
  });
};

/**
 * 更新系统配置的Hook
 */
export const useUpdateSystemConfig = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (config: SystemConfig) => apiService.updateSystemConfig(config),
    onSuccess: () => {
      message.success('系统配置更新成功！');
      // 刷新系统配置
      queryClient.invalidateQueries(queryKeys.systemConfig);
    },
    onError: (error) => {
      message.error(`系统配置更新失败: ${error.message}`);
    },
  });
};

/**
 * 验证配置的Hook
 */
export const useValidateConfig = () => {
  return useMutation({
    mutationFn: (config: any) => apiService.validateConfig(config),
    onSuccess: (result) => {
      if (result.valid) {
        message.success('配置验证通过！');
      } else {
        message.error(`配置验证失败: ${result.errors?.join(', ') || '未知错误'}`);
      }
    },
    onError: (error) => {
      message.error(`配置验证出错: ${error.message}`);
    },
  });
};

/**
 * API状态监控Hook
 */
export const useApiHealth = () => {
  const { data: systemStatus, isLoading, error } = useSystemStatus();

  return {
    isHealthy: !error && systemStatus,
    isLoading,
    lastCheck: systemStatus?.timestamp,
    uptime: systemStatus?.uptime,
    version: systemStatus?.version,
  };
};

export default {
  useTestConnection,
  useInitializeProject,
  useSystemStatus,
  useSystemLogs,
  useProviders,
  useTestProvider,
  useUpdateProvider,
  useSystemConfig,
  useUpdateSystemConfig,
  useValidateConfig,
  useApiHealth,
  queryKeys,
};