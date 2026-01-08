package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gingfrederik/docx"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	apiKey       string
	model        string
	docsService  *docs.Service
	driveService *drive.Service
}

func NewGeminiClient(apiKey string, googleCredsFile string) (*GeminiClient, error) {
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-pro"
	}

	ctx := context.Background()

	// 1. Baca File Credential
	credsData, err := os.ReadFile(googleCredsFile)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file credential: %w", err)
	}

	// 2. Inisialisasi Google Docs Service (Untuk edit isi dokumen)
	docsSrv, err := docs.NewService(ctx, option.WithCredentialsJSON(credsData))
	if err != nil {
		return nil, fmt.Errorf("gagal init docs service: %w", err)
	}

	driveSrv, err := drive.NewService(ctx, option.WithCredentialsJSON(credsData))
	if err != nil {
		return nil, fmt.Errorf("gagal init drive service: %w", err)
	}

	return &GeminiClient{
		apiKey:       apiKey,
		model:        model,
		docsService:  docsSrv,
		driveService: driveSrv,
	}, nil
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (c *GeminiClient) ExtractContent(filePath string, fileType string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(
		"Extract and summarize the key information from this %s document:\n\n%s",
		fileType,
		string(content),
	)

	return c.callGemini(prompt)
}

// func (c *GeminiClient) CreateGoogleDoc(title string, content string, folderID string) (string, error) {
// 	if c.docsService == nil || c.driveService == nil {
// 		return "", errors.New("google services not initialized")
// 	}

// 	// A. BUAT DOKUMEN KOSONG (Default masuk ke root folder Service Account)
// 	doc := &docs.Document{
// 		Title: title,
// 	}
// 	createdDoc, err := c.docsService.Documents.Create(doc).Do()
// 	if err != nil {
// 		return "", fmt.Errorf("gagal membuat dokumen: %w", err)
// 	}

// 	// B. PINDAHKAN KE FOLDER ID (Jika folderID diisi)
// 	if folderID != "" {
// 		// 1. Ambil ID parent saat ini (biasanya root)
// 		file, err := c.driveService.Files.Get(createdDoc.DocumentId).Fields("parents").Do()
// 		if err == nil {
// 			previousParents := ""
// 			if len(file.Parents) > 0 {
// 				previousParents = file.Parents[0]
// 			}

// 			// 2. Pindahkan: AddParents ke folder tujuan, RemoveParents dari root
// 			_, err = c.driveService.Files.Update(createdDoc.DocumentId, nil).
// 				AddParents(folderID).
// 				RemoveParents(previousParents).
// 				Do()

// 			if err != nil {
// 				// Kita log error tapi jangan return fail, karena doc sebenarnya sudah terbuat
// 				fmt.Printf("Warning: Gagal memindahkan file ke folder %s: %v\n", folderID, err)
// 			}
// 		}
// 	}

// 	// C. ISI KONTEN TEKS
// 	requests := []*docs.Request{
// 		{
// 			InsertText: &docs.InsertTextRequest{
// 				Text: content,
// 				Location: &docs.Location{
// 					Index: 1,
// 				},
// 			},
// 		},
// 	}

// 	batchUpdate := &docs.BatchUpdateDocumentRequest{
// 		Requests: requests,
// 	}

// 	_, err = c.docsService.Documents.BatchUpdate(createdDoc.DocumentId, batchUpdate).Do()
// 	if err != nil {
// 		return "", fmt.Errorf("gagal mengisi konten: %w", err)
// 	}

// 	// D. RETURN LINK
// 	return fmt.Sprintf("https://docs.google.com/document/d/%s/edit", createdDoc.DocumentId), nil
// }

func (c *GeminiClient) GenerateDocxFile(content string, filename string) (string, error) {
	// 1. Inisialisasi File Docx Baru
	f := docx.NewFile()

	// 2. Format Teks Gemini ke Paragraf Docx
	// Karena Gemini outputnya Markdown, kita pecah per baris agar rapi
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		// Skip baris kosong berlebihan
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Tambahkan paragraf baru
		para := f.AddParagraph()

		// Sedikit trik: Membersihkan format Markdown sederhana (opsional)
		// Misal menghapus tanda ** atau # agar lebih bersih di Word
		cleanLine := strings.ReplaceAll(line, "**", "")
		cleanLine = strings.ReplaceAll(cleanLine, "##", "")

		para.AddText(cleanLine)
	}

	// 3. Tentukan Lokasi Simpan
	// Pastikan folder "outputs" sudah dibuat manual di root project
	outputPath := fmt.Sprintf("./outputs/%s.docx", filename)

	// 4. Simpan File
	err := f.Save(outputPath)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan file docx: %w", err)
	}

	return outputPath, nil
}

func (c *GeminiClient) CreateGoogleDoc(title string, content string, folderID string) (string, error) {
	if c.docsService == nil || c.driveService == nil {
		return "", errors.New("google services not initialized")
	}

	// --- PERUBAHAN UTAMA DI SINI ---
	// Cara Lama: Buat Docs (di root) -> Pindah ke Folder (Sering Error 403)
	// Cara Baru: Buat File via Drive API langsung di dalam Folder ID

	fileMetadata := &drive.File{
		Name:     title,
		MimeType: "application/vnd.google-apps.document", // Tipe Google Doc
	}

	// Jika folderID ada, langsung set sebagai 'Parents' saat pembuatan
	if folderID != "" {
		fileMetadata.Parents = []string{folderID}
	}

	// 1. Eksekusi Pembuatan File
	createdFile, err := c.driveService.Files.Create(fileMetadata).Do()
	if err != nil {
		return "", fmt.Errorf("gagal membuat file di drive: %w", err)
	}

	// 2. Isi Konten (Menggunakan Docs API berdasarkan ID file yang baru dibuat)
	// Kita perlu sedikit jeda agar file terpropagasi, tapi biasanya instan
	requests := []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Text: content,
				Location: &docs.Location{
					Index: 1,
				},
			},
		},
	}

	batchUpdate := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err = c.docsService.Documents.BatchUpdate(createdFile.Id, batchUpdate).Do()
	if err != nil {
		return "", fmt.Errorf("gagal mengisi konten: %w", err)
	}

	// 3. Return Link
	return fmt.Sprintf("https://docs.google.com/document/d/%s/edit", createdFile.Id), nil
}

func (c *GeminiClient) GenerateSRS(brdContent string) (string, error) {
	prompt := fmt.Sprintf(`Analisis file BRD ini dan buatkan Draft SRS yang sangat detail dalam format Markdown.`, brdContent)

	return c.callGemini(prompt)
}

func (c *GeminiClient) AnalyzeDocument(content string) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Analyze this document and extract key information in JSON format:

%s

Return a JSON object with: title, summary, key_points (array), requirements (array), stakeholders (array)`, content)

	response, err := c.callGemini(prompt)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *GeminiClient) callGemini(prompt string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("Gemini API key not configured")
	}

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.model, c.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini API error: %s", string(body))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no response from Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
