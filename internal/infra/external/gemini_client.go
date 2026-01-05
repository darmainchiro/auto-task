package external

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type GeminiClient struct {
	apiKey string
	model  string
}

func NewGeminiClient(apiKey string) *GeminiClient {
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-pro"
	}
	return &GeminiClient{
		apiKey: apiKey,
		model:  model,
	}
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

func (c *GeminiClient) GenerateSRS(brdContent string) (string, error) {
	prompt := fmt.Sprintf(`Generate a comprehensive Software Requirements Specification (SRS) document based on the following Business Requirements Document (BRD):

%s

Please structure the SRS with the following sections:
1. Introduction
2. Overall Description
3. Functional Requirements
4. Non-Functional Requirements
5. System Features
6. External Interface Requirements
7. Other Requirements

Provide detailed and technical specifications.`, brdContent)

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
