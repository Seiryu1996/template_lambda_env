package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/weather-lambda/internal/config"
	"github.com/weather-lambda/internal/handlers"
	"github.com/weather-lambda/internal/services"
)

// Event represents the input to the Lambda function
type Event struct {
	Detail string `json:"detail,omitempty"`
}

// Response represents the output from the Lambda function
type Response struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Data       interface{} `json:"data,omitempty"`
}

// Handler is the Lambda function handler
type Handler struct {
	weatherService  *services.WeatherService
	s3Handler       *handlers.S3Handler
	dynamoDBHandler *handlers.DynamoDBHandler
	config          *config.Config
}

// NewHandler creates a new handler instance
func NewHandler() (*Handler, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate required configuration
	if cfg.Weather.APIKey == "" {
		return nil, fmt.Errorf("WEATHER_API_KEY environment variable is required")
	}
	if cfg.AWS.S3Bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET environment variable is required")
	}
	if cfg.AWS.DynamoDBTable == "" {
		return nil, fmt.Errorf("DYNAMODB_TABLE environment variable is required")
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWS.Region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Initialize services and handlers
	weatherService := services.NewWeatherService(cfg)
	s3Handler := handlers.NewS3Handler(cfg, sess)
	dynamoDBHandler, err := handlers.NewDynamoDBHandler(cfg, sess)
	if err != nil {
		return nil, fmt.Errorf("failed to create DynamoDB handler: %w", err)
	}

	return &Handler{
		weatherService:  weatherService,
		s3Handler:       s3Handler,
		dynamoDBHandler: dynamoDBHandler,
		config:          cfg,
	}, nil
}

// HandleRequest handles the Lambda function request
func (h *Handler) HandleRequest(ctx context.Context, event Event) (*Response, error) {
	log.Printf("Processing weather data collection for city: %s", h.config.Weather.CityName)

	// Fetch weather data from API
	weatherResponse, err := h.weatherService.GetWeatherData()
	if err != nil {
		log.Printf("Error fetching weather data: %v", err)
		return &Response{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to fetch weather data: %v", err),
		}, nil
	}

	log.Printf("Successfully fetched weather data for %s: %.2f°C, %s",
		weatherResponse.Name,
		weatherResponse.Main.Temp,
		weatherResponse.Weather[0].Description,
	)

	// Convert to internal models
	weatherRecord := h.weatherService.ConvertToWeatherRecord(weatherResponse)
	s3Data := h.weatherService.ConvertToS3Data(weatherResponse, weatherRecord)

	// Store to DynamoDB
	if err := h.dynamoDBHandler.StoreWeatherRecord(weatherRecord); err != nil {
		log.Printf("Error storing to DynamoDB: %v", err)
		return &Response{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to store to DynamoDB: %v", err),
		}, nil
	}
	log.Printf("Successfully stored weather record to DynamoDB: %s", weatherRecord.ID)

	// Store to S3
	if err := h.s3Handler.StoreWeatherData(s3Data); err != nil {
		log.Printf("Error storing to S3: %v", err)
		return &Response{
			StatusCode: 500,
			Message:    fmt.Sprintf("Failed to store to S3: %v", err),
		}, nil
	}
	log.Printf("Successfully stored weather data to S3 for record: %s", weatherRecord.ID)

	return &Response{
		StatusCode: 200,
		Message:    "Weather data processed successfully",
		Data: map[string]interface{}{
			"city":        weatherRecord.CityName,
			"temperature": weatherRecord.Temperature,
			"description": weatherRecord.Description,
			"timestamp":   weatherRecord.Timestamp,
			"recordId":    weatherRecord.ID,
		},
	}, nil
}

func main() {
	// Initialize handler
	handler, err := NewHandler()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to initialize handler: %v", err))
	}

	// Start Lambda function
	lambda.Start(handler.HandleRequest)
}