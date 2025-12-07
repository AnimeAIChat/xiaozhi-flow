package webapi

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"xiaozhi-server-go/internal/domain/auth"
	"xiaozhi-server-go/internal/domain/auth/store"
	"xiaozhi-server-go/internal/platform/config"
	"xiaozhi-server-go/internal/platform/storage"
	"xiaozhi-server-go/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication related HTTP requests
type AuthHandler struct {
	logger *utils.Logger
	config *config.Config
	authManager *auth.Manager
	tokenManager *auth.AuthToken
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(logger *utils.Logger, config *config.Config) (*AuthHandler, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	// Initialize auth manager with memory store for simplicity
	// In production, this should use SQLite or Redis store
	authStoreConfig := store.Config{
		TTL: 7 * 24 * time.Hour,
	}
	authStore := store.NewMemory(authStoreConfig)

	authManager, err := auth.NewManager(auth.Options{
		Store:           authStore,
		Logger:          logger,
		SessionTTL:      7 * 24 * time.Hour, // 7 days as per requirements
		CleanupInterval: 1 * time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Initialize token manager
	tokenManager := auth.NewAuthToken(config.Server.Token)

	return &AuthHandler{
		logger:       logger,
		config:       config,
		authManager:  authManager,
		tokenManager: tokenManager,
	}, nil
}

// Close releases resources
func (h *AuthHandler) Close() error {
	if h.authManager != nil {
		return h.authManager.Close()
	}
	return nil
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token     string      `json:"token"`
	ExpiresAt int64       `json:"expires_at"`
	User      *UserInfo   `json:"user"`
}

// UserInfo represents user information
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

// handleLogin handles user login
// @Summary 用户登录
// @Description 使用用户名和密码进行身份验证
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录凭据"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} object
// @Failure 401 {object} object
// @Router /auth/login [post]
func (s *Service) handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate input
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		s.respondError(c, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Get database connection
	db := storage.GetDB()
	if db == nil {
		s.respondError(c, http.StatusInternalServerError, "Database not available")
		return
	}

	// Find user by username
	var user storage.User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		s.logger.InfoTag("Auth", "Login attempt for non-existent user: %s", req.Username)
		s.respondError(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Check if user is active
	if user.Status != 1 {
		s.respondError(c, http.StatusUnauthorized, "User account is disabled")
		return
	}

	// Verify password - handle both plain text and bcrypt hashed passwords
	var passwordValid bool
	// First try plain text comparison (for users created during system setup)
	if user.Password == req.Password {
		passwordValid = true
		s.logger.InfoTag("Auth", "Using plain text password comparison for user: %s", req.Username)
	} else {
		// If plain text fails, try bcrypt comparison (for newly registered users)
		if bcryptErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); bcryptErr == nil {
			passwordValid = true
			s.logger.InfoTag("Auth", "Using bcrypt password comparison for user: %s", req.Username)
		} else {
			passwordValid = false
			s.logger.InfoTag("Auth", "Both plain text and bcrypt comparison failed for user: %s", req.Username)
		}
	}

	if !passwordValid {
		s.logger.InfoTag("Auth", "Invalid password for user: %s", req.Username)
		s.respondError(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Generate JWT token
	tokenManager := auth.NewAuthToken(s.config.Server.Token).WithTTL(7 * 24 * time.Hour)
	clientID := fmt.Sprintf("web_%d_%d", user.ID, time.Now().Unix())

	token, err := tokenManager.GenerateToken(clientID)
	if err != nil {
		s.logger.ErrorTag("Auth", "Failed to generate token for user %s: %v", req.Username, err)
		s.respondError(c, http.StatusInternalServerError, "Failed to generate authentication token")
		return
	}

	// Store client session using auth manager
	authHandler, err := NewAuthHandler(s.logger, s.config)
	if err == nil && authHandler.authManager != nil {
		clientInfo := auth.ClientInfo{
			ClientID: clientID,
			Username: user.Username,
			Password: "", // Don't store password in session
			IP:        c.ClientIP(),
			DeviceID:  "web",
			Metadata: map[string]any{
				"user_id":   user.ID,
				"user_role": user.Role,
				"user_email": user.Email,
			},
		}

		// Store session in background, don't fail login if this fails
		if err := authHandler.authManager.RegisterClient(c.Request.Context(), clientInfo); err != nil {
			s.logger.WarnTag("Auth", "Failed to store client session for user %s: %v", req.Username, err)
		}

		authHandler.Close()
	}

	// Prepare response
	expiresAt := time.Now().Add(7 * 24 * time.Hour).Unix()
	response := AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Nickname: user.Nickname,
			Role:     user.Role,
		},
	}

	s.logger.InfoTag("Auth", "User logged in successfully: %s", req.Username)
	s.respondSuccess(c, http.StatusOK, response, "Login successful")
}

// handleRegister handles user registration
// @Summary 用户注册
// @Description 注册新用户账户
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} object
// @Failure 409 {object} object
// @Router /auth/register [post]
func (s *Service) handleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.respondError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate input
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		s.respondError(c, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	if len(req.Password) < 6 {
		s.respondError(c, http.StatusBadRequest, "Password must be at least 6 characters long")
		return
	}

	// Get database connection
	db := storage.GetDB()
	if db == nil {
		s.respondError(c, http.StatusInternalServerError, "Database not available")
		return
	}

	// Check if username already exists
	var existingUser storage.User
	if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		s.respondError(c, http.StatusConflict, "Username already exists")
		return
	}

	// Check if email already exists
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		s.respondError(c, http.StatusConflict, "Email already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.ErrorTag("Auth", "Failed to hash password for user %s: %v", req.Username, err)
		s.respondError(c, http.StatusInternalServerError, "Failed to process registration")
		return
	}

	// Create user
	user := storage.User{
		Username:  req.Username,
		Password:  string(hashedPassword),
		Email:     req.Email,
		Nickname:  req.Username, // Default nickname to username
		Role:      "user",       // Default role
		Status:    1,            // Active
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&user).Error; err != nil {
		s.logger.ErrorTag("Auth", "Failed to create user %s: %v", req.Username, err)
		s.respondError(c, http.StatusInternalServerError, "Failed to create user account")
		return
	}

	// Auto-login after registration
	tokenManager := auth.NewAuthToken(s.config.Server.Token).WithTTL(7 * 24 * time.Hour)
	clientID := fmt.Sprintf("web_%d_%d", user.ID, time.Now().Unix())

	token, err := tokenManager.GenerateToken(clientID)
	if err != nil {
		s.logger.ErrorTag("Auth", "Failed to generate token for new user %s: %v", req.Username, err)
		// Registration succeeded, but token generation failed
		s.respondSuccess(c, http.StatusCreated, gin.H{
			"message": "User registered successfully, but auto-login failed",
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"nickname": user.Nickname,
				"role":     user.Role,
			},
		}, "Registration successful")
		return
	}

	// Store client session
	authHandler, err := NewAuthHandler(s.logger, s.config)
	if err == nil && authHandler.authManager != nil {
		clientInfo := auth.ClientInfo{
			ClientID: clientID,
			Username: user.Username,
			Password: "",
			IP:        c.ClientIP(),
			DeviceID:  "web",
			Metadata: map[string]any{
				"user_id":   user.ID,
				"user_role": user.Role,
				"user_email": user.Email,
			},
		}

		if err := authHandler.authManager.RegisterClient(c.Request.Context(), clientInfo); err != nil {
			s.logger.WarnTag("Auth", "Failed to store client session for new user %s: %v", req.Username, err)
		}

		authHandler.Close()
	}

	// Prepare response
	expiresAt := time.Now().Add(7 * 24 * time.Hour).Unix()
	response := AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Nickname: user.Nickname,
			Role:     user.Role,
		},
	}

	s.logger.InfoTag("Auth", "User registered successfully: %s", req.Username)
	s.respondSuccess(c, http.StatusCreated, response, "Registration successful")
}

// handleMe handles getting current user information
// @Summary 获取当前用户信息
// @Description 获取当前已验证用户的信息
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserInfo
// @Failure 401 {object} object
// @Router /auth/me [get]
func (s *Service) handleMe(c *gin.Context) {
	// Extract user info from context (set by authMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		s.respondError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	username, exists := c.Get("username")
	if !exists {
		s.respondError(c, http.StatusUnauthorized, "User information not available")
		return
	}

	userRole, exists := c.Get("user_role")
	if !exists {
		userRole = "user"
	}

	userEmail, exists := c.Get("user_email")
	if !exists {
		userEmail = ""
	}

	// Convert userID to uint if it's stored as uint
	var userIDUint uint
	switch v := userID.(type) {
	case uint:
		userIDUint = v
	case int:
		userIDUint = uint(v)
	case string:
		if id, err := strconv.ParseUint(v, 10, 32); err == nil {
			userIDUint = uint(id)
		}
	default:
		s.respondError(c, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	userInfo := UserInfo{
		ID:       userIDUint,
		Username: username.(string),
		Email:    userEmail.(string),
		Nickname: username.(string), // Default to username
		Role:     userRole.(string),
	}

	s.respondSuccess(c, http.StatusOK, userInfo, "User information retrieved successfully")
}

// handleRefresh handles token refresh
// @Summary 刷新令牌
// @Description 刷新身份验证令牌
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AuthResponse
// @Failure 401 {object} object
// @Router /auth/refresh [post]
func (s *Service) handleRefresh(c *gin.Context) {
	// Get current user info
	userID, exists := c.Get("user_id")
	if !exists {
		s.respondError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	username, exists := c.Get("username")
	if !exists {
		s.respondError(c, http.StatusUnauthorized, "User information not available")
		return
	}

	// Generate new token
	tokenManager := auth.NewAuthToken(s.config.Server.Token).WithTTL(7 * 24 * time.Hour)
	clientID := fmt.Sprintf("web_%v_%d", userID, time.Now().Unix())

	token, err := tokenManager.GenerateToken(clientID)
	if err != nil {
		s.logger.ErrorTag("Auth", "Failed to refresh token for user %v: %v", username, err)
		s.respondError(c, http.StatusInternalServerError, "Failed to refresh token")
		return
	}

	// Store new client session
	authHandler, err := NewAuthHandler(s.logger, s.config)
	if err == nil && authHandler.authManager != nil {
		clientInfo := auth.ClientInfo{
			ClientID: clientID,
			Username: username.(string),
			Password: "",
			IP:        c.ClientIP(),
			DeviceID:  "web",
			Metadata: map[string]any{
				"user_id":    userID,
				"user_role":  c.GetString("user_role"),
				"user_email": c.GetString("user_email"),
			},
		}

		if err := authHandler.authManager.RegisterClient(c.Request.Context(), clientInfo); err != nil {
			s.logger.WarnTag("Auth", "Failed to store refresh client session for user %v: %v", username, err)
		}

		authHandler.Close()
	}

	// Prepare response
	expiresAt := time.Now().Add(7 * 24 * time.Hour).Unix()
	response := AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: &UserInfo{
			ID:       userID.(uint),
			Username: username.(string),
			Email:    c.GetString("user_email"),
			Nickname: username.(string),
			Role:     c.GetString("user_role"),
		},
	}

	s.respondSuccess(c, http.StatusOK, response, "Token refreshed successfully")
}

// handleLogout handles user logout
// @Summary 用户登出
// @Description 登出当前用户并使会话失效
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object
// @Router /auth/logout [delete]
func (s *Service) handleLogout(c *gin.Context) {
	clientID, exists := c.Get("client_id")
	if !exists {
		s.respondError(c, http.StatusUnauthorized, "Client not authenticated")
		return
	}

	// Remove client session using auth manager
	authHandler, err := NewAuthHandler(s.logger, s.config)
	if err == nil && authHandler.authManager != nil {
		if err := authHandler.authManager.Remove(c.Request.Context(), clientID.(string)); err != nil {
			s.logger.WarnTag("Auth", "Failed to remove client session %v: %v", clientID, err)
		}
		authHandler.Close()
	}

	s.logger.InfoTag("Auth", "User logged out: %v", c.GetString("username"))
	s.respondSuccess(c, http.StatusOK, nil, "Logout successful")
}

// handleLogoutAll handles logout from all devices
// @Summary 从所有设备登出
// @Description 从所有活动会话中登出用户
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object
// @Router /auth/logout-all [delete]
func (s *Service) handleLogoutAll(c *gin.Context) {
	username := c.GetString("username")
	if username == "" {
		s.respondError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get all clients and remove those belonging to this user
	authHandler, err := NewAuthHandler(s.logger, s.config)
	if err == nil && authHandler.authManager != nil {
		clients, err := authHandler.authManager.List(c.Request.Context())
		if err == nil {
			for _, clientID := range clients {
				if client, err := authHandler.authManager.Get(c.Request.Context(), clientID); err == nil {
					if client.Username == username {
						authHandler.authManager.Remove(c.Request.Context(), clientID)
					}
				}
			}
		}
		authHandler.Close()
	}

	s.logger.InfoTag("Auth", "User logged out from all devices: %s", username)
	s.respondSuccess(c, http.StatusOK, nil, "Logged out from all devices successfully")
}

// hashPassword creates a SHA-256 hash of the password (for backwards compatibility)
// Note: New users should use bcrypt hashing
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}