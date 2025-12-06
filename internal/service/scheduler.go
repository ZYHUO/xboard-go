package service

import (
	"log"
	"time"

	"xboard/internal/model"
	"xboard/internal/repository"
)

// SchedulerService 定时任务服务
type SchedulerService struct {
	userRepo    *repository.UserRepository
	orderRepo   *repository.OrderRepository
	statRepo    *repository.StatRepository
	mailService *MailService
	tgService   *TelegramService
}

func NewSchedulerService(
	userRepo *repository.UserRepository,
	orderRepo *repository.OrderRepository,
	statRepo *repository.StatRepository,
	mailService *MailService,
	tgService *TelegramService,
) *SchedulerService {
	return &SchedulerService{
		userRepo:    userRepo,
		orderRepo:   orderRepo,
		statRepo:    statRepo,
		mailService: mailService,
		tgService:   tgService,
	}
}

// Start 启动定时任务
func (s *SchedulerService) Start() {
	// 每天凌晨执行
	go s.runDaily()

	// 每小时执行
	go s.runHourly()

	// 每分钟执行
	go s.runMinutely()
}

// runDaily 每天执行的任务
func (s *SchedulerService) runDaily() {
	// 计算到明天凌晨的时间
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	time.Sleep(next.Sub(now))

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		s.dailyTasks()
		<-ticker.C
	}
}

// runHourly 每小时执行的任务
func (s *SchedulerService) runHourly() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.hourlyTasks()
	}
}

// runMinutely 每分钟执行的任务
func (s *SchedulerService) runMinutely() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.minutelyTasks()
	}
}

// dailyTasks 每日任务
func (s *SchedulerService) dailyTasks() {
	log.Println("[Scheduler] Running daily tasks...")

	// 1. 重置流量（每月1号）
	if time.Now().Day() == 1 {
		s.resetMonthlyTraffic()
	}

	// 2. 发送到期提醒
	s.sendExpireReminders()

	// 3. 清理过期订单
	s.cleanExpiredOrders()

	// 4. 生成每日统计
	s.generateDailyStats()
}

// hourlyTasks 每小时任务
func (s *SchedulerService) hourlyTasks() {
	// 1. 发送流量预警
	s.sendTrafficWarnings()
}

// minutelyTasks 每分钟任务
func (s *SchedulerService) minutelyTasks() {
	// 可以添加需要频繁执行的任务
}

// resetMonthlyTraffic 重置月流量
func (s *SchedulerService) resetMonthlyTraffic() {
	log.Println("[Scheduler] Resetting monthly traffic...")

	users, err := s.userRepo.GetUsersNeedTrafficReset(model.ResetTrafficFirstDayMonth)
	if err != nil {
		log.Printf("[Scheduler] Failed to get users for traffic reset: %v", err)
		return
	}

	for _, user := range users {
		user.U = 0
		user.D = 0
		if err := s.userRepo.Update(&user); err != nil {
			log.Printf("[Scheduler] Failed to reset traffic for user %d: %v", user.ID, err)
		}
	}

	log.Printf("[Scheduler] Reset traffic for %d users", len(users))
}

// sendExpireReminders 发送到期提醒
func (s *SchedulerService) sendExpireReminders() {
	log.Println("[Scheduler] Sending expire reminders...")

	// 获取即将到期的用户（3天内）
	expireTime := time.Now().Add(3 * 24 * time.Hour).Unix()
	users, err := s.userRepo.GetUsersExpiringSoon(expireTime)
	if err != nil {
		log.Printf("[Scheduler] Failed to get expiring users: %v", err)
		return
	}

	for _, user := range users {
		if user.RemindExpire == nil || *user.RemindExpire == 0 {
			continue
		}

		daysLeft := 0
		if user.ExpiredAt != nil {
			daysLeft = int((*user.ExpiredAt - time.Now().Unix()) / 86400)
		}

		// 发送邮件
		if err := s.mailService.SendExpireReminder(&user, daysLeft); err != nil {
			log.Printf("[Scheduler] Failed to send expire email to %s: %v", user.Email, err)
		}

		// 发送 Telegram
		if err := s.tgService.NotifyExpire(&user, daysLeft); err != nil {
			log.Printf("[Scheduler] Failed to send expire telegram to user %d: %v", user.ID, err)
		}
	}

	log.Printf("[Scheduler] Sent expire reminders to %d users", len(users))
}

// sendTrafficWarnings 发送流量预警
func (s *SchedulerService) sendTrafficWarnings() {
	// 获取流量使用超过 80% 的用户
	users, err := s.userRepo.GetUsersWithHighTrafficUsage(80)
	if err != nil {
		return
	}

	for _, user := range users {
		if user.RemindTraffic == nil || *user.RemindTraffic == 0 {
			continue
		}

		usedPercent := 0
		if user.TransferEnable > 0 {
			usedPercent = int((user.U + user.D) * 100 / user.TransferEnable)
		}

		// 发送邮件
		s.mailService.SendTrafficWarning(&user, usedPercent)

		// 发送 Telegram
		s.tgService.NotifyTrafficWarning(&user, usedPercent)
	}
}

// cleanExpiredOrders 清理过期订单
func (s *SchedulerService) cleanExpiredOrders() {
	log.Println("[Scheduler] Cleaning expired orders...")

	// 取消超过 24 小时未支付的订单
	expireTime := time.Now().Add(-24 * time.Hour).Unix()
	count, err := s.orderRepo.CancelExpiredOrders(expireTime)
	if err != nil {
		log.Printf("[Scheduler] Failed to cancel expired orders: %v", err)
		return
	}

	log.Printf("[Scheduler] Cancelled %d expired orders", count)
}

// generateDailyStats 生成每日统计
func (s *SchedulerService) generateDailyStats() {
	log.Println("[Scheduler] Generating daily stats...")

	now := time.Now()
	yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	recordAt := yesterday.Unix()

	// 统计订单
	orderCount, orderTotal, _ := s.orderRepo.GetDailyStats(recordAt)

	// 统计注册
	registerCount, _ := s.userRepo.CountByDateRange(recordAt, recordAt+86400)

	stat := &model.Stat{
		RecordAt:      recordAt,
		RecordType:    "d",
		OrderCount:    orderCount,
		OrderTotal:    orderTotal,
		RegisterCount: registerCount,
	}

	s.statRepo.CreateOrUpdateStat(stat)

	log.Printf("[Scheduler] Daily stats generated: orders=%d, total=%d, registers=%d",
		orderCount, orderTotal, registerCount)
}
