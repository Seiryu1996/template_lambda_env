package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/weather-lambda/internal/config"
	"github.com/weather-lambda/internal/handlers"
	"github.com/weather-lambda/internal/models"
)

type Handler struct {
	dynamoHandler *handlers.DynamoDBHandler
	config        *config.Config
}

type WeatherHistoryResponse struct {
	StatusCode int                    `json:"statusCode"`
	Message    string                 `json:"message"`
	Data       []models.WeatherRecord `json:"data"`
	Count      int                    `json:"count"`
	Period     string                 `json:"period"`
	StartTime  string                 `json:"startTime"`
	EndTime    string                 `json:"endTime"`
}

func NewHandler() (*Handler, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	dynamoHandler, err := handlers.NewDynamoDBHandler(cfg)
	if err != nil {
		return nil, err
	}

	return &Handler{
		dynamoHandler: dynamoHandler,
		config:        cfg,
	}, nil
}

func (h *Handler) HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// CORS headers
	headers := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
		"Access-Control-Allow-Methods": "GET,OPTIONS",
		"Content-Type":                 "application/json",
	}

	// Handle OPTIONS request for CORS preflight
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    headers,
			Body:       "",
		}, nil
	}

	// Only allow GET requests
	if request.HTTPMethod != "GET" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Headers:    headers,
			Body:       `{"error": "Method not allowed"}`,
		}, nil
	}

	// Parse query parameters
	period := request.QueryStringParameters["period"]
	if period == "" {
		period = "6h" // Default to 6 hours
	}

	city := request.QueryStringParameters["city"]
	if city == "" {
		city = h.config.Weather.CityName // Use default city from config
	}

	var records []models.WeatherRecord
	var err error
	var startTime, endTime time.Time

	switch period {
	case "6h":
		endTime = time.Now()
		startTime = endTime.Add(-6 * time.Hour)
		records, err = h.dynamoHandler.GetWeatherHistory(ctx, city, startTime, endTime)
	case "24h", "1d":
		endTime = time.Now()
		startTime = endTime.Add(-24 * time.Hour)
		records, err = h.dynamoHandler.GetWeatherHistory(ctx, city, startTime, endTime)
	default:
		// Custom period in hours
		if hours, parseErr := strconv.Atoi(period); parseErr == nil && hours > 0 && hours <= 168 { // Max 7 days
			endTime = time.Now()
			startTime = endTime.Add(-time.Duration(hours) * time.Hour)
			records, err = h.dynamoHandler.GetWeatherHistory(ctx, city, startTime, endTime)
		} else {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Headers:    headers,
				Body:       `{"error": "Invalid period. Use '6h', '24h', or number of hours (1-168)"}`,
			}, nil
		}
	}

	if err != nil {
		log.Printf("Error getting weather history: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error": "Failed to retrieve weather history"}`,
		}, nil
	}

	response := WeatherHistoryResponse{
		StatusCode: http.StatusOK,
		Message:    "Weather history retrieved successfully",
		Data:       records,
		Count:      len(records),
		Period:     period,
		StartTime:  startTime.Format(time.RFC3339),
		EndTime:    endTime.Format(time.RFC3339),
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error": "Failed to create response"}`,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    headers,
		Body:       string(responseBody),
	}, nil
}

func main() {
	handler, err := NewHandler()
	if err != nil {
		log.Fatalf("Failed to initialize handler: %v", err)
	}

	lambda.Start(handler.HandleRequest)
}