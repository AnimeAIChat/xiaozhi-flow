import React from 'react';
import { Typography, Card } from 'antd';

const { Title } = Typography;

const Login: React.FC = () => {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <Card className="w-96">
        <Title level={3} className="text-center">登录</Title>
        <p>登录功能正在开发中...</p>
      </Card>
    </div>
  );
};

export default Login;