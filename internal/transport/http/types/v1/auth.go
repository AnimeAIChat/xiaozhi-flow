package v1

import "time"

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone,omitempty"`
}

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int64     `json:"expires_in"`
	User        UserInfo  `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       int64         `json:"id"`
	Username string         `json:"username"`
	Email    string         `json:"email"`
	Phone    string         `json:"phone,omitempty"`
	Avatar   string         `json:"avatar,omitempty"`
	Role     string         `json:"role"`
	Status   string         `json:"status"` // active, inactive, suspended
	LastLogin *time.Time    `json:"last_login,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// UserProfile 用户档案
type UserProfile struct {
	ID       int64                 `json:"id"`
	Username string                 `json:"username"`
	Email    string                 `json:"email"`
	Phone    string                 `json:"phone,omitempty"`
	Avatar   string                 `json:"avatar,omitempty"`
	Bio      string                 `json:"bio,omitempty"`
	Settings UserSettings          `json:"settings"`
	Status   string                 `json:"status"`
	LastLogin *time.Time            `json:"last_login,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// UserSettings 用户设置
type UserSettings struct {
	Language    string `json:"language,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
	Theme       string `json:"theme,omitempty"`
	Notifications bool   `json:"notifications"`
}

// UpdateProfileRequest 更新档案请求
type UpdateProfileRequest struct {
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Bio      string `json:"bio,omitempty"`
	Settings *UserSettings `json:"settings,omitempty"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword  string `json:"confirm_password" binding:"required"`
}

// UserListQuery 用户列表查询参数
type UserListQuery struct {
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=20"`
	Role       string `form:"role"`
	Status     string `form:"status"`
	Search     string `form:"search"`
	SortBy     string `form:"sort_by,default=created_at"`
	SortOrder  string `form:"sort_order,default=desc"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users      []UserProfile `json:"users"`
	Pagination Pagination  `json:"pagination"`
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

// SessionSession 会话信息
type SessionSession struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	IPAddress string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`
}

// SessionListQuery 会话列表查询
type SessionListQuery struct {
	UserID   int64  `form:"user_id"`
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
	Active   *bool  `form:"active"`
}

// SessionListResponse 会话列表响应
type SessionListResponse struct {
	Sessions  []SessionSession `json:"sessions"`
	Pagination Pagination        `json:"pagination"`
}