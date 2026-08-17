package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AIAnalysisResult struct {
	BusinessModel         string `json:"business_model"`
	Moat                  string `json:"moat"`
	RedFlags              string `json:"red_flags"`
	ManagementAssumptions string `json:"management_assumptions"`
	KeyRisks              string `json:"key_risks"`
	CustomerConcentration string `json:"customer_concentration"`
	IndustryOutlook       string `json:"industry_outlook"`
	DebtAssessment        string `json:"debt_assessment"`
	PromoterRisk          string `json:"promoter_risk"`
	LegalCases            string `json:"legal_cases"`
}

type GeminiClient struct {
	client *genai.Client
}

func NewGeminiClient(ctx context.Context) (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return &GeminiClient{client: client}, nil
}

func (c *GeminiClient) Close() {
	c.client.Close()
}

// AnalyzeChunk sends a single chunk of text to Gemini using the analyst prompt.
func (c *GeminiClient) AnalyzeChunk(ctx context.Context, text string, prompt string) (*AIAnalysisResult, error) {
	model := c.client.GenerativeModel("gemini-3.5-flash-lite") // Map to the appropriate model based on limits
	
	// Enforce JSON response type
	model.ResponseMIMEType = "application/json"

	fullPrompt := fmt.Sprintf("%s\n\nDocument Chunk to Analyze:\n%s", prompt, text)

	resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response returned from model")
	}

	part := resp.Candidates[0].Content.Parts[0]
	var textResp string
	if t, ok := part.(genai.Text); ok {
		textResp = string(t)
	} else {
		return nil, fmt.Errorf("unexpected response type")
	}

	// Remove markdown code blocks if any (e.g. ```json ... ```)
	textResp = strings.TrimPrefix(textResp, "```json")
	textResp = strings.TrimPrefix(textResp, "```")
	textResp = strings.TrimSuffix(textResp, "```")
	textResp = strings.TrimSpace(textResp)

	var result AIAnalysisResult
	if err := json.Unmarshal([]byte(textResp), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response: %w\nRaw response: %s", err, textResp)
	}

	return &result, nil
}

// MergeAnalyses takes multiple partial JSON analyses and merges them using Gemini.
func (c *GeminiClient) MergeAnalyses(ctx context.Context, partials []AIAnalysisResult, mergePrompt string) (*AIAnalysisResult, error) {
	model := c.client.GenerativeModel("gemini-3.5-flash-lite")
	model.ResponseMIMEType = "application/json"

	partialsJSON, _ := json.MarshalIndent(partials, "", "  ")
	fullPrompt := fmt.Sprintf("%s\n\nPartial Analyses:\n%s", mergePrompt, string(partialsJSON))

	// Slight delay before calling to respect limits if we just finished mapping
	time.Sleep(2 * time.Second)

	resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content during merge: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response returned from merge model")
	}

	part := resp.Candidates[0].Content.Parts[0]
	var textResp string
	if t, ok := part.(genai.Text); ok {
		textResp = string(t)
	} else {
		return nil, fmt.Errorf("unexpected response type")
	}

	textResp = strings.TrimPrefix(textResp, "```json")
	textResp = strings.TrimPrefix(textResp, "```")
	textResp = strings.TrimSuffix(textResp, "```")
	textResp = strings.TrimSpace(textResp)

	var result AIAnalysisResult
	if err := json.Unmarshal([]byte(textResp), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON merge response: %w", err)
	}

	return &result, nil
}

type AIModuleScores struct {
	PromoterScore  int    `json:"promoter_score"`
	PromoterReason string `json:"promoter_reason"`
	IndustryScore  int    `json:"industry_score"`
	IndustryReason string `json:"industry_reason"`
	RiskScore      int    `json:"risk_score"`
	RiskReason     string `json:"risk_reason"`
}

func (c *GeminiClient) ScoreModules(ctx context.Context, aiAnalysisJSON string, prompt string) (*AIModuleScores, error) {
	model := c.client.GenerativeModel("gemini-3.5-flash-lite")
	model.ResponseMIMEType = "application/json"

	fullPrompt := fmt.Sprintf("%s\n\nAI Analysis JSON:\n%s", prompt, aiAnalysisJSON)

	resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate scoring content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response returned from scoring model")
	}

	part := resp.Candidates[0].Content.Parts[0]
	var textResp string
	if t, ok := part.(genai.Text); ok {
		textResp = string(t)
	} else {
		return nil, fmt.Errorf("unexpected response type")
	}

	textResp = strings.TrimPrefix(textResp, "```json")
	textResp = strings.TrimPrefix(textResp, "```")
	textResp = strings.TrimSuffix(textResp, "```")
	textResp = strings.TrimSpace(textResp)

	var result AIModuleScores
	if err := json.Unmarshal([]byte(textResp), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scoring response: %w\nRaw: %s", err, textResp)
	}

	return &result, nil
}

type SectorInferenceResult struct {
	Sector string `json:"sector"`
}

// InferSector uses Gemini to classify the company description into one of the allowed sectors.
func (c *GeminiClient) InferSector(ctx context.Context, companyDescription string, allowedSectors []string) (string, error) {
	model := c.client.GenerativeModel("gemini-3.5-flash-lite")
	model.ResponseMIMEType = "application/json"

	prompt := fmt.Sprintf(`Given the following company description, classify the company into EXACTLY ONE of the following sectors:
%s

Respond ONLY with a JSON object containing the key "sector" and the chosen sector as the value. If none fit well, choose the closest or return "Others".
Example: {"sector": "IT Services"}

Company Description:
%s`, strings.Join(allowedSectors, ", "), companyDescription)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate sector inference: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response returned from inference model")
	}

	part := resp.Candidates[0].Content.Parts[0]
	var textResp string
	if t, ok := part.(genai.Text); ok {
		textResp = string(t)
	} else {
		return "", fmt.Errorf("unexpected response type")
	}

	textResp = strings.TrimPrefix(textResp, "```json")
	textResp = strings.TrimPrefix(textResp, "```")
	textResp = strings.TrimSuffix(textResp, "```")
	textResp = strings.TrimSpace(textResp)

	var result SectorInferenceResult
	if err := json.Unmarshal([]byte(textResp), &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal sector inference response: %w\nRaw: %s", err, textResp)
	}

	return result.Sector, nil
}
