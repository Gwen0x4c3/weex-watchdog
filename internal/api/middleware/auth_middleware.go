package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"weex-watchdog/pkg/crypto"
)

// CORSMiddleware 创建一个CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// AuthMiddleware 创建一个Token验证中间件
func AuthMiddleware(aesKey []byte, validUsername, validPassword string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Authorization token required"})
			c.Abort()
			return
		}

		decrypted, err := crypto.DecryptAES(aesKey, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid token"})
			c.Abort()
			return
		}

		parts := strings.Split(decrypted, ":")
		if len(parts) != 2 || parts[0] != validUsername || parts[1] != validPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token validation failed"})
			c.Abort()
			return
		}

		c.Next()
	}
}
