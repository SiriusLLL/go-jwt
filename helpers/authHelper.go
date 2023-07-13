package helpers

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// 检查用户类型
func CheckUserType(c *gin.Context, role string) (err error) {
	userType := c.GetString("userType")
	err = nil
	if userType != role {
		err = errors.New("unauthorized to access this resource")
		return err
	}
	return err
}

// 检查用户类型和用户ID是否与所需的类型和ID匹配
func MatchUserTypeUid(c *gin.Context, userId string) (err error) {
	userType := c.GetString("userType")
	uid := c.GetString("uid")
	err = nil
	if userType == "USER" && uid != userId {
		err = errors.New("unauthorized to access this resource")
		return err
	}
	err = CheckUserType(c, userType)
	return err
}
