package service

import (
	"encoding/json"
	"errors"
	"srs-automation/internal/core/domain"
	"srs-automation/internal/core/ports"
)

type SRSService struct {
	srsRepo   ports.SRSRepository
	docRepo   ports.DocumentRepository
	aiService ports.AIService
}

func NewSRSService(
	srsRepo ports.SRSRepository,
	docRepo ports.DocumentRepository,
	aiService ports.AIService,
) *SRSService {
	return &SRSService{
		srsRepo:   srsRepo,
		docRepo:   docRepo,
		aiService: aiService,
	}
}

func (s *SRSService) GenerateSRS(documentID uint, title string) (*domain.SRS, error) {
	// Get source document
	doc, err := s.docRepo.FindByID(documentID)
	if err != nil {
		return nil, err
	}

	if doc.Status != domain.StatusCompleted {
		return nil, errors.New("document must be processed first")
	}

	// Generate SRS using AI
	srsContent, err := s.aiService.GenerateSRS(doc.Content)
	if err != nil {
		return nil, err
	}

	// Parse sections (simplified)
	sections := []domain.SRSSection{
		{Title: "Introduction", Content: "Generated introduction"},
		{Title: "Functional Requirements", Content: "Generated requirements"},
		{Title: "Non-Functional Requirements", Content: "Generated non-functional requirements"},
	}

	sectionsJSON, _ := json.Marshal(sections)

	// Create SRS record
	srs := &domain.SRS{
		SourceDocumentID: documentID,
		Title:            title,
		Version:          "1.0",
		Content:          srsContent,
		Sections:         string(sectionsJSON),
		Status:           "DRAFT",
	}

	if err := s.srsRepo.Create(srs); err != nil {
		return nil, err
	}

	return srs, nil
}

func (s *SRSService) GetSRS(id uint) (*domain.SRS, error) {
	return s.srsRepo.FindByID(id)
}

func (s *SRSService) GetSRSByDocument(docID uint) ([]domain.SRS, error) {
	return s.srsRepo.FindByDocumentID(docID)
}

func (s *SRSService) GetAllSRS() ([]domain.SRS, error) {
	return s.srsRepo.FindAll()
}

func (s *SRSService) UpdateSRS(id uint, content string, status string) error {
	srs, err := s.srsRepo.FindByID(id)
	if err != nil {
		return err
	}

	if content != "" {
		srs.Content = content
	}
	if status != "" {
		srs.Status = status
	}

	return s.srsRepo.Update(srs)
}

func (s *SRSService) DeleteSRS(id uint) error {
	return s.srsRepo.Delete(id)
}
