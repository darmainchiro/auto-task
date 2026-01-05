package repository

import (
	"srs-automation/internal/core/domain"

	"gorm.io/gorm"
)

type SRSRepository struct {
	db *gorm.DB
}

func NewSRSRepository(db *gorm.DB) *SRSRepository {
	return &SRSRepository{db: db}
}

func (r *SRSRepository) Create(srs *domain.SRS) error {
	return r.db.Create(srs).Error
}

func (r *SRSRepository) FindByID(id uint) (*domain.SRS, error) {
	var srs domain.SRS
	err := r.db.Preload("SourceDocument").First(&srs, id).Error
	return &srs, err
}

func (r *SRSRepository) FindByDocumentID(docID uint) ([]domain.SRS, error) {
	var srsList []domain.SRS
	err := r.db.Where("source_document_id = ?", docID).Order("created_at DESC").Find(&srsList).Error
	return srsList, err
}

func (r *SRSRepository) FindAll() ([]domain.SRS, error) {
	var srsList []domain.SRS
	err := r.db.Preload("SourceDocument").Order("created_at DESC").Find(&srsList).Error
	return srsList, err
}

func (r *SRSRepository) Update(srs *domain.SRS) error {
	return r.db.Save(srs).Error
}

func (r *SRSRepository) Delete(id uint) error {
	return r.db.Delete(&domain.SRS{}, id).Error
}
