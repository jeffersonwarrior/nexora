package internal

// TestGreptileFunction is a simple test function to verify Greptile code review
func TestGreptileFunction() string {
	// This function intentionally has issues that Greptile should catch:
	// 1. Exported function without doc comment matching the name
	// 2. No error handling
	// 3. Could use a better variable name
	x := "test"
	return x
}
