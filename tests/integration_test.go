//go:build integration

package tests

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/weather-lambda/internal/config"
	"github.com/weather-lambda/internal/handlers"
	"github.com/weather-lambda/internal/models"
	"github.com/weather-lambda/internal/services"
)

func TestAWSIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping AWS integration test")
	}

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create AWS session (uses real AWS resources)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.AWS.Region),
	})
	if err != nil {
		t.Fatalf("Failed to create AWS session: %v", err)
	}

	// Initialize handlers
	s3Handler := handlers.NewS3Handler(cfg, sess)
	dynamoDBHandler := handlers.NewDynamoDBHandler(cfg, sess)

	t.Run("TestS3Handler_AWS", func(t *testing.T) {
		// Create test data
		now := time.Now()
		testRecord := &models.WeatherRecord{
			ID:          "test-integration-" + now.Format("20060102-150405"),
			Timestamp:   now.Format(time.RFC3339),
			CityName:    "Tokyo",
			Temperature: 25.5,
			Description: "Integration test data",
			Humidity:    60,
			Pressure:    1013,
			WindSpeed:   3.5,
			Country:     "JP",
			CreatedAt:   now,
			TTL:         now.Add(24 * time.Hour).Unix(),
		}

		testS3Data := &models.S3WeatherData{
			WeatherRecord: *testRecord,
			RawResponse: models.WeatherResponse{
				Name: "Tokyo",
				Main: models.Main{
					Temp:     25.5,
					Humidity: 60,
					Pressure: 1013,
				},
				Weather: []models.Weather{
					{Description: "Integration test data"},
				},
			},
		}

		// Test storing data to S3
		err := s3Handler.StoreWeatherData(testS3Data)
		if err != nil {
			t.Errorf("Failed to store data to S3: %v", err)
		}
		t.Logf("Successfully stored test data to S3 with ID: %s", testRecord.ID)
	})

	t.Run("TestDynamoDBHandler_AWS", func(t *testing.T) {
		// Create test record
		now := time.Now()
		testRecord := &models.WeatherRecord{
			ID:          "test-integration-" + now.Format("20060102-150405"),
			Timestamp:   now.Format(time.RFC3339),
			CityName:    "Tokyo",
			Temperature: 23.0,
			Description: "Integration test data",
			Humidity:    65,
			Pressure:    1015,
			WindSpeed:   2.8,
			Country:     "JP",
			CreatedAt:   now,
			TTL:         now.Add(24 * time.Hour).Unix(),
		}

		// Test storing record to DynamoDB
		err := dynamoDBHandler.StoreWeatherRecord(testRecord)
		if err != nil {
			t.Errorf("Failed to store record to DynamoDB: %v", err)
		}
		t.Logf("Successfully stored test record to DynamoDB with ID: %s", testRecord.ID)

		// Test retrieving record from DynamoDB
		retrieved, err := dynamoDBHandler.GetWeatherRecord(testRecord.ID, testRecord.Timestamp)
		if err != nil {
			t.Errorf("Failed to retrieve record from DynamoDB: %v", err)
			return
		}

		// Verify retrieved data
		if retrieved.ID != testRecord.ID {
			t.Errorf("Expected ID %s, got %s", testRecord.ID, retrieved.ID)
		}
		if retrieved.Temperature != testRecord.Temperature {
			t.Errorf("Expected temperature %f, got %f", testRecord.Temperature, retrieved.Temperature)
		}
		t.Logf("Successfully retrieved and verified test record from DynamoDB")
	})
}

func TestWeatherService(t *testing.T) {
	// Unit test - no AWS resources needed
	cfg := &config.Config{
		Weather: config.WeatherConfig{
			APIKey:   "test-key",
			APIURL:   "https://api.openweathermap.org/data/2.5/weather",
			CityName: "Tokyo",
		},
	}

	service := services.NewWeatherService(cfg)

	t.Run("TestConvertToWeatherRecord", func(t *testing.T) {
		// Mock weather response
		mockResponse := &models.WeatherResponse{
			Name: "Tokyo",
			Main: models.Main{
				Temp:     25.5,
				Humidity: 60,
				Pressure: 1013,
			},
			Weather: []models.Weather{
				{Description: "Clear sky"},
			},
			Wind: models.Wind{
				Speed: 3.5,
			},
			Sys: models.Sys{
				Country: "JP",
			},
		}

		record := service.ConvertToWeatherRecord(mockResponse)

		if record.CityName != "Tokyo" {
			t.Errorf("Expected city name Tokyo, got %s", record.CityName)
		}
		if record.Temperature != 25.5 {
			t.Errorf("Expected temperature 25.5, got %f", record.Temperature)
		}
		if record.Description != "Clear sky" {
			t.Errorf("Expected description 'Clear sky', got %s", record.Description)
		}
	})
}