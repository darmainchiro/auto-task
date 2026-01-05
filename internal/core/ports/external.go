package ports

// AIService defines the interface for AI processing
type AIService interface {
	ExtractContent(filePath string, fileType string) (string, error)
	GenerateSRS(brdContent string) (string, error)
	AnalyzeDocument(content string) (map[string]interface{}, error)
}

// FileStorageService defines the interface for file operations
type FileStorageService interface {
	SaveFile(filename string, data []byte) (string, error)
	GetFile(filepath string) ([]byte, error)
	DeleteFile(filepath string) error
}
