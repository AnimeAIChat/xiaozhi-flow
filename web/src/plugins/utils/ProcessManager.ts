// 简单的事件发射器实现，兼容浏览器环境
export class SimpleEventEmitter {
  private listeners: Map<string, Function[]> = new Map();

  on(event: string, listener: Function): void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event)!.push(listener);
  }

  emit(event: string, data?: any): void {
    const eventListeners = this.listeners.get(event);
    if (eventListeners) {
      eventListeners.forEach(listener => listener(data));
    }
  }

  off(event: string, listener: Function): void {
    const eventListeners = this.listeners.get(event);
    if (eventListeners) {
      const index = eventListeners.indexOf(listener);
      if (index > -1) {
        eventListeners.splice(index, 1);
      }
    }
  }
}

export interface ProcessOptions {
  command: string;
  args: string[];
  cwd?: string;
  env?: Record<string, string>;
  timeout?: number;
  stdio?: 'pipe' | 'inherit' | 'ignore' | Array<'pipe' | 'inherit' | 'ignore'>;
  shell?: boolean;
  detached?: boolean;
  uid?: number;
  gid?: number;
}

export interface ProcessInfo {
  pid: number;
  command: string;
  args: string[];
  cwd: string;
  startTime: Date;
  status: 'running' | 'stopped' | 'error';
  exitCode?: number;
  signal?: string;
  memoryUsage?: number;
  cpuUsage?: number;
}

export interface ProcessResult {
  stdout: string;
  stderr: string;
  code: number | null;
  signal: string | null;
}

export interface RunningProcess {
  pid: number;
  process: any; // 后端进程信息
  startTime: Date;
  status: 'starting' | 'running' | 'stopping' | 'stopped' | 'error';
  lastActivity: Date;
}

/**
 * 浏览器兼容的进程管理器
 * 通过API与后端通信，管理远程进程
 */
export class ProcessManager extends SimpleEventEmitter {
  private processes: Map<number, RunningProcess> = new Map();
  private cleanupInterval: number;

  constructor() {
    super();

    // 定期清理已结束的进程
    this.cleanupInterval = window.setInterval(() => {
      this.cleanup();
    }, 30000); // 每30秒清理一次
  }

  /**
   * 执行命令并等待完成
   */
  async executeCommand(options: ProcessOptions): Promise<ProcessResult> {
    try {
      const response = await fetch('/api/process/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(options)
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const result = await response.json();
      return result;
    } catch (error) {
      throw new Error(`Failed to execute command: ${error}`);
    }
  }

  /**
   * 启动进程并返回控制
   */
  async startProcess(options: ProcessOptions): Promise<RunningProcess> {
    try {
      const response = await fetch('/api/process/start', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(options)
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const processData = await response.json();

      const runningProcess: RunningProcess = {
        pid: processData.pid,
        process: processData, // 存储后端返回的进程信息
        startTime: new Date(processData.startTime),
        status: 'starting',
        lastActivity: new Date()
      };

      this.processes.set(processData.pid, runningProcess);
      this.emit('process-started', { pid: processData.pid, command: options.command, args: options.args });

      // 开始监控进程状态
      this.monitorProcess(processData.pid);

      return runningProcess;
    } catch (error) {
      throw new Error(`Failed to start process: ${error}`);
    }
  }

  /**
   * 杀死进程
   */
  async killProcess(pid: number, signal: string = 'SIGTERM'): Promise<void> {
    const runningProcess = this.processes.get(pid);
    if (!runningProcess) {
      throw new Error(`Process ${pid} not found`);
    }

    try {
      const response = await fetch('/api/process/kill', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ pid, signal })
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      // 更新本地状态
      runningProcess.status = 'stopping';
      this.emit('process-killed', { pid, signal });
    } catch (error) {
      throw new Error(`Failed to kill process ${pid}: ${error}`);
    }
  }

  /**
   * 检查进程是否运行
   */
  async isProcessRunning(pid: number): Promise<boolean> {
    try {
      const response = await fetch(`/api/process/${pid}/status`);

      if (!response.ok) {
        return false;
      }

      const status = await response.json();
      return status.status === 'running';
    } catch (error) {
      return false;
    }
  }

  /**
   * 获取进程信息
   */
  async getProcessInfo(pid: number): Promise<ProcessInfo> {
    try {
      const response = await fetch(`/api/process/${pid}/info`);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const info = await response.json();
      return info;
    } catch (error) {
      // 返回默认错误信息
      return {
        pid,
        command: 'unknown',
        args: [],
        cwd: '',
        startTime: new Date(),
        status: 'error'
      };
    }
  }

  /**
   * 获取进程内存使用情况
   */
  async getProcessMemoryUsage(pid: number): Promise<number> {
    try {
      const response = await fetch(`/api/process/${pid}/memory`);

      if (!response.ok) {
        return 0;
      }

      const data = await response.json();
      return data.memoryUsage || 0;
    } catch (error) {
      return 0;
    }
  }

  /**
   * 获取进程CPU使用情况
   */
  async getProcessCpuUsage(pid: number): Promise<number> {
    try {
      const response = await fetch(`/api/process/${pid}/cpu`);

      if (!response.ok) {
        return 0;
      }

      const data = await response.json();
      return data.cpuUsage || 0;
    } catch (error) {
      return 0;
    }
  }

  /**
   * 获取所有正在运行的进程
   */
  getRunningProcesses(): RunningProcess[] {
    return Array.from(this.processes.values());
  }

  /**
   * 获取进程数量
   */
  getProcessCount(): number {
    return this.processes.size;
  }

  /**
   * 监控进程状态
   */
  private async monitorProcess(pid: number): Promise<void> {
    const monitor = async () => {
      const runningProcess = this.processes.get(pid);
      if (!runningProcess) {
        return;
      }

      try {
        const isRunning = await this.isProcessRunning(pid);

        if (!isRunning && runningProcess.status !== 'stopped' && runningProcess.status !== 'error') {
          // 进程已停止，更新状态
          runningProcess.status = 'stopped';
          this.processes.delete(pid);
          this.emit('process-exit', { pid, code: null, signal: null });
        } else if (isRunning && runningProcess.status === 'starting') {
          // 进程已启动
          runningProcess.status = 'running';
        }

        runningProcess.lastActivity = new Date();

        // 如果进程还在运行，继续监控
        if (isRunning) {
          setTimeout(monitor, 5000); // 每5秒检查一次
        }
      } catch (error) {
        // 监控出错，标记为错误状态
        runningProcess.status = 'error';
        this.processes.delete(pid);
        this.emit('process-error', { pid, error });
      }
    };

    // 开始监控
    setTimeout(monitor, 1000); // 1秒后开始第一次检查
  }

  /**
   * 清理已结束的进程
   */
  cleanup(): void {
    const deadPids: number[] = [];

    for (const [pid, process] of this.processes) {
      const now = new Date();
      const timeSinceLastActivity = now.getTime() - process.lastActivity.getTime();

      // 清理超过5分钟没有活动且状态为停止的进程
      if (
        (process.status === 'stopped' || process.status === 'error') &&
        timeSinceLastActivity > 5 * 60 * 1000
      ) {
        deadPids.push(pid);
      }
    }

    deadPids.forEach(pid => {
      this.processes.delete(pid);
    });
  }

  /**
   * 杀死所有进程
   */
  async killAll(): Promise<void> {
    const killPromises = Array.from(this.processes.keys()).map(pid =>
      this.killProcess(pid).catch(error =>
        console.error(`Failed to kill process ${pid}:`, error)
      )
    );

    await Promise.all(killPromises);
    this.processes.clear();
  }

  /**
   * 销毁进程管理器
   */
  destroy(): void {
    if (this.cleanupInterval) {
      window.clearInterval(this.cleanupInterval);
    }

    this.killAll();
  }
}

// 全局进程管理器实例
export const processManager = new ProcessManager();