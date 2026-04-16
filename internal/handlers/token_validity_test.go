package handlers

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"
)

func buildUnsignedJWT(exp int64) string {
	header := `{"alg":"none","typ":"JWT"}`
	payload := fmt.Sprintf(`{"exp":%d}`, exp)
	return base64.RawURLEncoding.EncodeToString([]byte(header)) + "." +
		base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".signature"
}

func TestParseJWTExpiry(t *testing.T) {
	now := time.Now().Unix()
	token := buildUnsignedJWT(now + 300)

	expiry, ok := parseJWTExpiry(token)
	if !ok {
		t.Fatalf("parseJWTExpiry() should parse valid JWT exp")
	}

	if expiry.Unix() != now+300 {
		t.Fatalf("parseJWTExpiry() got %d, want %d", expiry.Unix(), now+300)
	}
}

func TestParseJWTExpiry_InvalidToken(t *testing.T) {
	if _, ok := parseJWTExpiry("invalid.token"); ok {
		t.Fatalf("parseJWTExpiry() should fail for malformed token")
	}
}

func TestShouldRefreshToken_UsesJWTExp(t *testing.T) {
	now := time.Now()
	notExpiringSoon := buildUnsignedJWT(now.Add(5 * time.Minute).Unix())
	expiringSoon := buildUnsignedJWT(now.Add(15 * time.Second).Unix())

	if shouldRefreshToken(notExpiringSoon, "", jwtTokenRefreshLeadTime, accessTokenFallbackTTL, accessTokenFallbackLeadTime, now) {
		t.Fatalf("shouldRefreshToken() should be false for token with sufficient lifetime")
	}

	if !shouldRefreshToken(expiringSoon, "", jwtTokenRefreshLeadTime, accessTokenFallbackTTL, accessTokenFallbackLeadTime, now) {
		t.Fatalf("shouldRefreshToken() should be true for token near expiry")
	}
}

func TestShouldRefreshToken_FallbackTimestamp(t *testing.T) {
	now := time.Now()
	recentRefresh := fmt.Sprintf("%d", now.Unix())
	oldRefresh := fmt.Sprintf("%d", now.Add(-2*time.Hour).Unix())

	if shouldRefreshToken("not-a-jwt", recentRefresh, jwtTokenRefreshLeadTime, accessTokenFallbackTTL, accessTokenFallbackLeadTime, now) {
		t.Fatalf("shouldRefreshToken() fallback should consider recent refresh valid")
	}

	if !shouldRefreshToken("not-a-jwt", oldRefresh, jwtTokenRefreshLeadTime, accessTokenFallbackTTL, accessTokenFallbackLeadTime, now) {
		t.Fatalf("shouldRefreshToken() fallback should consider old refresh expired")
	}
}

func TestNextTokenValidationCheck(t *testing.T) {
	now := time.Now()
	token := buildUnsignedJWT(now.Add(2 * time.Minute).Unix())

	nextCheck, ok := nextTokenValidationCheck(token, "", jwtTokenRefreshLeadTime, accessTokenFallbackTTL, accessTokenFallbackLeadTime, now)
	if !ok {
		t.Fatalf("nextTokenValidationCheck() should return next check for JWT token")
	}

	if nextCheck.Before(now.Add(minCredentialValidationInterval)) {
		t.Fatalf("nextTokenValidationCheck() should respect minimum validation interval")
	}

	expected := now.Add(2*time.Minute - jwtTokenRefreshLeadTime)
	delta := nextCheck.Sub(expected)
	if delta < 0 {
		delta = -delta
	}
	if delta > 2*time.Second {
		t.Fatalf("nextTokenValidationCheck() unexpected schedule delta: %s", delta)
	}
}

func TestDecodeJWTComponent_WithPadding(t *testing.T) {
	value := "hello-world"
	encoded := base64.URLEncoding.EncodeToString([]byte(value))
	encoded = strings.TrimRight(encoded, "=")

	decoded, err := decodeJWTComponent(encoded)
	if err != nil {
		t.Fatalf("decodeJWTComponent() error: %v", err)
	}

	if string(decoded) != value {
		t.Fatalf("decodeJWTComponent() = %q, want %q", string(decoded), value)
	}
}
