package models

import "time"

// Location represents geographic coordinates
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// DeliveryEvent represents an incoming order lifecycle event
type DeliveryEvent struct {
	OrderID        string    `json:"order_id"`
	EventType      string    `json:"event_type"`
	EventTimestamp time.Time `json:"event_timestamp"`
	CustomerID     string    `json:"customer_id"`
	RestaurantID   string    `json:"restaurant_id"`
	DriverID       string    `json:"driver_id"`
	Location       Location  `json:"location"`
}

// StoredEvent represents what we store in the database
type StoredEvent struct {
	ID               int     `json:"id"`
	OrderID          string  `json:"order_id"`
	EventType        string  `json:"event_type"`
	EventTimestamp   int64   `json:"event_timestamp"`
	ReceivedAt       int64   `json:"received_at"`
	CustomerID       string  `json:"customer_id"`
	RestaurantID     string  `json:"restaurant_id"`
	DriverID         string  `json:"driver_id"`
	LocationLat      float64 `json:"location_lat"`
	LocationLng      float64 `json:"location_lng"`
	PlatformToken    string  `json:"platform_token"`
	ValidationStatus string  `json:"validation_status"`
	ValidationError  string  `json:"validation_error,omitempty"`
}
