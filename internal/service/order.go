package service

import (
	"errors"
	"time"

	"xboard/internal/model"
	"xboard/internal/repository"

	"github.com/google/uuid"
)

type OrderService struct {
	orderRepo *repository.OrderRepository
	userRepo  *repository.UserRepository
	planRepo  *repository.PlanRepository
}

func NewOrderService(orderRepo *repository.OrderRepository, userRepo *repository.UserRepository, planRepo *repository.PlanRepository) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		userRepo:  userRepo,
		planRepo:  planRepo,
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(userID, planID int64, period string) (*model.Order, error) {
	// 获取套餐
	plan, err := s.planRepo.FindByID(planID)
	if err != nil {
		return nil, errors.New("plan not found")
	}

	// 获取价格
	price := plan.GetPriceByPeriod(period)
	if price <= 0 {
		return nil, errors.New("invalid period")
	}

	// 获取用户
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 确定订单类型
	orderType := model.OrderTypeNewPurchase
	if user.PlanID != nil {
		if *user.PlanID == planID {
			orderType = model.OrderTypeRenewal
		} else {
			orderType = model.OrderTypeUpgrade
		}
	}

	order := &model.Order{
		UserID:      userID,
		PlanID:      planID,
		Period:      period,
		TradeNo:     uuid.New().String(),
		TotalAmount: price,
		Type:        orderType,
		Status:      model.OrderStatusPending,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	// 设置邀请人
	if user.InviteUserID != nil {
		order.InviteUserID = user.InviteUserID
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, err
	}

	return order, nil
}

// GetByTradeNo 根据交易号获取订单
func (s *OrderService) GetByTradeNo(tradeNo string) (*model.Order, error) {
	return s.orderRepo.FindByTradeNo(tradeNo)
}

// GetUserOrders 获取用户订单列表
func (s *OrderService) GetUserOrders(userID int64) ([]model.Order, error) {
	return s.orderRepo.FindByUserID(userID)
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(orderID int64, userID int64) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return errors.New("order not found")
	}

	if order.UserID != userID {
		return errors.New("permission denied")
	}

	if order.Status != model.OrderStatusPending {
		return errors.New("order cannot be cancelled")
	}

	order.Status = model.OrderStatusCancelled
	return s.orderRepo.Update(order)
}

// CompleteOrder 完成订单（支付成功后调用）
func (s *OrderService) CompleteOrder(tradeNo string, callbackNo string) error {
	order, err := s.orderRepo.FindByTradeNo(tradeNo)
	if err != nil {
		return errors.New("order not found")
	}

	if order.Status != model.OrderStatusPending {
		return errors.New("order already processed")
	}

	// 获取套餐
	plan, err := s.planRepo.FindByID(order.PlanID)
	if err != nil {
		return errors.New("plan not found")
	}

	// 获取用户
	user, err := s.userRepo.FindByID(order.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	// 计算过期时间
	days := model.GetPeriodDays(order.Period)
	var expiredAt int64
	if days > 0 {
		if user.ExpiredAt != nil && *user.ExpiredAt > time.Now().Unix() {
			expiredAt = *user.ExpiredAt + int64(days*86400)
		} else {
			expiredAt = time.Now().Unix() + int64(days*86400)
		}
	}

	// 更新用户
	user.PlanID = &order.PlanID
	user.GroupID = plan.GroupID
	user.TransferEnable = plan.TransferEnable * 1024 * 1024 * 1024 // GB to Bytes
	if days > 0 {
		user.ExpiredAt = &expiredAt
	}
	if plan.SpeedLimit != nil {
		user.SpeedLimit = plan.SpeedLimit
	}
	if plan.DeviceLimit != nil {
		user.DeviceLimit = plan.DeviceLimit
	}

	// 重置流量（新购或升级）
	if order.Type == model.OrderTypeNewPurchase || order.Type == model.OrderTypeUpgrade {
		user.U = 0
		user.D = 0
	}

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	// 更新订单状态
	now := time.Now().Unix()
	order.Status = model.OrderStatusCompleted
	order.PaidAt = &now
	order.CallbackNo = &callbackNo

	return s.orderRepo.Update(order)
}
