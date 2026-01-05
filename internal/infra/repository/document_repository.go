package repository

import (
	"srs-automation/internal/core/domain"

	"gorm.io/gorm"
)

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(doc *domain.Document) error {
	return r.db.Create(doc).Error
}

func (r *DocumentRepository) FindByID(id uint) (*domain.Document, error) {
	var doc domain.Document
	err := r.db.First(&doc, id).Error
	return &doc, err
}

func (r *DocumentRepository) FindAll() ([]domain.Document, error) {
	var docs []domain.Document
	err := r.db.Order("created_at DESC").Find(&docs).Error
	return docs, err
}

func (r *DocumentRepository) Update(doc *domain.Document) error {
	return r.db.Save(doc).Error
}

func (r *DocumentRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Document{}, id).Error
}
