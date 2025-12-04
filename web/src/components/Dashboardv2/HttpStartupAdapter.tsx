import axios from 'axios';
import { SimpleNode, SimpleConnection } from './SimpleReteEditor';
import { BaseNode, NodeData } from './nodes';

/**
 * 基于 HTTP API 的启动流程适配器
 * 不依赖 WebSocket，通过 REST API 与后端交互
 */
export class HttpStartupAdapter {
  private baseUrl: string;

  constructor(baseUrl?: string) {
    // 如果明确提供了 baseUrl，直接使用
    if (baseUrl) {
      this.baseUrl = baseUrl.endsWith('/') ? baseUrl.slice(0, -1) : baseUrl;
      return;
    }

    // 根据当前环境构建API URL
    const protocol = window.location.protocol;
    const host = window.location.hostname;

    // 获取当前端口，如果是前端开发服务器端口，则需要映射到后端端口
    const currentPort = window.location.port;
    let apiPort = '8080'; // 默认后端端口

    // 如果当前端口是前端开发端口（如3000, 5173等），使用默认后端端口
    if (currentPort && ['3000', '5173', '8080', '8081'].includes(currentPort)) {
      apiPort = currentPort === '8080' ? '8080' : '8080';
    } else if (currentPort) {
      // 其他情况，假设前后端同端口
      apiPort = currentPort;
    }

    this.baseUrl = `${protocol}//${host}:${apiPort}/api/startup`;
    console.log('初始化 HttpStartupAdapter，baseUrl:', this.baseUrl);
  }

  /**
   * 获取工作流定义
   */
  async getWorkflow(workflowId: string): Promise<any> {
    console.log(`正在获取工作流 ${workflowId}，API地址: ${this.baseUrl}/workflows/${workflowId}`);

    try {
      const response = await axios.get(`${this.baseUrl}/workflows/${workflowId}`, {
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        },
        timeout: 5000 // 5秒超时
      });

      console.log('获取到工作流数据:', response.data);

      // 检查返回的数据结构
      if (response.data && typeof response.data === 'object' && !this.isHtmlContent(response.data)) {
        // 处理后端返回的标准API格式：{success: true, data: {...}, message: '获取成功', code: 200}
        if (response.data.success && response.data.data) {
          console.log('提取工作流数据:', response.data.data);
          return response.data.data;
        }
        // 如果直接返回工作流数据，直接使用
        return response.data;
      } else {
        console.error('API返回的数据格式不正确，可能是HTML页面:', response.data);
        return this.getDefaultWorkflow(workflowId);
      }
    } catch (error: any) {
      console.error(`获取工作流失败 (${this.baseUrl}/workflows/${workflowId}):`, error.message);

      // 如果是网络错误或404错误，尝试使用默认工作流
      if (error.code === 'ECONNREFUSED' || error.response?.status === 404) {
        console.log('API服务不可用，使用默认工作流');
        return this.getDefaultWorkflow(workflowId);
      }

      // 其他错误也使用默认工作流
      return this.getDefaultWorkflow(workflowId);
    }
  }

  /**
   * 检查返回的内容是否为HTML
   */
  private isHtmlContent(data: any): boolean {
    if (typeof data === 'string') {
      return data.includes('<!doctype html>') || data.includes('<html>');
    }
    return false;
  }

  /**
   * 执行工作流
   */
  async executeWorkflow(workflowId: string, inputs?: Record<string, any>): Promise<string> {
    try {
      const response = await axios.post(`${this.baseUrl}/workflows/${workflowId}/execute`, {
        inputs: inputs || {}
      });
      return response.data.execution_id;
    } catch (error) {
      console.error('执行工作流失败:', error);
      // 返回模拟执行ID
      return `exec-${Date.now()}`;
    }
  }

  /**
   * 获取执行状态
   */
  async getExecutionStatus(executionId: string): Promise<any> {
    try {
      const response = await axios.get(`${this.baseUrl}/executions/${executionId}/status`);
      return response.data;
    } catch (error) {
      console.error('获取执行状态失败:', error);
      return null;
    }
  }

  /**
   * 取消执行
   */
  async cancelExecution(executionId: string): Promise<void> {
    try {
      await axios.post(`${this.baseUrl}/executions/${executionId}/cancel`);
    } catch (error) {
      console.error('取消执行失败:', error);
      throw error;
    }
  }

  /**
   * 暂停执行
   */
  async pauseExecution(executionId: string): Promise<void> {
    try {
      await axios.post(`${this.baseUrl}/executions/${executionId}/pause`);
    } catch (error) {
      console.error('暂停执行失败:', error);
      throw error;
    }
  }

  /**
   * 恢复执行
   */
  async resumeExecution(executionId: string): Promise<void> {
    try {
      await axios.post(`${this.baseUrl}/executions/${executionId}/resume`);
    } catch (error) {
      console.error('恢复执行失败:', error);
      throw error;
    }
  }

  /**
   * 轮询执行状态
   */
  async pollExecutionStatus(
    executionId: string,
    onProgress: (status: any) => void,
    interval: number = 1000
  ): Promise<void> {
    const poll = async () => {
      try {
        const status = await this.getExecutionStatus(executionId);
        if (status) {
          onProgress(status);

          // 如果执行还在进行中，继续轮询
          if (status.status === 'running' || status.status === 'paused') {
            setTimeout(poll, interval);
          }
        }
      } catch (error) {
        console.error('轮询状态失败:', error);
      }
    };

    poll();
  }

  /**
   * 将后端工作流转换为编辑器节点
   */
  convertWorkflowToEditorNodes(workflow: any): SimpleNode[] {
    // 添加防护性检查
    if (!workflow) {
      console.error('工作流数据为空');
      return [];
    }

    // 处理可能的包装数据结构
    let workflowData = workflow;
    if (workflow.success && workflow.data) {
      console.log('检测到包装数据结构，提取实际工作流数据');
      workflowData = workflow.data;
    }

    if (!workflowData.nodes || !Array.isArray(workflowData.nodes)) {
      console.error('工作流节点数据无效:', workflowData);
      return [];
    }

    console.log('转换节点数据，节点数量:', workflowData.nodes.length);
    return workflowData.nodes.map((node: any) => this.convertStartupNodeToEditorNode(node));
  }

  /**
   * 将单个后端节点转换为编辑器节点
   */
  private convertStartupNodeToEditorNode(startupNode: any): SimpleNode {
    // 映射后端节点类型到编辑器类型
    const typeMap: Record<string, NodeData['type']> = {
      'storage': 'database',
      'config': 'config',
      'service': 'api',
      'auth': 'api',        // 认证相关节点映射为API节点
      'plugin': 'cloud',     // 插件节点映射为云服务节点
      'database': 'database',
      'api': 'api',
      'ai': 'ai',
      'cloud': 'cloud',
      'observability': 'config',  // 可观测性节点映射为配置节点
      'mcp': 'ai',              // MCP管理器映射为AI节点
      'transport': 'api',       // 传输服务映射为API节点
      'system': 'config'        // 系统服务映射为配置节点
    };

    const nodeType = typeMap[startupNode.type] || 'api';

    // 映射状态
    const statusMap: Record<string, NodeData['status']> = {
      'pending': 'stopped',
      'running': 'running',
      'completed': 'running',
      'failed': 'warning',
      'paused': 'warning',
      'cancelled': 'stopped'
    };

    const status = statusMap[startupNode.status] || 'stopped';

    // 构建指标
    const metrics: Record<string, string | number> = {};

    if (startupNode.critical) {
      metrics['关键节点'] = '是';
    }
    if (startupNode.optional) {
      metrics['可选节点'] = '是';
    }
    if (startupNode.timeout) {
      metrics['超时时间'] = `${Math.round(startupNode.timeout / 1000)}s`;
    }
    if (startupNode.progress !== undefined) {
      metrics['进度'] = `${Math.round(startupNode.progress * 100)}%`;
    }
    if (startupNode.duration) {
      metrics['执行时间'] = `${Math.round(startupNode.duration / 1000)}s`;
    }
    if (startupNode.error) {
      metrics['错误'] = startupNode.error;
    }

    // 优化坐标映射，将大的 x 坐标压缩到前端画布适合的范围
    let x = Math.random() * 600 + 100; // 默认随机位置
    let y = Math.random() * 300 + 100;

    if (startupNode.position) {
      // 如果后端坐标很大，进行压缩映射
      if (startupNode.position.x > 1000) {
        // 将大坐标映射到 100-1200 范围内
        x = 100 + (startupNode.position.x / 2500) * 1100;
      } else {
        // 小坐标直接使用，限制在合理范围内
        x = Math.max(100, Math.min(1200, startupNode.position.x));
      }

      // y 坐标处理，提供更好的垂直分布
      if (startupNode.position.y > 0) {
        y = Math.max(100, Math.min(500, startupNode.position.y));
      }
    }

    return {
      id: startupNode.id,
      data: {
        label: startupNode.name,
        type: nodeType,
        status,
        description: startupNode.description,
        metrics
      },
      x,
      y
    } as SimpleNode;
  }

  /**
   * 将后端工作流边转换为编辑器连接
   */
  convertWorkflowToEditorConnections(workflow: any): SimpleConnection[] {
    // 添加防护性检查
    if (!workflow) {
      console.error('工作流数据为空');
      return [];
    }

    // 处理可能的包装数据结构
    let workflowData = workflow;
    if (workflow.success && workflow.data) {
      workflowData = workflow.data;
    }

    if (!workflowData.edges || !Array.isArray(workflowData.edges)) {
      console.error('工作流边数据无效:', workflowData);
      return [];
    }

    console.log('转换连接数据，连接数量:', workflowData.edges.length);
    return workflowData.edges.map((edge: any) => ({
      id: edge.id,
      from: edge.from,
      to: edge.to,
      fromOutput: 'data',
      toInput: 'input'
    }));
  }

  /**
   * 根据执行状态更新节点
   */
  updateNodesByExecution(nodes: SimpleNode[], execution: any): SimpleNode[] {
    if (!execution || !execution.nodes) return nodes;

    return nodes.map(node => {
      const executionNode = execution.nodes.find((n: any) => n.id === node.id);
      if (!executionNode) return node;

      // 映射状态
      const statusMap: Record<string, NodeData['status']> = {
        'pending': 'stopped',
        'running': 'running',
        'completed': 'running',
        'failed': 'warning',
        'paused': 'warning',
        'cancelled': 'stopped'
      };

      const status = statusMap[executionNode.status] || 'stopped';

      // 更新指标
      const metrics: Record<string, string | number> = { ...node.data.metrics };

      if (executionNode.progress !== undefined) {
        metrics['进度'] = `${Math.round(executionNode.progress * 100)}%`;
      }

      if (executionNode.duration) {
        metrics['执行时间'] = `${Math.round(executionNode.duration / 1000)}s`;
      }

      if (executionNode.error) {
        metrics['错误'] = executionNode.error;
      }

      return {
        ...node,
        data: {
          ...node.data,
          status,
          metrics
        }
      };
    });
  }

  /**
   * 生成执行统计信息
   */
  generateExecutionStats(execution: any): Record<string, any> {
    if (!execution) return null;

    const statusCount = execution.nodes?.reduce((acc: Record<string, number>, node: any) => {
      acc[node.status] = (acc[node.status] || 0) + 1;
      return acc;
    }, {}) || {};

    return {
      total: execution.total_nodes || execution.nodes?.length || 0,
      completed: execution.completed_nodes || statusCount.completed || 0,
      failed: execution.failed_nodes || statusCount.failed || 0,
      running: statusCount.running || 0,
      pending: statusCount.pending || 0,
      paused: statusCount.paused || 0,
      progress: Math.round((execution.progress || 0) * 100),
      duration: Math.round((execution.duration || 0) / 1000),
      startTime: execution.start_time ? new Date(execution.start_time).toLocaleString() : new Date().toLocaleString(),
      endTime: execution.end_time ? new Date(execution.end_time).toLocaleString() : null,
      status: execution.status || 'pending',
      criticalFailed: execution.nodes?.filter((n: any) => n.critical && n.status === 'failed').length || 0,
      optionalSkipped: execution.nodes?.filter((n: any) => n.optional && n.status === 'pending' && execution.status === 'completed').length || 0
    };
  }

  /**
   * 获取默认的xiaozhi-flow-default-startup工作流
   */
  private getDefaultWorkflow(workflowId: string): any {
    return {
      id: workflowId,
      name: 'XiaoZhi Flow 默认启动流程',
      description: '系统启动和初始化的标准流程',
      version: '1.0.0',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      tags: ['startup', 'system', 'default'],
      nodes: [
        {
          id: 'storage-init',
          name: '存储初始化',
          type: 'storage',
          description: '初始化数据库连接和存储系统',
          status: 'pending',
          timeout: 30000,
          critical: true,
          optional: false,
          position: { x: 100, y: 100 },
          config: {
            database_url: 'mysql://localhost:3306/xiaozhi',
            redis_url: 'redis://localhost:6379'
          },
          metadata: {
            version: '1.0.0',
            author: 'system'
          },
          depends_on: []
        },
        {
          id: 'config-load',
          name: '配置加载',
          type: 'config',
          description: '加载系统配置和环境变量',
          status: 'pending',
          timeout: 10000,
          critical: true,
          optional: false,
          position: { x: 400, y: 100 },
          config: {
            config_path: '/etc/xiaozhi/config.yaml',
            env_path: '/etc/xiaozhi/.env'
          },
          metadata: {
            version: '1.0.0',
            author: 'system'
          },
          depends_on: ['storage-init']
        },
        {
          id: 'service-start',
          name: '服务启动',
          type: 'service',
          description: '启动核心服务组件',
          status: 'pending',
          timeout: 60000,
          critical: true,
          optional: false,
          position: { x: 700, y: 100 },
          config: {
            services: ['api', 'scheduler', 'processor'],
            port: 8080
          },
          metadata: {
            version: '1.0.0',
            author: 'system'
          },
          depends_on: ['config-load']
        },
        {
          id: 'auth-setup',
          name: '认证设置',
          type: 'auth',
          description: '配置认证和授权系统',
          status: 'pending',
          timeout: 20000,
          critical: true,
          optional: false,
          position: { x: 100, y: 300 },
          config: {
            auth_type: 'jwt',
            secret_key: 'your-secret-key'
          },
          metadata: {
            version: '1.0.0',
            author: 'system'
          },
          depends_on: ['storage-init']
        },
        {
          id: 'plugin-load',
          name: '插件加载',
          type: 'plugin',
          description: '加载和初始化插件系统',
          status: 'pending',
          timeout: 30000,
          critical: false,
          optional: true,
          position: { x: 400, y: 300 },
          config: {
            plugin_dir: '/plugins',
            auto_load: true
          },
          metadata: {
            version: '1.0.0',
            author: 'system'
          },
          depends_on: ['config-load', 'auth-setup']
        }
      ],
      edges: [
        {
          id: 'edge-1',
          from: 'storage-init',
          to: 'config-load',
          label: '配置依赖'
        },
        {
          id: 'edge-2',
          from: 'config-load',
          to: 'service-start',
          label: '启动服务'
        },
        {
          id: 'edge-3',
          from: 'storage-init',
          to: 'auth-setup',
          label: '认证设置'
        },
        {
          id: 'edge-4',
          from: 'config-load',
          to: 'plugin-load',
          label: '加载插件'
        },
        {
          id: 'edge-5',
          from: 'auth-setup',
          to: 'plugin-load',
          label: '插件认证'
        }
      ],
      config: {
        timeout: 120000,
        max_retries: 3,
        parallel_limit: 5,
        enable_log: true,
        environment: {
          NODE_ENV: 'production',
          LOG_LEVEL: 'info'
        },
        variables: {
          APP_NAME: 'XiaoZhi Flow',
          VERSION: '1.0.0'
        },
        on_failure: 'stop_all'
      }
    };
  }

  /**
   * 模拟执行数据
   */
  createMockExecution(workflowId: string, nodes: SimpleNode[]): any {
    return {
      id: `exec-${Date.now()}`,
      workflow_id: workflowId,
      workflow_name: 'XiaoZhi Flow 默认启动流程',
      status: 'running',
      start_time: new Date().toISOString(),
      duration: 0,
      progress: 0,
      total_nodes: nodes.length,
      completed_nodes: 0,
      failed_nodes: 0,
      current_nodes: [],
      nodes: nodes.map(node => ({
        id: node.id,
        name: node.data.label,
        type: node.data.type,
        status: 'pending',
        start_time: null,
        end_time: null,
        duration: 0,
        progress: 0,
        error: null
      }))
    };
  }
}