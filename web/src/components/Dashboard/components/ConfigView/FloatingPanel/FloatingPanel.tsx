/**
 * 悬浮面板基础组件
 * 支持拖拽、调整位置、固定等功能
 */

import React, { useRef, useEffect, useState, useCallback } from 'react';
import { Button, Space } from 'antd';
import { PushpinOutlined, CloseOutlined, ExpandOutlined } from '@ant-design/icons';
import { useSidebarState } from '../hooks/useSidebarState';
import './floating.css';

export interface FloatingPanelProps {
  visible: boolean;
  title?: string;
  pinned?: boolean;
  position?: { x: number; y: number };
  width?: number;
  height?: number;
  className?: string;
  children: React.ReactNode;
  onClose?: () => void;
  onPin?: (pinned: boolean) => void;
  onPositionChange?: (position: { x: number; y: number }) => void;
  resizable?: boolean;
  minimizable?: boolean;
  showHeader?: boolean;
}

const FloatingPanel: React.FC<FloatingPanelProps> = ({
  visible,
  title = '悬浮面板',
  pinned: initialPinned = false,
  position: initialPosition = { x: 100, y: 100 },
  width = 320,
  height = 400,
  className = '',
  children,
  onClose,
  onPin,
  onPositionChange,
  resizable = false,
  minimizable = false,
  showHeader = true,
}) => {
  const {
    panelPinned,
    panelPosition,
    setPanelPin,
    hidePanel,
    setPanelPosition,
  } = useSidebarState();

  const [isDragging, setIsDragging] = useState(false);
  const [isPinned, setIsPinned] = useState(initialPinned || panelPinned);
  const [position, setPosition] = useState(initialPosition || panelPosition);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [isAnimating, setIsAnimating] = useState(false);

  const panelRef = useRef<HTMLDivElement>(null);
  const headerRef = useRef<HTMLDivElement>(null);

  // 同步外部状态
  useEffect(() => {
    setIsPinned(initialPinned || panelPinned);
  }, [initialPinned, panelPinned]);

  useEffect(() => {
    setPosition(initialPosition || panelPosition);
  }, [initialPosition, panelPosition]);

  // 点击外部关闭（未固定时）
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        visible &&
        !isPinned &&
        panelRef.current &&
        !panelRef.current.contains(event.target as Node)
      ) {
        handleClose();
      }
    };

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && visible && !isPinned) {
        handleClose();
      }
    };

    if (visible && !isPinned) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleEscape);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
    };
  }, [visible, isPinned]);

  // 边界检查
  const checkBounds = useCallback((x: number, y: number) => {
    const maxX = window.innerWidth - (panelRef.current?.offsetWidth || width);
    const maxY = window.innerHeight - (panelRef.current?.offsetHeight || height);

    return {
      x: Math.max(0, Math.min(x, maxX)),
      y: Math.max(0, Math.min(y, maxY)),
    };
  }, [width, height]);

  // 拖拽开始
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (resizable || minimizable) return;

    setIsDragging(true);
    setDragStart({
      x: e.clientX - position.x,
      y: e.clientY - position.y,
    });

    if (panelRef.current) {
      panelRef.current.classList.add('dragging');
    }
  }, [position, resizable, minimizable]);

  // 拖拽移动
  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!isDragging) return;

      const newPosition = checkBounds(
        e.clientX - dragStart.x,
        e.clientY - dragStart.y
      );

      setPosition(newPosition);
      setPanelPosition(newPosition);
      onPositionChange?.(newPosition);
    };

    const handleMouseUp = () => {
      if (isDragging) {
        setIsDragging(false);
        if (panelRef.current) {
          panelRef.current.classList.remove('dragging');
        }
      }
    };

    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
    }

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isDragging, dragStart, checkBounds, setPanelPosition, onPositionChange]);

  // 处理关闭
  const handleClose = useCallback(() => {
    setIsAnimating(true);
    setTimeout(() => {
      hidePanel();
      onClose?.();
      setIsAnimating(false);
    }, 200);
  }, [hidePanel, onClose]);

  // 处理固定
  const handlePin = useCallback(() => {
    const newPinned = !isPinned;
    setIsPinned(newPinned);
    setPanelPin(newPinned);
    onPin?.(newPinned);
  }, [isPinned, setPanelPin, onPin]);

  // 处理最大化
  const handleMaximize = useCallback(() => {
    // 实现最大化逻辑
    setPosition({ x: 0, y: 0 });
    setPanelPosition({ x: 0, y: 0 });
  }, [setPanelPosition]);

  if (!visible) {
    return null;
  }

  return (
    <div
      ref={panelRef}
      className={`floating-panel floating-component-library backdrop-blur-enhanced glass-border ${visible ? 'visible' : 'hidden'} ${isAnimating ? 'exiting' : 'entering'} ${isPinned ? 'pinned' : ''} ${className}`}
      style={{
        left: position.x,
        top: position.y,
        width,
        height,
        zIndex: isPinned ? 1001 : 1000,
      }}
    >
      {/* 拖拽指示器 */}
      <div className="drag-indicator" />

      {/* 面板头部 */}
      {showHeader && (
        <div
          ref={headerRef}
          className={`floating-header ${isDragging ? 'dragging' : ''}`}
          onMouseDown={handleMouseDown}
        >
          <h4>{title}</h4>
          <div className="floating-actions">
            <Space size="small">
              {minimizable && (
                <Button
                  type="text"
                  icon={<ExpandOutlined />}
                  className="floating-action-btn"
                  onClick={handleMaximize}
                  title="最大化"
                />
              )}
              <Button
                type="text"
                icon={<PushpinOutlined />}
                className={`floating-action-btn ${isPinned ? 'pinned' : ''}`}
                onClick={handlePin}
                title={isPinned ? '取消固定' : '固定面板'}
              />
              <Button
                type="text"
                icon={<CloseOutlined />}
                className="floating-action-btn close"
                onClick={handleClose}
                title="关闭面板"
              />
            </Space>
          </div>
        </div>
      )}

      {/* 面板内容 */}
      <div className="floating-content">
        {children}
      </div>
    </div>
  );
};

export default FloatingPanel;