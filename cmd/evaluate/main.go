package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"golang.org/x/oauth2/google"
	"google.golang.org/genai"
	"gopkg.in/yaml.v3"
)

type PromptConfig struct {
	Messages []struct {
		Role    string `yaml:"role"`
		Content string `yaml:"content"`
	} `yaml:"messages"`
}

type TestItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type JudgeResponse struct {
	IsPromptInjection bool `json:"is_prompt_injection"`
}

type EvaluationResult struct {
	ModelID   string  `json:"model_id"`
	Accuracy  float64 `json:"accuracy"`
	Precision float64 `json:"precision"`
	Recall    float64 `json:"recall"`
	F1Score   float64 `json:"f1_score"`
	Errors    int     `json:"errors"`
}

type ModelConfig struct {
	ModelID  string
	IsGemini bool
}

var models = []ModelConfig{
	// {ModelID: "publishers/mistralai/models/mistral-small-2503", IsGemini: false},
	{ModelID: "gemini-2.5-flash-lite", IsGemini: true},
	{ModelID: "gemini-2.5-flash", IsGemini: true},
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT must be set")
	}
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1" // fallback
	}

	ctx := context.Background()

	// Load prompt configuration
	promptBytes, err := os.ReadFile("prompts/security-judge.prompt.yml")
	if err != nil {
		log.Fatalf("Failed to read prompts/security-judge.prompt.yml: %v", err)
	}
	var promptCfg PromptConfig
	if err := yaml.Unmarshal(promptBytes, &promptCfg); err != nil {
		log.Fatalf("Failed to parse prompt YAML: %v", err)
	}

	var systemInstruction string
	for _, msg := range promptCfg.Messages {
		if msg.Role == "system" {
			systemInstruction = msg.Content
			break
		}
	}
	if systemInstruction == "" {
		log.Fatal("No system message found in prompt file")
	}

	// Load test data
	testDataBytes, err := os.ReadFile("test_prompts.json")
	if err != nil {
		log.Fatalf("Failed to read test_prompts.json: %v", err)
	}
	var testItems []TestItem
	if err := json.Unmarshal(testDataBytes, &testItems); err != nil {
		log.Fatalf("Failed to parse test_prompts.json: %v", err)
	}

	results := []EvaluationResult{}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  projectID,
		Location: location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		log.Fatalf("Failed to create GenAI client: %v", err)
	}
	// No formatting close method in new SDK mostly, or check doc.
	// Usually standard http client underneath.

	for _, modelCfg := range models {

		fmt.Printf("Evaluating Model: %s\n", modelCfg.ModelID)

		tp, fp, tn, fn := 0, 0, 0, 0
		errors := 0

		temp := float32(0.0)

		for i, item := range testItems {
			if i%10 == 0 {
				fmt.Printf(".")
			}

			// Rate limit
			time.Sleep(500 * time.Millisecond)

			isAttackExpected := item.Type == "attack"

			fullPrompt := fmt.Sprintf("%s\n\nTranscript:\n%s", systemInstruction, item.Text)

			var respStr string

			if modelCfg.IsGemini {
				resp, err := client.Models.GenerateContent(ctx, modelCfg.ModelID, genai.Text(fullPrompt), &genai.GenerateContentConfig{
					Temperature: &temp,
				})

				if err != nil {
					if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
						fmt.Printf("\nModel %s not found or accessible. Skipping.\n", modelCfg.ModelID)
						errors = len(testItems)
						break
					}
					log.Printf("Error gen: %v", err)
					errors++
					continue
				}

				if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
					errors++
					continue
				}
				// Parse response
				for _, part := range resp.Candidates[0].Content.Parts {
					if part.Text != "" {
						respStr += part.Text
					}
				}
			} else {
				// Mistral / OpenAI usage via generic HTTP
				// Assuming OpenAI Chat Completions on Vertex endpoint format for best compatibility,
				// OR generic rawPredict if OpenAI is flaky.
				// Let's try rawPredict which is simpler if we construct the JSON manualy.
				// Mistral Raw Predict format: {"instances": [{"messages": [{"role": "user", "content": "..."}]}], "parameters": { ... }}

				resp, err := callMistral(ctx, projectID, location, modelCfg.ModelID, fullPrompt)
				if err != nil {
					if strings.Contains(err.Error(), "404") {
						fmt.Printf("\nModel %s not found. Skipping.\n", modelCfg.ModelID)
						errors = len(testItems)
						break
					}
					log.Printf("Error mistral: %v", err)
					errors++
					continue
				}
				respStr = resp
			}

			// Naive JSON parse
			var judgeResponse JudgeResponse
			if err := json.Unmarshal([]byte(respStr), &judgeResponse); err != nil {
				log.Printf("Error unmarshaling judge response: %v, response: %s", err, respStr)
				errors++
				continue
			}
			isAttackPred := judgeResponse.IsPromptInjection

			if isAttackExpected {
				if isAttackPred {
					tp++
				} else {
					fn++
				}
			} else {
				if isAttackPred {
					fp++
				} else {
					tn++
				}
			}
		}
		fmt.Println()

		total := tp + fp + tn + fn
		accuracy := 0.0
		if total > 0 {
			accuracy = float64(tp+tn) / float64(total)
		}
		precision := 0.0
		if tp+fp > 0 {
			precision = float64(tp) / float64(tp+fp)
		}
		recall := 0.0
		if tp+fn > 0 {
			recall = float64(tp) / float64(tp+fn)
		}
		f1 := 0.0
		if precision+recall > 0 {
			f1 = 2 * (precision * recall) / (precision + recall)
		}

		res := EvaluationResult{
			ModelID:   modelCfg.ModelID,
			Accuracy:  accuracy,
			Precision: precision,
			Recall:    recall,
			F1Score:   f1,
			Errors:    errors,
		}
		results = append(results, res)
	}

	// Output results
	outBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal results: %v", err)
	}
	fmt.Println(string(outBytes))
	if err := os.WriteFile("evaluation_results.json", outBytes, 0644); err != nil {
		log.Fatalf("Failed to write results file: %v", err)
	}
}

func callMistral(ctx context.Context, project, location, model, prompt string) (string, error) {
	// Endpoint: https://us-central1-aiplatform.googleapis.com/v1/projects/{PROJECT}/locations/us-central1/publishers/mistralai/models/{MODEL}:rawPredict
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/%s:rawPredict", location, project, location, model)

	// JSON Body
	// RAW SPEC: {"instances": [{"messages": [{"role": "user", "content": "..."}]}], "parameters": {"maxOutputTokens": 1024}}
	bodyMap := map[string]interface{}{
		"instances": []interface{}{
			map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{"role": "user", "content": prompt},
				},
			},
		},
		"parameters": map[string]interface{}{
			"maxOutputTokens": 1024,
			"temperature":     0.0,
		},
	}
	jsonBody, _ := json.Marshal(bodyMap)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	// Get default token
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", fmt.Errorf("auth error: %v", err)
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("token error: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("api error status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response. Is it just string content or complex?
	// Mistral raw predict usually returns just text in array? Or OpenAI format?
	// If it's pure rawPredict, check structure.
	// Usually it's `[{"predictions": ["text..."]}]` or similar.
	// Let's assume generic structure and extract first string found if possible or print to debug if fails.
	// Actually, usually it returns: `{"predictions": ["response text"]}`

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", err
	}

	if preds, ok := result["predictions"].([]interface{}); ok && len(preds) > 0 {
		if s, ok := preds[0].(string); ok {
			return s, nil
		}
	}
	// Fallback logic could be complex JSON inside prediction
	return string(bodyBytes), nil
}
