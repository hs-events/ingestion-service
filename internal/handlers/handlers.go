package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"ingestion-service/internal/logger"
	"ingestion-service/internal/models"
	"ingestion-service/internal/storage"
	"ingestion-service/internal/validation"
)

// DeliveryEventsHandler routes between POST (receive events) and GET (query events)
func DeliveryEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		handlePostDeliveryEvent(w, r)
	} else if r.Method == http.MethodGet {
		handleGetDeliveryEvents(w, r)
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePostDeliveryEvent processes incoming delivery events
func handlePostDeliveryEvent(w http.ResponseWriter, r *http.Request) {
	logger.Info("POST /delivery-events requested", nil)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("failed to read request body", map[string]interface{}{"error": err.Error()})
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var event models.DeliveryEvent
	if err := json.Unmarshal(body, &event); err != nil {
		logger.Error("failed to unmarshal event", map[string]interface{}{"error": err.Error()})
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := validateEvent(event); err != nil {
		logger.Error("event validation failed", map[string]interface{}{"error": err.Error()})
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	platformToken := r.Header.Get("X-Platform-Token")
	if platformToken == "" {
		logger.Error("missing platform token", nil)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	validTokens, err := validation.FetchPlatformTokens()
	if err != nil {
		logger.Error("failed to fetch platform tokens", map[string]interface{}{"error": err.Error()})
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if !validation.ValidatePlatformToken(platformToken, validTokens) {
		logger.Error("invalid platform token", map[string]interface{}{"token": platformToken})
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err = storage.StoreEvent(event, platformToken, "valid")
	if err != nil {
		logger.Error("failed to store event", map[string]interface{}{"error": err.Error()})
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleGetDeliveryEvents returns stored events for debugging/inspection
func handleGetDeliveryEvents(w http.ResponseWriter, r *http.Request) {
	logger.Info("GET /delivery-events requested", nil)

	// Get optional limit parameter (default 100)
	limitStr := r.URL.Query().Get("limit")
	filtersStr := r.URL.Query().Get("filters")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	events, err := storage.QueryEvents(limit, filtersStr)
	if err != nil {
		logger.Error("failed to query events", map[string]interface{}{"error": err.Error()})
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	logger.Info("returning events", map[string]interface{}{"count": len(events)})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// HealthHandler returns service health status
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// validateEvent checks if the delivery event has all required fields and valid data
func validateEvent(event models.DeliveryEvent) error {
	if event.OrderID == "" {
		return fmt.Errorf("order_id is required")
	}
	if event.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if event.EventTimestamp.IsZero() {
		return fmt.Errorf("event_timestamp is required and must be valid")
	}
	if event.CustomerID == "" {
		return fmt.Errorf("customer_id is required")
	}
	if event.RestaurantID == "" {
		return fmt.Errorf("restaurant_id is required")
	}
	if event.DriverID == "" {
		return fmt.Errorf("driver_id is required")
	}
	// Validate location coordinates
	if event.Location.Lat < -90 || event.Location.Lat > 90 {
		return fmt.Errorf("location lat must be between -90 and 90")
	}
	if event.Location.Lng < -180 || event.Location.Lng > 180 {
		return fmt.Errorf("location lng must be between -180 and 180")
	}
	return nil
}
