package v1

import "time"

// DeviceRegistrationRequest 设备注册请求
type DeviceRegistrationRequest struct {
	DeviceID   string            `json:"device_id" binding:"required"`
	DeviceName string            `json:"device_name" binding:"required"`
	DeviceType string            `json:"device_type" binding:"required"`
	Model      string            `json:"model,omitempty"`
	Version    string            `json:"version,omitempty"`
	Location   *DeviceLocation   `json:"location,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// DeviceLocation 设备位置
type DeviceLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude,omitempty"`
	Address   string  `json:"address,omitempty"`
	City      string  `json:"city,omitempty"`
	Province  string  `json:"province,omitempty"`
	Country   string  `json:"country,omitempty"`
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	ID            int64              `json:"id"`
	DeviceID      string             `json:"device_id"`
	DeviceName    string             `json:"device_name"`
	DeviceType    string             `json:"device_type"`
	Model         string             `json:"model"`
	Version       string             `json:"version"`
	Status        string             `json:"status"`        // online, offline, error, unknown
	Location      *DeviceLocation    `json:"location"`
	LastSeen      *time.Time         `json:"last_seen,omitempty"`
	Firmware      *FirmwareInfo      `json:"firmware,omitempty"`
	Configuration map[string]interface{} `json:"configuration"`
	Metadata      map[string]interface{} `json:"metadata"`
	IsActive      bool               `json:"is_active"`
	IsActivated   bool               `json:"is_activated"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// FirmwareInfo 固件信息
type FirmwareInfo struct {
	Version       string    `json:"version"`
	URL          string    `json:"url"`
	Checksum      string    `json:"checksum"`
	Size          int64     `json:"size"`
	ReleaseDate   time.Time `json:"release_date"`
	DownloadCount int       `json:"download_count"`
	Description   string    `json:"description"`
}

// DeviceActivationRequest 设备激活请求
type DeviceActivationRequest struct {
	ActivationCode string            `json:"activation_code" binding:"required"`
	DeviceName     string            `json:"device_name,omitempty"`
	Configuration  map[string]interface{} `json:"configuration,omitempty"`
}

// DeviceActivationResponse 设备激活响应
type DeviceActivationResponse struct {
	Success      bool       `json:"success"`
	DeviceToken   string     `json:"device_token"`
	AccessToken  string     `json:"access_token"`
	ExpiresIn     int64      `json:"expires_in"`
	Message       string     `json:"message"`
	DeviceInfo    DeviceInfo  `json:"device_info,omitempty"`
}

// DeviceStatusRequest 设备状态管理请求
type DeviceStatusRequest struct {
	DeviceID string `json:"device_id" binding:"required"` // 设备MAC地址
	IsActive *bool  `json:"is_active" binding:"required"` // 激活状态：true激活，false禁用
}

// DeviceStatusResponse 设备状态管理响应
type DeviceStatusResponse struct {
	Success   bool       `json:"success"`
	Message   string     `json:"message"`
	DeviceInfo DeviceInfo `json:"device_info"`
}

// DeviceUpdateRequest 设备更新请求
type DeviceUpdateRequest struct {
	DeviceName    string                `json:"device_name,omitempty"`
	Location     *DeviceLocation       `json:"location,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	IsActive     *bool                 `json:"is_active,omitempty"`
}

// DeviceQuery 设备查询参数
type DeviceQuery struct {
	Page       int      `form:"page,default=1"`
	Limit      int      `form:"limit,default=20"`
	Status     string   `form:"status"`
	DeviceType string   `form:"device_type"`
	Search     string   `form:"search"`
	SortBy     string   `form:"sort_by,default=created_at"`
	SortOrder  string   `form:"sort_order,default=desc"`
	Location  bool     `form:"location"`
}

// Pagination 分页信息
type Pagination struct {
	Page      int64 `json:"page"`
	Limit     int64 `json:"limit"`
	Total     int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
	HasNext   bool  `json:"has_next"`
	HasPrev   bool  `json:"has_prev"`
}

// DeviceListResponse 设备列表响应
type DeviceListResponse struct {
	Devices    []DeviceInfo `json:"devices"`
	Pagination Pagination  `json:"pagination"`
}


// WebSocketInfo WebSocket信息
type WebSocketInfo struct {
	URL        string    `json:"url"`
	Path       string    `json:"path"`
	Protocol   string    `json:"protocol"`
	Status     string    `json:"status"`
	Connected  bool      `json:"connected"`
	Clients    int       `json:"clients"`
	StartTime  time.Time `json:"start_time"`
}

// MQTTInfo MQTT信息
type MQTTInfo struct {
	Broker     string    `json:"broker"`
	Port       int       `json:"port"`
	Protocol   string    `json:"protocol"`
	Status     string    `json:"status"`
	Connected  bool      `json:"connected"`
	Clients    int       `json:"clients"`
	Uptime     int64     `json:"uptime"`
}

// ServerTimeInfo 服务器时间信息
type ServerTimeInfo struct {
	CurrentTime time.Time `json:"current_time"`
	Timezone    string    `json:"timezone"`
	Uptime      int64     `json:"uptime"`
	Load        float64   `json:"load"`
}

// Activation 激活信息
type Activation struct {
	Code        string    `json:"code"`
	DeviceID    string    `json:"device_id"`
	ActivatedAt time.Time `json:"activated_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	IsActive    bool      `json:"is_active"`
}

// OTARequest 统一OTA请求（包含设备注册和固件更新）
type OTARequest struct {
	// 设备注册相关
	DeviceID      string                 `json:"device_id" binding:"required"`
	DeviceName    string                 `json:"device_name" binding:"required"`
	DeviceType    string                 `json:"device_type" binding:"required"`
	Model         string                 `json:"model,omitempty"`
	Version       string                 `json:"version,omitempty"`
	Location      *DeviceLocation        `json:"location,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`

	// 固件更新相关
	Action          string `json:"action" binding:"required"` // register, update, activate
	FirmwareVersion string `json:"firmware_version,omitempty"`
	ForceUpdate     bool   `json:"force_update,omitempty"`
	ActivationCode  string `json:"activation_code,omitempty"`
}

// BatchCommandRequest 批量命令执行请求
type BatchCommandRequest struct {
	DeviceIDs []string          `json:"device_ids" binding:"required"`
	Command   string            `json:"command" binding:"required"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Timeout   int               `json:"timeout,omitempty"`
}

// BatchCommandResponse 批量命令执行响应
type BatchCommandResponse struct {
	BatchID   string        `json:"batch_id"`
	TotalCmds int           `json:"total_cmds"`
	Commands  []CommandTask `json:"commands"`
	CreatedAt time.Time     `json:"created_at"`
}

// CommandTask 命令任务
type CommandTask struct {
	DeviceID   string                 `json:"device_id"`
	CommandID  string                 `json:"command_id"`
	Command    string                 `json:"command"`
	Params     map[string]interface{} `json:"params"`
	Status     string                 `json:"status"`
	SentAt     time.Time              `json:"sent_at"`
}

// BatchConfigRequest 批量配置更新请求
type BatchConfigRequest struct {
	DeviceIDs     []string                 `json:"device_ids" binding:"required"`
	Configuration map[string]interface{}   `json:"configuration" binding:"required"`
	Merge         bool                     `json:"merge,omitempty"`
	Restart       bool                     `json:"restart,omitempty"`
}

// BatchConfigResponse 批量配置更新响应
type BatchConfigResponse struct {
	BatchID      string         `json:"batch_id"`
	TotalConfigs int            `json:"total_configs"`
	Configs      []ConfigTask   `json:"configs"`
	CreatedAt    time.Time      `json:"created_at"`
}

// ConfigTask 配置任务
type ConfigTask struct {
	DeviceID     string                 `json:"device_id"`
	ConfigID     string                 `json:"config_id"`
	Configuration map[string]interface{} `json:"configuration"`
	Merge        bool                   `json:"merge"`
	Restart      bool                   `json:"restart"`
	Status       string                 `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
}

// OTAResponse OTA统一响应
type OTAResponse struct {
	Success       bool           `json:"success"`
	Message       string         `json:"message"`
	Data          interface{}    `json:"data,omitempty"`
	ErrorCode     string         `json:"error_code,omitempty"`
	DeviceToken   string         `json:"device_token,omitempty"`
	WebSocketInfo *WebSocketInfo `json:"websocket_info,omitempty"`
	MQTTInfo      *MQTTInfo      `json:"mqtt_info,omitempty"`
	ServerTime    *ServerTimeInfo `json:"server_time,omitempty"`
	Activation    *Activation     `json:"activation,omitempty"`
}