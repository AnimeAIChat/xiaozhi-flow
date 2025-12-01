import React from 'react';
import { FullscreenLayout } from '../../../layout';
import { LoadingStateProps } from '../../types';

const LoadingState: React.FC<LoadingStateProps> = ({ message = '加载中...' }) => {
  return (
    <FullscreenLayout>
      <div className="flex items-center justify-center min-h-screen bg-white">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-gray-200 rounded-full animate-spin border-t-blue-500 border-r-blue-500 mx-auto mb-4"></div>
          <div className="text-lg text-gray-600">{message}</div>
        </div>
      </div>
    </FullscreenLayout>
  );
};

export default LoadingState;