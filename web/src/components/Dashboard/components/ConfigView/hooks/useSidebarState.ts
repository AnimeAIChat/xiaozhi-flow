/**
 * 配置页面侧边栏状态管理Hook
 * 提供配置侧边栏和组件库面板的状态管理功能
 */

import { useCallback } from 'react';
import { useAppStore } from '../../../../../stores/useAppStore';

export interface SidebarState {
  // 配置侧边栏状态
  collapsed: boolean;
  width: number;
  defaultWidth: number;
  minWidth: number;
  maxWidth: number;

  // 组件库面板状态
  panelVisible: boolean;
  panelPosition: { x: number; y: number };
  panelPinned: boolean;
}

export interface SidebarActions {
  // 配置侧边栏操作
  toggleSidebar: () => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
  setSidebarWidth: (width: number) => void;

  // 组件库面板操作
  togglePanel: () => void;
  showPanel: (position?: { x: number; y: number }) => void;
  hidePanel: () => void;
  setPanelPosition: (position: { x: number; y: number }) => void;
  togglePanelPin: () => void;
  setPanelPin: (pinned: boolean) => void;

  // 组合操作
  showPanelAtCenter: () => void;
  resetSidebar: () => void;
}

export const useSidebarState = (): SidebarState & SidebarActions => {
  // 获取状态和操作方法，添加安全检查
  const ui = useAppStore((state) => state.ui);

  // 安全访问状态，提供默认值
  const configSidebar = ui?.configSidebar || {
    collapsed: false,
    width: 280,
    defaultWidth: 280,
    minWidth: 240,
    maxWidth: 400,
  };

  const componentLibraryPanel = ui?.componentLibraryPanel || {
    visible: false,
    position: { x: 100, y: 100 },
    pinned: false,
  };

  const responsive = ui?.responsive || {
    isMobile: false,
    isTablet: false,
    isDesktop: true,
    screenSize: { width: 1920, height: 1080 },
  };

  // 获取操作方法
  const toggleConfigSidebar = useAppStore((state) => state.toggleConfigSidebar);
  const setConfigSidebarCollapsed = useAppStore((state) => state.setConfigSidebarCollapsed);
  const setConfigSidebarWidth = useAppStore((state) => state.setConfigSidebarWidth);
  const toggleComponentLibraryPanel = useAppStore((state) => state.toggleComponentLibraryPanel);
  const showComponentLibraryPanel = useAppStore((state) => state.showComponentLibraryPanel);
  const hideComponentLibraryPanel = useAppStore((state) => state.hideComponentLibraryPanel);
  const setComponentLibraryPanelPosition = useAppStore((state) => state.setComponentLibraryPanelPosition);
  const toggleComponentLibraryPanelPin = useAppStore((state) => state.toggleComponentLibraryPanelPin);
  const setComponentLibraryPanelPin = useAppStore((state) => state.setComponentLibraryPanelPin);

  // 组合操作：在屏幕中央显示面板
  const showPanelAtCenter = useCallback(() => {
    const centerX = window.innerWidth / 2 - 160; // 面板宽度的一半
    const centerY = window.innerHeight / 2 - 300; // 面板高度的一半
    showComponentLibraryPanel({ x: Math.max(20, centerX), y: Math.max(20, centerY) });
  }, [showComponentLibraryPanel]);

  // 组合操作：重置侧边栏状态
  const resetSidebar = useCallback(() => {
    setConfigSidebarCollapsed(false);
    setConfigSidebarWidth(280); // 使用默认值而不是可能未定义的状态
    hideComponentLibraryPanel();
    setComponentLibraryPanelPin(false);
  }, [
    setConfigSidebarCollapsed,
    setConfigSidebarWidth,
    hideComponentLibraryPanel,
    setComponentLibraryPanelPin,
  ]);

  return {
    // 配置侧边栏状态
    collapsed: configSidebar.collapsed,
    width: configSidebar.width,
    defaultWidth: configSidebar.defaultWidth,
    minWidth: configSidebar.minWidth,
    maxWidth: configSidebar.maxWidth,

    // 组件库面板状态
    panelVisible: componentLibraryPanel.visible,
    panelPosition: componentLibraryPanel.position,
    panelPinned: componentLibraryPanel.pinned,

    // 配置侧边栏操作
    toggleSidebar: toggleConfigSidebar,
    setSidebarCollapsed: setConfigSidebarCollapsed,
    setSidebarWidth: setConfigSidebarWidth,

    // 组件库面板操作
    togglePanel: toggleComponentLibraryPanel,
    showPanel: showComponentLibraryPanel,
    hidePanel: hideComponentLibraryPanel,
    setPanelPosition: setComponentLibraryPanelPosition,
    togglePanelPin: toggleComponentLibraryPanelPin,
    setPanelPin: setComponentLibraryPanelPin,

    // 组合操作
    showPanelAtCenter,
    resetSidebar,
  };
};

export default useSidebarState;