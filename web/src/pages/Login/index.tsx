import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import {
  Form,
  Input,
  Button,
  Card,
  Typography,
  Alert,
  Checkbox,
  Divider,
  Row,
  Col,
} from 'antd';
import {
  UserOutlined,
  LockOutlined,
  EyeInvisibleOutlined,
  EyeTwoTone,
} from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';
import { LoginRequest } from '../../types/auth';

const { Title, Text } = Typography;

interface LoginForm {
  username: string;
  password: string;
  remember: boolean;
}

const Login: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, isAuthenticated, isLoading, error, clearError } = useAuth();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [loginSuccess, setLoginSuccess] = useState(false);

  // 如果已经登录且不在加载状态，重定向到目标页面
  useEffect(() => {
    if (isAuthenticated && !isLoading && !loading) {
      const from = location.state?.from?.pathname || '/dashboard';
      navigate(from, { replace: true });
    }
  }, [isAuthenticated, isLoading, loading, loginSuccess, navigate, location.state]);

  // 清除错误信息
  useEffect(() => {
    return () => {
      clearError();
    };
  }, [clearError]);

  // 处理登录表单提交
  const handleSubmit = async (values: LoginForm) => {
    setLoading(true);
    clearError();

    try {
      const loginData: LoginRequest = {
        username: values.username.trim(),
        password: values.password,
      };

      await login(loginData);

      setLoginSuccess(true);

      // 登录成功后立即跳转
      const from = location.state?.from?.pathname || '/dashboard';
      navigate(from, { replace: true });

    } catch (error) {
      console.error('登录失败:', error);
      // 错误信息由AuthContext处理
    } finally {
      setLoading(false);
    }
  };

  // 处理表单验证失败
  const handleSubmitFailed = (errorInfo: any) => {
    console.log('表单验证失败:', errorInfo);
  };

  // 获取初始值
  const getInitialValues = (): Partial<LoginForm> => {
    const savedUsername = localStorage.getItem('remembered_username');
    return {
      username: savedUsername || '',
      password: '',
      remember: !!savedUsername,
    };
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p className="mt-4 text-gray-600">正在验证身份...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        {/* 登录卡片 */}
        <Card className="shadow-lg border-0">
          <div className="text-center mb-8">
            <Title level={2} className="text-gray-900 mb-2">
              欢迎回来
            </Title>
            <Text type="secondary">
              登录到您的账户以继续使用系统
            </Text>
          </div>

          {/* 错误提示 */}
          {error && (
            <Alert
              message="登录失败"
              description={error}
              type="error"
              showIcon
              closable
              onClose={clearError}
              className="mb-6"
            />
          )}

          {/* 成功提示 */}
          {loginSuccess && (
            <Alert
              message="登录成功"
              description="正在跳转到主页..."
              type="success"
              showIcon
              className="mb-6"
            />
          )}

          {/* 登录表单 */}
          <Form
            form={form}
            name="login"
            initialValues={getInitialValues()}
            onFinish={handleSubmit}
            onFinishFailed={handleSubmitFailed}
            size="large"
            layout="vertical"
            disabled={loginSuccess}
          >
            <Form.Item
              name="username"
              label="用户名"
              rules={[
                {
                  required: true,
                  message: '请输入用户名',
                },
                {
                  min: 3,
                  message: '用户名至少3个字符',
                },
                {
                  max: 50,
                  message: '用户名不能超过50个字符',
                },
              ]}
            >
              <Input
                prefix={<UserOutlined className="text-gray-400" />}
                placeholder="请输入用户名"
                autoComplete="username"
                disabled={loading || loginSuccess}
              />
            </Form.Item>

            <Form.Item
              name="password"
              label="密码"
              rules={[
                {
                  required: true,
                  message: '请输入密码',
                },
                {
                  min: 6,
                  message: '密码至少6个字符',
                },
              ]}
            >
              <Input.Password
                prefix={<LockOutlined className="text-gray-400" />}
                placeholder="请输入密码"
                autoComplete="current-password"
                iconRender={(visible) =>
                  visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
                }
                disabled={loading || loginSuccess}
              />
            </Form.Item>

            <Form.Item>
              <Row justify="space-between" align="middle">
                <Col>
                  <Form.Item name="remember" valuePropName="checked" noStyle>
                    <Checkbox disabled={loading || loginSuccess}>
                      记住用户名
                    </Checkbox>
                  </Form.Item>
                </Col>
                <Col>
                  {/* 后续可以添加忘记密码功能 */}
                  <Text type="secondary" className="text-sm">
                    忘记密码？
                  </Text>
                </Col>
              </Row>
            </Form.Item>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                disabled={loginSuccess}
                block
                className="h-12 text-base font-medium"
              >
                {loginSuccess ? '登录成功' : '登录'}
              </Button>
            </Form.Item>
          </Form>

          <Divider plain>
            <Text type="secondary" className="text-sm">
              或
            </Text>
          </Divider>

          {/* 注册链接 */}
          <div className="text-center">
            <Text type="secondary">
              还没有账户？
              {' '}
              <Link
                to="/register"
                className="text-blue-600 hover:text-blue-500 font-medium"
              >
                立即注册
              </Link>
            </Text>
          </div>
        </Card>

        {/* 版权信息 */}
        <div className="text-center mt-8">
          <Text type="secondary" className="text-sm">
            © 2024 XiaoZhi Flow. 保留所有权利.
          </Text>
        </div>
      </div>
    </div>
  );
};

export default Login;