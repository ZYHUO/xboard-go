package service

import (
	"errors"
	"time"

	"xboard/internal/model"
	"xboard/internal/repository"
	"xboard/pkg/cache"
	"xboard/pkg/utils"
)

type UserService struct {
	userRepo *repository.UserRepository
	cache    *cache.Client
}

func NewUserService(userRepo *repository.UserRepository, cache *cache.Client) *UserService {
	return &UserService{
		userRepo: userRepo,
		cache:    cache,
	}
}

// Register 用户注册
func (s *UserService) Register(email, password string, inviteUserID *int64) (*model.User, error) {
	// 检查邮箱是否已存在
	existing, _ := s.userRepo.FindByEmail(email)
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        email,
		Password:     hashedPassword,
		UUID:         utils.GenerateUUID(),
		Token:        utils.GenerateToken(32),
		InviteUserID: inviteUserID,
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(email, password string) (*model.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("invalid email or password")
	}

	if user.Banned {
		return nil, errors.New("account is banned")
	}

	// 更新最后登录时间
	now := time.Now().Unix()
	user.LastLoginAt = &now
	s.userRepo.Update(user)

	return user, nil
}

// GetByID 根据 ID 获取用户
func (s *UserService) GetByID(id int64) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

// GetByToken 根据 Token 获取用户
func (s *UserService) GetByToken(token string) (*model.User, error) {
	return s.userRepo.FindByToken(token)
}

// GetByUUID 根据 UUID 获取用户
func (s *UserService) GetByUUID(uuid string) (*model.User, error) {
	return s.userRepo.FindByUUID(uuid)
}

// UpdateTraffic 更新用户流量
func (s *UserService) UpdateTraffic(userID int64, u, d int64) error {
	return s.userRepo.UpdateTraffic(userID, u, d)
}

// TrafficFetch 批量处理流量上报
func (s *UserService) TrafficFetch(server *model.Server, trafficData map[int64][2]int64) error {
	// 计算倍率
	rate := server.Rate
	if rate <= 0 {
		rate = 1
	}

	// 应用倍率
	for userID, traffic := range trafficData {
		u := int64(float64(traffic[0]) * rate)
		d := int64(float64(traffic[1]) * rate)
		trafficData[userID] = [2]int64{u, d}
	}

	return s.userRepo.BatchUpdateTraffic(trafficData)
}

// ResetToken 重置用户 Token
func (s *UserService) ResetToken(userID int64) (string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", err
	}

	user.Token = utils.GenerateToken(32)
	if err := s.userRepo.Update(user); err != nil {
		return "", err
	}

	return user.Token, nil
}

// ResetUUID 重置用户 UUID
func (s *UserService) ResetUUID(userID int64) (string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", err
	}

	user.UUID = utils.GenerateUUID()
	if err := s.userRepo.Update(user); err != nil {
		return "", err
	}

	return user.UUID, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("invalid old password")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	return s.userRepo.Update(user)
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(user *model.User) map[string]interface{} {
	return map[string]interface{}{
		"id":              user.ID,
		"email":           user.Email,
		"uuid":            user.UUID,
		"token":           user.Token,
		"balance":         user.Balance,
		"plan_id":         user.PlanID,
		"group_id":        user.GroupID,
		"transfer_enable": user.TransferEnable,
		"u":               user.U,
		"d":               user.D,
		"expired_at":      user.ExpiredAt,
		"is_admin":        user.IsAdmin,
		"is_staff":        user.IsStaff,
		"created_at":      user.CreatedAt,
	}
}
