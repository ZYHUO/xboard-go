package handler

import (
	"net/http"
	"strconv"

	"xboard/internal/model"
	"xboard/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminListServers 获取服务器列表
func AdminListServers(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		servers, err := services.Server.GetAllServers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": servers})
	}
}

// AdminCreateServer 创建服务器
func AdminCreateServer(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var server model.Server
		if err := c.ShouldBindJSON(&server); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// TODO: 实现创建逻辑
		c.JSON(http.StatusOK, gin.H{"data": server})
	}
}

// AdminUpdateServer 更新服务器
func AdminUpdateServer(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		var server model.Server
		if err := c.ShouldBindJSON(&server); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		server.ID = id
		// TODO: 实现更新逻辑
		c.JSON(http.StatusOK, gin.H{"data": server})
	}
}

// AdminDeleteServer 删除服务器
func AdminDeleteServer(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		// TODO: 实现删除逻辑
		_ = id
		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminGetServerStatus 获取服务器状态
func AdminGetServerStatus(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		server, err := services.Server.FindServer(id, "")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
			return
		}

		status, err := services.NodeSync.GetNodeStatus(server)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"online": false,
					"error":  err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": status})
	}
}

// AdminSyncServerUsers 手动同步服务器用户
func AdminSyncServerUsers(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		server, err := services.Server.FindServer(id, "")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
			return
		}

		endpoint := service.NodeEndpoint{
			Server: server,
		}

		// 从 protocol_settings 获取 SSMAPI 配置
		if ps := server.ProtocolSettings; ps != nil {
			if apiURL, ok := ps["ssmapi_url"].(string); ok {
				endpoint.BaseURL = apiURL
			}
			if token, ok := ps["ssmapi_token"].(string); ok {
				endpoint.BearerToken = token
			}
		}

		if endpoint.BaseURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "SSMAPI URL not configured"})
			return
		}

		if err := services.NodeSync.SyncUsers(endpoint); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminListUsers 获取用户列表
func AdminListUsers(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		// TODO: 实现分页查询
		_ = page
		_ = pageSize
		c.JSON(http.StatusOK, gin.H{"data": []model.User{}, "total": 0})
	}
}

// AdminGetUser 获取用户详情
func AdminGetUser(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		user, err := services.User.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": services.User.GetUserInfo(user)})
	}
}

// AdminUpdateUser 更新用户
func AdminUpdateUser(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		// TODO: 实现更新逻辑
		_ = id
		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminListPlans 获取套餐列表
func AdminListPlans(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		plans, err := services.Plan.GetAll()
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

// AdminCreatePlan 创建套餐
func AdminCreatePlan(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var plan model.Plan
		if err := c.ShouldBindJSON(&plan); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Plan.Create(&plan); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": plan})
	}
}

// AdminUpdatePlan 更新套餐
func AdminUpdatePlan(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		plan, err := services.Plan.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}

		if err := c.ShouldBindJSON(plan); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := services.Plan.Update(plan); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": plan})
	}
}

// AdminDeletePlan 删除套餐
func AdminDeletePlan(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.Plan.Delete(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminListOrders 获取订单列表
func AdminListOrders(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现订单列表
		c.JSON(http.StatusOK, gin.H{"data": []model.Order{}, "total": 0})
	}
}

// AdminGetSettings 获取系统设置
func AdminGetSettings(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		settings, err := services.Setting.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": settings})
	}
}

// AdminUpdateSettings 更新系统设置
func AdminUpdateSettings(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		var settings map[string]string
		if err := c.ShouldBindJSON(&settings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		for key, value := range settings {
			if err := services.Setting.Set(key, value); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminListTickets 获取工单列表
func AdminListTickets(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		var status *int
		if s := c.Query("status"); s != "" {
			v, _ := strconv.Atoi(s)
			status = &v
		}

		tickets, total, err := services.Ticket.GetAllTickets(status, page, pageSize)
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

// AdminTicketDetail 获取工单详情
func AdminTicketDetail(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		detail, err := services.Ticket.GetTicketDetail(id, user.ID, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": detail})
	}
}

// AdminReplyTicket 管理员回复工单
func AdminReplyTicket(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		var req struct {
			Message string `json:"message" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		message, err := services.Ticket.ReplyTicket(id, user.ID, req.Message, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": message})
	}
}

// AdminCloseTicket 管理员关闭工单
func AdminCloseTicket(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getUserFromContext(c)
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		if err := services.Ticket.CloseTicket(id, user.ID, true); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

// AdminStatsOverview 获取统计概览
func AdminStatsOverview(services *service.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现统计数据
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"total_users":    0,
				"active_users":   0,
				"total_orders":   0,
				"total_income":   0,
				"pending_tickets": 0,
			},
		})
	}
}

func getUserFromContext(c *gin.Context) *model.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*model.User)
}
