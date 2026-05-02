package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ingestion-service/internal/logger"
	"ingestion-service/internal/models"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDatabase(databaseURL string) error {
	var err error
	DB, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}

	DB.SetMaxOpenConns(50)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)
	DB.SetConnMaxIdleTime(1 * time.Minute)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := DB.PingContext(pingCtx); err != nil {
		return err
	}

	logger.Info("database connection established", nil)
	return nil
}

func StoreEvent(ctx context.Context, event models.DeliveryEvent, platformToken, validationStatus string) error {
	_, err := DB.ExecContext(ctx, `
		INSERT INTO events (
			order_id, event_type, event_timestamp, received_at,
			customer_id, restaurant_id, driver_id, location_lat, location_lng,
			platform_token, validation_status, validation_error
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, '')
	`,
		event.OrderID, event.EventType, event.EventTimestamp.Unix(), time.Now().Unix(),
		event.CustomerID, event.RestaurantID, event.DriverID,
		event.Location.Lat, event.Location.Lng,
		platformToken, validationStatus,
	)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}

func QueryEvents(ctx context.Context, limit int) ([]models.StoredEvent, error) {
	rows, err := DB.QueryContext(ctx, `
		SELECT id, order_id, event_type, event_timestamp, received_at,
		       customer_id, restaurant_id, driver_id, location_lat, location_lng,
		       platform_token, validation_status, validation_error
		FROM events
		ORDER BY received_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]models.StoredEvent, 0, limit)
	for rows.Next() {
		var e models.StoredEvent
		var validationError sql.NullString
		if err := rows.Scan(
			&e.ID, &e.OrderID, &e.EventType, &e.EventTimestamp, &e.ReceivedAt,
			&e.CustomerID, &e.RestaurantID, &e.DriverID, &e.LocationLat, &e.LocationLng,
			&e.PlatformToken, &e.ValidationStatus, &validationError,
		); err != nil {
			return nil, err
		}
		if validationError.Valid {
			e.ValidationError = validationError.String
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
