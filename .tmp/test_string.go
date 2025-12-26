package main

import (
	"strings"
	"fmt"
)

func main() {
	s := "operation timed out"
	sub := "timeout"
	
	fmt.Printf("String: %q\n", s)
	fmt.Printf("Substr: %q\n", sub)
	fmt.Printf("strings.Contains: %v\n", strings.Contains(s, sub))
	
	// Let's check byte by byte
	for i, b := range []byte(s) {
		fmt.Printf("%d: %c (%d)\n", i, b, b)
	}
	
	fmt.Println("\nLooking for 't':")
	for i, r := range s {
		if r == 't' {
			fmt.Printf("Found 't' at index %d\n", i)
		}
	}
}