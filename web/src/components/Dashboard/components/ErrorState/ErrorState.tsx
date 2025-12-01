import React from 'react';
import { DatabaseOutlined } from '@ant-design/icons';
import { FullscreenLayout } from '../../../layout';
import { ErrorStateProps } from '../../types';

const ErrorState: React.FC<ErrorStateProps> = ({ error }) => {
  return (
    <FullscreenLayout>
      <div className="flex items-center justify-center min-h-screen bg-white">
        <div className="text-center p-8 bg-red-50 rounded-lg border border-red-200">
          <DatabaseOutlined className="text-4xl text-red-500 mb-4" />
          <div className="text-lg text-red-600 mb-2">数据库表结构加载失败</div>
          <div className="text-sm text-red-500">{error}</div>
        </div>
      </div>
    </FullscreenLayout>
  );
};

export default ErrorState;