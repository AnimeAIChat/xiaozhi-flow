package utils

// API错误码常量定义
const (
	// 通用错误码 (1000-1999)
	ErrorCodeValidationFailed     = "VALIDATION_FAILED"
	ErrorCodeInvalidInput         = "INVALID_INPUT"
	ErrorCodeAuthenticationFailed = "AUTHENTICATION_FAILED"
	ErrorCodeAuthorizationFailed  = "AUTHORIZATION_FAILED"
	ErrorCodeResourceNotFound     = "RESOURCE_NOT_FOUND"
	ErrorCodeInternalServer       = "INTERNAL_SERVER_ERROR"
	ErrorCodeBadRequest          = "BAD_REQUEST"
	ErrorCodeConflict            = "CONFLICT"
	ErrorCodeUnsupportedVersion   = "UNSUPPORTED_API_VERSION"
	ErrorCodeRateLimited         = "RATE_LIMITED"
	ErrorCodeRequestTooLarge     = "REQUEST_TOO_LARGE"
	ErrorCodeTimeout             = "TIMEOUT"

	// 认证相关错误码 (2000-2999)
	ErrorCodeInvalidToken       = "INVALID_TOKEN"
	ErrorCodeTokenExpired      = "TOKEN_EXPIRED"
	ErrorCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrorCodeAccountLocked     = "ACCOUNT_LOCKED"
	ErrorCodeAccountDisabled    = "ACCOUNT_DISABLED"
	ErrorCodeUserNotFound       = "USER_NOT_FOUND"
	ErrorCodeUserExists         = "USER_EXISTS"
	ErrorCodeEmailExists        = "EMAIL_EXISTS"
	ErrorCodeInvalidPassword    = "INVALID_PASSWORD"
	ErrorCodePasswordMismatch   = "PASSWORD_MISMATCH"

	// 工作流相关错误码 (3000-3999)
	ErrorCodeWorkflowNotFound     = "WORKFLOW_NOT_FOUND"
	ErrorCodeExecutionNotFound    = "EXECUTION_NOT_FOUND"
	ErrorCodeExecutionFailed      = "EXECUTION_FAILED"
	ErrorCodeExecutionRunning     = "EXECUTION_RUNNING"
	ErrorCodeExecutionCompleted   = "EXECUTION_COMPLETED"
	ErrorCodeExecutionCancelled   = "EXECUTION_CANCELLED"
	ErrorCodeInvalidWorkflowState = "INVALID_WORKFLOW_STATE"
	ErrorCodeNodeExecutionFailed  = "NODE_EXECUTION_FAILED"
	ErrorCodeWorkflowTimeout      = "WORKFLOW_TIMEOUT"

	// 设备相关错误码 (4000-4999)
	ErrorCodeDeviceNotFound     = "DEVICE_NOT_FOUND"
	ErrorCodeDeviceOffline      = "DEVICE_OFFLINE"
	ErrorCodeDeviceBusy         = "DEVICE_BUSY"
	ErrorCodeDeviceExists       = "DEVICE_EXISTS"
	ErrorCodeDeviceUpdating     = "DEVICE_UPDATING"
	ErrorCodeDeviceActivated    = "DEVICE_ACTIVATED"
	ErrorCodeFirmwareNotFound   = "FIRMWARE_NOT_FOUND"
	ErrorCodeFirmwareCorrupted  = "FIRMWARE_CORRUPTED"
	ErrorCodeUpdateFailed       = "UPDATE_FAILED"
	ErrorCodeOTACompleted       = "OTA_COMPLETED"
	ErrorCodeOTAFailed          = "OTA_FAILED"
	ErrorCodeActivationFailed   = "ACTIVATION_FAILED"
	ErrorCodeInvalidActivationCode = "INVALID_ACTIVATION_CODE"
	ErrorCodeInvalidDeviceId    = "INVALID_DEVICE_ID"

	// 配置相关错误码 (5000-5999)
	ErrorCodeConfigNotFound       = "CONFIG_NOT_FOUND"
	ErrorCodeConfigInvalid        = "CONFIG_INVALID"
	ErrorCodeConfigUpdateFailed   = "CONFIG_UPDATE_FAILED"
	ErrorCodeProviderExists       = "PROVIDER_EXISTS"
	ErrorCodeDatabaseConnection   = "DATABASE_CONNECTION_FAILED"
	ErrorCodeDatabaseQuery        = "DATABASE_QUERY_FAILED"
	ErrorCodeSchemaValidation     = "SCHEMA_VALIDATION_FAILED"

	// 视觉分析相关错误码 (6000-6999)
	ErrorCodeVisionProcessingFailed = "VISION_PROCESSING_FAILED"
	ErrorCodeImageTooLarge         = "IMAGE_TOO_LARGE"
	ErrorCodeInvalidImageFormat     = "INVALID_IMAGE_FORMAT"
	ErrorCodeImageUploadFailed     = "IMAGE_UPLOAD_FAILED"
	ErrorCodeAnalysisTimeout        = "ANALYSIS_TIMEOUT"
	ErrorCodeVisionServiceUnavailable = "VISION_SERVICE_UNAVAILABLE"

	// 系统相关错误码 (7000-7999)
	ErrorCodeSystemNotInitialized   = "SYSTEM_NOT_INITIALIZED"
	ErrorCodeSystemMaintenance      = "SYSTEM_MAINTENANCE"
	ErrorCodeServiceUnavailable     = "SERVICE_UNAVAILABLE"
	ErrorCodeDependencyFailed       = "DEPENDENCY_FAILED"
	ErrorCodeQuotaExceeded         = "QUOTA_EXCEEDED"
	ErrorCodeInsufficientResources  = "INSUFFICIENT_RESOURCES"
)

// ErrorMessages 错误消息映射
var ErrorMessages = map[string]string{
	// 通用错误
	ErrorCodeValidationFailed:     "请求参数验证失败",
	ErrorCodeInvalidInput:         "输入参数无效",
	ErrorCodeAuthenticationFailed: "身份验证失败",
	ErrorCodeAuthorizationFailed:  "权限不足",
	ErrorCodeResourceNotFound:     "资源不存在",
	ErrorCodeInternalServer:       "内部服务器错误",
	ErrorCodeBadRequest:          "请求格式错误",
	ErrorCodeConflict:            "资源冲突",
	ErrorCodeUnsupportedVersion:   "不支持的API版本",
	ErrorCodeRateLimited:         "请求频率过高",
	ErrorCodeRequestTooLarge:     "请求体过大",
	ErrorCodeTimeout:             "请求超时",

	// 认证错误
	ErrorCodeInvalidToken:       "无效的访问令牌",
	ErrorCodeTokenExpired:      "访问令牌已过期",
	ErrorCodeInvalidCredentials: "用户名或密码错误",
	ErrorCodeAccountLocked:     "账户已被锁定",
	ErrorCodeAccountDisabled:    "账户已被禁用",
	ErrorCodeUserNotFound:       "用户不存在",
	ErrorCodeUserExists:         "用户已存在",
	ErrorCodeEmailExists:        "邮箱已被注册",
	ErrorCodeInvalidPassword:    "密码格式不正确",
	ErrorCodePasswordMismatch:   "密码确认不匹配",

	// 工作流错误
	ErrorCodeWorkflowNotFound:     "工作流不存在",
	ErrorCodeExecutionNotFound:    "执行记录不存在",
	ErrorCodeExecutionFailed:      "工作流执行失败",
	ErrorCodeExecutionRunning:     "工作流正在执行中",
	ErrorCodeExecutionCompleted:   "工作流已完成",
	ErrorCodeExecutionCancelled:   "工作流已取消",
	ErrorCodeInvalidWorkflowState: "工作流状态无效",
	ErrorCodeNodeExecutionFailed:  "节点执行失败",
	ErrorCodeWorkflowTimeout:      "工作流执行超时",

	// 设备错误
	ErrorCodeDeviceNotFound:       "设备不存在",
	ErrorCodeDeviceOffline:        "设备离线",
	ErrorCodeDeviceBusy:           "设备忙碌",
	ErrorCodeDeviceExists:         "设备已存在",
	ErrorCodeDeviceUpdating:       "设备正在更新中，无法删除",
	ErrorCodeDeviceActivated:      "设备已激活",
	ErrorCodeFirmwareNotFound:     "固件文件不存在",
	ErrorCodeFirmwareCorrupted:    "固件文件损坏",
	ErrorCodeUpdateFailed:         "固件更新失败",
	ErrorCodeOTACompleted:         "OTA更新已完成，无法取消",
	ErrorCodeOTAFailed:            "OTA更新已失败",
	ErrorCodeActivationFailed:     "设备激活失败",
	ErrorCodeInvalidActivationCode: "无效的激活码",
	ErrorCodeInvalidDeviceId:      "无效的设备ID",

	// 配置错误
	ErrorCodeConfigNotFound:     "配置不存在",
	ErrorCodeConfigInvalid:      "配置格式无效",
	ErrorCodeConfigUpdateFailed: "配置更新失败",
	ErrorCodeProviderExists:     "供应商已存在",
	ErrorCodeDatabaseConnection: "数据库连接失败",
	ErrorCodeDatabaseQuery:      "数据库查询失败",
	ErrorCodeSchemaValidation:   "数据模式验证失败",

	// 视觉分析错误
	ErrorCodeVisionProcessingFailed: "视觉分析处理失败",
	ErrorCodeImageTooLarge:         "图片文件过大",
	ErrorCodeInvalidImageFormat:     "不支持的图片格式",
	ErrorCodeImageUploadFailed:     "图片上传失败",
	ErrorCodeAnalysisTimeout:        "分析处理超时",
	ErrorCodeVisionServiceUnavailable: "视觉分析服务不可用",

	// 系统错误
	ErrorCodeSystemNotInitialized:  "系统未初始化",
	ErrorCodeSystemMaintenance:     "系统维护中",
	ErrorCodeServiceUnavailable:    "服务不可用",
	ErrorCodeDependencyFailed:      "依赖服务失败",
	ErrorCodeQuotaExceeded:        "配额已超出",
	ErrorCodeInsufficientResources: "资源不足",
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(errorCode string) string {
	if message, exists := ErrorMessages[errorCode]; exists {
		return message
	}
	return "未知错误"
}