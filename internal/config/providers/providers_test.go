package providers

import "testing"

// TestAllProviders_NoConflictingAuthHeaders validates all providers don't have
// conflicting Authorization headers that would cause 401 errors.
func TestAllProviders_NoConflictingAuthHeaders(t *testing.T) {
	t.Parallel()
	for name, factory := range allProviderFuncs() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			provider := factory(nil)
			if provider.ID == "" {
				return // Skip empty providers
			}
			requireNoConflictingAuthHeaders(t, provider)
		})
	}
}
