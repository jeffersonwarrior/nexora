package shell

import (
	"bytes"
	"sync"
	"testing"
)

// BenchmarkSyncWriterWrite measures thread-safe write performance
func BenchmarkSyncWriterWrite(b *testing.B) {
	var mu sync.RWMutex
	sw := &syncWriter{buf: &bytes.Buffer{}, mu: &mu}
	data := []byte("test output line from shell execution\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sw.Write(data)
	}
}

// BenchmarkSyncWriterString measures thread-safe read performance
func BenchmarkSyncWriterString(b *testing.B) {
	var mu sync.RWMutex
	sw := &syncWriter{buf: &bytes.Buffer{}, mu: &mu}
	sw.Write([]byte("some initial content\n"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sw.String()
	}
}

// BenchmarkSyncWriterConcurrent measures concurrent read/write performance
func BenchmarkSyncWriterConcurrent(b *testing.B) {
	var mu sync.RWMutex
	sw := &syncWriter{buf: &bytes.Buffer{}, mu: &mu}
	data := []byte("concurrent output line\n")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sw.Write(data)
			_ = sw.String()
		}
	})
}

// BenchmarkBackgroundShellGetOutput measures output retrieval performance
func BenchmarkBackgroundShellGetOutput(b *testing.B) {
	var mu sync.RWMutex
	bs := &BackgroundShell{
		stdout: &syncWriter{buf: &bytes.Buffer{}, mu: &mu},
		stderr: &syncWriter{buf: &bytes.Buffer{}, mu: &mu},
		done:   make(chan struct{}),
	}
	bs.stdout.Write([]byte("stdout content here\n"))
	bs.stderr.Write([]byte("stderr content here\n"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = bs.GetOutput()
	}
}

// BenchmarkBackgroundShellIsDone measures done check performance
func BenchmarkBackgroundShellIsDone(b *testing.B) {
	bs := &BackgroundShell{
		done: make(chan struct{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bs.IsDone()
	}
}
