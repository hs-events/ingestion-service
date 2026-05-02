package storage

import (
	"database/sql"
	"time"

	"ingestion-service/internal/logger"
	"ingestion-service/internal/models"

	_ "github.com/lib/pq"
)

var DB *sql.DB

type eventRecord struct {
	event            models.DeliveryEvent
	platformToken    string
	validationStatus string
	receivedAt       int64
}

var eventQueue = make(chan eventRecord, 10000)

func StartWorker() {
	go func() {
		for record := range eventQueue {
			writeEvent(record)
		}
	}()
}

func writeEvent(record eventRecord) {
	_, err := DB.Exec(`
		INSERT INTO events (order_id, event_type, event_timestamp, received_at,
		                    customer_id, restaurant_id, driver_id, location_lat, location_lng,
		                    platform_token, validation_status, validation_error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, '')
	`,
		record.event.OrderID,
		record.event.EventType,
		record.event.EventTimestamp.Unix(),
		record.receivedAt,
		record.event.CustomerID,
		record.event.RestaurantID,
		record.event.DriverID,
		record.event.Location.Lat,
		record.event.Location.Lng,
		record.platformToken,
		record.validationStatus,
	)
	if err != nil {
		logger.Error("failed to write event", map[string]interface{}{"error": err.Error()})
	}
}

// StoreEvent queues an event for async writing and returns immediately
func StoreEvent(event models.DeliveryEvent, platformToken, validationStatus string) error {
	record := eventRecord{
		event:            event,
		platformToken:    platformToken,
		validationStatus: validationStatus,
		receivedAt:       time.Now().Unix(),
	}

	select {
	case eventQueue <- record:
	default:
		logger.Error("event queue full, dropping event", nil)
	}

	return nil
}

// QueryEvents retrieves events from the database
func QueryEvents(limit int, filtersStr string) ([]models.StoredEvent, error) {
	rows, err := DB.Query(`
		SELECT id, order_id, event_type, event_timestamp, received_at,
		       customer_id, restaurant_id, driver_id, location_lat, location_lng,
		       platform_token, validation_status, validation_error
		FROM events
		`+filtersStr+`
		ORDER BY received_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.StoredEvent
	for rows.Next() {
		var e models.StoredEvent
		var validationError sql.NullString
		if err := rows.Scan(&e.ID, &e.OrderID, &e.EventType, &e.EventTimestamp, &e.ReceivedAt,
			&e.CustomerID, &e.RestaurantID, &e.DriverID, &e.LocationLat, &e.LocationLng,
			&e.PlatformToken, &e.ValidationStatus, &validationError); err != nil {
			logger.Error("failed to scan row", map[string]interface{}{"error": err.Error()})
			continue
		}
		if validationError.Valid {
			e.ValidationError = validationError.String
		}
		events = append(events, e)
	}

	return events, nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func InitDatabase(databaseURL string) error {
	var err error
	DB, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	logger.Info("database connection established", nil)
	return nil
}
