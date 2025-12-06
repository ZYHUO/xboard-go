package service

import (
	"time"

	"xboard/internal/model"
	"xboard/internal/repository"
)

// NoticeService 公告服务
type NoticeService struct {
	noticeRepo *repository.NoticeRepository
}

func NewNoticeService(noticeRepo *repository.NoticeRepository) *NoticeService {
	return &NoticeService{noticeRepo: noticeRepo}
}

// GetAll 获取所有公告
func (s *NoticeService) GetAll() ([]model.Notice, error) {
	return s.noticeRepo.GetAll()
}

// GetVisible 获取可见公告
func (s *NoticeService) GetVisible() ([]model.Notice, error) {
	return s.noticeRepo.GetVisible()
}

// GetByID 根据 ID 获取公告
func (s *NoticeService) GetByID(id int64) (*model.Notice, error) {
	return s.noticeRepo.FindByID(id)
}

// Create 创建公告
func (s *NoticeService) Create(notice *model.Notice) error {
	notice.CreatedAt = time.Now().Unix()
	notice.UpdatedAt = time.Now().Unix()
	return s.noticeRepo.Create(notice)
}

// Update 更新公告
func (s *NoticeService) Update(notice *model.Notice) error {
	notice.UpdatedAt = time.Now().Unix()
	return s.noticeRepo.Update(notice)
}

// Delete 删除公告
func (s *NoticeService) Delete(id int64) error {
	return s.noticeRepo.Delete(id)
}

// KnowledgeService 知识库服务
type KnowledgeService struct {
	knowledgeRepo *repository.KnowledgeRepository
}

func NewKnowledgeService(knowledgeRepo *repository.KnowledgeRepository) *KnowledgeService {
	return &KnowledgeService{knowledgeRepo: knowledgeRepo}
}

// GetAll 获取所有知识库文章
func (s *KnowledgeService) GetAll() ([]model.Knowledge, error) {
	return s.knowledgeRepo.GetAll()
}

// GetVisible 获取可见文章
func (s *KnowledgeService) GetVisible(language string) ([]model.Knowledge, error) {
	return s.knowledgeRepo.GetVisible(language)
}

// GetByCategory 按分类获取文章
func (s *KnowledgeService) GetByCategory(category, language string) ([]model.Knowledge, error) {
	return s.knowledgeRepo.GetByCategory(category, language)
}

// GetByID 根据 ID 获取文章
func (s *KnowledgeService) GetByID(id int64) (*model.Knowledge, error) {
	return s.knowledgeRepo.FindByID(id)
}

// Create 创建文章
func (s *KnowledgeService) Create(knowledge *model.Knowledge) error {
	knowledge.CreatedAt = time.Now().Unix()
	knowledge.UpdatedAt = time.Now().Unix()
	return s.knowledgeRepo.Create(knowledge)
}

// Update 更新文章
func (s *KnowledgeService) Update(knowledge *model.Knowledge) error {
	knowledge.UpdatedAt = time.Now().Unix()
	return s.knowledgeRepo.Update(knowledge)
}

// Delete 删除文章
func (s *KnowledgeService) Delete(id int64) error {
	return s.knowledgeRepo.Delete(id)
}

// GetCategories 获取所有分类
func (s *KnowledgeService) GetCategories(language string) ([]string, error) {
	return s.knowledgeRepo.GetCategories(language)
}
