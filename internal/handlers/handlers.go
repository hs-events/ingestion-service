package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ingestion-service/internal/logger"
	"ingestion-service/internal/models"
	"ingestion-service/internal/storage"
	"ingestion-service/internal/validation"
)

func DeliveryEventsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handlePostDeliveryEvent(w, r)
	case http.MethodGet:
		handleGetDeliveryEvents(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handlePostDeliveryEvent(w http.ResponseWriter, r *http.Request) {
	var event models.DeliveryEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	platformToken := r.Header.Get("X-Platform-Token")
	if platformToken == "" {
		http.Error(w, "missing X-Platform-Token", http.StatusUnauthorized)
		return
	}

	validTokens, err := validation.FetchPlatformTokens(r.Context())
	if err != nil {
		logger.Error("control server fetch failed", map[string]interface{}{"error": err.Error()})
		http.Error(w, "validation unavailable", http.StatusServiceUnavailable)
		return
	}

	if !validation.ValidatePlatformToken(platformToken, validTokens) {
		http.Error(w, "invalid platform token", http.StatusUnauthorized)
		return
	}

	if err := storage.StoreEvent(r.Context(), event, platformToken, "valid"); err != nil {
		logger.Error("store event failed", map[string]interface{}{"error": err.Error()})
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleGetDeliveryEvents(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	events, err := storage.QueryEvents(r.Context(), limit)
	if err != nil {
		logger.Error("failed to query events", map[string]interface{}{"error": err.Error()})
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
