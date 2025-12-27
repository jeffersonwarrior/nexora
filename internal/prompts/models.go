package prompts

import (
	"database/sql"

	"github.com/nexora/nexora/internal/db"
)

// Prompt represents a prompt in the library
type Prompt struct {
	ID             string
	Category       string
	Subcategory    sql.NullString
	Title          string
	Description    string
	Content        string
	ContentHash    sql.NullString
	Tags           sql.NullString
	Variables      sql.NullString
	Author         sql.NullString
	Source         sql.NullString
	SourceURL      sql.NullString
	Votes          sql.NullInt64
	Rating         sql.NullFloat64
	UsageCount     sql.NullInt64
	SuccessRate    sql.NullFloat64
	AvgTokens      sql.NullInt64
	AvgLatencyMs   sql.NullInt64
	LastUsedAt     sql.NullInt64
	FavoritesCount sql.NullInt64
	CreatedAt      int64
	UpdatedAt      int64
}

// FromDBPromptLibrary converts a db.PromptLibrary to a Prompt
func FromDBPromptLibrary(dbPrompt db.PromptLibrary) Prompt {
	return Prompt{
		ID:             dbPrompt.ID,
		Category:       dbPrompt.Category,
		Subcategory:    dbPrompt.Subcategory,
		Title:          dbPrompt.Title,
		Description:    dbPrompt.Description,
		Content:        dbPrompt.Content,
		ContentHash:    dbPrompt.ContentHash,
		Tags:           dbPrompt.Tags,
		Variables:      dbPrompt.Variables,
		Author:         dbPrompt.Author,
		Source:         dbPrompt.Source,
		SourceURL:      dbPrompt.SourceUrl,
		Votes:          dbPrompt.Votes,
		Rating:         dbPrompt.Rating,
		UsageCount:     dbPrompt.UsageCount,
		SuccessRate:    dbPrompt.SuccessRate,
		AvgTokens:      dbPrompt.AvgTokens,
		AvgLatencyMs:   dbPrompt.AvgLatencyMs,
		LastUsedAt:     dbPrompt.LastUsedAt,
		FavoritesCount: dbPrompt.FavoritesCount,
		CreatedAt:      dbPrompt.CreatedAt,
		UpdatedAt:      dbPrompt.UpdatedAt,
	}
}

// ListOptions holds options for listing prompts
type ListOptions struct {
	Limit  int64
	Offset int64
}

// SearchOptions holds options for searching prompts
type SearchOptions struct {
	Query string
	Limit int64
}
