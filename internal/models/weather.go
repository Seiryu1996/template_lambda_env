package models

import "time"

// WeatherResponse represents the weather API response
type WeatherResponse struct {
	Name   string `json:"name"`
	Coord  Coord  `json:"coord"`
	Main   Main   `json:"main"`
	Weather []Weather `json:"weather"`
	Wind   Wind   `json:"wind"`
	Clouds Clouds `json:"clouds"`
	Sys    Sys    `json:"sys"`
	Dt     int64  `json:"dt"`
}

// Coord represents coordinates
type Coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

// Main represents main weather data
type Main struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Pressure  int     `json:"pressure"`
	Humidity  int     `json:"humidity"`
}

// Weather represents weather condition
type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// Wind represents wind data
type Wind struct {
	Speed float64 `json:"speed"`
	Deg   int     `json:"deg"`
}

// Clouds represents cloud data
type Clouds struct {
	All int `json:"all"`
}

// Sys represents system data
type Sys struct {
	Country string `json:"country"`
	Sunrise int64  `json:"sunrise"`
	Sunset  int64  `json:"sunset"`
}

// WeatherRecord represents data to be stored in DynamoDB
type WeatherRecord struct {
	ID          string    `json:"id" dynamodbav:"id"`
	Timestamp   string    `json:"timestamp" dynamodbav:"timestamp"`
	CityName    string    `json:"cityName" dynamodbav:"cityName"`
	Temperature float64   `json:"temperature" dynamodbav:"temperature"`
	Description string    `json:"description" dynamodbav:"description"`
	Humidity    int       `json:"humidity" dynamodbav:"humidity"`
	Pressure    int       `json:"pressure" dynamodbav:"pressure"`
	WindSpeed   float64   `json:"windSpeed" dynamodbav:"windSpeed"`
	Country     string    `json:"country" dynamodbav:"country"`
	CreatedAt   time.Time `json:"createdAt" dynamodbav:"createdAt"`
	TTL         int64     `json:"ttl" dynamodbav:"ttl"` // Time to live (30 days from creation)
}

// S3WeatherData represents data to be stored in S3
type S3WeatherData struct {
	WeatherRecord
	RawResponse WeatherResponse `json:"rawResponse"`
}