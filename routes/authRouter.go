package routes

import (
	"github.com/SiriusLLL/go-jwt/controllers"
	"github.com/gin-gonic/gin"
)

// 身份验证相关的API路由
func AuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("users/signup", controllers.SignUp())
	incomingRoutes.POST("users/login", controllers.Login())
}
