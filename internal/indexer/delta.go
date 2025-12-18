package indexer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/ncruces/go-sqlite3"
)

// DeltaHandler manages incremental index updates
type DeltaHandler struct {
	indexer *Indexer
	parser  *ASTParser
	engine  *EmbeddingEngine

	// Delta tracking
	lastSync time.Time
	batch    DeltaBatch
}

// DeltaBatch represents a batch of changes to apply
type DeltaBatch struct {
	Added    []string `json:"added"`    // File paths added
	Modified []string `json:"modified"` // File paths modified
	Removed  []string `json:"removed"`  // File paths removed
}

// NewDeltaHandler creates a new delta handler
func NewDeltaHandler(indexer *Indexer, parser *ASTParser, engine *EmbeddingEngine) (*DeltaHandler, error) {
	// Get last sync time from database
	lastSync := time.Time{}
	if indexer != nil {
		var syncTime sql.NullTime
		err := indexer.db.QueryRow("SELECT MAX(updated_at) FROM symbols").Scan(&syncTime)
		if err == nil && syncTime.Valid {
			lastSync = syncTime.Time
		}
	}

	return &DeltaHandler{
		indexer:  indexer,
		parser:   parser,
		engine:   engine,
		lastSync: lastSync,
	}, nil
}

// ProcessDelta processes a batch of file changes
func (dh *DeltaHandler) ProcessDelta(ctx context.Context, batch DeltaBatch) error {
	slog.Info("Processing delta batch",
		"added", len(batch.Added),
		"modified", len(batch.Modified),
		"removed", len(batch.Removed))

	// Start transaction for atomic updates
	tx, err := dh.indexer.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Process removed files
	for _, filePath := range batch.Removed {
		if err := dh.removeFile(ctx, tx, filePath); err != nil {
			slog.Warn("Failed to remove file", "file", filePath, "error", err)
		}
	}

	// Process added and modified files
	allUpdates := append(batch.Added, batch.Modified...)
	for _, filePath := range allUpdates {
		if err := dh.updateFile(ctx, tx, filePath); err != nil {
			slog.Warn("Failed to update file", "file", filePath, "error", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit delta transaction: %w", err)
	}

	// Update last sync time
	dh.lastSync = time.Now()

	return nil
}

// removeFile removes all symbols and embeddings for a file
func (dh *DeltaHandler) removeFile(ctx context.Context, tx *sql.Tx, filePath string) error {
	// Remove symbols
	_, err := tx.ExecContext(ctx, "DELETE FROM symbols WHERE file = ?", filePath)
	if err != nil {
		return fmt.Errorf("failed to remove symbols for %s: %w", filePath, err)
	}

	// Remove embeddings (metadata contains file path)
	_, err = tx.ExecContext(ctx, "DELETE FROM embeddings WHERE metadata LIKE ?", "%"+filePath+"%")
	if err != nil {
		return fmt.Errorf("failed to remove embeddings for %s: %w", filePath, err)
	}

	slog.Info("Removed file from index", "file", filePath)
	return nil
}

// updateFile updates symbols and embeddings for a file
func (dh *DeltaHandler) updateFile(ctx context.Context, tx *sql.Tx, filePath string) error {
	// Remove old symbols for this file first
	_, err := tx.ExecContext(ctx, "DELETE FROM symbols WHERE file = ?", filePath)
	if err != nil {
		return fmt.Errorf("failed to remove old symbols for %s: %w", filePath, err)
	}

	// Parse file for new symbols
	symbols, err := dh.parser.ParseFile(ctx, filePath)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	if len(symbols) == 0 {
		slog.Debug("No symbols found in file", "file", filePath)
		return nil
	}

	// Store new symbols using transaction
	for _, symbol := range symbols {
		if err := dh.storeSymbolInTx(ctx, tx, symbol); err != nil {
			slog.Warn("Failed to store symbol", "symbol", symbol.Name, "file", filePath, "error", err)
			continue
		}
	}

	// Generate and store embeddings
	if dh.engine != nil {
		embeddings, err := dh.engine.GenerateSymbolEmbeddings(ctx, symbols)
		if err != nil {
			slog.Warn("Failed to generate embeddings", "file", filePath, "error", err)
		} else {
			for _, embedding := range embeddings {
				if err := dh.storeEmbeddingInTx(ctx, tx, embedding); err != nil {
					slog.Warn("Failed to store embedding", "symbol", embedding.ID, "file", filePath, "error", err)
				}
			}
		}
	}

	slog.Info("Updated file in index", "file", filePath, "symbols", len(symbols))
	return nil
}

// storeSymbolInTx stores a symbol within a transaction
func (dh *DeltaHandler) storeSymbolInTx(ctx context.Context, tx *sql.Tx, symbol Symbol) error {
	importsJSON, _ := json.Marshal(symbol.Imports)
	callersJSON, _ := json.Marshal(symbol.Callers)
	callsJSON, _ := json.Marshal(symbol.Calls)
	paramsJSON, _ := json.Marshal(symbol.Params)
	returnsJSON, _ := json.Marshal(symbol.Returns)
	fieldsJSON, _ := json.Marshal(symbol.Fields)
	methodsJSON, _ := json.Marshal(symbol.Methods)

	_, err := tx.ExecContext(ctx, `
		INSERT OR REPLACE INTO symbols (
			id, name, type, package, file, line, column, signature, doc,
			imports, callers, calls, public, params, returns, fields, methods, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		fmt.Sprintf("%s:%s:%d", symbol.Package, symbol.Name, symbol.Line),
		symbol.Name, symbol.Type, symbol.Package, symbol.File,
		symbol.Line, symbol.Column, symbol.Signature, symbol.Doc,
		string(importsJSON), string(callersJSON), string(callsJSON),
		symbol.Public, string(paramsJSON), string(returnsJSON),
		string(fieldsJSON), string(methodsJSON), time.Now(),
	)
	return err
}

// storeEmbeddingInTx stores an embedding within a transaction
func (dh *DeltaHandler) storeEmbeddingInTx(ctx context.Context, tx *sql.Tx, embedding Embedding) error {
	vectorJSON, _ := json.Marshal(embedding.Vector)
	metadataJSON, _ := json.Marshal(embedding.Metadata)

	_, err := tx.ExecContext(ctx, `
		INSERT OR REPLACE INTO embeddings (
			id, type, text, vector, metadata, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		embedding.ID, embedding.Type, embedding.Text,
		string(vectorJSON), string(metadataJSON), embedding.Created,
	)
	return err
}

// GetLastSync returns the timestamp of the last sync
func (dh *DeltaHandler) GetLastSync() time.Time {
	return dh.lastSync
}

// SetLastSync manually sets the last sync time
func (dh *DeltaHandler) SetLastSync(t time.Time) {
	dh.lastSync = t
}

// CreateDeltaBatch creates a delta batch from file change lists
func CreateDeltaBatch(added, modified, removed []string) DeltaBatch {
	return DeltaBatch{
		Added:    added,
		Modified: modified,
		Removed:  removed,
	}
}

// Checkpoint creates a checkpoint to mark a successful sync
func (dh *DeltaHandler) Checkpoint(ctx context.Context) error {
	if dh.indexer == nil {
		return nil
	}

	// Store checkpoint in metadata table
	_, err := dh.indexer.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS checkpoints (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create checkpoints table: %w", err)
	}

	_, err = dh.indexer.db.ExecContext(ctx, `
		INSERT INTO checkpoints (timestamp) VALUES (?)
	`, dh.lastSync)
	if err != nil {
		return fmt.Errorf("failed to store checkpoint: %w", err)
	}

	// Clean up old checkpoints (keep last 10)
	_, err = dh.indexer.db.ExecContext(ctx, `
		DELETE FROM checkpoints WHERE id NOT IN (
			SELECT id FROM checkpoints ORDER BY timestamp DESC LIMIT 10
		)
	`)
	return err
}

// GetLastCheckpoint gets the timestamp of the last successful checkpoint
func (dh *DeltaHandler) GetLastCheckpoint(ctx context.Context) (time.Time, error) {
	if dh.indexer == nil {
		return time.Time{}, nil
	}

	var checkpointTime sql.NullTime
	err := dh.indexer.db.QueryRowContext(ctx, `
		SELECT timestamp FROM checkpoints ORDER BY timestamp DESC LIMIT 1
	`).Scan(&checkpointTime)

	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last checkpoint: %w", err)
	}

	if checkpointTime.Valid {
		return checkpointTime.Time, nil
	}
	return time.Time{}, nil
}
