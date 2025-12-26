package main

import (
	"fmt"
)

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func main() {
	fmt.Println("contains('operation timed out', 'timeout'):", contains("operation timed out", "timeout"))
	fmt.Println("findSubstring('operation timed out', 'timeout'):", findSubstring("operation timed out", "timeout"))
	
	// Check each index
	for i := 0; i <= len("operation timed out")-len("timeout"); i++ {
		sub := string([]rune("operation timed out")[i:i+len("timeout")])
		fmt.Printf("Index %d: '%s' == 'timeout'? %v\n", i, sub, sub == "timeout")
	}
}