package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/redis/go-redis/v9"
	"github.com/satishpolireddy/car-rental-ai/config"
	"github.com/satishpolireddy/car-rental-ai/internal/models"
	"github.com/satishpolireddy/car-rental-ai/internal/repository"
	log "github.com/sirupsen/logrus"
)

const (
	cachePrefix = "ai:rec:"
	cacheTTL    = 10 * time.Minute
	systemPrompt = `You are a helpful car rental assistant for DriveAI Car Rentals.
Your job is to recommend the best car options based on the customer's needs.
Always be concise, friendly, and specific. When recommending cars, mention:
- Why the car suits their needs
- Key features relevant to their trip
- Approximate cost if daily rate is known
Return your response as JSON with fields: "message" (string) and "recommended_car_ids" (array of integers).`
)

// AIService wraps Azure OpenAI and Redis for recommendation logic.
// Redis caching prevents redundant API calls for identical queries, reducing cost and latency.
type AIService struct {
	client     *azopenai.Client
	deployment string
	redis      *redis.Client
	carRepo    *repository.CarRepository
}

func NewAIService(cfg config.AzureConfig, rdb *redis.Client, carRepo *repository.CarRepository) (*AIService, error) {
	keyCredential := azcore.NewKeyCredential(cfg.OpenAIKey)
	client, err := azopenai.NewClientWithKeyCredential(cfg.OpenAIEndpoint, keyCredential, nil)
	if err != nil {
		return nil, fmt.Errorf("create azure openai client: %w", err)
	}
	return &AIService{
		client:     client,
		deployment: cfg.DeploymentName,
		redis:      rdb,
		carRepo:    carRepo,
	}, nil
}

// Recommend returns AI-generated car recommendations based on a natural language query.
// Results are cached in Redis using the query as the key.
func (s *AIService) Recommend(ctx context.Context, req models.AIQueryRequest, availableCars []models.Car) (*AIResponse, error) {
	cacheKey := cachePrefix + hashQuery(req.Query)

	// Check Redis cache first
	if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var resp AIResponse
		if json.Unmarshal([]byte(cached), &resp) == nil {
			log.WithField("cache_hit", true).Info("AI recommendation served from cache")
			return &resp, nil
		}
	}

	// Build context message with available cars
	carContext := buildCarContext(availableCars)
	userMessage := fmt.Sprintf("Customer query: %s\n\nAvailable cars:\n%s", req.Query, carContext)
	if req.Context != "" {
		userMessage += "\n\nAdditional context: " + req.Context
	}

	start := time.Now()
	resp, err := s.client.GetChatCompletions(ctx, azopenai.ChatCompletionsOptions{
		DeploymentName: &s.deployment,
		Messages: []azopenai.ChatRequestMessageClassification{
			&azopenai.ChatRequestSystemMessage{Content: azopenai.NewChatRequestSystemMessageContent(systemPrompt)},
			&azopenai.ChatRequestUserMessage{Content: azopenai.NewChatRequestUserMessageContent(userMessage)},
		},
		ResponseFormat: &azopenai.ChatCompletionsJSONResponseFormat{},
		MaxTokens:      toPtr(int32(800)),
		Temperature:    toPtr(float32(0.3)),
	}, nil)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return nil, fmt.Errorf("azure openai chat completions: %w", err)
	}

	rawContent := *resp.Choices[0].Message.Content
	tokensUsed := int(*resp.Usage.TotalTokens)

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(rawContent), &aiResp); err != nil {
		// Fallback: return raw message if JSON parse fails
		aiResp = AIResponse{Message: rawContent}
	}
	aiResp.LatencyMs = latency
	aiResp.TokensUsed = tokensUsed

	// Cache the result
	if b, err := json.Marshal(aiResp); err == nil {
		s.redis.Set(ctx, cacheKey, string(b), cacheTTL)
	}

	log.WithFields(log.Fields{
		"latency_ms":  latency,
		"tokens_used": tokensUsed,
		"cache_hit":   false,
	}).Info("AI recommendation generated")

	return &aiResp, nil
}

type AIResponse struct {
	Message            string `json:"message"`
	RecommendedCarIDs  []uint `json:"recommended_car_ids"`
	LatencyMs          int    `json:"latency_ms,omitempty"`
	TokensUsed         int    `json:"tokens_used,omitempty"`
}

func buildCarContext(cars []models.Car) string {
	var sb strings.Builder
	for _, c := range cars {
		sb.WriteString(fmt.Sprintf(
			"ID:%d | %d %s %s | %s | $%.2f/day | %d seats | %s | %s\n",
			c.ID, c.Year, c.Make, c.Model, c.Category, c.DailyRate,
			c.Seats, c.Transmission, c.Location,
		))
	}
	return sb.String()
}

func hashQuery(q string) string {
	// Simple normalisation — lowercase + trim
	return strings.ToLower(strings.TrimSpace(q))
}

func toPtr[T any](v T) *T { return &v }
