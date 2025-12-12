package indexer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseDirectory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tmpdir := t.TempDir()

	// Create a test Go file
	testFile := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}

type TestStruct struct {
	Field string
}

func (t *TestStruct) Method() string {
	return t.Field
}
`

	// Write the test file
	err := os.WriteFile(filepath.Join(tmpdir, "test.go"), []byte(testFile), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test current implementation
	p := NewASTParser()
	symbols, err := p.ParseDirectory(ctx, tmpdir)
	if err != nil {
		t.Fatalf("ParseDirectory failed: %v", err)
	}

	// Verify symbols were extracted
	if len(symbols) == 0 {
		t.Fatal("No symbols found")
	}

	// Check for expected symbols
	var foundMain, foundStruct, foundMethod bool
	for _, symbol := range symbols {
		switch symbol.Name {
		case "main":
			foundMain = true
		case "TestStruct":
			foundStruct = true
		case "TestStruct.Method":
			foundMethod = true
		}
	}

	if !foundMain {
		t.Error("Expected to find main function")
	}
	if !foundStruct {
		t.Error("Expected to find TestStruct")
	}
	if !foundMethod {
		t.Error("Expected to find Method")
	}
}

func TestParseDirectoryWithGoMod(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tmpdir := t.TempDir()

	// Create a go.mod file
	goModContent := `module test.com/example

go 1.21
`
	err := os.WriteFile(filepath.Join(tmpdir, "go.mod"), []byte(goModContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a test Go file with build tags
	testFile := `//go:build testtag

package main

import "fmt"

// This function should only be included when testtag is set
func taggedFunction() {
	fmt.Println("tagged")
}

type TestStruct struct {
	Field string
}
`

	// Write the test file
	err = os.WriteFile(filepath.Join(tmpdir, "test.go"), []byte(testFile), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test implementation with go.mod present
	p := NewASTParser()
	symbols, err := p.ParseDirectory(ctx, tmpdir)
	if err != nil {
		t.Fatalf("ParseDirectory with go.mod failed: %v", err)
	}

	// Should still parse the file even with build tag
	if len(symbols) == 0 {
		t.Fatal("No symbols found with go.mod")
	}

	// Look for expected symbols
	var foundStruct bool
	for _, symbol := range symbols {
		switch symbol.Name {
		case "taggedFunction":
			// Found tagged function (may not be included due to build tag)
		case "TestStruct":
			foundStruct = true
		}
	}

	if !foundStruct {
		t.Error("Expected to find TestStruct even with build tags")
	}
}
