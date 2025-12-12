import { defineConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';

// Docs: https://rsbuild.rs/config/
export default defineConfig({
  plugins: [pluginReact()],
  source: {
    entry: {
      index: './src/index.tsx',
    },
  },
  output: {
    sourceMap: {
      // 在开发环境启用详细 sourcemap，在生产环境启用安全版本
      js:
        process.env.NODE_ENV === 'development'
          ? 'source-map'
          : 'hidden-source-map',
      css: process.env.NODE_ENV === 'development' ? true : false,
    },
    // 生成 source map 文件用于调试
    generateAssetsForBackground: true,
  },
  tools: {
    postcss: {
      config: {
        path: './postcss.config.js',
      },
    },
    // 增强 webpack 配置以支持更好的调试
    bundlerChain: (chain, { env }) => {
      if (env === 'development') {
        // 开发环境：启用所有调试功能
        chain.devtool('eval-source-map');

        // 优化开发构建性能
        chain.optimization.merge({
          splitChunks: {
            chunks: 'all',
            cacheGroups: {
              vendor: {
                test: /[\\/]node_modules[\\/]/,
                name: 'vendors',
                chunks: 'all',
              },
            },
          },
        });
      } else {
        // 生产环境：安全的 sourcemap 配置
        chain.devtool('hidden-source-map');

        // 生产环境优化
        chain.optimization.merge({
          minimize: true,
          splitChunks: {
            chunks: 'all',
            cacheGroups: {
              vendor: {
                test: /[\\/]node_modules[\\/]/,
                name: 'vendors',
                chunks: 'all',
                priority: 10,
              },
              common: {
                name: 'common',
                minChunks: 2,
                chunks: 'all',
                priority: 5,
              },
            },
          },
        });
      }
    },
  },
  css: {
    modules: {
      localsConvention: 'camelCaseOnly',
    },
    // 为 CSS 也生成 sourcemap
    sourceMap: process.env.NODE_ENV === 'development',
  },
  dev: {
    // 开发服务器配置
    hot: true,
    liveReload: true,
    overlay: {
      errors: true,
      warnings: false,
    },
    // 启用更详细的编译信息
    progressBar: true,
    printUrls: true,
  },
  performance: {
    // 性能监控配置
    chunkSplit: {
      strategy: 'split-by-experience',
      override: {
        chunks: 'all',
        minSize: 20000,
        maxSize: 244000,
        minChunks: 1,
        maxAsyncRequests: 30,
        maxInitialRequests: 25,
        automaticNameDelimiter: '~',
        cacheGroups: {
          reactLib: {
            test: /[\\/]node_modules[\\/](react|react-dom)[\\/]/,
            name: 'react-lib',
            chunks: 'all',
          },
          antdLib: {
            test: /[\\/]node_modules[\\/]antd[\\/]/,
            name: 'antd-lib',
            chunks: 'all',
          },
        },
      },
    },
  },
});
