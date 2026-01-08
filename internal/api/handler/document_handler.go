package handler

import (
	"fmt"
	"srs-automation/internal/core/domain"
	"srs-automation/internal/core/service"
	"time"

	"github.com/gofiber/fiber/v2"
)

type DocumentHandler struct {
	service *service.DocumentService
}

func NewDocumentHandler(service *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{service: service}
}

func (h *DocumentHandler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File is required",
		})
	}

	docType := c.FormValue("type")
	if docType == "" {
		docType = string(domain.DocumentTypeBRD)
	}

	// // Read file content
	fileContent, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read file",
		})
	}
	defer fileContent.Close()

	fileData := make([]byte, file.Size)
	if _, err := fileContent.Read(fileData); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read file content",
		})
	}

	// Upload document
	doc, err := h.service.UploadDocument(file.Filename, domain.DocumentType(docType), fileData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	go func(id uint) {
		// Logika ini berjalan TERPISAH dari request user
		fmt.Printf("🔄 [Background] Memulai analisis AI untuk ID: %d...\n", id)

		err := h.service.ProcessDocument(id)
		if err != nil {
			// Jika error, kita hanya bisa log di terminal server karena user sudah pergi
			fmt.Printf("❌ [Background] Gagal memproses ID %d: %v\n", id, err)
		} else {
			fmt.Printf("✅ [Background] Sukses! Google Doc dibuat untuk ID %d\n", id)
		}
	}(doc.ID) // Kita kirim ID dokumen yang baru saja dibuat

	// 5. Response Cepat ke User
	// User langsung dapat balasan detik itu juga
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Upload berhasil. Dokumen sedang diproses oleh AI di latar belakang.",
		"data": fiber.Map{
			"id":        doc.ID,
			"filename":  doc.Filename,
			"status":    "PROCESSING", // Beritahu user statusnya
			"timestamp": time.Now(),
		},
	})
}

func (h *DocumentHandler) Process(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid document ID",
		})
	}

	if err := h.service.ProcessDocument(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Document processing started",
	})
}

func (h *DocumentHandler) GetByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid document ID",
		})
	}

	doc, err := h.service.GetDocument(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": doc,
	})
}

func (h *DocumentHandler) GetAll(c *fiber.Ctx) error {
	docs, err := h.service.GetAllDocuments()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": docs,
	})
}

func (s *DocumentHandler) GetDocumentByID(id uint) (*domain.Document, error) {
	// Memanggil service untuk mencari data di database
	doc, err := s.service.GetDocument(id)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (h *DocumentHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid document ID",
		})
	}

	if err := h.service.DeleteDocument(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Document deleted successfully",
	})
}

// Endpoint: GET /api/v1/documents/:id/download
func (h *DocumentHandler) DownloadResult(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")

	doc, err := h.service.GetDocument(uint(id)) // Use the correct service method
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Dokumen tidak ditemukan"})
	}

	if doc.Status != "COMPLETED" {
		return c.Status(400).JSON(fiber.Map{"error": "Dokumen belum selesai diproses"})
	}

	// doc.GoogleDocLink sekarang berisi path file lokal "./outputs/SRS-namafile.docx"
	return c.Download(doc.GoogleDocLink)
}
