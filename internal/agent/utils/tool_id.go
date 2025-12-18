package utils

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// GenerateToolCallID creates a provider-compliant tool call ID
// Mistral: only a-z, A-Z, 0-9, exactly 9 chars
// OpenAI: call_[a-zA-Z0-9]+ (current format)
func GenerateToolCallID(provider string) string {
	switch provider {
	case "mistral", "mistral-native":
		return generateMistralID()
	default:
		// Default to OpenAI format (current behavior)
		return fmt.Sprintf("call_%s", generateRandomID(12))
	}
}

// generateMistralID creates exactly 9 alphanumeric characters
func generateMistralID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 9)
	for i := range b {
		// Read cryptographically secure random byte
		var n [1]byte
		_, err := rand.Read(n[:])
		if err != nil {
			// Fallback to simple random if crypto fails
			b[i] = charset[i%len(charset)]
			continue
		}
		b[i] = charset[int(n[0])%len(charset)]
	}
	return string(b)
}

// generateRandomID creates a random hex string of specified length
func generateRandomID(length int) string {
	const charset = "abcdef0123456789"
	b := make([]byte, length)
	for i := range b {
		var n [1]byte
		_, err := rand.Read(n[:])
		if err != nil {
			b[i] = charset[i%len(charset)]
			continue
		}
		b[i] = charset[int(n[0])%len(charset)]
	}
	return string(b)
}

// ValidateMistralID checks if ID meets Mistral requirements
func ValidateMistralID(id string) bool {
	if len(id) != 9 {
		return false
	}
	for _, char := range id {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	return true
}

// SanitizeToolCallID converts existing ID to provider format if needed
func SanitizeToolCallID(id, provider string) string {
	switch provider {
	case "mistral", "mistral-native":
		// Extract digits/letters from existing ID and truncate/pad to 9
		clean := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}, id)

		// Pad or truncate to 9 chars
		for len(clean) < 9 {
			clean += generateRandomID(1)
		}
		if len(clean) > 9 {
			clean = clean[:9]
		}
		return clean
	default:
		return id // Keep existing format
	}
}
