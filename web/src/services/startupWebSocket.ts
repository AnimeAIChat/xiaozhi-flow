/**
 * 启动流程WebSocket管理器
 * 负责与后端启动流程WebSocket服务建立连接并处理实时数据
 */

export interface StartupWorkflowNode {
  id: string;
  name: string;
  type: string;
  description: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'paused';
  timeout: number;
  critical: boolean;
  optional: boolean;
  position: { x: number; y: number };
  config: Record<string, any>;
  metadata: Record<string, string>;
  depends_on: string[];
  start_time?: string;
  end_time?: string;
  duration?: number;
  error?: string;
  progress?: number;
  metrics?: Record<string, any>;
}

export interface StartupWorkflowEdge {
  id: string;
  from: string;
  to: string;
  label?: string;
}

export interface StartupWorkflow {
  id: string;
  name: string;
  description: string;
  version: string;
  created_at: string;
  updated_at: string;
  tags: string[];
  nodes: StartupWorkflowNode[];
  edges: StartupWorkflowEdge[];
  config: {
    timeout: number;
    max_retries: number;
    parallel_limit: number;
    enable_log: boolean;
    environment: Record<string, any>;
    variables: Record<string, any>;
    on_failure: string;
  };
}

export interface StartupExecution {
  id: string;
  workflow_id: string;
  workflow_name: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'paused' | 'cancelled';
  start_time: string;
  end_time?: string;
  duration: number;
  progress: number;
  total_nodes: number;
  completed_nodes: number;
  failed_nodes: number;
  current_nodes: string[];
  error?: string;
  context: Record<string, any>;
  nodes: StartupWorkflowNode[];
}

export interface WebSocketMessage {
  type: string;
  event_id: string;
  timestamp: string;
  data: Record<string, any>;
}

export type WebSocketMessageHandler = (message: WebSocketMessage) => void;

export class StartupWebSocketManager {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private isConnecting = false;
  private isManualClose = false;
  private messageHandlers: Map<string, WebSocketMessageHandler[]> = new Map();
  private connectionId: string | null = null;
  private pingInterval: NodeJS.Timeout | null = null;
  private subscriptions: Set<string> = new Set();

  constructor(baseUrl?: string) {
    // 根据当前环境构建WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.hostname;
    const port = baseUrl ?
      (baseUrl.includes(':') ? baseUrl.split(':')[1] : '8080') :
      (window.location.port || '8080');

    this.url = `${protocol}//${host}:${port}/api/startup/ws`;
  }

  /**
   * 建立WebSocket连接
   */
  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      if (this.isConnecting) {
        reject(new Error('Connection already in progress'));
        return;
      }

      this.isConnecting = true;
      this.isManualClose = false;

      try {
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          console.log('📡 启动流程WebSocket连接已建立');
          this.isConnecting = false;
          this.reconnectAttempts = 0;
          this.startPing();
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('解析WebSocket消息失败:', error);
          }
        };

        this.ws.onclose = (event) => {
          console.log('📡 启动流程WebSocket连接已关闭', event.code, event.reason);
          this.isConnecting = false;
          this.stopPing();
          this.connectionId = null;

          if (!this.isManualClose && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect();
          }
        };

        this.ws.onerror = (error) => {
          console.error('📡 启动流程WebSocket连接错误:', error);
          this.isConnecting = false;
          reject(error);
        };

      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  /**
   * 断开WebSocket连接
   */
  disconnect(): void {
    this.isManualClose = true;
    this.stopPing();

    if (this.ws) {
      this.ws.close(1000, 'Manual disconnect');
      this.ws = null;
    }

    this.connectionId = null;
    this.subscriptions.clear();
    this.messageHandlers.clear();
  }

  /**
   * 检查连接状态
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * 发送消息
   */
  send(message: Record<string, any>): void {
    if (!this.isConnected()) {
      throw new Error('WebSocket未连接');
    }

    this.ws!.send(JSON.stringify(message));
  }

  /**
   * 订阅执行事件
   */
  subscribe(executionId: string): void {
    if (!this.isConnected()) {
      throw new Error('WebSocket未连接');
    }

    this.subscriptions.add(executionId);
    this.send({
      type: 'subscribe',
      execution_id: executionId
    });
  }

  /**
   * 取消订阅执行事件
   */
  unsubscribe(executionId: string): void {
    if (!this.isConnected()) {
      return;
    }

    this.subscriptions.delete(executionId);
    this.send({
      type: 'unsubscribe',
      execution_id: executionId
    });
  }

  /**
   * 执行启动工作流
   */
  async executeWorkflow(workflowId: string, inputs?: Record<string, any>): Promise<string> {
    return new Promise((resolve, reject) => {
      if (!this.isConnected()) {
        reject(new Error('WebSocket未连接'));
        return;
      }

      const messageId = `execute_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

      // 设置临时处理器来接收响应
      const handleResponse = (message: WebSocketMessage) => {
        if (message.type === 'execution_start' && message.data.execution_id) {
          this.off('execution_start', handleResponse);
          resolve(message.data.execution_id);
        } else if (message.type === 'error' && message.data.error.includes('execute')) {
          this.off('execution_start', handleResponse);
          this.off('error', handleResponse);
          reject(new Error(message.data.error));
        }
      };

      this.on('execution_start', handleResponse);
      this.on('error', handleResponse);

      // 发送执行请求
      this.send({
        type: 'execute_workflow',
        workflow_id: workflowId,
        inputs: inputs || {}
      });

      // 设置超时
      setTimeout(() => {
        this.off('execution_start', handleResponse);
        this.off('error', handleResponse);
        reject(new Error('执行请求超时'));
      }, 10000);
    });
  }

  /**
   * 获取执行状态
   */
  getExecutionStatus(executionId: string): void {
    if (!this.isConnected()) {
      throw new Error('WebSocket未连接');
    }

    this.send({
      type: 'get_execution_status',
      execution_id: executionId
    });
  }

  /**
   * 取消执行
   */
  cancelExecution(executionId: string): void {
    if (!this.isConnected()) {
      throw new Error('WebSocket未连接');
    }

    this.send({
      type: 'cancel_execution',
      execution_id: executionId
    });
  }

  /**
   * 暂停执行
   */
  pauseExecution(executionId: string): void {
    if (!this.isConnected()) {
      throw new Error('WebSocket未连接');
    }

    this.send({
      type: 'pause_execution',
      execution_id: executionId
    });
  }

  /**
   * 恢复执行
   */
  resumeExecution(executionId: string): void {
    if (!this.isConnected()) {
      throw new Error('WebSocket未连接');
    }

    this.send({
      type: 'resume_execution',
      execution_id: executionId
    });
  }

  /**
   * 注册消息处理器
   */
  on(messageType: string, handler: WebSocketMessageHandler): void {
    if (!this.messageHandlers.has(messageType)) {
      this.messageHandlers.set(messageType, []);
    }
    this.messageHandlers.get(messageType)!.push(handler);
  }

  /**
   * 取消注册消息处理器
   */
  off(messageType: string, handler: WebSocketMessageHandler): void {
    const handlers = this.messageHandlers.get(messageType);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
      if (handlers.length === 0) {
        this.messageHandlers.delete(messageType);
      }
    }
  }

  /**
   * 获取连接统计信息
   */
  getConnectionStats(): Record<string, any> {
    return {
      connected: this.isConnected(),
      connection_id: this.connectionId,
      subscriptions: Array.from(this.subscriptions),
      reconnect_attempts: this.reconnectAttempts,
      handlers_count: Array.from(this.messageHandlers.values()).reduce((total, handlers) => total + handlers.length, 0)
    };
  }

  private handleMessage(message: WebSocketMessage): void {
    // 处理连接建立消息
    if (message.type === 'connection_established') {
      this.connectionId = message.data.connection_id;
      console.log('📡 WebSocket连接已确认:', this.connectionId);
      return;
    }

    // 处理ping消息
    if (message.type === 'ping') {
      this.send({ type: 'pong' });
      return;
    }

    // 触发注册的处理器
    const handlers = this.messageHandlers.get(message.type);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(message);
        } catch (error) {
          console.error(`WebSocket消息处理器错误 (${message.type}):`, error);
        }
      });
    }

    // 触发通用处理器
    const allHandlers = this.messageHandlers.get('*');
    if (allHandlers) {
      allHandlers.forEach(handler => {
        try {
          handler(message);
        } catch (error) {
          console.error('WebSocket通用消息处理器错误:', error);
        }
      });
    }
  }

  private scheduleReconnect(): void {
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);

    console.log(`📡 ${delay}ms后尝试重连 (${this.reconnectAttempts + 1}/${this.maxReconnectAttempts})`);

    setTimeout(() => {
      if (!this.isManualClose && !this.isConnected()) {
        this.reconnectAttempts++;
        this.connect().catch(error => {
          console.error('📡 重连失败:', error);
        });
      }
    }, delay);
  }

  private startPing(): void {
    this.stopPing();
    this.pingInterval = setInterval(() => {
      if (this.isConnected()) {
        this.send({ type: 'ping' });
      }
    }, 30000); // 30秒ping一次
  }

  private stopPing(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }
}

// 创建全局实例
export const startupWebSocketManager = new StartupWebSocketManager();

// 导出类型
export type { StartupWorkflow, StartupExecution, StartupWorkflowNode, StartupWorkflowEdge, WebSocketMessage };