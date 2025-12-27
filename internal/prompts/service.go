package prompts

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/nexora/internal/db"
)

// Service defines the interface for prompt management
type Service interface {
	Create(ctx context.Context, params CreatePromptParams) (Prompt, error)
	Get(ctx context.Context, id string) (Prompt, error)
	List(ctx context.Context, opts ListOptions) ([]Prompt, error)
	ListByCategory(ctx context.Context, category string, limit int64) ([]Prompt, error)
	Search(ctx context.Context, opts SearchOptions) ([]Prompt, error)
	Update(ctx context.Context, params UpdatePromptParams) error
	Delete(ctx context.Context, id string) error
	IncrementUsage(ctx context.Context, id string) error
	UpdateRating(ctx context.Context, id string, rating float64) error
	GetByTag(ctx context.Context, tag string, limit int64) ([]Prompt, error)
	ListTop(ctx context.Context, limit int64) ([]Prompt, error)
}

type service struct {
	q db.Querier
}

// NewService creates a new prompt service
func NewService(q db.Querier) Service {
	return &service{q: q}
}

// CreatePromptParams holds parameters for creating a prompt
type CreatePromptParams struct {
	Category    string
	Subcategory string
	Title       string
	Description string
	Content     string
	Tags        string
	Author      string
}

func (s *service) Create(ctx context.Context, params CreatePromptParams) (Prompt, error) {
	id := uuid.New().String()

	// Generate content hash
	hash := sha256.Sum256([]byte(params.Content))
	contentHash := fmt.Sprintf("%x", hash)

	dbPrompt, err := s.q.CreatePrompt(ctx, db.CreatePromptParams{
		ID:          id,
		Category:    params.Category,
		Subcategory: toNullString(params.Subcategory),
		Title:       params.Title,
		Description: params.Description,
		Content:     params.Content,
		ContentHash: toNullString(contentHash),
		Tags:        toNullString(params.Tags),
		Author:      toNullString(params.Author),
	})
	if err != nil {
		return Prompt{}, fmt.Errorf("failed to create prompt: %w", err)
	}

	return FromDBPromptLibrary(dbPrompt), nil
}

func (s *service) Get(ctx context.Context, id string) (Prompt, error) {
	dbPrompt, err := s.q.GetPrompt(ctx, id)
	if err != nil {
		return Prompt{}, err
	}

	return FromDBPromptLibrary(dbPrompt), nil
}

func (s *service) List(ctx context.Context, opts ListOptions) ([]Prompt, error) {
	if opts.Limit == 0 {
		opts.Limit = 10
	}

	dbPrompts, err := s.q.ListPrompts(ctx, db.ListPromptsParams{
		Limit:  opts.Limit,
		Offset: opts.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}

	prompts := make([]Prompt, 0, len(dbPrompts))
	for _, dbPrompt := range dbPrompts {
		prompts = append(prompts, FromDBPromptLibrary(dbPrompt))
	}

	return prompts, nil
}

func (s *service) ListByCategory(ctx context.Context, category string, limit int64) ([]Prompt, error) {
	if limit == 0 {
		limit = 10
	}

	dbPrompts, err := s.q.ListPromptsByCategory(ctx, db.ListPromptsByCategoryParams{
		Category: category,
		Limit:    limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts by category: %w", err)
	}

	prompts := make([]Prompt, 0, len(dbPrompts))
	for _, dbPrompt := range dbPrompts {
		prompts = append(prompts, FromDBPromptLibrary(dbPrompt))
	}

	return prompts, nil
}

func (s *service) Search(ctx context.Context, opts SearchOptions) ([]Prompt, error) {
	if opts.Limit == 0 {
		opts.Limit = 10
	}

	// FTS5 query - search across all indexed columns
	dbPrompts, err := s.q.SearchPrompts(ctx, db.SearchPromptsParams{
		Title:       opts.Query,
		Description: opts.Query,
		Content:     opts.Query,
		Tags:        opts.Query,
		Limit:       opts.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search prompts: %w", err)
	}

	prompts := make([]Prompt, 0, len(dbPrompts))
	for _, dbPrompt := range dbPrompts {
		prompts = append(prompts, FromDBPromptLibrary(dbPrompt))
	}

	return prompts, nil
}

// UpdatePromptParams holds parameters for updating a prompt
type UpdatePromptParams struct {
	ID          string
	Title       string
	Description string
	Content     string
	Tags        string
}

func (s *service) Update(ctx context.Context, params UpdatePromptParams) error {
	// Generate new content hash
	hash := sha256.Sum256([]byte(params.Content))
	contentHash := fmt.Sprintf("%x", hash)

	err := s.q.UpdatePrompt(ctx, db.UpdatePromptParams{
		Title:       params.Title,
		Description: params.Description,
		Content:     params.Content,
		ContentHash: toNullString(contentHash),
		Tags:        toNullString(params.Tags),
		ID:          params.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to update prompt: %w", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	err := s.q.DeletePrompt(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete prompt: %w", err)
	}

	return nil
}

func (s *service) IncrementUsage(ctx context.Context, id string) error {
	err := s.q.IncrementUsage(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to increment usage: %w", err)
	}

	return nil
}

func (s *service) UpdateRating(ctx context.Context, id string, rating float64) error {
	err := s.q.UpdateRating(ctx, db.UpdateRatingParams{
		Rating: sql.NullFloat64{Float64: rating, Valid: true},
		ID:     id,
	})
	if err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}

	return nil
}

func (s *service) GetByTag(ctx context.Context, tag string, limit int64) ([]Prompt, error) {
	if limit == 0 {
		limit = 10
	}

	dbPrompts, err := s.q.GetPromptsByTag(ctx, db.GetPromptsByTagParams{
		Tag:   sql.NullString{String: tag, Valid: true},
		Limit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get prompts by tag: %w", err)
	}

	prompts := make([]Prompt, 0, len(dbPrompts))
	for _, dbPrompt := range dbPrompts {
		prompts = append(prompts, FromDBPromptLibrary(dbPrompt))
	}

	return prompts, nil
}

func (s *service) ListTop(ctx context.Context, limit int64) ([]Prompt, error) {
	if limit == 0 {
		limit = 10
	}

	dbPrompts, err := s.q.ListTopPrompts(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list top prompts: %w", err)
	}

	prompts := make([]Prompt, 0, len(dbPrompts))
	for _, dbPrompt := range dbPrompts {
		prompts = append(prompts, FromDBPromptLibrary(dbPrompt))
	}

	return prompts, nil
}

// Helper function to convert string to sql.NullString
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
