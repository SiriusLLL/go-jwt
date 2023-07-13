package middleware

import (
	"fmt"
	"net/http"

	"github.com/SiriusLLL/go-jwt/helpers"
	"github.com/gin-gonic/gin"
)

// 用于验证请求的token
func Authenticate() gin.HandlerFunc {
	// 匿名函数，接收请求的上下文
	return func(c *gin.Context) {
		// 获取请求头中的token
		clientToken := c.Request.Header.Get("token")
		// 如果token为空，则终止请求的处理
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintln("no authorization header provided")})
			c.Abort()
			return
		}

		claims, err := helpers.ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}
		c.Set("email", claims.Email)
		c.Set("firstName", claims.FirstName)
		c.Set("lastName", claims.LastName)
		c.Set("userId", claims.Uid)
		c.Set("userType", claims.UserType)
		c.Next()
	}
}
