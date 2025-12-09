package v1

import (
	"xiaozhi-server-go/internal/platform/logging"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/transport/http/types/v1"
	httpUtils "xiaozhi-server-go/internal/transport/http/utils"
)

// AuthServiceV1 V1版本认证服务
type AuthServiceV1 struct {
	logger *logging.Logger
	config *config.Config
	// TODO: 添加实际的业务逻辑依赖
}

// NewAuthServiceV1 创建认证服务V1实例
func NewAuthServiceV1(config *config.Config, logger *logging.Logger) (*AuthServiceV1, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &AuthServiceV1{
		logger: logger,
		config: config,
	}, nil
}

// Register 注册认证API路由
func (s *AuthServiceV1) Register(router *gin.RouterGroup) {
	// 认证相关路由（公开）
	auth := router.Group("/auth")
	{
		auth.POST("/login", s.login)          // 用户登录
		auth.POST("/register", s.register)    // 用户注册
		auth.POST("/refresh", s.refreshToken) // 刷新Token
	}
}

// RegisterSecure 注册需要认证的API路由
func (s *AuthServiceV1) RegisterSecure(router *gin.RouterGroup) {
	// 用户管理路由（需要认证）
	users := router.Group("/users")
	{
		users.GET("", s.listUsers)                    // 获取用户列表
		users.GET("/profile", s.getUserProfile)       // 获取用户档案
		users.PUT("/profile", s.updateProfile)        // 更新用户档案
		users.POST("/change-password", s.changePassword) // 修改密码
		users.GET("/:id", s.getUser)                  // 获取特定用户信息
	}

	// 会话管理路由（需要认证）
	sessions := router.Group("/sessions")
	{
		sessions.GET("", s.listSessions)      // 获取会话列表
		sessions.DELETE("/:id", s.revokeSession) // 撤销会话
	}
}

// login 用户登录
// @Summary 用户登录
// @Description 用户登录验证，返回访问令牌和用户信息
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body v1.LoginRequest true "登录信息"
// @Success 200 {object} httptransport.APIResponse{data=v1.LoginResponse}
// @Failure 400 {object} httptransport.APIResponse
// @Failure 401 {object} httptransport.APIResponse
// @Router /v1/auth/login [post]
func (s *AuthServiceV1) login(c *gin.Context) {
	var request v1.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "用户登录",
		"username", request.Username,
		"request_id", getRequestID(c),
	)

	// 模拟用户验证
	user := s.getMockUser(request.Username, request.Password)
	if user == nil {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInvalidCredentials, "用户名或密码错误")
		return
	}

	// 生成Token（模拟）
	accessToken := fmt.Sprintf("access_token_%d", time.Now().UnixNano())
	refreshToken := fmt.Sprintf("refresh_token_%d", time.Now().UnixNano())
	expiresIn := int64(3600) // 1小时

	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now

	response := v1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        *user,
	}

	httpUtils.Response.Success(c, response, "登录成功")
}

// register 用户注册
// @Summary 用户注册
// @Description 新用户注册
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body v1.RegisterRequest true "注册信息"
// @Success 201 {object} httptransport.APIResponse{data=v1.UserInfo}
// @Failure 400 {object} httptransport.APIResponse
// @Failure 409 {object} httptransport.APIResponse
// @Router /v1/auth/register [post]
func (s *AuthServiceV1) register(c *gin.Context) {
	var request v1.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "用户注册",
		"username", request.Username,
		"email", request.Email,
		"request_id", getRequestID(c),
	)

	// 检查用户是否已存在
	if s.getMockUser(request.Username, "") != nil {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeUserExists, "用户名已存在")
		return
	}

	if s.getMockUserByEmail(request.Email) != nil {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeEmailExists, "邮箱已被注册")
		return
	}

	// 创建新用户（模拟）
	user := v1.UserInfo{
		ID:        time.Now().UnixNano(),
		Username:  request.Username,
		Email:     request.Email,
		Phone:     request.Phone,
		Role:      "user",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	httpUtils.Response.Created(c, user, "注册成功")
}

// refreshToken 刷新Token
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body v1.RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} httptransport.APIResponse{data=v1.LoginResponse}
// @Failure 400 {object} httptransport.APIResponse
// @Failure 401 {object} httptransport.APIResponse
// @Router /v1/auth/refresh [post]
func (s *AuthServiceV1) refreshToken(c *gin.Context) {
	var request v1.RefreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "刷新Token",
		"refresh_token", request.RefreshToken[:min(len(request.RefreshToken), 20)]+"...",
		"request_id", getRequestID(c),
	)

	// 模拟刷新令牌验证
	if !s.validateMockRefreshToken(request.RefreshToken) {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInvalidToken, "刷新令牌无效或已过期")
		return
	}

	// 生成新的Token
	accessToken := fmt.Sprintf("access_token_%d", time.Now().UnixNano())
	refreshToken := fmt.Sprintf("refresh_token_%d", time.Now().UnixNano())
	expiresIn := int64(3600) // 1小时

	// 模拟用户信息
	user := v1.UserInfo{
		ID:        1,
		Username:  "demo_user",
		Email:     "demo@example.com",
		Role:      "user",
		Status:    "active",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	response := v1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        user,
	}

	httpUtils.Response.Success(c, response, "Token刷新成功")
}

// listUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表，支持分页和过滤
// @Tags Users
// @Produce json
// @Param role query string false "按角色过滤"
// @Param status query string false "按状态过滤"
// @Param search query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param sort_by query string false "排序字段" default(created_at)
// @Param sort_order query string false "排序方向" default(desc)
// @Success 200 {object} httptransport.APIResponse{data=v1.UserListResponse}
// @Router /v1/users [get]
func (s *AuthServiceV1) listUsers(c *gin.Context) {
	var query v1.UserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "获取用户列表",
		"role", query.Role,
		"status", query.Status,
		"search", query.Search,
		"page", query.Page,
		"limit", query.Limit,
		"request_id", getRequestID(c),
	)

	// 模拟获取用户列表
	users, pagination := s.getMockUserList(query)

	response := v1.UserListResponse{
		Users:      users,
		Pagination: pagination,
	}

	httpUtils.Response.Success(c, response, "获取用户列表成功")
}

// getUser 获取特定用户信息
// @Summary 获取用户信息
// @Description 根据ID获取特定用户的详细信息
// @Tags Users
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} httptransport.APIResponse{data=v1.UserInfo}
// @Failure 404 {object} httptransport.APIResponse
// @Router /v1/users/{id} [get]
func (s *AuthServiceV1) getUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		httpUtils.Response.BadRequest(c, "用户ID不能为空")
		return
	}

	s.logger.InfoTag("API", "获取用户信息",
		"user_id", userID,
		"request_id", getRequestID(c),
	)

	// 模拟获取用户信息
	user := s.getMockUserByID(userID)
	if user == nil {
		httpUtils.Response.NotFound(c, "用户")
		return
	}

	httpUtils.Response.Success(c, user, "获取用户信息成功")
}

// getUserProfile 获取用户档案
// @Summary 获取用户档案
// @Description 获取当前登录用户的详细档案信息
// @Tags Users
// @Produce json
// @Success 200 {object} httptransport.APIResponse{data=v1.UserProfile}
// @Failure 401 {object} httptransport.APIResponse
// @Router /v1/users/profile [get]
func (s *AuthServiceV1) getUserProfile(c *gin.Context) {
	s.logger.InfoTag("API", "获取用户档案",
		"request_id", getRequestID(c),
	)

	// 模拟获取当前用户档案
	// 实际实现中应该从JWT Token中获取用户ID
	profile := s.getMockUserProfile(1) // 模拟用户ID为1

	httpUtils.Response.Success(c, profile, "获取用户档案成功")
}

// updateProfile 更新用户档案
// @Summary 更新用户档案
// @Description 更新当前登录用户的档案信息
// @Tags Users
// @Accept json
// @Produce json
// @Param request body v1.UpdateProfileRequest true "档案更新信息"
// @Success 200 {object} httptransport.APIResponse{data=v1.UserProfile}
// @Failure 400 {object} httptransport.APIResponse
// @Failure 401 {object} httptransport.APIResponse
// @Router /v1/users/profile [put]
func (s *AuthServiceV1) updateProfile(c *gin.Context) {
	var request v1.UpdateProfileRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "更新用户档案",
		"email", request.Email,
		"phone", request.Phone,
		"request_id", getRequestID(c),
	)

	// 获取当前用户档案并更新
	profile := s.getMockUserProfile(1) // 模拟用户ID为1
	if profile.Email != request.Email && request.Email != "" {
		// 检查邮箱是否已被使用
		if s.getMockUserByEmail(request.Email) != nil {
			httpUtils.Response.Error(c, httpUtils.ErrorCodeEmailExists, "邮箱已被使用")
			return
		}
		profile.Email = request.Email
	}
	if request.Phone != "" {
		profile.Phone = request.Phone
	}
	if request.Avatar != "" {
		profile.Avatar = request.Avatar
	}
	if request.Bio != "" {
		profile.Bio = request.Bio
	}
	if request.Settings != nil {
		profile.Settings = *request.Settings
	}

	profile.UpdatedAt = time.Now()

	httpUtils.Response.Success(c, profile, "档案更新成功")
}

// changePassword 修改密码
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags Users
// @Accept json
// @Produce json
// @Param request body v1.ChangePasswordRequest true "密码修改信息"
// @Success 200 {object} httptransport.APIResponse
// @Failure 400 {object} httptransport.APIResponse
// @Failure 401 {object} httptransport.APIResponse
// @Router /v1/users/change-password [post]
func (s *AuthServiceV1) changePassword(c *gin.Context) {
	var request v1.ChangePasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "修改密码",
		"request_id", getRequestID(c),
	)

	// 验证确认密码
	if request.NewPassword != request.ConfirmPassword {
		httpUtils.Response.Error(c, httpUtils.ErrorCodePasswordMismatch, "新密码与确认密码不一致")
		return
	}

	// 模拟验证当前密码
	currentUser := s.getMockUserProfile(1) // 模拟用户ID为1
	if !s.validateMockPassword(currentUser.Username, request.CurrentPassword) {
		httpUtils.Response.Error(c, httpUtils.ErrorCodeInvalidCredentials, "当前密码错误")
		return
	}

	// 模拟密码更新
	httpUtils.Response.Success(c, nil, "密码修改成功")
}

// listSessions 获取会话列表
// @Summary 获取会话列表
// @Description 获取当前用户的会话列表
// @Tags Sessions
// @Produce json
// @Param user_id query int64 false "用户ID"
// @Param active query bool false "是否只返回活跃会话"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} httptransport.APIResponse{data=v1.SessionListResponse}
// @Router /v1/sessions [get]
func (s *AuthServiceV1) listSessions(c *gin.Context) {
	var query v1.SessionListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		httpUtils.Response.ValidationError(c, err)
		return
	}

	s.logger.InfoTag("API", "获取会话列表",
		"user_id", query.UserID,
		"active", query.Active,
		"page", query.Page,
		"limit", query.Limit,
		"request_id", getRequestID(c),
	)

	// 模拟获取会话列表
	sessions, pagination := s.getMockSessionList(query)

	response := v1.SessionListResponse{
		Sessions:  sessions,
		Pagination: pagination,
	}

	httpUtils.Response.Success(c, response, "获取会话列表成功")
}

// revokeSession 撤销会话
// @Summary 撤销会话
// @Description 撤销指定的用户会话
// @Tags Sessions
// @Produce json
// @Param id path string true "会话ID"
// @Success 200 {object} httptransport.APIResponse
// @Failure 404 {object} httptransport.APIResponse
// @Router /v1/sessions/{id} [delete]
func (s *AuthServiceV1) revokeSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		httpUtils.Response.BadRequest(c, "会话ID不能为空")
		return
	}

	s.logger.InfoTag("API", "撤销会话",
		"session_id", sessionID,
		"request_id", getRequestID(c),
	)

	// 模拟撤销会话
	if !s.sessionExists(sessionID) {
		httpUtils.Response.NotFound(c, "会话")
		return
	}

	httpUtils.Response.Success(c, gin.H{"session_id": sessionID}, "会话撤销成功")
}

// ========== 模拟数据方法 ==========
// TODO: 实际实现中应该从数据库中获取真实数据

func (s *AuthServiceV1) getMockUser(username, password string) *v1.UserInfo {
	// 模拟用户数据库
	users := map[string]v1.UserInfo{
		"admin": {
			ID:       1,
			Username: "admin",
			Email:    "admin@example.com",
			Role:     "admin",
			Status:   "active",
		},
		"user": {
			ID:       2,
			Username: "user",
			Email:    "user@example.com",
			Role:     "user",
			Status:   "active",
		},
	}

	if user, exists := users[username]; exists {
		// 简单的密码验证（实际应该使用bcrypt等）
		if password == "" || password == "password" {
			return &user
		}
	}
	return nil
}

func (s *AuthServiceV1) getMockUserByEmail(email string) *v1.UserInfo {
	// 简单模拟邮箱查找
	if email == "admin@example.com" || email == "user@example.com" {
		return &v1.UserInfo{
			ID:       1,
			Username: "existing_user",
			Email:    email,
			Role:     "user",
			Status:   "active",
		}
	}
	return nil
}

func (s *AuthServiceV1) getMockUserByID(userID string) *v1.UserInfo {
	// 简单模拟ID查找
	if userID == "1" || userID == "2" {
		return &v1.UserInfo{
			ID:       1,
			Username: "user_" + userID,
			Email:    "user" + userID + "@example.com",
			Role:     "user",
			Status:   "active",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		}
	}
	return nil
}

func (s *AuthServiceV1) getMockUserProfile(userID int64) *v1.UserProfile {
	return &v1.UserProfile{
		ID:       userID,
		Username: "demo_user",
		Email:    "demo@example.com",
		Phone:    "+86 138 0000 0000",
		Avatar:   "https://example.com/avatar.jpg",
		Bio:      "这是我的个人简介",
		Settings: v1.UserSettings{
			Language:     "zh-CN",
			Timezone:     "Asia/Shanghai",
			Theme:        "light",
			Notifications: true,
		},
		Status:   "active",
		CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt: time.Now(),
	}
}

func (s *AuthServiceV1) getMockUserList(query v1.UserListQuery) ([]v1.UserProfile, v1.Pagination) {
	// 模拟用户列表数据
	users := []v1.UserProfile{
		{
			ID:       1,
			Username: "admin",
			Email:    "admin@example.com",
			Status:   "active",
			Settings: v1.UserSettings{Language: "zh-CN", Notifications: true},
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt: time.Now(),
		},
		{
			ID:       2,
			Username: "user1",
			Email:    "user1@example.com",
			Status:   "active",
			Settings: v1.UserSettings{Language: "zh-CN", Notifications: true},
			CreatedAt: time.Now().Add(-20 * 24 * time.Hour),
			UpdatedAt: time.Now(),
		},
		{
			ID:       3,
			Username: "user2",
			Email:    "user2@example.com",
			Status:   "inactive",
			Settings: v1.UserSettings{Language: "zh-CN", Notifications: false},
			CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
			UpdatedAt: time.Now(),
		},
	}

	// 简单的过滤逻辑
	var filtered []v1.UserProfile
	for _, user := range users {
		if query.Status != "" && user.Status != query.Status {
			continue
		}
		if query.Search != "" {
			// 简单搜索逻辑
			match := false
			for _, field := range []string{user.Username, user.Email} {
				if len(field) >= len(query.Search) {
					if field[:len(query.Search)] == query.Search {
						match = true
						break
					}
				}
			}
			if !match {
				continue
			}
		}
		filtered = append(filtered, user)
	}

	// 分页逻辑
	total := int64(len(filtered))
	totalPages := (total + int64(query.Limit) - 1) / int64(query.Limit)
	start := (query.Page - 1) * query.Limit
	end := start + query.Limit
	if end > len(filtered) {
		end = len(filtered)
	}
	if start >= len(filtered) {
		return []v1.UserProfile{}, v1.Pagination{
			Page:       int64(query.Page),
			Limit:      int64(query.Limit),
			Total:      total,
			TotalPages: totalPages,
			HasNext:    false,
			HasPrev:    query.Page > 1,
		}
	}

	pagedUsers := filtered[start:end]
	pagination := v1.Pagination{
		Page:       int64(query.Page),
		Limit:      int64(query.Limit),
		Total:      total,
		TotalPages: totalPages,
		HasNext:    int64(query.Page) < totalPages,
		HasPrev:    query.Page > 1,
	}

	return pagedUsers, pagination
}

func (s *AuthServiceV1) getMockSessionList(query v1.SessionListQuery) ([]v1.SessionSession, v1.Pagination) {
	// 模拟会话数据
	sessions := []v1.SessionSession{
		{
			ID:        "session_1",
			UserID:    1,
			Token:     "token_1",
			IPAddress: "192.168.1.100",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now().Add(-30 * time.Minute),
			LastSeen:  time.Now().Add(-5 * time.Minute),
		},
		{
			ID:        "session_2",
			UserID:    1,
			Token:     "token_2",
			IPAddress: "192.168.1.101",
			UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			ExpiresAt: time.Now().Add(2 * time.Hour),
			CreatedAt: time.Now().Add(-2 * time.Hour),
			LastSeen:  time.Now().Add(-10 * time.Minute),
		},
	}

	// 简单过滤逻辑
	var filtered []v1.SessionSession
	for _, session := range sessions {
		if query.UserID != 0 && session.UserID != query.UserID {
			continue
		}
		if query.Active != nil {
			isActive := time.Now().Before(session.ExpiresAt)
			if *query.Active != isActive {
				continue
			}
		}
		filtered = append(filtered, session)
	}

	// 分页逻辑
	total := int64(len(filtered))
	totalPages := (total + int64(query.Limit) - 1) / int64(query.Limit)
	start := (query.Page - 1) * query.Limit
	end := start + query.Limit
	if end > len(filtered) {
		end = len(filtered)
	}
	if start >= len(filtered) {
		return []v1.SessionSession{}, v1.Pagination{
			Page:       int64(query.Page),
			Limit:      int64(query.Limit),
			Total:      total,
			TotalPages: totalPages,
			HasNext:    false,
			HasPrev:    query.Page > 1,
		}
	}

	pagedSessions := filtered[start:end]
	pagination := v1.Pagination{
		Page:       int64(query.Page),
		Limit:      int64(query.Limit),
		Total:      total,
		TotalPages: totalPages,
		HasNext:    int64(query.Page) < totalPages,
		HasPrev:    query.Page > 1,
	}

	return pagedSessions, pagination
}

func (s *AuthServiceV1) validateMockRefreshToken(refreshToken string) bool {
	// 简单的刷新令牌验证逻辑
	return len(refreshToken) > 10 && refreshToken[:13] == "refresh_token_"
}

func (s *AuthServiceV1) validateMockPassword(username, password string) bool {
	// 简单的密码验证逻辑
	return password == "password"
}

func (s *AuthServiceV1) sessionExists(sessionID string) bool {
	// 简单的会话存在性检查
	return sessionID == "session_1" || sessionID == "session_2"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getRequestID 从上下文中获取请求ID
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

