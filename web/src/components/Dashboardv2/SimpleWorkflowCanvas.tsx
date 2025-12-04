import React from 'react';
import { Card, Typography, Space, Button, Tag } from 'antd';
import { DatabaseOutlined, ApiOutlined, RobotOutlined, CloudOutlined, SettingOutlined, PlayCircleOutlined } from '@ant-design/icons';

const { Title, Paragraph } = Typography;

interface SimpleWorkflowCanvasProps {
  title: string;
  description: string;
}

export const SimpleWorkflowCanvas: React.FC<SimpleWorkflowCanvasProps> = ({
  title,
  description
}) => {
  return (
    <div className="w-full h-full flex items-center justify-center bg-gray-50">
      <div className="max-w-4xl mx-auto p-8">
        <Card className="bg-white shadow-lg border-0">
          <div className="text-center mb-8">
            <Title level={2} className="text-gray-800 mb-4">
              {title}
            </Title>
            <Paragraph className="text-gray-600 text-lg">
              {description}
            </Paragraph>
          </div>

          {/* 示例节点展示 */}
          <div className="mb-8">
            <Title level={4} className="text-gray-700 mb-4">可用节点类型</Title>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
              <NodeCard
                icon={<DatabaseOutlined style={{ fontSize: '24px' }} />}
                title="数据库节点"
                description="连接各种数据库，执行查询操作"
                color="#1890ff"
              />
              <NodeCard
                icon={<ApiOutlined style={{ fontSize: '24px' }} />}
                title="API 节点"
                description="调用 REST API 和 GraphQL 服务"
                color="#52c41a"
              />
              <NodeCard
                icon={<RobotOutlined style={{ fontSize: '24px' }} />}
                title="AI 节点"
                description="集成 AI 服务，智能数据处理"
                color="#722ed1"
              />
              <NodeCard
                icon={<CloudOutlined style={{ fontSize: '24px' }} />}
                title="云服务节点"
                description="连接云平台服务"
                color="#fa8c16"
              />
              <NodeCard
                icon={<SettingOutlined style={{ fontSize: '24px' }} />}
                title="配置节点"
                description="工作流配置和控制"
                color="#eb2f96"
              />
            </div>
          </div>

          {/* 特性展示 */}
          <div className="mb-8">
            <Title level={4} className="text-gray-700 mb-4">Rete.js 特性</Title>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FeatureCard
                title="模块化架构"
                description="基于插件的可扩展架构"
              />
              <FeatureCard
                title="可视化编辑"
                description="直观的拖拽式节点编辑器"
              />
              <FeatureCard
                title="实时连接"
                description="节点间的实时数据连接和传输"
              />
              <FeatureCard
                title="自定义渲染"
                description="完全自定义的节点外观和行为"
              />
            </div>
          </div>

          {/* 操作按钮 */}
          <div className="text-center">
            <Space size="large">
              <Button
                type="primary"
                size="large"
                icon={<PlayCircleOutlined />}
                className="bg-blue-500 hover:bg-blue-600"
              >
                开始使用
              </Button>
              <Button size="large">
                查看文档
              </Button>
            </Space>
          </div>

          {/* 状态信息 */}
          <div className="mt-8 p-4 bg-blue-50 rounded-lg">
            <div className="flex items-center justify-between">
              <div>
                <Tag color="green">Rete.js 2.0 已安装</Tag>
                <Tag color="blue">编辑器就绪</Tag>
              </div>
              <div className="text-sm text-gray-600">
                提示：完整的 Rete.js 编辑器实现正在开发中
              </div>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
};

interface NodeCardProps {
  icon: React.ReactNode;
  title: string;
  description: string;
  color: string;
}

const NodeCard: React.FC<NodeCardProps> = ({ icon, title, description, color }) => {
  return (
    <Card
      className="text-center hover:shadow-md transition-shadow duration-200"
      size="small"
    >
      <div className="mb-2" style={{ color }}>{icon}</div>
      <Title level={5} className="mb-2">{title}</Title>
      <Paragraph className="text-gray-600 text-sm mb-0">
        {description}
      </Paragraph>
    </Card>
  );
};

interface FeatureCardProps {
  title: string;
  description: string;
}

const FeatureCard: React.FC<FeatureCardProps> = ({ title, description }) => {
  return (
    <Card size="small" className="h-full">
      <Title level={5} className="mb-2 text-blue-600">{title}</Title>
      <Paragraph className="text-gray-600 text-sm mb-0">
        {description}
      </Paragraph>
    </Card>
  );
};