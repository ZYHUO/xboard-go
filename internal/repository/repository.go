package repository

import "gorm.io/gorm"

type Repositories struct {
	User          *UserRepository
	Server        *ServerRepository
	Plan          *PlanRepository
	Order         *OrderRepository
	Setting       *SettingRepository
	Stat          *StatRepository
	Ticket        *TicketRepository
	TicketMessage *TicketMessageRepository
	Payment       *PaymentRepository
	Coupon        *CouponRepository
	InviteCode    *InviteCodeRepository
	CommissionLog *CommissionLogRepository
	Notice        *NoticeRepository
	Knowledge     *KnowledgeRepository
	Host          *HostRepository
	ServerNode    *ServerNodeRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:          NewUserRepository(db),
		Server:        NewServerRepository(db),
		Plan:          NewPlanRepository(db),
		Order:         NewOrderRepository(db),
		Setting:       NewSettingRepository(db),
		Stat:          NewStatRepository(db),
		Ticket:        NewTicketRepository(db),
		TicketMessage: NewTicketMessageRepository(db),
		Payment:       NewPaymentRepository(db),
		Coupon:        NewCouponRepository(db),
		InviteCode:    NewInviteCodeRepository(db),
		CommissionLog: NewCommissionLogRepository(db),
		Notice:        NewNoticeRepository(db),
		Knowledge:     NewKnowledgeRepository(db),
		Host:          NewHostRepository(db),
		ServerNode:    NewServerNodeRepository(db),
	}
}
