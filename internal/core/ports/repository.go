package ports

import "srs-automation/internal/core/domain"

// DocumentRepository defines the interface for document data access
type DocumentRepository interface {
	Create(doc *domain.Document) error
	FindByID(id uint) (*domain.Document, error)
	FindAll() ([]domain.Document, error)
	Update(doc *domain.Document) error
	Delete(id uint) error
}

// SRSRepository defines the interface for SRS data access
type SRSRepository interface {
	Create(srs *domain.SRS) error
	FindByID(id uint) (*domain.SRS, error)
	FindByDocumentID(docID uint) ([]domain.SRS, error)
	FindAll() ([]domain.SRS, error)
	Update(srs *domain.SRS) error
	Delete(id uint) error
}
