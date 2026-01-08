package service

import (
	"bytes"
	"fmt"
	"os"
	"srs-automation/internal/core/domain"
	"srs-automation/internal/core/ports"
	"strings"

	"github.com/ledongthuc/pdf"
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

func extractTextFromPDF(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
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

	// fileContent, err := os.ReadFile(doc.FilePath)
	// if err != nil {
	// 	return fmt.Errorf("gagal baca file fisik: %w", err)
	// }

	var cleanContent string

	// 1. Ekstraksi (Hanya di RAM, tidak disimpan ke Disk)
	if strings.HasSuffix(strings.ToLower(doc.Filename), ".pdf") {
		text, err := extractTextFromPDF(doc.FilePath)
		if err != nil {
			// Fallback jika gagal baca PDF
			raw, _ := os.ReadFile(doc.FilePath)
			cleanContent = string(raw)
		} else {
			cleanContent = text
		}
	} else {
		rawBytes, err := os.ReadFile(doc.FilePath)
		if err != nil {
			return err
		}
		cleanContent = string(rawBytes)
	}

	// 2. Potong Teks (Agar Token AI tidak Jebol & Hemat RAM)
	// Kita ambil 15.000 karakter pertama saja (sekitar 5-7 halaman padat)
	// Ini biasanya sudah CUKUP untuk SRS (Intro + Functional Req biasanya di awal)
	if len(cleanContent) > 15000 {
		fmt.Println("⚠️ Teks terlalu panjang, mengambil 15.000 karakter awal...")
		cleanContent = cleanContent[:15000]
	}

	// 3. Generate SRS Menggunakan Gemini AI
	// Kita kirim konten BRD (doc.Content) ke AI
	fmt.Println("🤖 Gemini sedang menganalisis...")
	srsContent, err := s.aiService.GenerateSRS(string(cleanContent))
	if err != nil {
		doc.Status = "FAILED"
		s.repo.Update(doc)
		return fmt.Errorf("gagal generate SRS: %w", err)
	}

	// folderID := os.Getenv("GOOGLE_DRIVE_FOLDER_ID")

	title := fmt.Sprintf("SRS Draft - %s", doc.Filename)

	savedPath, err := s.aiService.GenerateDocxFile(srsContent, title)
	if err != nil {
		return fmt.Errorf("gagal membuat file docx: %w", err)
	}

	// fmt.Println("📄 Membuat dokumen cloud...")
	// docLink, err := s.aiService.CreateGoogleDoc(title, srsContent, folderID)
	// if err != nil {
	// 	fmt.Printf("Error membuat dokumen: %v\n", err)
	// 	// Kita tidak return error agar data text tetap tersimpan
	// }

	doc.ExtractedData = []byte(srsContent)
	doc.GoogleDocLink = savedPath
	doc.Status = "COMPLETED"

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
