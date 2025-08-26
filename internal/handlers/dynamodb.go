package handlers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/weather-lambda/internal/config"
	"github.com/weather-lambda/internal/models"
)

// DynamoDBHandler handles DynamoDB operations
type DynamoDBHandler struct {
	client    *dynamodb.DynamoDB
	tableName string
}

// NewDynamoDBHandler creates a new DynamoDB handler
func NewDynamoDBHandler(cfg *config.Config, sess ...*session.Session) (*DynamoDBHandler, error) {
	var sess_client *session.Session
	var err error
	
	if len(sess) > 0 && sess[0] != nil {
		sess_client = sess[0]
	} else {
		sess_client, err = session.NewSession(&aws.Config{
			Region: aws.String(cfg.AWS.Region),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS session: %w", err)
		}
	}

	return &DynamoDBHandler{
		client:    dynamodb.New(sess_client),
		tableName: cfg.AWS.DynamoDBTable,
	}, nil
}

// StoreWeatherRecord stores a weather record to DynamoDB
func (h *DynamoDBHandler) StoreWeatherRecord(record *models.WeatherRecord) error {
	// Convert the record to DynamoDB attribute value map
	av, err := dynamodbattribute.MarshalMap(record)
	if err != nil {
		return fmt.Errorf("failed to marshal weather record: %w", err)
	}

	// Put the item into DynamoDB
	_, err = h.client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(h.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put item to DynamoDB: %w", err)
	}

	return nil
}

// GetWeatherRecord retrieves a weather record from DynamoDB
func (h *DynamoDBHandler) GetWeatherRecord(id, timestamp string) (*models.WeatherRecord, error) {
	result, err := h.client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(h.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
			"timestamp": {
				S: aws.String(timestamp),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get item from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("weather record not found")
	}

	var record models.WeatherRecord
	err = dynamodbattribute.UnmarshalMap(result.Item, &record)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal weather record: %w", err)
	}

	return &record, nil
}

// QueryWeatherRecordsByCity queries weather records by city name
func (h *DynamoDBHandler) QueryWeatherRecordsByCity(cityName string, limit int64) ([]*models.WeatherRecord, error) {
	// This would require a GSI (Global Secondary Index) on cityName
	// For now, we'll use scan with filter (not recommended for production)
	input := &dynamodb.ScanInput{
		TableName:        aws.String(h.tableName),
		FilterExpression: aws.String("cityName = :city"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":city": {
				S: aws.String(cityName),
			},
		},
	}

	if limit > 0 {
		input.Limit = aws.Int64(limit)
	}

	result, err := h.client.Scan(input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan DynamoDB table: %w", err)
	}

	var records []*models.WeatherRecord
	for _, item := range result.Items {
		var record models.WeatherRecord
		err := dynamodbattribute.UnmarshalMap(item, &record)
		if err != nil {
			continue // Skip invalid records
		}
		records = append(records, &record)
	}

	return records, nil
}

// QueryRecentWeatherRecords queries recent weather records
func (h *DynamoDBHandler) QueryRecentWeatherRecords(id string, limit int64) ([]*models.WeatherRecord, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(h.tableName),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id": {
				S: aws.String(id),
			},
		},
		ScanIndexForward: aws.Bool(false), // Sort by timestamp descending (most recent first)
	}

	if limit > 0 {
		input.Limit = aws.Int64(limit)
	}

	result, err := h.client.Query(input)
	if err != nil {
		return nil, fmt.Errorf("failed to query DynamoDB table: %w", err)
	}

	var records []*models.WeatherRecord
	for _, item := range result.Items {
		var record models.WeatherRecord
		err := dynamodbattribute.UnmarshalMap(item, &record)
		if err != nil {
			continue // Skip invalid records
		}
		records = append(records, &record)
	}

	return records, nil
}

// GetWeatherHistory retrieves weather records for a specific city within a time range
func (h *DynamoDBHandler) GetWeatherHistory(ctx context.Context, cityName string, startTime, endTime time.Time) ([]models.WeatherRecord, error) {
	// Since we're using city-timestamp as the ID structure, we need to scan with filters
	// In production, consider adding a GSI with cityName as partition key and timestamp as sort key
	
	startTimeStr := startTime.Format(time.RFC3339)
	endTimeStr := endTime.Format(time.RFC3339)
	
	input := &dynamodb.ScanInput{
		TableName:        aws.String(h.tableName),
		FilterExpression: aws.String("cityName = :city AND #ts BETWEEN :start AND :end"),
		ExpressionAttributeNames: map[string]*string{
			"#ts": aws.String("timestamp"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":city": {
				S: aws.String(cityName),
			},
			":start": {
				S: aws.String(startTimeStr),
			},
			":end": {
				S: aws.String(endTimeStr),
			},
		},
	}

	var records []models.WeatherRecord
	err := h.client.ScanPagesWithContext(ctx, input, func(page *dynamodb.ScanOutput, lastPage bool) bool {
		for _, item := range page.Items {
			var record models.WeatherRecord
			err := dynamodbattribute.UnmarshalMap(item, &record)
			if err != nil {
				continue // Skip invalid records
			}
			records = append(records, record)
		}
		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan weather history: %w", err)
	}

	// Sort records by timestamp (oldest first)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp < records[j].Timestamp
	})

	return records, nil
}

// GetWeatherHistoryByCityID retrieves weather records for a specific city ID within a time range
// This is more efficient if you know the exact city ID format
func (h *DynamoDBHandler) GetWeatherHistoryByCityID(ctx context.Context, cityID string, startTime, endTime time.Time) ([]models.WeatherRecord, error) {
	startTimeStr := startTime.Format(time.RFC3339)
	endTimeStr := endTime.Format(time.RFC3339)
	
	input := &dynamodb.QueryInput{
		TableName:              aws.String(h.tableName),
		KeyConditionExpression: aws.String("id = :id AND #ts BETWEEN :start AND :end"),
		ExpressionAttributeNames: map[string]*string{
			"#ts": aws.String("timestamp"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id": {
				S: aws.String(cityID),
			},
			":start": {
				S: aws.String(startTimeStr),
			},
			":end": {
				S: aws.String(endTimeStr),
			},
		},
		ScanIndexForward: aws.Bool(true), // Sort by timestamp ascending (oldest first)
	}

	var records []models.WeatherRecord
	err := h.client.QueryPagesWithContext(ctx, input, func(page *dynamodb.QueryOutput, lastPage bool) bool {
		for _, item := range page.Items {
			var record models.WeatherRecord
			err := dynamodbattribute.UnmarshalMap(item, &record)
			if err != nil {
				continue // Skip invalid records
			}
			records = append(records, record)
		}
		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query weather history: %w", err)
	}

	return records, nil
}