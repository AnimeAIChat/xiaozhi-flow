/**
 * 配置编辑器页面
 */

import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { FullscreenLayout } from '../../components/layout';
import ConfigCanvasWrapper from '../../components/ConfigEditor/ConfigCanvas';
import { configService } from '../../services/configService';
import { log } from '../../utils/logger';

const ConfigEditor: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const initEditor = async () => {
      try {
        log.info('初始化配置编辑器', null, 'config', 'ConfigEditor');

        // 检查配置服务是否可用
        await configService.getConfigs({ limit: 1 });

        setLoading(false);
        setError(null);
      } catch (err) {
        console.error('配置编辑器初始化失败:', err);
        setError(err instanceof Error ? err.message : '配置编辑器初始化失败');
        setLoading(false);
      }
    };

    initEditor();
  }, []);

  if (loading) {
    return (
      <FullscreenLayout>
        <div className="flex items-center justify-center min-h-screen bg-white">
          <div className="text-center">
            <div className="w-12 h-12 border-4 border-gray-200 rounded-full animate-spin border-t-blue-500 border-r-blue-500 mx-auto mb-4"></div>
            <div className="text-lg text-gray-600">正在初始化配置编辑器...</div>
          </div>
        </div>
      </FullscreenLayout>
    );
  }

  if (error) {
    return (
      <FullscreenLayout>
        <div className="flex items-center justify-center min-h-screen bg-white">
          <div className="text-center p-8 bg-red-50 rounded-lg border border-red-200 max-w-md">
            <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <span className="text-2xl">⚠️</span>
            </div>
            <div className="text-lg text-red-600 mb-2">配置编辑器初始化失败</div>
            <div className="text-sm text-red-500 mb-4">{error}</div>
            <div className="space-x-2">
              <button
                onClick={() => window.location.reload()}
                className="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600"
              >
                重试
              </button>
              <button
                onClick={() => navigate('/dashboard')}
                className="px-4 py-2 bg-gray-300 text-gray-700 rounded hover:bg-gray-400"
              >
                返回
              </button>
            </div>
          </div>
        </div>
      </FullscreenLayout>
    );
  }

  return (
    <FullscreenLayout>
      <ConfigCanvasWrapper onClose={() => navigate('/dashboard')} />
    </FullscreenLayout>
  );
};

export default ConfigEditor;