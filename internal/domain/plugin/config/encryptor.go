package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"xiaozhi-server-go/internal/platform/errors"
)

// ConfigEncryptor 配置加密器
type ConfigEncryptor struct {
	gcm cipher.AEAD
}

// NewConfigEncryptor 创建配置加密器
func NewConfigEncryptor(key string) (*ConfigEncryptor, error) {
	if len(key) != 32 {
		return nil, errors.New(errors.KindDomain, "config_encryptor.new", "encryption key must be 32 characters long")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, errors.Wrap(errors.KindDomain, "config_encryptor.new", "failed to create cipher block", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(errors.KindDomain, "config_encryptor.new", "failed to create GCM", err)
	}

	return &ConfigEncryptor{
		gcm: gcm,
	}, nil
}

// Encrypt 加密配置数据
func (e *ConfigEncryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.Wrap(errors.KindDomain, "config_encryptor.encrypt", "failed to generate nonce", err)
	}

	ciphertext := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密配置数据
func (e *ConfigEncryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", errors.Wrap(errors.KindDomain, "config_encryptor.decrypt", "failed to decode base64", err)
	}

	nonceSize := e.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New(errors.KindDomain, "config_encryptor.decrypt", "ciphertext too short")
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", errors.Wrap(errors.KindDomain, "config_encryptor.decrypt", "failed to decrypt", err)
	}

	return string(plaintext), nil
}

// GenerateKey 生成随机加密密钥
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", errors.Wrap(errors.KindDomain, "config_encryptor.generate_key", "failed to generate key", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}