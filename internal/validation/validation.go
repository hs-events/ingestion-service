package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var ControlServerURL string

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

var (
	tokenCache    []string
	tokenCachedAt time.Time
	tokenMu       sync.RWMutex
	cacheTTL      = 30 * time.Second
)

func FetchPlatformTokens() ([]string, error) {
	tokenMu.RLock()
	if tokenCache != nil && time.Since(tokenCachedAt) < cacheTTL {
		tokens := tokenCache
		tokenMu.RUnlock()
		return tokens, nil
	}
	tokenMu.RUnlock()

	tokenMu.Lock()
	defer tokenMu.Unlock()

	// Re-check after acquiring write lock (another goroutine may have refreshed)
	if tokenCache != nil && time.Since(tokenCachedAt) < cacheTTL {
		return tokenCache, nil
	}

	resp, err := httpClient.Get(ControlServerURL + "/platform-tokens")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch platform tokens: %w", err)
	}
	defer resp.Body.Close()

	var result map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	tokens, ok := result["platform_tokens"]
	if !ok {
		return nil, fmt.Errorf("platform_tokens not found in response")
	}

	tokenCache = tokens
	tokenCachedAt = time.Now()

	return tokens, nil
}

func ValidatePlatformToken(token string, validTokens []string) bool {
	for _, validToken := range validTokens {
		if validToken == token {
			return true
		}
	}
	return false
}
