package indexer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// Indexer handles storage and retrieval of indexed code data
type Indexer struct {
	db        *sql.DB
	initOnce  sync.Once
	initMu    sync.Mutex
	initError error
}

// NewIndexer creates a new indexer with SQLite backend
func NewIndexer(dbPath string) (*Indexer, error) {
	// Parse the path to separate filename from query params if needed
	// Or trust the driver to handle it.
	// But tests show that dbPath is used literally as filename if params are appended.
	// The standard sqlite3 driver supports URI filenames if "file:" prefix is used or config is set.
	// Here we just append params.

	// If we are testing, we might want to avoid WAL mode or be careful about filenames.
	// But better: use "file:" prefix to force URI interpretation.
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000", dbPath)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	indexer := &Indexer{db: db}
	indexer.initOnce.Do(func() {
		indexer.initMu.Lock()
		defer indexer.initMu.Unlock()
		if err := indexer.initSchema(); err != nil {
			indexer.initError = err
		}
	})

	if indexer.initError != nil {
		return nil, indexer.initError
	}

	return indexer, nil
}

// initSchema creates the necessary database tables
func (i *Indexer) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS symbols (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			package TEXT NOT NULL,
			file TEXT NOT NULL,
			line INTEGER NOT NULL,
			column INTEGER NOT NULL,
			signature TEXT,
			doc TEXT,
			imports TEXT,
			callers TEXT,
			calls TEXT,
			public BOOLEAN NOT NULL,
			params TEXT,
			returns TEXT,
			fields TEXT,
			methods TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS embeddings (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			text TEXT NOT NULL,
			vector BLOB NOT NULL,
			metadata TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS relationships (
			from_id TEXT NOT NULL,
			to_id TEXT NOT NULL,
			relationship_type TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (from_id, to_id, relationship_type)
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS symbols_fts USING fts5(
			id UNINDEXED, name, type, package, signature, doc
		)`,
		`CREATE INDEX IF NOT EXISTS idx_symbols_name ON symbols(name)`,
		`CREATE INDEX IF NOT EXISTS idx_symbols_type ON symbols(type)`,
		`CREATE INDEX IF NOT EXISTS idx_symbols_package ON symbols(package)`,
		`CREATE INDEX IF NOT EXISTS idx_symbols_file ON symbols(file)`,
	}

	for _, query := range queries {
		if _, err := i.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	return nil
}

// StoreSymbols saves symbols to the database
func (i *Indexer) StoreSymbols(ctx context.Context, symbols []Symbol) error {
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO symbols (
			id, name, type, package, file, line, column, signature, doc,
			imports, callers, calls, public, params, returns, fields, methods
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, symbol := range symbols {
		importsJSON, _ := json.Marshal(symbol.Imports)
		callersJSON, _ := json.Marshal(symbol.Callers)
		callsJSON, _ := json.Marshal(symbol.Calls)
		paramsJSON, _ := json.Marshal(symbol.Params)
		returnsJSON, _ := json.Marshal(symbol.Returns)
		fieldsJSON, _ := json.Marshal(symbol.Fields)
		methodsJSON, _ := json.Marshal(symbol.Methods)

		id := fmt.Sprintf("%s:%s:%d", symbol.Package, symbol.Name, symbol.Line)

		_, err := stmt.ExecContext(ctx,
			id, symbol.Name, symbol.Type, symbol.Package, symbol.File,
			symbol.Line, symbol.Column, symbol.Signature, symbol.Doc,
			string(importsJSON), string(callersJSON), string(callsJSON),
			symbol.Public, string(paramsJSON), string(returnsJSON),
			string(fieldsJSON), string(methodsJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert symbol %s: %w", symbol.Name, err)
		}

		// Add to FTS
		_, err = tx.ExecContext(ctx, `
			INSERT OR REPLACE INTO symbols_fts(id, name, type, package, signature, doc)
			VALUES (?, ?, ?, ?, ?, ?)
		`, id, symbol.Name, symbol.Type, symbol.Package, symbol.Signature, symbol.Doc)
		if err != nil {
			slog.Warn("Failed to add symbol to FTS", "symbol", symbol.Name, "error", err)
		}
	}

	return tx.Commit()
}

// StoreEmbeddings saves embeddings to the database
func (i *Indexer) StoreEmbeddings(ctx context.Context, embeddings []Embedding) error {
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO embeddings (id, type, text, vector, metadata)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, embedding := range embeddings {
		vectorJSON, err := json.Marshal(embedding.Vector)
		if err != nil {
			return fmt.Errorf("failed to marshal vector for embedding %s: %w", embedding.ID, err)
		}

		metadataJSON, err := json.Marshal(embedding.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for embedding %s: %w", embedding.ID, err)
		}

		_, err = stmt.ExecContext(ctx,
			embedding.ID, embedding.Type, embedding.Text, string(vectorJSON), string(metadataJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert embedding %s: %w", embedding.ID, err)
		}
	}

	return tx.Commit()
}

// SearchSymbols performs text search on symbols
func (i *Indexer) SearchSymbols(ctx context.Context, query string, limit int) ([]Symbol, error) {
	var rows *sql.Rows
	var err error

	if query == "" {
		// For empty query, return all symbols
		rows, err = i.db.QueryContext(ctx, `
			SELECT DISTINCT s.* FROM symbols s
			LIMIT ?
		`, limit)
	} else {
		// Use FTS for text search
		rows, err = i.db.QueryContext(ctx, `
			SELECT s.* FROM symbols s
			JOIN symbols_fts ON s.id = symbols_fts.id
			WHERE symbols_fts MATCH ?
			LIMIT ?
		`, query, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to search symbols: %w", err)
	}
	defer rows.Close()

	symbols := make([]Symbol, 0)
	for rows.Next() {
		var symbol Symbol
		var id string
		var importsJSON, callersJSON, callsJSON, paramsJSON, returnsJSON, fieldsJSON, methodsJSON string
		var createdAt time.Time

		err := rows.Scan(
			&id, &symbol.Name, &symbol.Type, &symbol.Package,
			&symbol.File, &symbol.Line, &symbol.Column, &symbol.Signature,
			&symbol.Doc, &importsJSON, &callersJSON, &callsJSON,
			&symbol.Public, &paramsJSON, &returnsJSON, &fieldsJSON,
			&methodsJSON, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan symbol row: %w", err)
		}

		// Parse JSON fields
		json.Unmarshal([]byte(importsJSON), &symbol.Imports)
		json.Unmarshal([]byte(callersJSON), &symbol.Callers)
		json.Unmarshal([]byte(callsJSON), &symbol.Calls)
		json.Unmarshal([]byte(paramsJSON), &symbol.Params)
		json.Unmarshal([]byte(returnsJSON), &symbol.Returns)
		json.Unmarshal([]byte(fieldsJSON), &symbol.Fields)
		json.Unmarshal([]byte(methodsJSON), &symbol.Methods)

		symbols = append(symbols, symbol)
	}

	return symbols, nil
}

// GetSymbol finds a specific symbol by ID
func (i *Indexer) GetSymbol(ctx context.Context, id string) (*Symbol, error) {
	// Try to get symbol by ID first
	row := i.db.QueryRowContext(ctx, `
		SELECT id, name, type, package, file, line, column, signature, doc,
		imports, callers, calls, public, params, returns, fields, methods, created_at
		FROM symbols WHERE id = ?
	`, id)

	var symbol Symbol
	var importsJSON, callersJSON, callsJSON, paramsJSON, returnsJSON, fieldsJSON, methodsJSON string
	var createdAt time.Time

	err := row.Scan(
		&id, &symbol.Name, &symbol.Type, &symbol.Package,
		&symbol.File, &symbol.Line, &symbol.Column, &symbol.Signature,
		&symbol.Doc, &importsJSON, &callersJSON, &callsJSON,
		&symbol.Public, paramsJSON, &returnsJSON, &fieldsJSON,
		&methodsJSON, &createdAt,
	)
	if err == nil {
		// Successfully found by ID
		return &symbol, nil
	}

	// If ID lookup failed, try name lookup
	row = i.db.QueryRowContext(ctx, `
		SELECT id, name, type, package, file, line, column, signature, doc,
		imports, callers, calls, public, params, returns, fields, methods, created_at
		FROM symbols WHERE name = ?
		LIMIT 1
	`, id)

	// Reset variables for name lookup
	var symbolID string
	var importsJSON2, callersJSON2, callsJSON2, paramsJSON2, returnsJSON2, fieldsJSON2, methodsJSON2 string
	var createdAt2 time.Time

	err = row.Scan(
		&symbolID, &symbol.Name, &symbol.Type, &symbol.Package,
		&symbol.File, &symbol.Line, &symbol.Column, &symbol.Signature,
		&symbol.Doc, &importsJSON2, &callersJSON2, &callsJSON2,
		&symbol.Public, &paramsJSON2, &returnsJSON2, &fieldsJSON2,
		&methodsJSON2, &createdAt2,
	)

	if err == nil {
		// Parse JSON fields for name lookup
		json.Unmarshal([]byte(importsJSON2), &symbol.Imports)
		json.Unmarshal([]byte(callersJSON2), &symbol.Callers)
		json.Unmarshal([]byte(callsJSON2), &symbol.Calls)
		json.Unmarshal([]byte(paramsJSON2), &symbol.Params)
		json.Unmarshal([]byte(returnsJSON2), &symbol.Returns)
		json.Unmarshal([]byte(fieldsJSON2), &symbol.Fields)
		json.Unmarshal([]byte(methodsJSON2), &symbol.Methods)

		return &symbol, nil
	}

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("symbol not found: %s", id)
	}
	return nil, fmt.Errorf("failed to get symbol: %w", err)
}

// FindCallers finds all symbols that call the given symbol
func (i *Indexer) FindCallers(ctx context.Context, targetSymbol string) ([]Symbol, error) {
	rows, err := i.db.QueryContext(ctx, `
		SELECT DISTINCT s.* FROM symbols s
		WHERE s.calls LIKE ?
		ORDER BY s.package, s.name
	`, "%"+targetSymbol+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to find callers: %w", err)
	}
	defer rows.Close()

	return i.scanSymbols(rows)
}

// FindCalled finds all symbols called by the given symbol
func (i *Indexer) FindCalled(ctx context.Context, sourceSymbol string) ([]Symbol, error) {
	rows, err := i.db.QueryContext(ctx, `
		SELECT DISTINCT s.* FROM symbols s
		WHERE s.callers LIKE ?
		ORDER BY s.package, s.name
	`, "%"+sourceSymbol+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to find called symbols: %w", err)
	}
	defer rows.Close()

	return i.scanSymbols(rows)
}

// GetAllEmbeddings retrieves all embeddings from storage
func (i *Indexer) GetAllEmbeddings(ctx context.Context) ([]Embedding, error) {
	rows, err := i.db.QueryContext(ctx, "SELECT id, type, text, vector, metadata, created_at FROM embeddings")
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}
	defer rows.Close()

	embeddings := make([]Embedding, 0)
	for rows.Next() {
		var embedding Embedding
		var vectorJSON, metadataJSON string

		err := rows.Scan(&embedding.ID, &embedding.Type, &embedding.Text, &vectorJSON, &metadataJSON, &embedding.Created)
		if err != nil {
			return nil, fmt.Errorf("failed to scan embedding row: %w", err)
		}

		json.Unmarshal([]byte(vectorJSON), &embedding.Vector)
		json.Unmarshal([]byte(metadataJSON), &embedding.Metadata)

		embeddings = append(embeddings, embedding)
	}

	return embeddings, nil
}

// Close closes the database connection
func (i *Indexer) Close() error {
	return i.db.Close()
}

// Helper function to scan symbols from database rows
func (i *Indexer) scanSymbols(rows *sql.Rows) ([]Symbol, error) {
	symbols := make([]Symbol, 0)
	for rows.Next() {
		var symbol Symbol
		var id string
		var importsJSON, callersJSON, callsJSON, paramsJSON, returnsJSON, fieldsJSON, methodsJSON string
		var createdAt time.Time

		err := rows.Scan(
			&id, &symbol.Name, &symbol.Type, &symbol.Package,
			&symbol.File, &symbol.Line, &symbol.Column, &symbol.Signature,
			&symbol.Doc, &importsJSON, &callersJSON, &callsJSON,
			&symbol.Public, &paramsJSON, &returnsJSON, &fieldsJSON,
			&methodsJSON, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan symbol row: %w", err)
		}

		// Parse JSON fields
		json.Unmarshal([]byte(importsJSON), &symbol.Imports)
		json.Unmarshal([]byte(callersJSON), &symbol.Callers)
		json.Unmarshal([]byte(callsJSON), &symbol.Calls)
		json.Unmarshal([]byte(paramsJSON), &symbol.Params)
		json.Unmarshal([]byte(returnsJSON), &symbol.Returns)
		json.Unmarshal([]byte(fieldsJSON), &symbol.Fields)
		json.Unmarshal([]byte(methodsJSON), &symbol.Methods)

		symbols = append(symbols, symbol)
	}
	return symbols, nil
}

// DeleteSymbolsByFile removes all symbols associated with a specific file
func (i *Indexer) DeleteSymbolsByFile(ctx context.Context, filePath string) error {
	// Delete symbols from the symbols table
	_, err := i.db.ExecContext(ctx, "DELETE FROM symbols WHERE file = ?", filePath)
	if err != nil {
		return fmt.Errorf("failed to delete symbols from file %s: %w", filePath, err)
	}

	// Delete associated embeddings
	_, err = i.db.ExecContext(ctx, "DELETE FROM embeddings WHERE metadata LIKE ?", "%"+filePath+"%")
	if err != nil {
		return fmt.Errorf("failed to delete embeddings from file %s: %w", filePath, err)
	}

	return nil
}
