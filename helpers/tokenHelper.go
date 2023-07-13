package helpers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/SiriusLLL/go-jwt/database"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 用于处理token

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	Uid       string
	UserType  string
	jwt.StandardClaims
}

// 从mongoDB数据库连接中获取用户集合
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

// 获取环境变量中的SECRET_KEY值，用于JWT的签名和验证
var SECRET_KEY string = os.Getenv("SECRET_KEY")

// 用于生成和刷新token
// 接收用户信息作为参数
// 返回生成的token，刷新的token和错误
func GenerateAllTokens(
	email string,
	firstName string,
	lastName string,
	userType string,
	uid string,
) (
	signedToken string,
	signedRefreshToken string,
	err error,
) {
	// 创建结构体实例
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Uid:       uid,
		UserType:  userType,
		StandardClaims: jwt.StandardClaims{
			// 设置token过期时间
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}
	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			// 设置刷新token的过期时间
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	// 创建JWT token，并使用密钥SECRET_KEY进行签名
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}
	return token, refreshToken, err
}

// 用于验证JWT token
// 确保令牌的完整性和有效性
// 接收token
// 返回claims SignedDetails结构体和错误信息msg
func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {

	// 解析JWT token并验证其有效性
	// 接收token，空结构体和一个回调函数
	token, err := jwt.ParseWithClaims(
		signedToken,
		// 空结构体用于存储解析后的声明信息
		&SignedDetails{},
		// 返回密钥和nil
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	// 将token的声明信息转换为SignedDetails类型
	// 如果转换失败，说明token无效
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = fmt.Sprintln("the token is invalid")
		// msg = err.Error()
		return
	}

	// 检查token是否过期
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintln("token is expired")
		// msg = err.Error()
		return
	}
	return claims, msg
}

// 用于刷新token
func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	// 创建context对象ctx，用于控制数据库控制的超时时间
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	// 用于存储更新操作的字段和值
	var updateObj primitive.D

	// 将对应字段和值添加到updateObj
	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refreshToken", Value: signedRefreshToken})
	UpdatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updatedAt", Value: UpdatedAt})

	upsert := true
	filter := bson.M{"userId": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	// 执行更新操作
	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			// 使用 $set 操作符来更新
			{Key: "$set", Value: updateObj},
		},
		&opt,
	)

	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}
}
