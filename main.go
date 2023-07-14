package main

import (
	"log"
	"os"

	"github.com/SiriusLLL/go-jwt/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// 入口文件，其中包含了主要的路由和处理函数
func main() {

	// 从env中获取端口
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	// 创建Gin路由实例，用于处理HTTP请求和定义路由。
	router := gin.New()
	// 将Gin的日志中间件添加到路由中，用于记录请求和响应的日志信息。
	router.Use(gin.Logger())

	// 路由
	routes.AuthRoutes(router)
	routes.UserRoutes(router)
	router.GET("/api-1", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access granted for api-1"})
	})
	router.GET("/api-2", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access granted for api-2"})
	})

	// 启动HTTP服务器，监听port
	router.Run(":" + port)

}
