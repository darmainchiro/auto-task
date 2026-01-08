package external

import (
	"context"
	"fmt"
	"strings"

	"github.com/gingfrederik/docx"
	openai "github.com/sashabaranov/go-openai"
)

type GroqClient struct {
	client *openai.Client
	model  string
}

func NewGroqClient(apiKey string) *GroqClient {
	// 1. Konfigurasi Client agar mengarah ke Server Groq
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"

	client := openai.NewClientWithConfig(config)

	return &GroqClient{
		client: client,
		// Model rekomendasi: llama3-70b-8192 atau mixtral-8x7b-32768
		model: "llama-3.3-70b-versatile",
	}
}

// Implementasi Interface: GenerateSRS
func (c *GroqClient) GenerateSRS(content string) (string, error) {
	prompt := fmt.Sprintf(`You are a Senior System Analyst. 
Buatlah Software Requirements Specification (SRS) yang komprehensif berdasarkan input teks di bawah ini.

CATATAN PENTING:
1. Input berupa teks yang diekstrak dari dokumen. Gambar atau diagram TIDAK disertakan.

2. Jika teks merujuk pada diagram yang hilang (misalnya, "lihat Gambar 1"), simpulkan logikanya dari konteks sekitarnya jika memungkinkan.

3. Keluarkan HANYA isi SRS.

Data Input:
%s

Struktur:
1. Pendahuluan
2. Persyaratan Fungsional
3. Persyaratan Non-Fungsional
4. Fitur Sistem`, content)

	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			// Groq support max tokens, sesuaikan kebutuhan
			MaxTokens: 4096,
		},
	)

	if err != nil {
		return "", fmt.Errorf("groq api error: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

// Implementasi Interface: GenerateDocxFile (Sama persis seperti sebelumnya)
func (c *GroqClient) GenerateDocxFile(content string, filename string) (string, error) {
	f := docx.NewFile()
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Bersihkan markdown simpel
		cleanLine := strings.ReplaceAll(line, "**", "")
		cleanLine = strings.ReplaceAll(cleanLine, "##", "")
		cleanLine = strings.ReplaceAll(cleanLine, "#", "")

		para := f.AddParagraph()
		para.AddText(cleanLine)
	}

	// Simpan ke folder outputs
	outputPath := fmt.Sprintf("./outputs/%s.docx", filename)
	err := f.Save(outputPath)
	if err != nil {
		return "", fmt.Errorf("gagal save docx: %w", err)
	}

	return outputPath, nil
}
