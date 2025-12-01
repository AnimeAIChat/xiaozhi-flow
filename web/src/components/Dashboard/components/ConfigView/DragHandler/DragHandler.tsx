/**
 * 拖拽处理器
 * 处理从组件库拖拽到ReactFlow画布的逻辑
 */

import React, { useCallback, useRef } from 'react';
import { useReactFlow } from 'reactflow';
// 简单的ID生成器，替代uuid
const generateId = () => {
  return 'node_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
};
import { message } from 'antd';
import type { ConfigNode } from '../../../../../types/config';

interface DragHandlerProps {
  children: React.ReactNode;
  onNodeCreate?: (node: ConfigNode) => void;
}

const DragHandler: React.FC<DragHandlerProps> = ({ children, onNodeCreate }) => {
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const { project, screenToFlowPosition } = useReactFlow();

  // 处理拖拽放置
  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();

      // 移除拖拽区域高亮样式
      const renderer = event.currentTarget.querySelector('.react-flow__renderer');
      if (renderer) {
        renderer.classList.remove('dropzone');
      }

      try {
        // 获取拖拽的组件模板数据
        const componentTemplate = JSON.parse(
          event.dataTransfer.getData('component-template')
        );

        if (!componentTemplate) {
          return;
        }

        // 计算放置位置
        const position = screenToFlowPosition({
          x: event.clientX,
          y: event.clientY,
        });

        // 创建新节点
        const newNode: ConfigNode = {
          id: generateId(),
          type: 'config',
          position,
          data: {
            key: `${componentTemplate.category.toLowerCase()}_${Date.now()}`,
            label: componentTemplate.label,
            description: componentTemplate.description,
            dataType: componentTemplate.dataType,
            value: componentTemplate.defaultValue,
            category: componentTemplate.category,
            subCategory: componentTemplate.subCategory,
            color: componentTemplate.color,
            editable: true,
            configCount: 0, // 非分组节点
            tags: componentTemplate.tags
          },
        };

        // 触发节点创建事件
        onNodeCreate?.(newNode);

        // 显示成功消息
        message.success(`已添加 ${componentTemplate.label} 节点`);

      } catch (error) {
        console.error('拖拽创建节点失败:', error);
        message.error('创建节点失败，请重试');
      }
    },
    [screenToFlowPosition, onNodeCreate]
  );

  // 处理拖拽悬停
  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';

    // 添加拖拽区域高亮样式
    const renderer = event.currentTarget.querySelector('.react-flow__renderer');
    if (renderer) {
      renderer.classList.add('dropzone');
    }
  }, []);

  // 处理拖拽离开
  const onDragLeave = useCallback((event: React.DragEvent) => {
    // 移除拖拽区域高亮样式
    const renderer = event.currentTarget.querySelector('.react-flow__renderer');
    if (renderer) {
      renderer.classList.remove('dropzone');
    }
  }, []);

  return (
    <div
      ref={reactFlowWrapper}
      style={{ width: '100%', height: '100%' }}
      onDrop={onDrop}
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
    >
      {children}
    </div>
  );
};

export default DragHandler;