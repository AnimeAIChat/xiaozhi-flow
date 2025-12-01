import React, { useState, useEffect, useCallback } from 'react';
import {
  ParameterDefinition,
  ParameterType
} from '../../plugins/types';
import {
  Form,
  Input,
  InputNumber,
  Switch,
  Select,
  Radio,
  Checkbox,
  Button,
  Space,
  Divider,
  Tooltip,
  message,
  Collapse,
  Row,
  Col
} from 'antd';
import {
  SettingOutlined,
  EyeInvisibleOutlined,
  EyeOutlined,
  PlusOutlined,
  DeleteOutlined
} from '@ant-design/icons';

const { TextArea } = Input;
const { Panel } = Collapse;
const { Option } = Select;

interface ParameterEditorProps {
  parameters: ParameterDefinition[];
  values: Record<string, any>;
  onChange?: (paramId: string, value: any) => void;
  onValidate?: (paramId: string, isValid: boolean, error?: string) => void;
  readOnly?: boolean;
  showAdvanced?: boolean;
  collapsible?: boolean;
}

interface DynamicOptionsState {
  [paramId: string]: Array<{ label: string; value: any }>;
}

/**
 * 参数输入组件
 */
const ParameterInput: React.FC<{
  parameter: ParameterDefinition;
  value: any;
  onChange: (value: any) => void;
  dynamicOptions?: Array<{ label: string; value: any }>;
  readOnly?: boolean;
}> = ({ parameter, value, onChange, dynamicOptions, readOnly }) => {
  const [showSecret, setShowSecret] = useState(false);

  const handleChange = useCallback((newValue: any) => {
    if (parameter.type === 'number') {
      onChange(newValue);
    } else if (parameter.type === 'boolean') {
      onChange(newValue);
    } else {
      onChange(newValue);
    }
  }, [onChange]);

  const renderInput = () => {
    if (readOnly) {
      return renderReadOnlyValue();
    }

    switch (parameter.type) {
      case 'string':
        if (parameter.ui?.component === 'textarea') {
          return (
            <TextArea
              value={value || ''}
              onChange={(e) => handleChange(e.target.value)}
              placeholder={parameter.ui?.placeholder}
              rows={parameter.ui?.height === 'large' ? 6 : parameter.ui?.height === 'small' ? 2 : 4}
              maxLength={parameter.constraints?.maxLength}
            />
          );
        }

        if (parameter.type === 'secret' || parameter.type === 'api-key') {
          return (
            <Input.Password
              value={value || ''}
              onChange={(e) => handleChange(e.target.value)}
              placeholder={parameter.ui?.placeholder}
              visibilityToggle={{
                visible: showSecret,
                onVisibleChange: setShowSecret
              }}
            />
          );
        }

        return (
          <Input
            value={value || ''}
            onChange={(e) => handleChange(e.target.value)}
            placeholder={parameter.ui?.placeholder}
            maxLength={parameter.constraints?.maxLength}
          />
        );

      case 'number':
        return (
          <InputNumber
            value={value !== undefined ? Number(value) : undefined}
            onChange={(newValue) => handleChange(newValue)}
            placeholder={parameter.ui?.placeholder}
            min={parameter.constraints?.min}
            max={parameter.constraints?.max}
            step={parameter.type === 'number' ? 0.1 : 1}
            style={{ width: '100%' }}
          />
        );

      case 'boolean':
        return (
          <Switch
            checked={value || false}
            onChange={(checked) => handleChange(checked)}
          />
        );

      case 'select':
        const options = dynamicOptions || parameter.constraints?.options || [];
        return (
          <Select
            value={value}
            onChange={(newValue) => handleChange(newValue)}
            placeholder={parameter.ui?.placeholder || `Select ${parameter.name}`}
            style={{ width: '100%' }}
            allowClear={!parameter.required}
          >
            {options.map((option) => (
              <Option key={option.value} value={option.value}>
                {option.label}
              </Option>
            ))}
          </Select>
        );

      case 'multiselect':
        const multiOptions = dynamicOptions || parameter.constraints?.options || [];
        return (
          <Select
            value={value || []}
            onChange={(newValue) => handleChange(newValue)}
            placeholder={parameter.ui?.placeholder || `Select ${parameter.name}`}
            mode="multiple"
            style={{ width: '100%' }}
            allowClear
          >
            {multiOptions.map((option) => (
              <Option key={option.value} value={option.value}>
                {option.label}
              </Option>
            ))}
          </Select>
        );

      case 'code':
        return (
          <TextArea
            value={typeof value === 'string' ? value : JSON.stringify(value, null, 2)}
            onChange={(e) => {
              try {
                const parsed = JSON.parse(e.target.value);
                handleChange(parsed);
              } catch {
                handleChange(e.target.value);
              }
            }}
            placeholder={parameter.ui?.placeholder || 'Enter JSON code...'}
            rows={6}
            style={{ fontFamily: 'monospace' }}
          />
        );

      case 'json':
        return (
          <TextArea
            value={typeof value === 'string' ? value : JSON.stringify(value, null, 2)}
            onChange={(e) => {
              try {
                const parsed = JSON.parse(e.target.value);
                handleChange(parsed);
              } catch {
                handleChange(e.target.value);
              }
            }}
            placeholder={parameter.ui?.placeholder || 'Enter JSON...'}
            rows={4}
            style={{ fontFamily: 'monospace' }}
          />
        );

      case 'array':
        if (Array.isArray(value)) {
          return (
            <div>
              {value.map((item, index) => (
                <div key={index} style={{ marginBottom: 8, display: 'flex', alignItems: 'center' }}>
                  <Input
                    value={String(item)}
                    onChange={(e) => {
                      const newArray = [...value];
                      newArray[index] = e.target.value;
                      handleChange(newArray);
                    }}
                    style={{ flex: 1 }}
                  />
                  <Button
                    type="text"
                    danger
                    icon={<DeleteOutlined />}
                    onClick={() => {
                      const newArray = value.filter((_, i) => i !== index);
                      handleChange(newArray);
                    }}
                  />
                </div>
              ))}
              <Button
                type="dashed"
                icon={<PlusOutlined />}
                onClick={() => {
                  const newArray = [...value, ''];
                  handleChange(newArray);
                }}
                style={{ width: '100%' }}
              >
                Add Item
              </Button>
            </div>
          );
        }
        return (
          <Input
            value={value || ''}
            onChange={(e) => {
              try {
                const parsed = JSON.parse(e.target.value);
                handleChange(parsed);
              } catch {
                handleChange(e.target.value);
              }
            }}
            placeholder={parameter.ui?.placeholder || 'Enter array...'}
          />
        );

      case 'object':
        return (
          <TextArea
            value={typeof value === 'string' ? value : JSON.stringify(value, null, 2)}
            onChange={(e) => {
              try {
                const parsed = JSON.parse(e.target.value);
                handleChange(parsed);
              } catch {
                handleChange(e.target.value);
              }
            }}
            placeholder={parameter.ui?.placeholder || 'Enter object...'}
            rows={4}
            style={{ fontFamily: 'monospace' }}
          />
        );

      case 'file':
        return (
          <Input
            type="file"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) {
                handleChange(file.path || file.name);
              }
            }}
          />
        );

      case 'directory':
        return (
          <Input
            value={value || ''}
            onChange={(e) => handleChange(e.target.value)}
            placeholder={parameter.ui?.placeholder || 'Enter directory path...'}
          />
        );

      default:
        return (
          <Input
            value={value || ''}
            onChange={(e) => handleChange(e.target.value)}
            placeholder={parameter.ui?.placeholder}
          />
        );
    }
  };

  const renderReadOnlyValue = () => {
    if (value === null || value === undefined || value === '') {
      return <span style={{ color: '#999' }}>Not set</span>;
    }

    switch (parameter.type) {
      case 'boolean':
        return value ? 'True' : 'False';

      case 'array':
        return Array.isArray(value) ? `${value.length} items` : String(value);

      case 'object':
        try {
          const jsonString = JSON.stringify(value, null, 2);
          return (
            <pre style={{ fontSize: '12px', maxHeight: '100px', overflow: 'auto' }}>
              {jsonString}
            </pre>
          );
        } catch {
          return String(value);
        }

      case 'secret':
      case 'api-key':
        return (
          <span style={{ fontFamily: 'monospace' }}>
            {'*'.repeat(Math.min(String(value).length, 8))}
          </span>
        );

      default:
        return String(value);
    }
  };

  return (
    <div>
      {renderInput()}
      {parameter.ui?.helpText && (
        <div style={{ fontSize: '12px', color: '#666', marginTop: 4 }}>
          {parameter.ui.helpText}
        </div>
      )}
    </div>
  );
};

/**
 * 参数编辑器主组件
 */
export const ParameterEditor: React.FC<ParameterEditorProps> = ({
  parameters,
  values,
  onChange,
  onValidate,
  readOnly = false,
  showAdvanced = false,
  collapsible = false
}) => {
  const [dynamicOptions, setDynamicOptions] = useState<DynamicOptionsState>({});
  const [activeKeys, setActiveKeys] = useState<string[]>([]);

  // 参数分组
  const groupedParameters = parameters.reduce((groups, param) => {
    const group = param.group || 'General';
    if (!groups[group]) {
      groups[group] = [];
    }
    groups[group].push(param);
    return groups;
  }, {} as Record<string, ParameterDefinition[]>);

  // 分离高级参数
  const basicParameters = parameters.filter(param => !param.ui?.advanced);
  const advancedParameters = parameters.filter(param => param.ui?.advanced);

  // 处理动态选项加载
  const loadDynamicOptions = useCallback(async (param: ParameterDefinition) => {
    if (!param.dynamic?.options) return;

    try {
      const options = await param.dynamic.options(values);
      setDynamicOptions(prev => ({
        ...prev,
        [param.id]: options
      }));
    } catch (error) {
      console.error(`Failed to load options for parameter ${param.id}:`, error);
      message.error(`Failed to load options for ${param.name}`);
    }
  }, [values]);

  // 参数变化处理
  const handleParameterChange = useCallback((paramId: string, newValue: any) => {
    const param = parameters.find(p => p.id === paramId);
    if (!param) return;

    // 验证参数
    let isValid = true;
    let error = undefined;

    try {
      // 类型验证
      switch (param.type) {
        case 'number':
          if (newValue !== undefined && isNaN(Number(newValue))) {
            isValid = false;
            error = 'Must be a number';
          }
          break;

        case 'boolean':
          if (newValue !== undefined && typeof newValue !== 'boolean') {
            isValid = false;
            error = 'Must be true or false';
          }
          break;
      }

      // 约束验证
      if (isValid && param.constraints) {
        const numValue = Number(newValue);

        if (param.constraints.min !== undefined && numValue < param.constraints.min) {
          isValid = false;
          error = `Must be >= ${param.constraints.min}`;
        }

        if (param.constraints.max !== undefined && numValue > param.constraints.max) {
          isValid = false;
          error = `Must be <= ${param.constraints.max}`;
        }

        if (param.constraints.pattern && typeof newValue === 'string') {
          if (!new RegExp(param.constraints.pattern).test(newValue)) {
            isValid = false;
            error = 'Invalid format';
          }
        }
      }

      // 自定义验证
      if (isValid && param.dynamic?.validation) {
        const validationResult = param.dynamic.validation(newValue, values);
        if (!validationResult.valid) {
          isValid = false;
          error = validationResult.message;
        }
      }

    } catch (err) {
      isValid = false;
      error = 'Validation error';
    }

    // 调用验证回调
    if (onValidate) {
      onValidate(paramId, isValid, error);
    }

    // 调用变化回调
    if (onChange) {
      onChange(paramId, newValue);
    }

    // 触发相关参数的动态选项重新加载
    parameters.forEach(p => {
      if (p.dynamic?.options && p.id !== paramId) {
        loadDynamicOptions(p);
      }
    });

  }, [parameters, values, onChange, onValidate, loadDynamicOptions]);

  // 初始化动态选项
  useEffect(() => {
    parameters.forEach(param => {
      if (param.dynamic?.options) {
        loadDynamicOptions(param);
      }
    });
  }, [parameters, loadDynamicOptions]);

  // 渲染参数组
  const renderParameterGroup = (groupName: string, params: ParameterDefinition[]) => {
    if (collapsible) {
      return (
        <Panel
          header={
            <Space>
              <span>{groupName}</span>
              <span style={{ color: '#666' }}>({params.length})</span>
            </Space>
          }
          key={groupName}
        >
          {renderParameters(params)}
        </Panel>
      );
    } else {
      return (
        <div key={groupName}>
          {groupName !== 'General' && (
            <>
              <Divider orientation="left">{groupName}</Divider>
              <div style={{ fontWeight: 'bold', marginBottom: 12 }}>{groupName}</div>
            </>
          )}
          {renderParameters(params)}
        </div>
      );
    }
  };

  // 渲染参数列表
  const renderParameters = (params: ParameterDefinition[]) => {
    return (
      <Form layout="vertical" style={{ width: '100%' }}>
        {params.map((parameter) => {
          const isVisible = parameter.dynamic?.visible
            ? parameter.dynamic.visible(values)
            : true;

          if (!isVisible) {
            return null;
          }

          return (
            <Form.Item
              key={parameter.id}
              label={
                <Space>
                  <span>
                    {parameter.name}
                    {parameter.required && <span style={{ color: '#ff4d4f' }}> *</span>}
                  </span>
                  {parameter.ui?.advanced && (
                    <Tooltip title="Advanced parameter">
                      <SettingOutlined style={{ color: '#999' }} />
                    </Tooltip>
                  )}
                </Space>
              }
              validateStatus={parameter.required && !values[parameter.id] ? 'error' : undefined}
              help={parameter.required && !values[parameter.id] ? 'This field is required' : undefined}
            >
              <ParameterInput
                parameter={parameter}
                value={values[parameter.id]}
                onChange={(newValue) => handleParameterChange(parameter.id, newValue)}
                dynamicOptions={dynamicOptions[parameter.id]}
                readOnly={readOnly}
              />
            </Form.Item>
          );
        })}
      </Form>
    );
  };

  if (collapsible && Object.keys(groupedParameters).length > 1) {
    return (
      <Collapse
        activeKey={activeKeys}
        onChange={setActiveKeys}
        ghost
        style={{ width: '100%' }}
      >
        {Object.entries(groupedParameters).map(([groupName, params]) =>
          renderParameterGroup(groupName, params)
        )}
      </Collapse>
    );
  }

  return (
    <div style={{ width: '100%' }}>
      {/* 基础参数 */}
      {basicParameters.length > 0 && renderParameters(basicParameters)}

      {/* 高级参数 */}
      {showAdvanced && advancedParameters.length > 0 && (
        <>
          <Divider orientation="left">
            <Space>
              Advanced Parameters
              <span style={{ color: '#666' }}>({advancedParameters.length})</span>
            </Space>
          </Divider>
          {renderParameters(advancedParameters)}
        </>
      )}

      {!collapsible && Object.keys(groupedParameters).length > 1 && (
        Object.entries(groupedParameters).map(([groupName, params]) =>
          groupName !== 'General' ? renderParameterGroup(groupName, params) : null
        )
      )}
    </div>
  );
};

export default ParameterEditor;