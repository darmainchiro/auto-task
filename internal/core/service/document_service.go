package service

import (
	"srs-automation/internal/core/domain"
	"srs-automation/internal/core/ports"
)

type DocumentService struct {
	repo           ports.DocumentRepository
	aiService      ports.AIService
	storageService ports.FileStorageService
}

func NewDocumentService(
	repo ports.DocumentRepository,
	aiService ports.AIService,
	storageService ports.FileStorageService,
) *DocumentService {
	return &DocumentService{
		repo:           repo,
		aiService:      aiService,
		storageService: storageService,
	}
}

func (s *DocumentService) UploadDocument(filename string, docType domain.DocumentType, data []byte) (*domain.Document, error) {
	// Save file first
	filePath, err := s.storageService.SaveFile(filename, data)
	if err != nil {
		return nil, err
	}

	// Create document record
	doc := &domain.Document{
		Filename: filename,
		FilePath: filePath,
		Type:     domain.DocumentTypeBRD,
		Status:   domain.StatusUploaded,
		// Content will be filled later during processing
	}

	if err := s.repo.Create(doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *DocumentService) ProcessDocument(id uint) error {
	doc, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// Update status to processing
	doc.Status = domain.StatusProcessing
	if err := s.repo.Update(doc); err != nil {
		return err
	}

	// Extract content using AI
	extractedContent, err := s.aiService.ExtractContent(doc.FilePath, string(doc.Type))
	if err != nil {
		doc.Status = domain.StatusFailed
		s.repo.Update(doc)
		return err
	}

	doc.Content = extractedContent
	doc.Status = domain.StatusCompleted
	return s.repo.Update(doc)
}

func (s *DocumentService) GetDocument(id uint) (*domain.Document, error) {
	return s.repo.FindByID(id)
}

func (s *DocumentService) GetAllDocuments() ([]domain.Document, error) {
	return s.repo.FindAll()
}

func (s *DocumentService) DeleteDocument(id uint) error {
	doc, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// Delete file
	if err := s.storageService.DeleteFile(doc.FilePath); err != nil {
		return err
	}

	return s.repo.Delete(id)
}
