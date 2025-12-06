package repository

import (
	"time"
	"xboard/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id int64) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByToken(token string) (*model.User, error) {
	var user model.User
	err := r.db.Where("token = ?", token).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUUID(uuid string) (*model.User, error) {
	var user model.User
	err := r.db.Where("uuid = ?", uuid).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAvailableUsers 获取指定权限组的可用用户
func (r *UserRepository) GetAvailableUsers(groupIDs []int64) ([]model.User, error) {
	var users []model.User
	err := r.db.
		Where("group_id IN ?", groupIDs).
		Where("u + d < transfer_enable").
		Where("(expired_at >= ? OR expired_at IS NULL)", getCurrentTimestamp()).
		Where("banned = ?", false).
		Select("id", "uuid", "speed_limit", "device_limit").
		Find(&users).Error
	return users, err
}

// UpdateTraffic 更新用户流量
func (r *UserRepository) UpdateTraffic(userID int64, u, d int64) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"u": gorm.Expr("u + ?", u),
			"d": gorm.Expr("d + ?", d),
			"t": getCurrentTimestamp(),
		}).Error
}

// BatchUpdateTraffic 批量更新用户流量
func (r *UserRepository) BatchUpdateTraffic(trafficData map[int64][2]int64) error {
	tx := r.db.Begin()
	for userID, traffic := range trafficData {
		if err := tx.Model(&model.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"u": gorm.Expr("u + ?", traffic[0]),
				"d": gorm.Expr("d + ?", traffic[1]),
				"t": getCurrentTimestamp(),
			}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *UserRepository) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	r.db.Model(&model.User{}).Count(&total)
	err := r.db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (r *UserRepository) CountByPlanID(planID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("plan_id = ?", planID).Count(&count).Error
	return count, err
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
