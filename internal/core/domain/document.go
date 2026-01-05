package domain

import "time"

// DocumentType represents the type of document
type DocumentType string

const (
	DocumentTypeBRD   DocumentType = "BRD"
	DocumentTypeSRS   DocumentType = "SRS"
	DocumentTypeOther DocumentType = "OTHER"
)

// DocumentStatus represents the processing status
type DocumentStatus string

const (
	StatusUploaded   DocumentStatus = "UPLOADED"
	StatusProcessing DocumentStatus = "PROCESSING"
	StatusCompleted  DocumentStatus = "COMPLETED"
	StatusFailed     DocumentStatus = "FAILED"
)

// Document represents a document entity
type Document struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Name          string         `json:"name" gorm:"not null"`
	Type          DocumentType   `json:"type" gorm:"not null"`
	FilePath      string         `json:"file_path" gorm:"not null"`
	Status        DocumentStatus `json:"status" gorm:"default:'UPLOADED'"`
	Content       string         `json:"content" gorm:"type:text"`
	ExtractedData string         `json:"extracted_data" gorm:"type:jsonb"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
