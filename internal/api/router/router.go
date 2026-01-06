package router

import (
	"srs-automation/internal/api/handler"
	"srs-automation/internal/core/service"
	"srs-automation/internal/infra/external"
	"srs-automation/internal/infra/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB, aiClient *external.GeminiClient) {
	// Initialize repositories
	docRepo := repository.NewDocumentRepository(db)
	srsRepo := repository.NewSRSRepository(db)

	// Initialize external services
	fileStorage := external.NewFileStorage()

	// Initialize services
	docService := service.NewDocumentService(docRepo, aiClient, fileStorage)
	srsService := service.NewSRSService(srsRepo, docRepo, aiClient)

	// Initialize handlers
	docHandler := handler.NewDocumentHandler(docService)
	srsHandler := handler.NewSRSHandler(srsService)

	app.Static("/uploads", "./uploads")

	// API routes
	api := app.Group("/api/v1")

	// Document routes
	documents := api.Group("/documents")
	documents.Post("/", docHandler.Upload)
	documents.Get("/", docHandler.GetAll)
	documents.Get("/:id", docHandler.GetByID)
	documents.Post("/:id/process", docHandler.Process)
	documents.Delete("/:id", docHandler.Delete)

	// SRS routes
	srs := api.Group("/srs")
	srs.Post("/", srsHandler.Generate)
	srs.Get("/", srsHandler.GetAll)
	srs.Get("/:id", srsHandler.GetByID)
	srs.Get("/document/:documentId", srsHandler.GetByDocument)
	srs.Put("/:id", srsHandler.Update)
	srs.Delete("/:id", srsHandler.Delete)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})
}
