package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/SiriusLLL/go-jwt/database"
	"github.com/SiriusLLL/go-jwt/helpers"
	"github.com/SiriusLLL/go-jwt/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// 打开名为user的MongoDB集合
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

// 定义验证器，用于验证数据有效性
var validate = validator.New()

// 将密码进行哈希处理
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

// 验证用户输入的密码是否与哈希密码匹配
func VerifyPassword(userPassword string, providedPasswrod string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPasswrod), []byte(userPassword))
	check := true
	msg := ""
	if err != nil {
		msg = fmt.Sprintln("email of password is incorrect")
		check = false
	}
	return check, msg
}

// 用户注册
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 创建上下文对象，处理超时
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		// 验证用户格式
		// 绑定用户变量
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 验证
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		// 检查是否存在相同的邮箱地址
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
		}
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email already exists"})
		}
		// 对用户密码进行哈希处理
		password := HashPassword(*user.Password)
		user.Password = &password
		// 检查是否存在相同的手机号码
		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone number"})
		}
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this phone number already exists"})
		}

		// 设置时间
		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		// 生成一个新的 MongoDB ObjectID
		user.ID = primitive.NewObjectID()
		// 转为十六进制字符串
		user.UserId = user.ID.Hex()
		// 生成和刷新token
		token, refreshToken, _ := helpers.GenerateAllTokens(
			*user.Email,
			*user.FirstName,
			*user.LastName,
			*user.UserType,
			user.UserId,
		)
		user.Token = &token
		user.RefreshToken = &refreshToken

		// 用 MongoDB 的 InsertOne 方法将用户信息插入到数据库中
		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintln("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

// 用户登录
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建上下文对象，处理超时
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// user存储从请求中解析得到的用户信息
		var user models.User
		// foundUser存储从数据库中检索到的用户信息
		var foundUser models.User

		// // 绑定用户变量
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error})
			return
		}

		// 用 MongoDB 的 FindOne 方法根据用户的邮箱地址检索用户信息，并将结果解码到 foundUser 变量中
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
			return
		}

		// 检查用户输入的密码是否与数据库中存储的密码匹配
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		// 用户不存在
		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		}

		// 生成token
		token, refreshToken, _ := helpers.GenerateAllTokens(
			*foundUser.Email,
			*foundUser.FirstName,
			*foundUser.LastName,
			*foundUser.UserType,
			foundUser.UserId,
		)
		// 更新用户token
		helpers.UpdateAllTokens(token, refreshToken, foundUser.UserId)

		// 根据用户的 ID 从数据库中重新检索用户信息，并将结果解码到 foundUser 变量中
		err = userCollection.FindOne(ctx, bson.M{"userId": foundUser.UserId}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, foundUser)
	}
}

// 获取用户列表
func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查用户类型是否为 ADMIN
		if err := helpers.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 创建上下文对象，处理超时
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		// 从请求参数中获取每页记录数，并将其转换为整数类型
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		// 获取页码
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		// 根据页码和每页记录数计算起始索引
		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		// 匹配所有文档
		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		// 用于分组和计算总记录数以及按照指定条件进行分组
		groupStage := bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}},
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}
		// 用于筛选结果
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
			}}}
		// 使用 MongoDB 的 Aggregate 方法执行聚合查询，传入聚合管道作为参数，并将结果保存到 result 变量中
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
		}

		// 用于存储查询结果
		var allusers []bson.M
		if err = result.All(ctx, &allusers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allusers[0])
	}
}

// 获取单个用户信息
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从路由参数中获取用户的 ID
		userId := c.Param("userId")

		// 检查用户ID
		if err := helpers.MatchUserTypeUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		}

		// 创建上下文对象
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		// user用于存储从数据库中检索到的用户信息
		var user models.User
		// 根据用户 ID 检索用户信息
		err := userCollection.FindOne(ctx, bson.M{"userId": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
