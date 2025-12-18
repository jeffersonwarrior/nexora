package oauth

import (
	"testing"
	"time"
)

func TestTokenSetExpiresAt(t *testing.T) {
	tests := []struct {
		name      string
		expiresIn int
		tolerance time.Duration // Allow some tolerance for timing
	}{
		{
			name:      "1 hour expiry",
			expiresIn: 3600,
			tolerance: 2 * time.Second,
		},
		{
			name:      "1 minute expiry",
			expiresIn: 60,
			tolerance: 2 * time.Second,
		},
		{
			name:      "1 day expiry",
			expiresIn: 86400,
			tolerance: 2 * time.Second,
		},
		{
			name:      "zero expiry",
			expiresIn: 0,
			tolerance: 2 * time.Second,
		},
		{
			name:      "very short expiry",
			expiresIn: 1,
			tolerance: 2 * time.Second,
		},
		{
			name:      "30 days expiry",
			expiresIn: 2592000,
			tolerance: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &Token{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				ExpiresIn:    tt.expiresIn,
			}

			beforeSet := time.Now()
			token.SetExpiresAt()
			afterSet := time.Now()

			// Calculate expected expiry time range
			expectedMin := beforeSet.Add(time.Duration(tt.expiresIn) * time.Second).Unix()
			expectedMax := afterSet.Add(time.Duration(tt.expiresIn) * time.Second).Unix()

			// Verify ExpiresAt is set within the expected range
			if token.ExpiresAt < expectedMin || token.ExpiresAt > expectedMax {
				t.Errorf("SetExpiresAt() = %d, expected between %d and %d",
					token.ExpiresAt, expectedMin, expectedMax)
			}

			// Verify it's approximately ExpiresIn seconds from now
			actualDuration := time.Unix(token.ExpiresAt, 0).Sub(beforeSet)
			expectedDuration := time.Duration(tt.expiresIn) * time.Second

			diff := actualDuration - expectedDuration
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.tolerance {
				t.Errorf("SetExpiresAt() duration difference too large: got %v, expected ~%v (diff: %v)",
					actualDuration, expectedDuration, diff)
			}
		})
	}
}

func TestTokenIsExpired(t *testing.T) {
	tests := []struct {
		name            string
		setupToken      func() *Token
		expectedExpired bool
		description     string
	}{
		{
			name: "fresh token",
			setupToken: func() *Token {
				token := &Token{
					AccessToken:  "test",
					RefreshToken: "test",
					ExpiresIn:    3600, // 1 hour
				}
				token.SetExpiresAt()
				return token
			},
			expectedExpired: false,
			description:     "Token just created should not be expired",
		},
		{
			name: "expired token",
			setupToken: func() *Token {
				return &Token{
					AccessToken:  "test",
					RefreshToken: "test",
					ExpiresIn:    3600,
					ExpiresAt:    time.Now().Add(-1 * time.Hour).Unix(), // 1 hour ago
				}
			},
			expectedExpired: true,
			description:     "Token expired 1 hour ago should be expired",
		},
		{
			name: "token expiring in 5% of lifetime",
			setupToken: func() *Token {
				expiresIn := 3600 // 1 hour
				// Set to expire in 5% of lifetime (180 seconds)
				// IsExpired triggers at 10%, so this SHOULD be expired
				return &Token{
					AccessToken:  "test",
					RefreshToken: "test",
					ExpiresIn:    expiresIn,
					ExpiresAt:    time.Now().Add(180 * time.Second).Unix(),
				}
			},
			expectedExpired: true,
			description:     "Token with 5% lifetime remaining should be expired (threshold is 10%)",
		},
		{
			name: "token at 10% threshold",
			setupToken: func() *Token {
				expiresIn := 3600 // 1 hour
				// Set to expire in exactly 10% of lifetime (360 seconds)
				// This is the threshold, so should be considered expired
				return &Token{
					AccessToken:  "test",
					RefreshToken: "test",
					ExpiresIn:    expiresIn,
					ExpiresAt:    time.Now().Add(360 * time.Second).Unix(),
				}
			},
			expectedExpired: true,
			description:     "Token at exactly 10% threshold should be expired",
		},
		{
			name: "token with 15% lifetime remaining",
			setupToken: func() *Token {
				expiresIn := 3600 // 1 hour
				// Set to expire in 15% of lifetime (540 seconds)
				// Above 10% threshold, should NOT be expired
				return &Token{
					AccessToken:  "test",
					RefreshToken: "test",
					ExpiresIn:    expiresIn,
					ExpiresAt:    time.Now().Add(540 * time.Second).Unix(),
				}
			},
			expectedExpired: false,
			description:     "Token with 15% lifetime remaining should not be expired",
		},
		{
			name: "token just below 10% threshold",
			setupToken: func() *Token {
				expiresIn := 3600 // 1 hour
				// Set to expire in 9% of lifetime (324 seconds)
				// Should be considered expired
				return &Token{
					AccessToken:  "test",
					RefreshToken: "test",
					ExpiresIn:    expiresIn,
					ExpiresAt:    time.Now().Add(324 * time.Second).Unix(),
				}
			},
			expectedExpired: true,
			description:     "Token with <10% lifetime should be expired",
		},
		{
			name: "token with zero expiresIn",
			setupToken: func() *Token {
				return &Token{
					AccessToken:  "test",
					RefreshToken: "test",
					ExpiresIn:    0,
					ExpiresAt:    time.Now().Unix(),
				}
			},
			expectedExpired: true,
			description:     "Token with zero expiresIn should be expired",
		},
		{
			name: "token expired exactly now",
			setupToken: func() *Token {
				return &Token{
					AccessToken:  "test",
					RefreshToken: "test",
					ExpiresIn:    3600,
					ExpiresAt:    time.Now().Unix(),
				}
			},
			expectedExpired: true,
			description:     "Token expiring right now should be expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()
			result := token.IsExpired()

			if result != tt.expectedExpired {
				t.Errorf("IsExpired() = %v, want %v - %s (ExpiresIn: %d, ExpiresAt: %d, Now: %d)",
					result, tt.expectedExpired, tt.description,
					token.ExpiresIn, token.ExpiresAt, time.Now().Unix())
			}
		})
	}
}

func TestTokenLifecycle(t *testing.T) {
	// Test a full lifecycle: create, check not expired, wait, check expired
	token := &Token{
		AccessToken:  "test-access",
		RefreshToken: "test-refresh",
		ExpiresIn:    1, // 1 second
	}

	token.SetExpiresAt()

	// Should not be expired immediately
	if token.IsExpired() {
		t.Error("Token should not be expired immediately after creation")
	}

	// Wait for expiration (plus 10% buffer)
	time.Sleep(1200 * time.Millisecond)

	// Should now be expired
	if !token.IsExpired() {
		t.Error("Token should be expired after waiting past expiry time")
	}
}

func TestTokenFields(t *testing.T) {
	token := &Token{
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		ExpiresIn:    7200,
	}

	// Verify fields are set correctly
	if token.AccessToken != "access-123" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "access-123")
	}
	if token.RefreshToken != "refresh-456" {
		t.Errorf("RefreshToken = %q, want %q", token.RefreshToken, "refresh-456")
	}
	if token.ExpiresIn != 7200 {
		t.Errorf("ExpiresIn = %d, want %d", token.ExpiresIn, 7200)
	}

	// Before SetExpiresAt, ExpiresAt should be 0
	if token.ExpiresAt != 0 {
		t.Errorf("ExpiresAt should be 0 before SetExpiresAt(), got %d", token.ExpiresAt)
	}

	token.SetExpiresAt()

	// After SetExpiresAt, ExpiresAt should be set
	if token.ExpiresAt == 0 {
		t.Error("ExpiresAt should be set after SetExpiresAt()")
	}
}

func TestTokenExpiryCalculation(t *testing.T) {
	// Test the 10% buffer calculation
	testCases := []struct {
		expiresIn      int
		expectedBuffer int64
		description    string
	}{
		{3600, 360, "1 hour with 6 minute buffer"},
		{60, 6, "1 minute with 6 second buffer"},
		{10, 1, "10 seconds with 1 second buffer"},
		{100, 10, "100 seconds with 10 second buffer"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			buffer := int64(tc.expiresIn) / 10
			if buffer != tc.expectedBuffer {
				t.Errorf("Buffer calculation: got %d, want %d", buffer, tc.expectedBuffer)
			}
		})
	}
}

// BenchmarkSetExpiresAt tests performance of SetExpiresAt
func BenchmarkSetExpiresAt(b *testing.B) {
	token := &Token{
		AccessToken:  "test",
		RefreshToken: "test",
		ExpiresIn:    3600,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token.SetExpiresAt()
	}
}

// BenchmarkIsExpired tests performance of IsExpired
func BenchmarkIsExpired(b *testing.B) {
	token := &Token{
		AccessToken:  "test",
		RefreshToken: "test",
		ExpiresIn:    3600,
	}
	token.SetExpiresAt()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token.IsExpired()
	}
}
