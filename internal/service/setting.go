package service

import (
	"strconv"
	"time"

	"xboard/internal/repository"
	"xboard/pkg/cache"
)

type SettingService struct {
	settingRepo *repository.SettingRepository
	cache       *cache.Client
}

func NewSettingService(settingRepo *repository.SettingRepository, cache *cache.Client) *SettingService {
	return &SettingService{
		settingRepo: settingRepo,
		cache:       cache,
	}
}

// Get 获取设置
func (s *SettingService) Get(key string) (string, error) {
	// 先从缓存获取
	cacheKey := "setting:" + key
	if val, err := s.cache.Get(cacheKey); err == nil {
		return val, nil
	}

	// 从数据库获取
	val, err := s.settingRepo.Get(key)
	if err != nil {
		return "", err
	}

	// 写入缓存
	s.cache.Set(cacheKey, val, time.Hour)
	return val, nil
}

// Set 设置
func (s *SettingService) Set(key, value string) error {
	if err := s.settingRepo.Set(key, value); err != nil {
		return err
	}

	// 更新缓存
	cacheKey := "setting:" + key
	return s.cache.Set(cacheKey, value, time.Hour)
}

// GetInt 获取整数设置
func (s *SettingService) GetInt(key string, defaultVal int) int {
	val, err := s.Get(key)
	if err != nil {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

// GetBool 获取布尔设置
func (s *SettingService) GetBool(key string, defaultVal bool) bool {
	val, err := s.Get(key)
	if err != nil {
		return defaultVal
	}
	return val == "1" || val == "true"
}

// GetAll 获取所有设置
func (s *SettingService) GetAll() (map[string]string, error) {
	return s.settingRepo.GetAll()
}

// 常用设置 key
const (
	SettingAppName           = "app_name"
	SettingAppURL            = "app_url"
	SettingServerPushInterval = "server_push_interval"
	SettingServerPullInterval = "server_pull_interval"
	SettingSubscribeURL      = "subscribe_url"
)
