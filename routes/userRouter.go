package routes

import (
	"github.com/SiriusLLL/go-jwt/controllers"
	"github.com/SiriusLLL/go-jwt/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controllers.GetUsers())
	incomingRoutes.GET("/users/:userId", controllers.GetUser())
}
