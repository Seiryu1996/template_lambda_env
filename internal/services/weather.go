package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/weather-lambda/internal/config"
	"github.com/weather-lambda/internal/models"
)

// WeatherService handles weather API interactions
type WeatherService struct {
	config *config.Config
	client *http.Client
}

// NewWeatherService creates a new weather service
func NewWeatherService(cfg *config.Config) *WeatherService {
	return &WeatherService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetWeatherData fetches weather data from the API
func (w *WeatherService) GetWeatherData() (*models.WeatherResponse, error) {
	baseURL, err := url.Parse(w.config.Weather.APIURL)
	if err != nil {
		return nil, fmt.Errorf("invalid weather API URL: %w", err)
	}

	params := url.Values{}
	params.Add("q", w.config.Weather.CityName)
	params.Add("appid", w.config.Weather.APIKey)
	params.Add("units", "metric") // Celsius temperature
	baseURL.RawQuery = params.Encode()

	resp, err := w.client.Get(baseURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to make weather API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("weather API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read weather API response: %w", err)
	}

	var weatherResponse models.WeatherResponse
	if err := json.Unmarshal(body, &weatherResponse); err != nil {
		return nil, fmt.Errorf("failed to parse weather API response: %w", err)
	}

	return &weatherResponse, nil
}

// ConvertToWeatherRecord converts API response to DynamoDB record
func (w *WeatherService) ConvertToWeatherRecord(response *models.WeatherResponse) *models.WeatherRecord {
	now := time.Now()
	ttl := now.Add(30 * 24 * time.Hour).Unix() // 30 days TTL

	record := &models.WeatherRecord{
		ID:        fmt.Sprintf("%s-%d", response.Name, now.Unix()),
		Timestamp: now.Format(time.RFC3339),
		CityName:  response.Name,
		Humidity:  response.Main.Humidity,
		Pressure:  response.Main.Pressure,
		Country:   response.Sys.Country,
		CreatedAt: now,
		TTL:       ttl,
	}

	// Set temperature
	record.Temperature = response.Main.Temp

	// Set weather description
	if len(response.Weather) > 0 {
		record.Description = response.Weather[0].Description
	}

	// Set wind speed
	record.WindSpeed = response.Wind.Speed

	return record
}

// ConvertToS3Data converts API response to S3 storage format
func (w *WeatherService) ConvertToS3Data(response *models.WeatherResponse, record *models.WeatherRecord) *models.S3WeatherData {
	return &models.S3WeatherData{
		WeatherRecord: *record,
		RawResponse:   *response,
	}
}