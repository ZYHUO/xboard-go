package handler

import (
	"net/http"
	"strconv"

	"xboard/internal/model"
	"xboard/internal/protocol"
	"xboard/internal/service"

	"github.com/gin-gonic/gin"
)

// UserInfo 获取用户信息
func UserInfo(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		info := services.User.GetUserInfo(user)
		c.JSON(http.StatusOK, gin.H{"data": info})
	}
}

// UserSubscribe 获取订阅信息
func UserSubscribe(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		servers, err := services.Server.GetAvailableServers(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"subscribe_url": "/api/v1/client/subscribe?token=" + user.Token,
				"servers":       servers,
			},
		})
	}
}

// UserResetToken 重置订阅 Token
func UserResetToken(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		newToken, err := services.User.ResetToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": gin.H{"token": newToken}})
	}
}

// UserResetUUID 重置 UUID
func UserResetUUID(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		newUUID, err := services.User.ResetUUID(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": gin.H{"uuid": newUUID}})
	}
}

// UserChangePassword 修改密码
func UserChangePassword(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req struct {
			OldPassword string `json:"old_password" binding:"required"`
			NewPassword string `json:"new_password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.User.ChangePassword(user.ID, req.OldPassword, req.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// UserOrders 获取用户订单列表
func UserOrders(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		orders, err := services.Order.GetUserOrders(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": orders})
	}
}

// UserCreateOrder 创建订单
func UserCreateOrder(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req struct {
			PlanID int64  `json:"plan_id" binding:"required"`
			Period string `json:"period" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order, err := services.Order.CreateOrder(user.ID, req.PlanID, req.Period)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": order})
	}
}

// UserCancelOrder 取消订单
func UserCancelOrder(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req struct {
			OrderID int64 `json:"order_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Order.CancelOrder(req.OrderID, user.ID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// ClientSubscribe 客户端订阅
func ClientSubscribe(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.String(http.StatusBadRequest, "token required")
			return
		}

		user, err := services.User.GetByToken(token)
		if err != nil {
			c.String(http.StatusNotFound, "user not found")
			return
		}

		if !user.IsActive() {
			c.String(http.StatusForbidden, "subscription expired")
			return
		}

		servers, err := services.Server.GetAvailableServers(user)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		// 根据 User-Agent 返回不同格式
		ua := c.GetHeader("User-Agent")
		format := c.Query("format")

		// 设置订阅信息头
		c.Header("subscription-userinfo", formatSubscriptionInfo(user))
		c.Header("profile-update-interval", "24")

		switch {
		case format == "singbox" || containsAny(ua, "sing-box", "hiddify", "sfm"):
			c.JSON(http.StatusOK, generateSingBoxConfig(servers, user))
		case format == "clash" || containsAny(ua, "clash", "stash"):
			c.String(http.StatusOK, generateClashConfig(servers, user))
		default:
			// 默认返回 base64 编码的链接
			c.String(http.StatusOK, generateBase64Links(servers, user))
		}
	}
}

func getUserFromContext(c *gin.Context) *model.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*model.User)
}

func formatSubscriptionInfo(user *model.User) string {
	expiredAt := int64(0)
	if user.ExpiredAt != nil {
		expiredAt = *user.ExpiredAt
	}
	return "upload=" + strconv.FormatInt(user.U, 10) +
		"; download=" + strconv.FormatInt(user.D, 10) +
		"; total=" + strconv.FormatInt(user.TransferEnable, 10) +
		"; expire=" + strconv.FormatInt(expiredAt, 10)
}

func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// 这些函数调用 protocol 包
func generateSingBoxConfig(servers []service.ServerInfo, user *model.User) map[string]interface{} {
	return protocol.GenerateSingBoxConfig(servers, user)
}

func generateClashConfig(servers []service.ServerInfo, user *model.User) string {
	return protocol.GenerateClashConfig(servers, user)
}

func generateBase64Links(servers []service.ServerInfo, user *model.User) string {
	return protocol.GenerateBase64Links(servers, user)
}

// UserTickets 获取用户工单列表
func UserTickets(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

		tickets, total, err := services.Ticket.GetUserTickets(user.ID, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  tickets,
			"total": total,
		})
	}
}

// UserTicketDetail 获取工单详情
func UserTicketDetail(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		detail, err := services.Ticket.GetTicketDetail(id, user.ID, user.IsAdmin)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": detail})
	}
}

// UserCreateTicket 创建工单
func UserCreateTicket(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req struct {
			Subject string `json:"subject" binding:"required"`
			Message string `json:"message" binding:"required"`
			Level   int    `json:"level"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ticket, err := services.Ticket.CreateTicket(user.ID, req.Subject, req.Message, req.Level)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": ticket})
	}
}

// UserReplyTicket 回复工单
func UserReplyTicket(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		var req struct {
			Message string `json:"message" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		message, err := services.Ticket.ReplyTicket(id, user.ID, req.Message, user.IsAdmin)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": message})
	}
}

// UserCloseTicket 关闭工单
func UserCloseTicket(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.Ticket.CloseTicket(id, user.ID, user.IsAdmin); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}
