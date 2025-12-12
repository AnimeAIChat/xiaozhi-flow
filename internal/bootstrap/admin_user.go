package bootstrap

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"xiaozhi-server-go/internal/platform/storage"
)

// createDefaultAdminUser 创建默认管理员用户
func createDefaultAdminUser(db *gorm.DB) error {
	// 检查是否已存在管理员用户
	var count int64
	if err := db.Model(&storage.User{}).Where("role = ?", "admin").Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check admin user count: %w", err)
	}

	// 如果已有管理员，则跳过
	if count > 0 {
		return nil
	}

	// 生成默认管理员密码哈希
	password := "admin123" // 默认密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建默认管理员用户
	adminUser := storage.User{
		Username:  "admin",
		Password:  string(hashedPassword),
		Nickname:  "Administrator",
		Email:     "admin@xiaozhi.local",
		Role:      "admin",
		Status:    1,
		HeadImg:   "",
		Extra:     "",
	}

	if err := db.Create(&adminUser).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// 使用 fmt.Printf 因为 logger 可能还没有初始化
	fmt.Printf("Default admin user created - username: admin, password: %s\n", password)

	return nil
}