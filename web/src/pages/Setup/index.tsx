import React from 'react';
import { Button } from 'antd';
import { useNavigate } from 'react-router-dom';
import { FullscreenLayout } from '../../components/layout';

const Setup: React.FC = () => {
  const navigate = useNavigate();

  const handleStartConfig = () => {
    navigate('/config');
  };

  return (
    <FullscreenLayout>
      <div className="min-h-screen flex flex-col items-center justify-center px-4">
        {/* 标题 */}
        <div className="text-center mb-24">
          <h1 className="text-black mb-4 font-medium text-6xl tracking-tight">
            Xiaozhi Flow
          </h1>
          <p className="text-gray-600 text-xl leading-relaxed max-w-xl mx-auto font-light">
            现代化 AI 助手配置平台
          </p>
        </div>

        {/* 操作区域 */}
        <div className="w-full max-w-md">
          <p className="text-gray-600 text-sm font-light text-center mb-8">
            准备开始您的 AI 配置流程
          </p>

          <Button
            type="primary"
            size="large"
            onClick={handleStartConfig}
            className="w-full bg-black text-white hover:bg-gray-800 border-none h-14 text-base font-medium rounded-lg transition-all duration-200"
          >
            开始配置
          </Button>

          <div className="mt-8 text-center">
            <a
              href="https://github.com/AnimeAIChat/xiaozhi-flow"
              target="_blank"
              rel="noopener noreferrer"
              className="text-gray-500 hover:text-gray-700 text-sm font-light transition-colors"
            >
              查看项目
            </a>
          </div>
        </div>

        {/* 底部信息 */}
        <div className="absolute bottom-8 text-center">
          <p className="text-gray-500 text-xs font-light">
            版本 0.0.1
          </p>
        </div>
      </div>
    </FullscreenLayout>
  );
};

export default Setup;