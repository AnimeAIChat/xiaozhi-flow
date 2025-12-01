/**
 * 悬浮组件库面板
 * 将现有ComponentLibrary功能集成到可拖拽的悬浮面板中
 */

import React, { useState, useRef, useEffect, useMemo, useCallback } from 'react';
import { Card, Button, Input, Space, Typography, Collapse, Badge, Empty, Spin, Tabs } from 'antd';
import {
  SearchOutlined,
  SettingOutlined,
  CodeOutlined,
  DatabaseOutlined,
  ApiOutlined,
  ClockCircleOutlined,
  FolderOpenOutlined,
  AppstoreOutlined,
  BranchesOutlined,
  CloudOutlined,
  RobotOutlined,
  SoundOutlined,
  VideoCameraOutlined,
  DesktopOutlined,
  GlobalOutlined,
  WifiOutlined,
  ToolOutlined,
  SaveOutlined,
  PlusOutlined,
  FilterOutlined
} from '@ant-design/icons';
import FloatingPanel from './FloatingPanel';
import { useSidebarState } from '../hooks/useSidebarState';
// 导入组件库样式以保持拖拽效果
import '../ComponentLibrary/ComponentLibrary.css';
import type { ConfigRecord } from '../../../../../types/config';
import { usePlugins } from '../../../../../stores/useAppStore';
import { nodeRegistry } from '../../../../../nodes/registry/NodeRegistry';
import { dynamicNodeFactory } from '../../../../../nodes/factory/DynamicNodeFactory';
import {
  NodeDefinition,
  IPlugin
} from '../../../../../plugins/types';

// 组件库节点模板定义（复用现有定义）
const COMPONENT_TEMPLATES = [
  {
    id: 'llm-node',
    category: 'LLM',
    label: '大语言模型',
    icon: <RobotOutlined style={{ color: '#1890ff' }} />,
    description: '配置大语言模型节点',
    dataType: 'object',
    color: '#1890ff',
    subCategory: 'AI模型',
    defaultValue: {
      model: 'gpt-3.5-turbo',
      temperature: 0.7,
      max_tokens: 1000
    },
    tags: ['AI', '聊天', '文本生成']
  },
  {
    id: 'asr-node',
    category: 'ASR',
    label: '语音识别',
    icon: <SoundOutlined style={{ color: '#fa8c16' }} />,
    description: '配置语音识别服务',
    dataType: 'object',
    color: '#fa8c16',
    subCategory: '音频处理',
    defaultValue: {
      provider: 'openai',
      language: 'zh-CN',
      model: 'whisper-1'
    },
    tags: ['语音', '识别', '转文字']
  },
  {
    id: 'tts-node',
    category: 'TTS',
    label: '语音合成',
    icon: <VideoCameraOutlined style={{ color: '#52c41a' }} />,
    description: '配置文字转语音服务',
    dataType: 'object',
    color: '#52c41a',
    subCategory: '音频处理',
    defaultValue: {
      provider: 'azure',
      voice: 'zh-CN-XiaoxiaoNeural',
      rate: 1.0
    },
    tags: ['语音', '合成', 'TTS']
  },
  {
    id: 'vllm-node',
    category: 'VLLM',
    label: '视觉语言模型',
    icon: <VideoCameraOutlined style={{ color: '#722ed1' }} />,
    description: '配置视觉语言模型',
    dataType: 'object',
    color: '#722ed1',
    subCategory: 'AI模型',
    defaultValue: {
      model: 'gpt-4-vision-preview',
      max_tokens: 1000
    },
    tags: ['视觉', '图像', '多模态']
  },
  {
    id: 'server-node',
    category: 'server',
    label: '服务器配置',
    icon: <DesktopOutlined style={{ color: '#13c2c2' }} />,
    description: '配置服务器相关参数',
    dataType: 'object',
    color: '#13c2c2',
    subCategory: '基础设施',
    defaultValue: {
      host: 'localhost',
      port: 8080,
      ssl: false
    },
    tags: ['服务器', '后端', '配置']
  },
  {
    id: 'web-node',
    category: 'web',
    label: 'Web配置',
    icon: <GlobalOutlined style={{ color: '#eb2f96' }} />,
    description: '配置Web相关参数',
    dataType: 'object',
    color: '#eb2f96',
    subCategory: '前端',
    defaultValue: {
      baseUrl: 'https://api.example.com',
      timeout: 5000
    },
    tags: ['Web', 'API', '前端']
  },
  {
    id: 'transport-node',
    category: 'transport',
    label: '传输配置',
    icon: <WifiOutlined style={{ color: '#faad14' }} />,
    description: '配置传输协议参数',
    dataType: 'object',
    color: '#faad14',
    subCategory: '网络',
    defaultValue: {
      protocol: 'http',
      retries: 3,
      timeout: 10000
    },
    tags: ['传输', '网络', '协议']
  },
  {
    id: 'system-node',
    category: 'system',
    label: '系统配置',
    icon: <ToolOutlined style={{ color: '#f5222d' }} />,
    description: '配置系统相关参数',
    dataType: 'object',
    color: '#f5222d',
    subCategory: '系统',
    defaultValue: {
      debug: false,
      logLevel: 'info'
    },
    tags: ['系统', '配置', '调试']
  },
  {
    id: 'audio-node',
    category: 'audio',
    label: '音频配置',
    icon: <SoundOutlined style={{ color: '#a0d911' }} />,
    description: '配置音频处理参数',
    dataType: 'object',
    color: '#a0d911',
    subCategory: '音频处理',
    defaultValue: {
      sampleRate: 44100,
      channels: 2,
      bitrate: 128
    },
    tags: ['音频', '处理', '编解码']
  },
  {
    id: 'database-node',
    category: 'database',
    label: '数据库配置',
    icon: <SaveOutlined style={{ color: '#2f54eb' }} />,
    description: '配置数据库连接',
    dataType: 'object',
    color: '#2f54eb',
    subCategory: '数据存储',
    defaultValue: {
      type: 'mysql',
      host: 'localhost',
      port: 3306
    },
    tags: ['数据库', '存储', '连接']
  }
];

interface FloatingComponentLibraryProps {
  onNodeDragStart?: (template: any) => void;
  onNodeDragEnd?: () => void;
  className?: string;
  databaseNodes?: ConfigRecord[]; // 数据库中的节点
}

// 获取类别图标函数（与画布节点保持一致）
const getCategoryIcon = (category: string) => {
  switch (category) {
    case 'ASR':
      return <span style={{ color: '#fa8c16' }}>🎤</span>;
    case 'TTS':
      return <span style={{ color: '#52c41a' }}>🔊</span>;
    case 'LLM':
      return <span style={{ color: '#1890ff' }}>🤖</span>;
    case 'VLLM':
      return <span style={{ color: '#722ed1' }}>👁️</span>;
    case 'server':
      return <span style={{ color: '#13c2c2' }}>🖥️</span>;
    case 'web':
      return <span style={{ color: '#eb2f96' }}>🌐</span>;
    case 'transport':
      return <span style={{ color: '#faad14' }}>📡</span>;
    case 'system':
      return <span style={{ color: '#f5222d' }}>⚙️</span>;
    case 'audio':
      return <span style={{ color: '#a0d911' }}>🎵</span>;
    case 'database':
      return <span style={{ color: '#2f54eb' }}>💾</span>;
    default:
      return <SettingOutlined style={{ color: '#666' }} />;
  }
};

const FloatingComponentLibrary: React.FC<FloatingComponentLibraryProps> = ({
  onNodeDragStart,
  onNodeDragEnd,
  className = '',
  databaseNodes = [],
}) => {
  const {
    panelVisible,
    panelPinned,
    panelPosition,
    hidePanel,
    setPanelPin,
    setPanelPosition,
  } = useSidebarState();

  const plugins = usePlugins();
  const [searchText, setSearchText] = useState('');
  const [activeCategory, setActiveCategory] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('builtin');
  const draggedItem = useRef<any>(null);
  const [loading, setLoading] = useState(false);

  // 获取动态节点定义
  const dynamicNodes = useMemo(() => {
    return nodeRegistry.getAllNodeDefinitions();
  }, []);

  // 将动态节点转换为组件模板格式
  const dynamicNodeTemplates = useMemo(() => {
    return dynamicNodes.map(node => ({
      id: node.id,
      category: node.category,
      label: node.displayName,
      icon: <SettingOutlined style={{ color: node.color }} />,
      description: node.description,
      dataType: 'object',
      color: node.color,
      subCategory: node.subCategory || '插件节点',
      defaultValue: node.parameters.reduce((acc, param) => ({
        ...acc,
        [param.id]: param.defaultValue
      }), {}),
      tags: node.tags,
      isDynamicNode: true,
      nodeDefinition: node
    }));
  }, [dynamicNodes]);

  // 将数据库节点转换为组件模板格式
  const databaseNodeTemplates = useMemo(() => {
    return databaseNodes.map((node, index) => ({
      id: `db-node-${node.key}`,
      category: node.category || '自定义',
      label: node.label || node.key,
      icon: getCategoryIcon(node.category), // 使用与画布相同的类别图标
      description: node.description || '从数据库加载的配置节点',
      dataType: node.dataType,
      color: node.color || '#1890ff',
      subCategory: '数据库节点',
      defaultValue: node.value,
      tags: node.tags || ['数据库', '自定义'],
      isDatabaseNode: true, // 标记为数据库节点
      originalNode: node, // 保存原始节点数据
    }));
  }, [databaseNodes]);

  // 合并所有组件模板
  const allComponentTemplates = useMemo(() => {
    return [...COMPONENT_TEMPLATES, ...databaseNodeTemplates];
  }, [databaseNodeTemplates]);

  // 根据当前Tab获取对应的组件模板
  const currentTemplates = useMemo(() => {
    switch (activeTab) {
      case 'builtin':
        return COMPONENT_TEMPLATES;
      case 'database':
        return databaseNodeTemplates;
      case 'plugins':
        return dynamicNodeTemplates;
      default:
        return allComponentTemplates;
    }
  }, [activeTab, COMPONENT_TEMPLATES, databaseNodeTemplates, dynamicNodeTemplates, allComponentTemplates]);

  // 按类别分组组件
  const categorizedComponents = useMemo(() => {
    return currentTemplates.reduce((acc, component) => {
      const category = component.category;
      if (!acc[category]) {
        acc[category] = [];
      }
      acc[category].push(component);
      return acc;
    }, {} as Record<string, typeof currentTemplates>);
  }, [currentTemplates]);

  // 过滤组件
  const filterComponents = (components: typeof currentTemplates) => {
    return components.filter(component => {
      const matchesSearch = !searchText ||
        component.label.toLowerCase().includes(searchText.toLowerCase()) ||
        component.description.toLowerCase().includes(searchText.toLowerCase()) ||
        component.tags.some(tag => tag.toLowerCase().includes(searchText.toLowerCase()));

      const matchesCategory = !activeCategory || component.category === activeCategory;

      return matchesSearch && matchesCategory;
    });
  };

  // 处理拖拽开始
  const handleDragStart = (e: React.DragEvent, template: any) => {
    draggedItem.current = template;
    e.dataTransfer.effectAllowed = 'copy';
    e.dataTransfer.setData('application/reactflow', 'node');

    // 根据节点类型准备模板数据
    let templateData;
    if (template.isDatabaseNode) {
      // 数据库节点
      templateData = {
        ...template,
        isDatabaseNode: true,
        originalNode: template.originalNode
      };
    } else if (template.isDynamicNode) {
      // 动态节点
      templateData = {
        ...template,
        isDynamicNode: true,
        nodeDefinition: template.nodeDefinition
      };
    } else {
      // 内置节点
      templateData = template;
    }

    e.dataTransfer.setData('component-template', JSON.stringify(templateData));

    // 创建自定义拖拽图像
    const dragImage = document.createElement('div');
    dragImage.innerHTML = `
      <div style="
        padding: 8px 12px;
        background: ${template.color};
        color: white;
        border-radius: 6px;
        font-size: 12px;
        font-weight: 500;
        box-shadow: 0 2px 8px rgba(0,0,0,0.15);
      ">
        ${template.isDatabaseNode ? getCategoryIcon(template.category) :
          template.isDynamicNode ? template.icon : template.icon} ${template.label}
      </div>
    `;
    dragImage.style.position = 'absolute';
    dragImage.style.top = '-1000px';
    document.body.appendChild(dragImage);
    e.dataTransfer.setDragImage(dragImage, 0, 0);

    setTimeout(() => {
      document.body.removeChild(dragImage);
    }, 0);

    onNodeDragStart?.(templateData);
  };

  // 处理拖拽结束
  const handleDragEnd = () => {
    draggedItem.current = null;
    onNodeDragEnd?.();
  };

  // 渲染组件项
  const renderComponentItem = (template: any) => (
    <div
      key={template.id}
      draggable
      onDragStart={(e) => handleDragStart(e, template)}
      onDragEnd={handleDragEnd}
      className="component-item"
      style={{
        padding: '12px',
        margin: '8px 0',
        backgroundColor: template.isDatabaseNode ? `${template.color}08` : '#ffffff',
        border: template.isDatabaseNode ? `1px solid ${template.color}30` : '1px solid #e8e8e8',
        borderRadius: '8px',
        cursor: 'grab',
        transition: 'all 0.2s ease',
        position: 'relative',
        overflow: 'hidden',
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.transform = 'translateY(-2px)';
        e.currentTarget.style.boxShadow = `0 4px 12px ${template.color}20`;
        e.currentTarget.style.borderColor = template.color;
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.transform = 'translateY(0)';
        e.currentTarget.style.boxShadow = '0 2px 4px rgba(0,0,0,0.1)';
        e.currentTarget.style.borderColor = template.isDatabaseNode ? `${template.color}30` : '#e8e8e8';
      }}
      onMouseDown={(e) => {
        e.currentTarget.style.cursor = 'grabbing';
      }}
      onMouseUp={(e) => {
        e.currentTarget.style.cursor = 'grab';
      }}
    >
      {/* 拖拽指示器 */}
      <div style={{
        position: 'absolute',
        left: 0,
        top: 0,
        bottom: 0,
        width: '3px',
        backgroundColor: template.color
      }} />

      {/* 数据库节点标识 */}
      {template.isDatabaseNode && (
        <div style={{
          position: 'absolute',
          top: '4px',
          right: '4px',
          background: template.color,
          color: 'white',
          fontSize: '10px',
          padding: '2px 6px',
          borderRadius: '4px',
          fontWeight: '500'
        }}>
          数据库
        </div>
      )}

      <div style={{ display: 'flex', alignItems: 'center', marginBottom: '8px' }}>
        <div style={{ marginRight: '8px', fontSize: '16px' }}>
          {template.isDatabaseNode ? getCategoryIcon(template.category) : template.icon}
        </div>
        <div style={{ flex: 1 }}>
          <div style={{
            fontWeight: 500,
            fontSize: '14px',
            color: '#262626',
            marginBottom: '2px'
          }}>
            {template.label}
          </div>
          <Badge
            color={template.color}
            text={template.subCategory}
            style={{ fontSize: '10px', color: '#8c8c8c' }}
          />
        </div>
      </div>

      <div style={{
        fontSize: '12px',
        color: '#8c8c8c',
        marginBottom: '8px',
        lineHeight: '1.4'
      }}>
        {template.description}
      </div>

      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px' }}>
        {template.tags.map((tag: string) => (
          <span
            key={tag}
            style={{
              fontSize: '10px',
              padding: '2px 6px',
              backgroundColor: `${template.color}10`,
              color: template.color,
              border: `1px solid ${template.color}30`,
              borderRadius: '4px'
            }}
          >
            {tag}
          </span>
        ))}
      </div>
    </div>
  );

  // 组件库内容
  const componentLibraryContent = (
    <>
      {/* Tab切换 */}
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        size="small"
        style={{ marginBottom: '12px' }}
        items={[
          {
            key: 'builtin',
            label: (
              <Space>
                <AppstoreOutlined />
                内置组件
              </Space>
            )
          },
          {
            key: 'database',
            label: (
              <Space>
                <DatabaseOutlined />
                数据库
                {databaseNodeTemplates.length > 0 && (
                  <Badge count={databaseNodeTemplates.length} size="small" />
                )}
              </Space>
            )
          },
          {
            key: 'plugins',
            label: (
              <Space>
                <PlusOutlined />
                插件节点
                {dynamicNodeTemplates.length > 0 && (
                  <Badge count={dynamicNodeTemplates.length} size="small" />
                )}
              </Space>
            )
          }
        ]}
      />

      {/* 搜索框 */}
      <div style={{ marginBottom: '16px', padding: '0 4px' }}>
        <Input.Search
          placeholder="搜索组件..."
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          allowClear
          style={{ fontSize: '12px' }}
        />
      </div>

      {/* 分类过滤 */}
      <div style={{ marginBottom: '16px', padding: '0 4px' }}>
        <Space wrap size={[4, 4]}>
          <Button
            size="small"
            type={!activeCategory ? 'primary' : 'default'}
            onClick={() => setActiveCategory(null)}
            style={{ fontSize: '11px', height: '24px' }}
          >
            全部
          </Button>
          {Object.keys(categorizedComponents).map(category => (
            <Button
              key={category}
              size="small"
              type={activeCategory === category ? 'primary' : 'default'}
              onClick={() => setActiveCategory(category)}
              style={{ fontSize: '11px', height: '24px' }}
            >
              {category}
            </Button>
          ))}
        </Space>
      </div>

      {/* 组件列表 */}
      <div style={{ flex: 1, overflowY: 'auto', paddingRight: '4px' }}>
        {Object.entries(categorizedComponents).map(([category, components]) => {
          const filteredComponents = filterComponents(components);
          if (filteredComponents.length === 0) return null;

          return (
            <div key={category} style={{ marginBottom: '16px' }}>
              <div style={{
                fontSize: '12px',
                fontWeight: 600,
                color: '#595959',
                marginBottom: '8px',
                display: 'flex',
                alignItems: 'center',
                padding: '4px 0'
              }}>
                <BranchesOutlined style={{ marginRight: '6px', fontSize: '11px' }} />
                {category}
                <span style={{
                  marginLeft: '8px',
                  fontSize: '10px',
                  color: '#bfbfbf',
                  fontWeight: 'normal'
                }}>
                  ({filteredComponents.length})
                </span>
              </div>
              {filteredComponents.map(renderComponentItem)}
            </div>
          );
        })}

        {Object.values(categorizedComponents).every(components =>
          filterComponents(components).length === 0
        ) && (
          <div style={{
            textAlign: 'center',
            padding: '40px 20px',
            color: '#bfbfbf'
          }}>
            <CloudOutlined style={{ fontSize: '48px', marginBottom: '16px' }} />
            <div style={{ fontSize: '14px', marginBottom: '8px' }}>
              没有找到匹配的组件
            </div>
            <div style={{ fontSize: '12px' }}>
              尝试调整搜索条件
            </div>
          </div>
        )}
      </div>

      {/* 底部提示 */}
      <div style={{
        marginTop: '16px',
        paddingTop: '12px',
        borderTop: '1px solid rgba(0, 0, 0, 0.06)',
        fontSize: '11px',
        color: '#8c8c8c',
        textAlign: 'center'
      }}>
        <Space separator={<span>•</span>}>
          <span>拖拽创建</span>
          <span>自定义配置</span>
          <span>实时预览</span>
        </Space>
      </div>
    </>
  );

  return (
    <FloatingPanel
      visible={panelVisible}
      title={
        <span style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <AppstoreOutlined style={{ color: '#1890ff' }} />
          组件库
        </span>
      }
      pinned={panelPinned}
      position={panelPosition}
      width={320}
      height={600}
      className={className}
      onPin={setPanelPin}
      onPositionChange={setPanelPosition}
      onClose={hidePanel}
    >
      {componentLibraryContent}
    </FloatingPanel>
  );
};

export default FloatingComponentLibrary;