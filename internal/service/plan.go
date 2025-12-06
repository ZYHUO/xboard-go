package service

import (
	"xboard/internal/model"
	"xboard/internal/repository"
)

type PlanService struct {
	planRepo *repository.PlanRepository
	userRepo *repository.UserRepository
}

func NewPlanService(planRepo *repository.PlanRepository, userRepo *repository.UserRepository) *PlanService {
	return &PlanService{
		planRepo: planRepo,
		userRepo: userRepo,
	}
}

// GetAll 获取所有套餐
func (s *PlanService) GetAll() ([]model.Plan, error) {
	return s.planRepo.GetAll()
}

// GetAvailable 获取可购买的套餐
func (s *PlanService) GetAvailable() ([]model.Plan, error) {
	return s.planRepo.GetAvailable()
}

// GetByID 根据 ID 获取套餐
func (s *PlanService) GetByID(id int64) (*model.Plan, error) {
	return s.planRepo.FindByID(id)
}

// Create 创建套餐
func (s *PlanService) Create(plan *model.Plan) error {
	return s.planRepo.Create(plan)
}

// Update 更新套餐
func (s *PlanService) Update(plan *model.Plan) error {
	return s.planRepo.Update(plan)
}

// Delete 删除套餐
func (s *PlanService) Delete(id int64) error {
	// 检查是否有用户使用该套餐
	count, err := s.userRepo.CountByPlanID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrPlanInUse
	}
	return s.planRepo.Delete(id)
}

// GetPlanInfo 获取套餐信息（包含价格列表）
func (s *PlanService) GetPlanInfo(plan *model.Plan) map[string]interface{} {
	periods := []string{
		model.PeriodMonthly,
		model.PeriodQuarterly,
		model.PeriodHalfYearly,
		model.PeriodYearly,
		model.PeriodTwoYearly,
		model.PeriodThreeYearly,
		model.PeriodOnetime,
	}

	prices := make(map[string]int64)
	for _, period := range periods {
		price := plan.GetPriceByPeriod(period)
		if price > 0 {
			prices[period] = price
		}
	}

	return map[string]interface{}{
		"id":                   plan.ID,
		"name":                 plan.Name,
		"group_id":             plan.GroupID,
		"transfer_enable":      plan.TransferEnable,
		"speed_limit":          plan.SpeedLimit,
		"device_limit":         plan.DeviceLimit,
		"show":                 plan.Show,
		"sell":                 plan.Sell,
		"renew":                plan.Renew,
		"content":              plan.Content,
		"prices":               prices,
		"reset_traffic_method": plan.ResetTrafficMethod,
		"capacity_limit":       plan.CapacityLimit,
	}
}

var ErrPlanInUse = &PlanError{Message: "plan is in use by users"}

type PlanError struct {
	Message string
}

func (e *PlanError) Error() string {
	return e.Message
}
