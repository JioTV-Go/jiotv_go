package handlers

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

const (
	jwtTokenRefreshLeadTime         = 30 * time.Second
	accessTokenFallbackTTL          = 2 * time.Hour
	accessTokenFallbackLeadTime     = 10 * time.Minute
	ssoTokenFallbackTTL             = 24 * time.Hour
	ssoTokenFallbackLeadTime        = 1 * time.Hour
	minCredentialValidationInterval = 5 * time.Second
	credentialRefreshRetryBackoff   = 20 * time.Second
)

func parseJWTExpiry(token string) (time.Time, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, false
	}

	payload, err := decodeJWTComponent(parts[1])
	if err != nil {
		return time.Time{}, false
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return time.Time{}, false
	}

	expRaw, exists := claims["exp"]
	if !exists {
		return time.Time{}, false
	}

	expUnix, ok := normalizeUnixTimestamp(expRaw)
	if !ok || expUnix <= 0 {
		return time.Time{}, false
	}

	return time.Unix(expUnix, 0), true
}

func decodeJWTComponent(component string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(component)
	if err == nil {
		return decoded, nil
	}

	if remainder := len(component) % 4; remainder != 0 {
		component += strings.Repeat("=", 4-remainder)
	}
	return base64.URLEncoding.DecodeString(component)
}

func normalizeUnixTimestamp(value any) (int64, bool) {
	switch v := value.(type) {
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	case int64:
		return v, true
	case int32:
		return int64(v), true
	case int:
		return int64(v), true
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i, true
		}
		f, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return 0, false
		}
		return int64(f), true
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}

func fallbackTokenRefreshThreshold(lastRefresh string, fallbackTTL, fallbackLead time.Duration) (time.Time, bool) {
	if lastRefresh == "" {
		return time.Time{}, false
	}

	lastRefreshUnix, err := strconv.ParseInt(lastRefresh, 10, 64)
	if err != nil {
		return time.Time{}, false
	}

	return time.Unix(lastRefreshUnix, 0).Add(fallbackTTL - fallbackLead), true
}

func shouldRefreshToken(token, lastRefresh string, jwtLead, fallbackTTL, fallbackLead time.Duration, now time.Time) bool {
	if token == "" {
		return true
	}

	if expiry, ok := parseJWTExpiry(token); ok {
		return !expiry.After(now.Add(jwtLead))
	}

	threshold, ok := fallbackTokenRefreshThreshold(lastRefresh, fallbackTTL, fallbackLead)
	if !ok {
		return true
	}

	return !threshold.After(now)
}

func nextTokenValidationCheck(token, lastRefresh string, jwtLead, fallbackTTL, fallbackLead time.Duration, now time.Time) (time.Time, bool) {
	minCheckTime := now.Add(minCredentialValidationInterval)

	if token != "" {
		if expiry, ok := parseJWTExpiry(token); ok {
			nextCheck := expiry.Add(-jwtLead)
			if nextCheck.Before(minCheckTime) {
				nextCheck = minCheckTime
			}
			return nextCheck, true
		}
	}

	threshold, ok := fallbackTokenRefreshThreshold(lastRefresh, fallbackTTL, fallbackLead)
	if !ok {
		return time.Time{}, false
	}

	if threshold.Before(minCheckTime) {
		threshold = minCheckTime
	}

	return threshold, true
}
