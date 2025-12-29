package delegation

import "fmt"

// MaxConcurrentDelegates is the maximum number of delegate sessions allowed per parent session
const MaxConcurrentDelegates = 10

// CanSpawnDelegate checks if a new delegate can be spawned given the current active count
func CanSpawnDelegate(activeCount int) error {
	if activeCount >= MaxConcurrentDelegates {
		return fmt.Errorf("maximum concurrent delegates reached (%d/%d)", activeCount, MaxConcurrentDelegates)
	}
	return nil
}
