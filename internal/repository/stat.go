package repository

import (
	"time"
	"xboard/internal/model"

	"gorm.io/gorm"
)

type StatRepository struct {
	db *gorm.DB
}

func NewStatRepository(db *gorm.DB) *StatRepository {
	return &StatRepository{db: db}
}

// RecordUserTraffic 记录用户流量统计
func (r *StatRepository) RecordUserTraffic(userID int64, serverRate float64, u, d int64, recordType string) error {
	now := time.Now()
	var recordAt int64
	switch recordType {
	case "d": // daily
		recordAt = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	case "m": // monthly
		recordAt = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()
	default:
		recordAt = now.Unix()
	}

	var stat model.StatUser
	err := r.db.Where("user_id = ? AND server_rate = ? AND record_at = ?", userID, serverRate, recordAt).First(&stat).Error
	if err == gorm.ErrRecordNotFound {
		stat = model.StatUser{
			UserID:     userID,
			ServerRate: serverRate,
			U:          u,
			D:          d,
			RecordType: recordType,
			RecordAt:   recordAt,
		}
		return r.db.Create(&stat).Error
	}
	if err != nil {
		return err
	}
	return r.db.Model(&stat).Updates(map[string]interface{}{
		"u": gorm.Expr("u + ?", u),
		"d": gorm.Expr("d + ?", d),
	}).Error
}

// RecordServerTraffic 记录节点流量统计
func (r *StatRepository) RecordServerTraffic(serverID int64, serverType string, u, d int64, recordType string) error {
	now := time.Now()
	var recordAt int64
	switch recordType {
	case "d": // daily
		recordAt = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	case "m": // monthly
		recordAt = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()
	default:
		recordAt = now.Unix()
	}

	var stat model.StatServer
	err := r.db.Where("server_id = ? AND server_type = ? AND record_at = ?", serverID, serverType, recordAt).First(&stat).Error
	if err == gorm.ErrRecordNotFound {
		stat = model.StatServer{
			ServerID:   serverID,
			ServerType: serverType,
			U:          u,
			D:          d,
			RecordType: recordType,
			RecordAt:   recordAt,
		}
		return r.db.Create(&stat).Error
	}
	if err != nil {
		return err
	}
	return r.db.Model(&stat).Updates(map[string]interface{}{
		"u": gorm.Expr("u + ?", u),
		"d": gorm.Expr("d + ?", d),
	}).Error
}

// GetUserStats 获取用户流量统计
func (r *StatRepository) GetUserStats(userID int64, startAt, endAt int64) ([]model.StatUser, error) {
	var stats []model.StatUser
	err := r.db.Where("user_id = ? AND record_at >= ? AND record_at <= ?", userID, startAt, endAt).Find(&stats).Error
	return stats, err
}

// GetServerStats 获取节点流量统计
func (r *StatRepository) GetServerStats(serverID int64, startAt, endAt int64) ([]model.StatServer, error) {
	var stats []model.StatServer
	err := r.db.Where("server_id = ? AND record_at >= ? AND record_at <= ?", serverID, startAt, endAt).Find(&stats).Error
	return stats, err
}
