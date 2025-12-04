import { StartupWorkflow, StartupExecution, StartupWorkflowNode, StartupWorkflowEdge, StartupWebSocketManager } from '../../services/startupWebSocket';
import { SimpleNode, SimpleConnection } from './SimpleReteEditor';
import { BaseNode, NodeData } from './nodes';

/**
 * 将后端启动流程数据转换为编辑器可用格式
 */
export class StartupFlowAdapter {
  private wsManager: StartupWebSocketManager;

  constructor() {
    this.wsManager = (window as any).startupWebSocketManager || new (require('../../services/startupWebSocket')).StartupWebSocketManager();
  }

  /**
   * 将后端工作流转换为编辑器节点
   */
  convertWorkflowToEditorNodes(workflow: StartupWorkflow): SimpleNode[] {
    return workflow.nodes.map(node => this.convertStartupNodeToEditorNode(node));
  }

  /**
   * 将单个后端节点转换为编辑器节点
   */
  private convertStartupNodeToEditorNode(startupNode: StartupWorkflowNode): SimpleNode {
    // 映射后端节点类型到编辑器类型
    const typeMap: Record<string, NodeData['type']> = {
      'storage': 'database',
      'config': 'config',
      'service': 'api',
      'auth': 'api',
      'plugin': 'cloud',
      'database': 'database',
      'api': 'api',
      'ai': 'ai',
      'cloud': 'cloud'
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

    return {
      id: startupNode.id,
      data: {
        label: startupNode.name,
        type: nodeType,
        status,
        description: startupNode.description,
        metrics
      },
      x: startupNode.position.x,
      y: startupNode.position.y
    } as SimpleNode;
  }

  /**
   * 将后端工作流边转换为编辑器连接
   */
  convertWorkflowToEditorConnections(workflow: StartupWorkflow): SimpleConnection[] {
    return workflow.edges.map(edge => ({
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
  updateNodesByExecution(nodes: SimpleNode[], execution: StartupExecution): SimpleNode[] {
    return nodes.map(node => {
      const executionNode = execution.nodes.find(n => n.id === node.id);
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
   * 根据执行状态更新连接动画
   */
  updateConnectionsByExecution(connections: SimpleConnection[], execution: StartupExecution): SimpleConnection[] {
    const currentNodes = new Set(execution.current_nodes);

    return connections.map(conn => {
      const isSourceActive = currentNodes.has(conn.from);
      const isTargetActive = currentNodes.has(conn.to);

      return {
        ...conn,
        // 可以添加动画状态字段
        animated: isSourceActive || isTargetActive
      };
    });
  }

  /**
   * 连接WebSocket
   */
  async connect(): Promise<void> {
    try {
      await this.wsManager.connect();
      console.log('✅ 启动流程WebSocket连接成功');
    } catch (error) {
      console.error('❌ 启动流程WebSocket连接失败:', error);
      throw error;
    }
  }

  /**
   * 断开WebSocket连接
   */
  disconnect(): void {
    this.wsManager.disconnect();
  }

  /**
   * 获取工作流数据
   */
  async getWorkflow(workflowId: string): Promise<StartupWorkflow | null> {
    try {
      // 这里需要实现获取工作流数据的逻辑
      // 暂时返回模拟数据
      return this.getDefaultWorkflow(workflowId);
    } catch (error) {
      console.error('获取工作流数据失败:', error);
      return null;
    }
  }

  /**
   * 执行工作流
   */
  async executeWorkflow(workflowId: string, inputs?: Record<string, any>): Promise<string> {
    try {
      const executionId = await this.wsManager.executeWorkflow(workflowId, inputs);
      console.log('✅ 工作流执行开始:', executionId);
      return executionId;
    } catch (error) {
      console.error('❌ 工作流执行失败:', error);
      throw error;
    }
  }

  /**
   * 取消执行
   */
  async cancelExecution(executionId: string): Promise<void> {
    try {
      this.wsManager.cancelExecution(executionId);
      console.log('✅ 工作流取消成功:', executionId);
    } catch (error) {
      console.error('❌ 工作流取消失败:', error);
      throw error;
    }
  }

  /**
   * 暂停执行
   */
  async pauseExecution(executionId: string): Promise<void> {
    try {
      this.wsManager.pauseExecution(executionId);
      console.log('✅ 工作流暂停成功:', executionId);
    } catch (error) {
      console.error('❌ 工作流暂停失败:', error);
      throw error;
    }
  }

  /**
   * 恢复执行
   */
  async resumeExecution(executionId: string): Promise<void> {
    try {
      this.wsManager.resumeExecution(executionId);
      console.log('✅ 工作流恢复成功:', executionId);
    } catch (error) {
      console.error('❌ 工作流恢复失败:', error);
      throw error;
    }
  }

  /**
   * 订阅执行事件
   */
  onExecutionStart(handler: (data: any) => void): () => void {
    this.wsManager.on('execution_start', (message) => handler(message.data));
    return () => this.wsManager.off('execution_start', (message) => handler(message.data));
  }

  /**
   * 订阅执行进度
   */
  onExecutionProgress(handler: (data: any) => void): () => void {
    this.wsManager.on('execution_progress', (message) => handler(message.data));
    return () => this.wsManager.off('execution_progress', (message) => handler(message.data));
  }

  /**
   * 订阅执行结束
   */
  onExecutionEnd(handler: (data: any) => void): () => void {
    this.wsManager.on('execution_end', (message) => handler(message.data));
    return () => this.wsManager.off('execution_end', (message) => handler(message.data));
  }

  /**
   * 获取默认的xiaozhi-flow-default-startup工作流
   */
  private getDefaultWorkflow(workflowId: string): StartupWorkflow {
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
   * 获取工作流执行统计信息
   */
  generateExecutionStats(execution: StartupExecution): Record<string, any> {
    const statusCount = execution.nodes.reduce((acc, node) => {
      acc[node.status] = (acc[node.status] || 0) + 1;
      return acc;
    }, {} as Record<string, number>);

    return {
      total: execution.total_nodes,
      completed: execution.completed_nodes,
      failed: execution.failed_nodes,
      running: statusCount.running || 0,
      pending: statusCount.pending || 0,
      paused: statusCount.paused || 0,
      progress: Math.round(execution.progress * 100),
      duration: Math.round(execution.duration / 1000),
      startTime: new Date(execution.start_time).toLocaleString(),
      endTime: execution.end_time ? new Date(execution.end_time).toLocaleString() : null,
      status: execution.status,
      criticalFailed: execution.nodes.filter(n => n.critical && n.status === 'failed').length,
      optionalSkipped: execution.nodes.filter(n => n.optional && n.status === 'pending' && execution.status === 'completed').length
    };
  }
}