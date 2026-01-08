package main

import (
	"log"
	"os"
	"srs-automation/internal/api/router"
	"srs-automation/internal/infra/database"
	"srs-automation/internal/infra/external"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize external services
	//Groq
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("GROQ_API_KEY is required in .env")
	}

	aiClient := external.NewGroqClient(apiKey)

	//Gemini
	// aiClient, err := external.NewGeminiClient(os.Getenv("GEMINI_API_KEY"), "google-credentials.json")
	// if err != nil {
	// 	log.Fatalf("Gagal inisialisasi Gemini/Google Client: %v", err)
	// }

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Setup routes
	router.SetupRoutes(app, db, aiClient)

	// Start server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
