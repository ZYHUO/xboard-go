package repository

import (
	"xboard/internal/model"

	"gorm.io/gorm"
)

type ServerRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

func (r *ServerRepository) Create(server *model.Server) error {
	return r.db.Create(server).Error
}

func (r *ServerRepository) Update(server *model.Server) error {
	return r.db.Save(server).Error
}

func (r *ServerRepository) Delete(id int64) error {
	return r.db.Delete(&model.Server{}, id).Error
}

func (r *ServerRepository) FindByID(id int64) (*model.Server, error) {
	var server model.Server
	err := r.db.First(&server, id).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *ServerRepository) FindByCode(serverType, code string) (*model.Server, error) {
	var server model.Server
	err := r.db.Where("type = ? AND code = ?", serverType, code).First(&server).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

// GetAllServers 获取所有服务器
func (r *ServerRepository) GetAllServers() ([]model.Server, error) {
	var servers []model.Server
	err := r.db.Order("sort ASC").Find(&servers).Error
	return servers, err
}

// GetAvailableServers 获取指定权限组的可用服务器
func (r *ServerRepository) GetAvailableServers(groupID int64) ([]model.Server, error) {
	var servers []model.Server
	// 使用 JSON_CONTAINS 查询包含指定 group_id 的服务器
	err := r.db.
		Where("JSON_CONTAINS(group_ids, ?)", groupID).
		Where("show = ?", true).
		Order("sort ASC").
		Find(&servers).Error
	return servers, err
}

// GetServersByType 按类型获取服务器
func (r *ServerRepository) GetServersByType(serverType string) ([]model.Server, error) {
	var servers []model.Server
	err := r.db.Where("type = ?", serverType).Order("sort ASC").Find(&servers).Error
	return servers, err
}

func (r *ServerRepository) List(page, pageSize int) ([]model.Server, int64, error) {
	var servers []model.Server
	var total int64

	r.db.Model(&model.Server{}).Count(&total)
	err := r.db.Order("sort ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&servers).Error
	return servers, total, err
}

// ServerGroup Repository
type ServerGroupRepository struct {
	db *gorm.DB
}

func NewServerGroupRepository(db *gorm.DB) *ServerGroupRepository {
	return &ServerGroupRepository{db: db}
}

func (r *ServerGroupRepository) FindByID(id int64) (*model.ServerGroup, error) {
	var group model.ServerGroup
	err := r.db.First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *ServerGroupRepository) GetAll() ([]model.ServerGroup, error) {
	var groups []model.ServerGroup
	err := r.db.Find(&groups).Error
	return groups, err
}

// ServerRoute Repository
type ServerRouteRepository struct {
	db *gorm.DB
}

func NewServerRouteRepository(db *gorm.DB) *ServerRouteRepository {
	return &ServerRouteRepository{db: db}
}

func (r *ServerRouteRepository) FindByIDs(ids []int64) ([]model.ServerRoute, error) {
	var routes []model.ServerRoute
	err := r.db.Where("id IN ?", ids).Find(&routes).Error
	return routes, err
}

func (r *ServerRouteRepository) GetAll() ([]model.ServerRoute, error) {
	var routes []model.ServerRoute
	err := r.db.Find(&routes).Error
	return routes, err
}
