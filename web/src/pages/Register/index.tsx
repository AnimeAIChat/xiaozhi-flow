import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
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
  MailOutlined,
  LockOutlined,
  EyeInvisibleOutlined,
  EyeTwoTone,
} from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';
import { RegisterRequest } from '../../types/auth';

const { Title, Text } = Typography;

interface RegisterForm {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  agree: boolean;
}

const Register: React.FC = () => {
  const navigate = useNavigate();
  const { register, isAuthenticated, isLoading, error, clearError } = useAuth();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [registerSuccess, setRegisterSuccess] = useState(false);

  // 如果已经登录，重定向到仪表板
  useEffect(() => {
    if (isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, navigate]);

  // 清除错误信息
  useEffect(() => {
    return () => {
      clearError();
    };
  }, [clearError]);

  // 处理注册表单提交
  const handleSubmit = async (values: RegisterForm) => {
    setLoading(true);
    clearError();

    try {
      const registerData: RegisterRequest = {
        username: values.username.trim(),
        email: values.email.trim().toLowerCase(),
        password: values.password,
      };

      await register(registerData);

      setRegisterSuccess(true);

      // 延迟导航，让用户看到成功消息
      setTimeout(() => {
        navigate('/dashboard', { replace: true });
      }, 2000);

    } catch (error) {
      console.error('注册失败:', error);
      // 错误信息由AuthContext处理
    } finally {
      setLoading(false);
    }
  };

  // 处理表单验证失败
  const handleSubmitFailed = (errorInfo: any) => {
    console.log('表单验证失败:', errorInfo);
  };

  // 验证用户名格式
  const validateUsername = (_: any, value: string) => {
    if (!value) {
      return Promise.reject(new Error('请输入用户名'));
    }
    if (value.length < 3) {
      return Promise.reject(new Error('用户名至少3个字符'));
    }
    if (value.length > 50) {
      return Promise.reject(new Error('用户名不能超过50个字符'));
    }
    if (!/^[a-zA-Z0-9_\u4e00-\u9fa5]+$/.test(value)) {
      return Promise.reject(new Error('用户名只能包含字母、数字、下划线和中文字符'));
    }
    return Promise.resolve();
  };

  // 验证邮箱格式
  const validateEmail = (_: any, value: string) => {
    if (!value) {
      return Promise.reject(new Error('请输入邮箱地址'));
    }
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(value)) {
      return Promise.reject(new Error('请输入有效的邮箱地址'));
    }
    return Promise.resolve();
  };

  // 验证密码格式
  const validatePassword = (_: any, value: string) => {
    if (!value) {
      return Promise.reject(new Error('请输入密码'));
    }
    if (value.length < 6) {
      return Promise.reject(new Error('密码至少6个字符'));
    }
    if (value.length > 128) {
      return Promise.reject(new Error('密码不能超过128个字符'));
    }
    // 检查密码复杂度（至少包含字母和数字）
    if (!/(?=.*[a-zA-Z])(?=.*\d)/.test(value)) {
      return Promise.reject(new Error('密码必须包含至少一个字母和一个数字'));
    }
    return Promise.resolve();
  };

  // 验证密码确认
  const validateConfirmPassword = (_: any, value: string) => {
    if (!value) {
      return Promise.reject(new Error('请确认密码'));
    }
    const password = form.getFieldValue('password');
    if (value !== password) {
      return Promise.reject(new Error('两次输入的密码不一致'));
    }
    return Promise.resolve();
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p className="mt-4 text-gray-600">正在创建账户...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        {/* 注册卡片 */}
        <Card className="shadow-lg border-0">
          <div className="text-center mb-8">
            <Title level={2} className="text-gray-900 mb-2">
              创建账户
            </Title>
            <Text type="secondary">
              填写以下信息来创建您的新账户
            </Text>
          </div>

          {/* 错误提示 */}
          {error && (
            <Alert
              message="注册失败"
              description={error}
              type="error"
              showIcon
              closable
              onClose={clearError}
              className="mb-6"
            />
          )}

          {/* 成功提示 */}
          {registerSuccess && (
            <Alert
              message="注册成功"
              description="账户创建成功，正在跳转到主页..."
              type="success"
              showIcon
              className="mb-6"
            />
          )}

          {/* 注册表单 */}
          <Form
            form={form}
            name="register"
            onFinish={handleSubmit}
            onFinishFailed={handleSubmitFailed}
            size="large"
            layout="vertical"
            disabled={registerSuccess}
          >
            <Form.Item
              name="username"
              label="用户名"
              rules={[
                { validator: validateUsername },
              ]}
            >
              <Input
                prefix={<UserOutlined className="text-gray-400" />}
                placeholder="请输入用户名"
                autoComplete="username"
                disabled={loading || registerSuccess}
              />
            </Form.Item>

            <Form.Item
              name="email"
              label="邮箱地址"
              rules={[
                { validator: validateEmail },
              ]}
            >
              <Input
                prefix={<MailOutlined className="text-gray-400" />}
                placeholder="请输入邮箱地址"
                autoComplete="email"
                disabled={loading || registerSuccess}
              />
            </Form.Item>

            <Form.Item
              name="password"
              label="密码"
              rules={[
                { validator: validatePassword },
              ]}
            >
              <Input.Password
                prefix={<LockOutlined className="text-gray-400" />}
                placeholder="请输入密码"
                autoComplete="new-password"
                iconRender={(visible) =>
                  visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
                }
                disabled={loading || registerSuccess}
              />
            </Form.Item>

            <Form.Item
              name="confirmPassword"
              label="确认密码"
              rules={[
                { validator: validateConfirmPassword },
              ]}
            >
              <Input.Password
                prefix={<LockOutlined className="text-gray-400" />}
                placeholder="请再次输入密码"
                autoComplete="new-password"
                iconRender={(visible) =>
                  visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
                }
                disabled={loading || registerSuccess}
              />
            </Form.Item>

            <Form.Item
              name="agree"
              valuePropName="checked"
              rules={[
                {
                  validator: (_, value) =>
                    value
                      ? Promise.resolve()
                      : Promise.reject(new Error('请同意服务条款和隐私政策')),
                },
              ]}
            >
              <Checkbox disabled={loading || registerSuccess}>
                我已阅读并同意
                {' '}
                <Link to="/terms" target="_blank" className="text-blue-600 hover:text-blue-500">
                  服务条款
                </Link>
                {' '}
                和
                {' '}
                <Link to="/privacy" target="_blank" className="text-blue-600 hover:text-blue-500">
                  隐私政策
                </Link>
              </Checkbox>
            </Form.Item>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                disabled={registerSuccess}
                block
                className="h-12 text-base font-medium"
              >
                {registerSuccess ? '创建成功' : '创建账户'}
              </Button>
            </Form.Item>
          </Form>

          <Divider plain>
            <Text type="secondary" className="text-sm">
              或
            </Text>
          </Divider>

          {/* 登录链接 */}
          <div className="text-center">
            <Text type="secondary">
              已有账户？
              {' '}
              <Link
                to="/login"
                className="text-blue-600 hover:text-blue-500 font-medium"
              >
                立即登录
              </Link>
            </Text>
          </div>
        </Card>

        {/* 安全提示 */}
        <Card className="bg-blue-50 border-blue-200">
          <div className="text-sm">
            <Text className="text-blue-800 font-medium">安全提示：</Text>
            <ul className="mt-2 space-y-1 text-blue-700">
              <li>• 请使用包含字母和数字的强密码</li>
              <li>• 定期更新您的密码以保证账户安全</li>
              <li>• 不要与他人分享您的登录凭据</li>
            </ul>
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

export default Register;