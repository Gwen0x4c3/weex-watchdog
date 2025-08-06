package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"weex-watchdog/pkg/crypto"
	"weex-watchdog/pkg/logger"
)

// AuthHandler 处理认证请求
type AuthHandler struct {
	username string
	password string
	aesKey   []byte
	logger   *logger.Logger
}

// NewAuthHandler 创建一个新的AuthHandler
func NewAuthHandler(username, password string, aesKey []byte, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		username: username,
		password: password,
		aesKey:   aesKey,
		logger:   logger,
	}
}

// LoginPayload 登录请求的结构
type LoginPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 处理用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var payload LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request payload"})
		return
	}

	if payload.Username == h.username && payload.Password == h.password {
		// 凭证正确，生成token
		plaintext := payload.Username + ":" + payload.Password
		token, err := crypto.EncryptAES(h.aesKey, plaintext)
		if err != nil {
			h.logger.Error("Failed to encrypt token", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "token": token})
	} else {
		// 凭证错误
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid username or password"})
	}
}
