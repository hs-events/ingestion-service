package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ControlServerURL is the URL of the control server
var ControlServerURL string

// Cache for platform tokens
var (
	tokenCache     []string
	cacheTimestamp time.Time
	cacheMutex     sync.RWMutex
	cacheDuration  = 5 * time.Minute // Cache for 5 minutes
)

// FetchPlatformTokens calls the control server to get valid platform tokens with caching
func FetchPlatformTokens() ([]string, error) {
	cacheMutex.RLock()
	if time.Since(cacheTimestamp) < cacheDuration && len(tokenCache) > 0 {
		tokens := make([]string, len(tokenCache))
		copy(tokens, tokenCache)
		cacheMutex.RUnlock()
		return tokens, nil
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if time.Since(cacheTimestamp) < cacheDuration && len(tokenCache) > 0 {
		tokens := make([]string, len(tokenCache))
		copy(tokens, tokenCache)
		return tokens, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(ControlServerURL + "/platform-tokens")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tokens: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("control server returned status %d", resp.StatusCode)
	}

	var result map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	tokens, ok := result["platform_tokens"]
	if !ok {
		return nil, fmt.Errorf("platform_tokens not found in response")
	}

	tokenCache = make([]string, len(tokens))
	copy(tokenCache, tokens)
	cacheTimestamp = time.Now()

	return tokens, nil
}

// ValidatePlatformToken checks if a token exists in the list of valid tokens
func ValidatePlatformToken(token string, validTokens []string) bool {
	for _, validToken := range validTokens {
		if validToken == token {
			return true
		}
	}
	return false
}
