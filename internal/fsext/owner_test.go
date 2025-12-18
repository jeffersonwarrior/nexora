package fsext

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestOwner(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("valid file", func(t *testing.T) {
		uid, err := Owner(tmpFile)
		if err != nil {
			t.Fatalf("Owner() error = %v", err)
		}

		if runtime.GOOS == "windows" {
			// On Windows, should return -1
			if uid != -1 {
				t.Errorf("Owner() on Windows = %d, want -1", uid)
			}
		} else {
			// On Unix-like systems, should return a valid UID
			// The UID should be non-negative
			if uid < 0 {
				t.Errorf("Owner() = %d, want non-negative UID", uid)
			}
			
			// Should match current user's UID
			currentUID := os.Getuid()
			if uid != currentUID {
				t.Logf("Owner() = %d, current UID = %d (may differ if running as different user)", uid, currentUID)
			}
		}
		
		t.Logf("Platform: %s, Owner UID: %d", runtime.GOOS, uid)
	})

	t.Run("valid directory", func(t *testing.T) {
		uid, err := Owner(tmpDir)
		if err != nil {
			t.Fatalf("Owner() error = %v", err)
		}

		if runtime.GOOS == "windows" {
			if uid != -1 {
				t.Errorf("Owner() on Windows = %d, want -1", uid)
			}
		} else {
			if uid < 0 {
				t.Errorf("Owner() = %d, want non-negative UID", uid)
			}
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		nonexistent := filepath.Join(tmpDir, "does-not-exist.txt")
		_, err := Owner(nonexistent)
		if err == nil {
			t.Error("Owner() for nonexistent file should return error")
		}
		
		if !os.IsNotExist(err) {
			t.Errorf("Owner() error should be IsNotExist, got: %v", err)
		}
	})
}

func TestOwnerCurrentDirectory(t *testing.T) {
	// Test with current directory
	uid, err := Owner(".")
	if err != nil {
		t.Fatalf("Owner('.') error = %v", err)
	}

	if runtime.GOOS == "windows" {
		if uid != -1 {
			t.Errorf("Owner() on Windows = %d, want -1", uid)
		}
	} else {
		if uid < 0 {
			t.Errorf("Owner() = %d, want non-negative UID", uid)
		}
		
		// Should be owned by current user (in most cases)
		currentUID := os.Getuid()
		t.Logf("Current directory owner: %d, current user: %d", uid, currentUID)
	}
}

func TestOwnerRootDirectory(t *testing.T) {
	// Test with root directory
	var rootDir string
	if runtime.GOOS == "windows" {
		rootDir = "C:\\"
	} else {
		rootDir = "/"
	}

	uid, err := Owner(rootDir)
	if err != nil {
		t.Fatalf("Owner('%s') error = %v", rootDir, err)
	}

	if runtime.GOOS == "windows" {
		if uid != -1 {
			t.Errorf("Owner() on Windows = %d, want -1", uid)
		}
	} else {
		// On Unix, root directory is typically owned by root (UID 0)
		// But we don't enforce this as it could be different in containers
		if uid < 0 {
			t.Errorf("Owner() = %d, want non-negative UID", uid)
		}
		t.Logf("Root directory owner UID: %d", uid)
	}
}

func TestOwnerSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Symlink test skipped on Windows (requires admin privileges)")
	}

	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	symlinkFile := filepath.Join(tmpDir, "symlink.txt")

	// Create target file
	err := os.WriteFile(targetFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// Create symlink
	err = os.Symlink(targetFile, symlinkFile)
	if err != nil {
		t.Skipf("Failed to create symlink (may not be supported): %v", err)
	}

	// Get owner of symlink (should follow the link)
	uid, err := Owner(symlinkFile)
	if err != nil {
		t.Fatalf("Owner() error = %v", err)
	}

	if uid < 0 {
		t.Errorf("Owner() = %d, want non-negative UID", uid)
	}
	
	t.Logf("Symlink owner UID: %d", uid)
}

func TestOwnerPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	
	tests := []struct {
		name string
		perm os.FileMode
	}{
		{"readable", 0644},
		{"writable", 0666},
		{"executable", 0755},
		{"restricted", 0600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.name+".txt")
			err := os.WriteFile(testFile, []byte("test"), tt.perm)
			if err != nil {
				t.Fatalf("Failed to create file: %v", err)
			}

			uid, err := Owner(testFile)
			if err != nil {
				t.Fatalf("Owner() error = %v", err)
			}

			if runtime.GOOS == "windows" {
				if uid != -1 {
					t.Errorf("Owner() on Windows = %d, want -1", uid)
				}
			} else {
				if uid < 0 {
					t.Errorf("Owner() = %d, want non-negative UID", uid)
				}
			}
		})
	}
}

func TestOwnerSpecialPaths(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "invalid path",
			path:    "/this/path/should/not/exist/hopefully",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Owner(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Owner() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestOwnerPlatformBehavior documents platform-specific behavior
func TestOwnerPlatformBehavior(t *testing.T) {
	t.Run("platform documentation", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "test.txt")
		os.WriteFile(tmpFile, []byte("test"), 0644)

		uid, _ := Owner(tmpFile)

		switch runtime.GOOS {
		case "windows":
			t.Log("Windows: Owner() always returns -1 (UID not applicable)")
			if uid != -1 {
				t.Errorf("Expected -1 on Windows, got %d", uid)
			}
		case "linux", "darwin", "freebsd":
			t.Logf("%s: Owner() returns actual UID from file stats", runtime.GOOS)
			if uid < 0 {
				t.Errorf("Expected non-negative UID on Unix-like systems, got %d", uid)
			}
		default:
			t.Logf("Unknown platform: %s, UID: %d", runtime.GOOS, uid)
		}
	})
}

// BenchmarkOwner tests performance of Owner function
func BenchmarkOwner(b *testing.B) {
	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "bench.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Owner(tmpFile)
	}
}
