package storage

import (
	"database/sql"
	"time"

	"ingestion-service/internal/logger"
	"ingestion-service/internal/models"

	_ "github.com/lib/pq"
)

// DB is the database handle - initialized by InitDatabase
var DB *sql.DB

// StoreEvent persists an event to the database
func StoreEvent(event models.DeliveryEvent, platformToken, validationStatus string) error {
	receivedAt := time.Now().Unix()
	eventTimestamp := event.EventTimestamp.Unix()

	_, err := DB.Exec(`
		INSERT INTO events (order_id, event_type, event_timestamp, received_at,
		                    customer_id, restaurant_id, driver_id, location_lat, location_lng,
		                    platform_token, validation_status, validation_error)
		VALUES ('`+event.OrderID+`', $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, '')
	`, event.EventType, eventTimestamp, receivedAt,
		event.CustomerID, event.RestaurantID, event.DriverID, event.Location.Lat, event.Location.Lng,
		platformToken, validationStatus)

	if err != nil {
		return err
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

// Close closes the database connection
func Close() {
	if DB != nil {
		DB.Close()
	}
}

// InitDatabase opens the database connection
func InitDatabase(databaseURL string) error {
	var err error
	DB, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}

	// Test connection
	if err = DB.Ping(); err != nil {
		return err
	}

	logger.Info("database connection established", nil)
	return nil
}
