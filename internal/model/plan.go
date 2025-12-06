package model

// Plan 套餐模型
type Plan struct {
	ID                 int64   `gorm:"primaryKey;column:id" json:"id"`
	GroupID            *int64  `gorm:"column:group_id" json:"group_id"`
	TransferEnable     int64   `gorm:"column:transfer_enable" json:"transfer_enable"`
	Name               string  `gorm:"column:name" json:"name"`
	SpeedLimit         *int    `gorm:"column:speed_limit" json:"speed_limit"`
	DeviceLimit        *int    `gorm:"column:device_limit" json:"device_limit"`
	Show               bool    `gorm:"column:show;default:false" json:"show"`
	Sell               bool    `gorm:"column:sell;default:true" json:"sell"`
	Renew              bool    `gorm:"column:renew;default:true" json:"renew"`
	Sort               *int    `gorm:"column:sort" json:"sort"`
	Content            *string `gorm:"column:content;type:text" json:"content"`
	Prices             JSONMap `gorm:"column:prices;type:json" json:"prices"`
	Tags               JSONArray `gorm:"column:tags;type:json" json:"tags"`
	ResetTrafficMethod *int    `gorm:"column:reset_traffic_method" json:"reset_traffic_method"`
	CapacityLimit      *int    `gorm:"column:capacity_limit" json:"capacity_limit"`
	CreatedAt          int64   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          int64   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Plan) TableName() string {
	return "v2_plan"
}

// 流量重置方式
const (
	ResetTrafficFollowSystem  = -1 // 跟随系统设置
	ResetTrafficFirstDayMonth = 0  // 每月1号
	ResetTrafficMonthly       = 1  // 按月重置
	ResetTrafficNever         = 2  // 不重置
	ResetTrafficFirstDayYear  = 3  // 每年1月1日
	ResetTrafficYearly        = 4  // 按年重置
)

// 订阅周期
const (
	PeriodMonthly     = "monthly"
	PeriodQuarterly   = "quarterly"
	PeriodHalfYearly  = "half_yearly"
	PeriodYearly      = "yearly"
	PeriodTwoYearly   = "two_yearly"
	PeriodThreeYearly = "three_yearly"
	PeriodOnetime     = "onetime"
	PeriodResetTraffic = "reset_traffic"
)

// GetPeriodDays 获取周期天数
func GetPeriodDays(period string) int {
	switch period {
	case PeriodMonthly:
		return 30
	case PeriodQuarterly:
		return 90
	case PeriodHalfYearly:
		return 180
	case PeriodYearly:
		return 365
	case PeriodTwoYearly:
		return 730
	case PeriodThreeYearly:
		return 1095
	case PeriodOnetime:
		return -1
	default:
		return 0
	}
}

// GetPriceByPeriod 获取指定周期的价格
func (p *Plan) GetPriceByPeriod(period string) int64 {
	if p.Prices == nil {
		return 0
	}
	if price, ok := p.Prices[period]; ok {
		switch v := price.(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		case int:
			return int64(v)
		}
	}
	return 0
}
