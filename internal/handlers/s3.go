package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/weather-lambda/internal/config"
	"github.com/weather-lambda/internal/models"
)

// S3Handler handles S3 operations
type S3Handler struct {
	client *s3.S3
	bucket string
}

// NewS3Handler creates a new S3 handler
func NewS3Handler(cfg *config.Config, sess *session.Session) *S3Handler {
	return &S3Handler{
		client: s3.New(sess),
		bucket: cfg.AWS.S3Bucket,
	}
}

// StoreWeatherData stores weather data to S3
func (h *S3Handler) StoreWeatherData(data *models.S3WeatherData) error {
	// Convert data to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal weather data: %w", err)
	}

	// Create S3 key with date-based prefix for organization
	now := time.Now()
	key := fmt.Sprintf("weather-data/%s/%s/%s.json",
		now.Format("2006"),
		now.Format("01-02"),
		data.ID,
	)

	// Upload to S3
	_, err = h.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(h.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(jsonData),
		ContentType: aws.String("application/json"),
		Metadata: map[string]*string{
			"city":        aws.String(data.CityName),
			"country":     aws.String(data.Country),
			"temperature": aws.String(fmt.Sprintf("%.2f", data.Temperature)),
			"timestamp":   aws.String(data.Timestamp),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// GetWeatherData retrieves weather data from S3
func (h *S3Handler) GetWeatherData(key string) (*models.S3WeatherData, error) {
	result, err := h.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(h.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	var data models.S3WeatherData
	if err := json.NewDecoder(result.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode weather data: %w", err)
	}

	return &data, nil
}

// ListWeatherData lists weather data files in S3
func (h *S3Handler) ListWeatherData(prefix string) ([]*s3.Object, error) {
	result, err := h.client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(h.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in S3: %w", err)
	}

	return result.Contents, nil
}