package domain

import "time"

// SRS represents a Software Requirements Specification
type SRS struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	SourceDocumentID uint      `json:"source_document_id" gorm:"not null"`
	Title            string    `json:"title" gorm:"not null"`
	Version          string    `json:"version" gorm:"default:'1.0'"`
	Content          string    `json:"content" gorm:"type:text"`
	Sections         string    `json:"sections" gorm:"type:jsonb"`
	Status           string    `json:"status" gorm:"default:'DRAFT'"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	SourceDocument Document `json:"source_document" gorm:"foreignKey:SourceDocumentID"`
}

// SRSSection represents a section in the SRS document
type SRSSection struct {
	Title       string       `json:"title"`
	Content     string       `json:"content"`
	Subsections []SRSSection `json:"subsections,omitempty"`
}
