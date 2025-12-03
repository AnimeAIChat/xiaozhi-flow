import { ContextMenuPlugin } from 'rete-context-menu-plugin';
import { Node } from 'rete';
import { log } from '../../../utils/logger';

/**
 * 创建启动流程专用的上下文菜单插件
 */
export const createStartupContextMenu = () => {
  return new ContextMenuPlugin({
    // 菜单项生成函数
    items: (node: Node | null) => {
      if (!node) {
        // 画布空白区域的菜单
        return [
          {
            label: '添加数据库节点',
            action: () => {
              log.info('添加数据库节点', null, 'ui', 'ContextMenu');
              // TODO: 实现添加节点逻辑
            },
            icon: '🗄️'
          },
          {
            label: '添加API节点',
            action: () => {
              log.info('添加API节点', null, 'ui', 'ContextMenu');
              // TODO: 实现添加节点逻辑
            },
            icon: '🔌'
          },
          {
            label: '添加AI节点',
            action: () => {
              log.info('添加AI节点', null, 'ui', 'ContextMenu');
              // TODO: 实现添加节点逻辑
            },
            icon: '🤖'
          },
          {
            label: '添加云服务节点',
            action: () => {
              log.info('添加云服务节点', null, 'ui', 'ContextMenu');
              // TODO: 实现添加节点逻辑
            },
            icon: '☁️'
          },
          { type: 'separator' },
          {
            label: '全部展开',
            action: () => {
              log.info('展开所有节点', null, 'ui', 'ContextMenu');
              // TODO: 实现展开逻辑
            },
            icon: '📂'
          },
          {
            label: '全部折叠',
            action: () => {
              log.info('折叠所有节点', null, 'ui', 'ContextMenu');
              // TODO: 实现折叠逻辑
            },
            icon: '📁'
          },
          { type: 'separator' },
          {
            label: '适应视图',
            action: () => {
              log.info('适应视图', null, 'ui', 'ContextMenu');
              // TODO: 实现适应视图逻辑
            },
            icon: '🎯'
          }
        ];
      }

      // 节点相关的菜单
      const nodeData = node.data as any;
      const nodeType = nodeData?.type || 'api';
      const nodeStatus = nodeData?.status || 'stopped';

      return [
        {
          label: '编辑配置',
          action: () => {
            log.info('编辑节点配置', { nodeId: node.id, nodeType }, 'ui', 'ContextMenu');
            // TODO: 实现编辑配置逻辑
            showNodeConfigDialog(node.id);
          },
          icon: '⚙️',
          shortcut: 'Ctrl+E'
        },
        {
          label: '查看详情',
          action: () => {
            log.info('查看节点详情', { nodeId: node.id }, 'ui', 'ContextMenu');
            // TODO: 实现查看详情逻辑
            showNodeDetails(node.id);
          },
          icon: '📋'
        },
        {
          label: '查看日志',
          action: () => {
            log.info('查看节点日志', { nodeId: node.id }, 'ui', 'ContextMenu');
            // TODO: 实现查看日志逻辑
            showNodeLogs(node.id);
          },
          icon: '📄',
          disabled: nodeStatus === 'stopped' // 停止状态下无法查看日志
        },
        { type: 'separator' },
        {
          label: nodeStatus === 'running' ? '停止节点' : '启动节点',
          action: () => {
            const action = nodeStatus === 'running' ? 'stop' : 'start';
            log.info(`${action === 'stop' ? '停止' : '启动'}节点`, { nodeId: node.id }, 'ui', 'ContextMenu');
            // TODO: 实现启动/停止逻辑
            toggleNodeStatus(node.id, action);
          },
          icon: nodeStatus === 'running' ? '⏹️' : '▶️',
          // 根据节点状态改变颜色
          style: {
            color: nodeStatus === 'running' ? '#ff4d4f' : '#52c41a'
          }
        },
        {
          label: '重启节点',
          action: () => {
            log.info('重启节点', { nodeId: node.id }, 'ui', 'ContextMenu');
            // TODO: 实现重启逻辑
            restartNode(node.id);
          },
          icon: '🔄',
          disabled: nodeStatus === 'stopped' // 停止状态下无法重启
        },
        { type: 'separator' },
        {
          label: '复制节点',
          action: () => {
            log.info('复制节点', { nodeId: node.id, nodeType }, 'ui', 'ContextMenu');
            // TODO: 实现复制节点逻辑
            duplicateNode(node.id);
          },
          icon: '📋',
          shortcut: 'Ctrl+D'
        },
        {
          label: '删除节点',
          action: () => {
            log.info('删除节点', { nodeId: node.id, nodeType }, 'ui', 'ContextMenu');
            // TODO: 实现删除节点逻辑
            deleteNode(node.id);
          },
          icon: '🗑️',
          style: {
            color: '#ff4d4f'
          },
          shortcut: 'Delete'
        },
        { type: 'separator' },
        {
          label: '查看依赖关系',
          action: () => {
            log.info('查看依赖关系', { nodeId: node.id }, 'ui', 'ContextMenu');
            // TODO: 实现查看依赖关系逻辑
            showNodeDependencies(node.id);
          },
          icon: '🔗'
        },
        {
          label: '高亮相关节点',
          action: () => {
            log.info('高亮相关节点', { nodeId: node.id }, 'ui', 'ContextMenu');
            // TODO: 实现高亮相关节点逻辑
            highlightRelatedNodes(node.id);
          },
          icon: '✨'
        }
      ];
    },
    // 菜单样式
    className: 'startup-context-menu',
    // 动画效果
    animation: 'fade',
    // 防止默认右键菜单
    preventDefault: true,
    // 菜单位置偏移
    offset: { x: 0, y: 0 }
  });
};

// 以下是菜单动作的占位符函数，实际实现时需要根据具体需求编写

/**
 * 显示节点配置对话框
 */
const showNodeConfigDialog = (nodeId: string) => {
  // TODO: 实现节点配置对话框
  console.log('显示节点配置对话框:', nodeId);
};

/**
 * 显示节点详情
 */
const showNodeDetails = (nodeId: string) => {
  // TODO: 实现节点详情显示
  console.log('显示节点详情:', nodeId);
};

/**
 * 显示节点日志
 */
const showNodeLogs = (nodeId: string) => {
  // TODO: 实现节点日志显示
  console.log('显示节点日志:', nodeId);
};

/**
 * 切换节点状态
 */
const toggleNodeStatus = (nodeId: string, action: 'start' | 'stop') => {
  // TODO: 实现节点状态切换
  console.log('切换节点状态:', nodeId, action);
};

/**
 * 重启节点
 */
const restartNode = (nodeId: string) => {
  // TODO: 实现节点重启
  console.log('重启节点:', nodeId);
};

/**
 * 复制节点
 */
const duplicateNode = (nodeId: string) => {
  // TODO: 实现节点复制
  console.log('复制节点:', nodeId);
};

/**
 * 删除节点
 */
const deleteNode = (nodeId: string) => {
  // TODO: 实现节点删除
  console.log('删除节点:', nodeId);
};

/**
 * 显示节点依赖关系
 */
const showNodeDependencies = (nodeId: string) => {
  // TODO: 实现节点依赖关系显示
  console.log('显示节点依赖关系:', nodeId);
};

/**
 * 高亮相关节点
 */
const highlightRelatedNodes = (nodeId: string) => {
  // TODO: 实现相关节点高亮
  console.log('高亮相关节点:', nodeId);
};