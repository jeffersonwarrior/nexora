package tools

import (
	"testing"
	"time"
)

func TestRecordFileRead(t *testing.T) {
	testPath := "/tmp/test-file-read.txt"

	// Record a file read
	recordFileRead(testPath)

	// Verify we can get the read time
	readTime := getLastReadTime(testPath)
	if readTime.IsZero() {
		t.Error("Expected non-zero read time after recording")
	}

	// Verify the time is recent (within last second)
	if time.Since(readTime) > time.Second {
		t.Error("Expected read time to be recent")
	}
}

func TestGetLastReadTime_NonExistent(t *testing.T) {
	testPath := "/tmp/non-existent-file-for-test.txt"

	// Get read time for file that was never read
	readTime := getLastReadTime(testPath)
	if !readTime.IsZero() {
		t.Error("Expected zero time for non-existent file record")
	}
}

func TestRecordFileWrite(t *testing.T) {
	testPath := "/tmp/test-file-write.txt"

	// Record a file write
	recordFileWrite(testPath)

	// Record is now internal, but we can verify through read/write sequence
	recordFileRead(testPath)
	readTime := getLastReadTime(testPath)
	if readTime.IsZero() {
		t.Error("Expected non-zero read time")
	}
}

func TestFileRecordConcurrency(t *testing.T) {
	testPath := "/tmp/test-concurrent-access.txt"

	// Run concurrent reads and writes
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			recordFileRead(testPath)
			recordFileWrite(testPath)
			_ = getLastReadTime(testPath)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify we can still read the time (mutex works)
	readTime := getLastReadTime(testPath)
	if readTime.IsZero() {
		t.Error("Expected non-zero read time after concurrent access")
	}
}
