package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var ControlServerURL string

var httpClient = &http.Client{
	Timeout: 2 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	},
}

// FetchPlatformTokens returns valid tokens as a set for O(1) lookup.
func FetchPlatformTokens(ctx context.Context) (map[string]struct{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ControlServerURL+"/platform-tokens", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("control server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("control server returned %d", resp.StatusCode)
	}

	var result struct {
		PlatformTokens []string `json:"platform_tokens"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	set := make(map[string]struct{}, len(result.PlatformTokens))
	for _, t := range result.PlatformTokens {
		set[t] = struct{}{}
	}
	return set, nil
}

func ValidatePlatformToken(token string, validTokens map[string]struct{}) bool {
	_, ok := validTokens[token]
	return ok
}
