import React from 'react';
import { Button, Space, Typography } from 'antd';
import { EditOutlined } from '@ant-design/icons';
import { QuickActionsProps } from '../../types';

const { Text } = Typography;

const QuickActions: React.FC<QuickActionsProps> = ({ onConfigEdit }) => {
  return (
    <>
      {/* 配置编辑器按钮 */}
      <div className="absolute top-4 right-4 z-10 bg-white rounded-lg shadow-sm border border-gray-200 p-2">
        <Space>
          <Button
            type="default"
            size="small"
            icon={<EditOutlined />}
            onClick={onConfigEdit}
            title="双击画布区域也可以进入配置编辑器"
          >
            配置编辑器
          </Button>
        </Space>
      </div>

      {/* 双击提示 */}
      <div className="absolute bottom-4 left-4 z-10 bg-white bg-opacity-90 rounded-lg shadow-sm border border-gray-200 px-3 py-2">
        <Space size="small">
          <Text type="secondary" style={{ fontSize: 12 }}>
            💡 双击画布区域打开配置编辑器
          </Text>
        </Space>
      </div>
    </>
  );
};

export default QuickActions;