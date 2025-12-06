package handler

import (
	"net/http"

	"xboard/internal/service"

	"github.com/gin-gonic/gin"
)

// GuestRegister 用户注册
func GuestRegister(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email        string `json:"email" binding:"required,email"`
			Password     string `json:"password" binding:"required,min=6"`
			InviteCode   string `json:"invite_code"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: 处理邀请码
		var inviteUserID *int64

		user, err := services.User.Register(req.Email, req.Password, inviteUserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := services.Auth.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"token": token,
			},
		})
	}
}

// GuestLogin 用户登录
func GuestLogin(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := services.User.Login(req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := services.Auth.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"token":    token,
				"is_admin": user.IsAdmin,
			},
		})
	}
}

// GuestGetPlans 获取可购买套餐列表
func GuestGetPlans(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		plans, err := services.Plan.GetAvailable()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		result := make([]map[string]interface{}, 0, len(plans))
		for _, plan := range plans {
			result = append(result, services.Plan.GetPlanInfo(&plan))
		}

		c.JSON(http.StatusOK, gin.H{"data": result})
	}
}

// PassportLogin Passport 登录
func PassportLogin(services *service.Services) gin.HandlerFunc {
	return GuestLogin(services)
}

// PassportRegister Passport 注册
func PassportRegister(services *service.Services) gin.HandlerFunc {
	return GuestRegister(services)
}
