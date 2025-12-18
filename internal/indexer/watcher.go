package indexer

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher monitors file system changes for real-time indexing updates
type FileWatcher struct {
	watcher     *fsnotify.Watcher
	indexer     *Indexer
	parser      *ASTParser
	embedEngine *EmbeddingEngine
	ctx         context.Context
	cancel      context.CancelFunc

	// Configuration
	paths         []string
	ignoredDirs   []string
	ignoredExts   []string
	debounceDelay time.Duration

	// State
	pendingFiles map[string]time.Time
	mu           sync.RWMutex
	batchSize    int

	// Events
	OnFileAdded   func(string)
	OnFileChanged func(string)
	OnFileRemoved func(string)
}

// NewFileWatcher creates a new file system watcher
func NewFileWatcher(indexer *Indexer, parser *ASTParser, embedEngine *EmbeddingEngine) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &FileWatcher{
		watcher:       watcher,
		indexer:       indexer,
		parser:        parser,
		embedEngine:   embedEngine,
		ctx:           ctx,
		cancel:        cancel,
		ignoredDirs:   []string{".git", "node_modules", "vendor", ".vscode", ".idea"},
		ignoredExts:   []string{".tmp", ".log", ".build", ".test"},
		debounceDelay: 2 * time.Second,
		pendingFiles:  make(map[string]time.Time),
		batchSize:     10,
	}, nil
}

// AddPath adds a directory to watch for changes
func (fw *FileWatcher) AddPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Add the directory itself
	err = fw.watcher.Add(absPath)
	if err != nil {
		return err
	}

	fw.paths = append(fw.paths, absPath)

	// Watch subdirectories recursively
	return filepath.Walk(absPath, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && fw.shouldWatchDir(walkPath) {
			if walkErr := fw.watcher.Add(walkPath); walkErr != nil {
				slog.Warn("Failed to watch directory", "path", walkPath, "error", walkErr)
			}
		}

		return nil
	})
}

// shouldWatchDir checks if a directory should be watched
func (fw *FileWatcher) shouldWatchDir(path string) bool {
	dirName := filepath.Base(path)
	for _, ignored := range fw.ignoredDirs {
		if strings.HasPrefix(dirName, ignored) {
			return false
		}
	}
	return true
}

// shouldProcessFile checks if a file should be processed for indexing
func (fw *FileWatcher) shouldProcessFile(path string) bool {
	// Only process Go files
	if !strings.HasSuffix(path, ".go") {
		return false
	}

	// Skip test files if desired
	if strings.HasSuffix(path, "_test.go") {
		return false // Could make this configurable
	}

	// Check ignored extensions
	for _, ext := range fw.ignoredExts {
		if strings.HasSuffix(path, ext) {
			return false
		}
	}

	return true
}

// Start begins watching for file system changes
func (fw *FileWatcher) Start() {
	go fw.watchLoop()
	go fw.debounceLoop()
}

// Stop stops the file watcher
func (fw *FileWatcher) Stop() {
	fw.cancel()
	fw.watcher.Close()
}

// watchLoop is the main event loop for file system changes
func (fw *FileWatcher) watchLoop() {
	for {
		select {
		case <-fw.ctx.Done():
			return

		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			fw.handleEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}

			slog.Error("File watcher error", "error", err)
		}
	}
}

// handleEvent processes a file system event
func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	if !fw.shouldProcessFile(event.Name) {
		return
	}

	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Add to pending files with timestamp
	fw.pendingFiles[event.Name] = time.Now()

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		if fw.OnFileAdded != nil {
			fw.OnFileAdded(event.Name)
		}

	case event.Op&fsnotify.Write == fsnotify.Write:
		if fw.OnFileChanged != nil {
			fw.OnFileChanged(event.Name)
		}

	case event.Op&fsnotify.Remove == fsnotify.Remove:
		if fw.OnFileRemoved != nil {
			fw.OnFileRemoved(event.Name)
		}
		delete(fw.pendingFiles, event.Name)
	}
}

// debounceLoop processes pending files after debounce delay
func (fw *FileWatcher) debounceLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-fw.ctx.Done():
			return

		case now := <-ticker.C:
			fw.processPendingFiles(now)
		}
	}
}

// processPendingFiles processes files that have been stable for debounce period
func (fw *FileWatcher) processPendingFiles(now time.Time) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	ready := make([]string, 0)

	// Find files that have been stable for debounce period
	for file, timestamp := range fw.pendingFiles {
		if now.Sub(timestamp) >= fw.debounceDelay {
			ready = append(ready, file)
			delete(fw.pendingFiles, file)
		}
	}

	// Process in batches to avoid overwhelming the system
	for i := 0; i < len(ready); i += fw.batchSize {
		end := i + fw.batchSize
		if end > len(ready) {
			end = len(ready)
		}

		batch := ready[i:end]
		go fw.processBatch(batch)
	}
}

// processBatch processes a batch of files
func (fw *FileWatcher) processBatch(files []string) {
	for _, file := range files {
		if err := fw.processFile(file); err != nil {
			slog.Error("Failed to process file", "file", file, "error", err)
		}
	}
}

// processFile updates the index for a single file
func (fw *FileWatcher) processFile(path string) error {
	// Remove old symbols for this file
	if err := fw.indexer.DeleteSymbolsByFile(fw.ctx, path); err != nil {
		slog.Warn("Failed to remove old symbols", "file", path, "error", err)
	}

	// Parse the file for new symbols
	symbols, err := fw.parser.ParseFile(fw.ctx, path)
	if err != nil {
		return err
	}

	if len(symbols) == 0 {
		return nil
	}

	// Store new symbols
	if err := fw.indexer.StoreSymbols(fw.ctx, symbols); err != nil {
		return err
	}

	// Generate embeddings for the symbols
	embeddings, err := fw.embedEngine.GenerateSymbolEmbeddings(fw.ctx, symbols)
	if err != nil {
		slog.Warn("Failed to generate embeddings", "file", path, "error", err)
		return nil // Don't fail the whole process for embedding issues
	}

	return fw.indexer.StoreEmbeddings(fw.ctx, embeddings)
}

// SetIgnoredDirs sets the list of directories to ignore
func (fw *FileWatcher) SetIgnoredDirs(dirs []string) {
	fw.ignoredDirs = dirs
}

// SetIgnoredExts sets the list of file extensions to ignore
func (fw *FileWatcher) SetIgnoredExts(exts []string) {
	fw.ignoredExts = exts
}

// SetDebounceDelay sets the debounce delay for processing files
func (fw *FileWatcher) SetDebounceDelay(delay time.Duration) {
	fw.debounceDelay = delay
}

// IsWatching returns true if currently watching any paths
func (fw *FileWatcher) IsWatching() bool {
	return len(fw.paths) > 0
}

// GetWatchedPaths returns the list of currently watched paths
func (fw *FileWatcher) GetWatchedPaths() []string {
	return append([]string{}, fw.paths...)
}
