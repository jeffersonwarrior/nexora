package prompts

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"testing"

	"github.com/nexora/nexora/internal/db"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sql.DB, Service) {
	t.Helper()

	// Create a temporary directory for test database
	tmpDir := t.TempDir()
	ctx := context.Background()
	database, err := db.Connect(ctx, tmpDir)
	require.NoError(t, err)

	queries := db.New(database)
	service := NewService(queries)

	return database, service
}

func TestService_Create(t *testing.T) {
	t.Parallel()

	database, service := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	t.Run("creates prompt with all required fields", func(t *testing.T) {
		prompt, err := service.Create(ctx, CreatePromptParams{
			Category:    "development",
			Title:       "Test Prompt",
			Description: "A test prompt",
			Content:     "This is the prompt content",
			Author:      "test-author",
		})

		require.NoError(t, err)
		require.NotEmpty(t, prompt.ID)
		require.Equal(t, "development", prompt.Category)
		require.Equal(t, "Test Prompt", prompt.Title)
		require.Equal(t, "A test prompt", prompt.Description)
		require.Equal(t, "This is the prompt content", prompt.Content)
		require.True(t, prompt.Author.Valid)
		require.Equal(t, "test-author", prompt.Author.String)
		require.NotZero(t, prompt.CreatedAt)
		require.NotZero(t, prompt.UpdatedAt)
	})

	t.Run("creates prompt with content hash", func(t *testing.T) {
		content := "Unique content for hashing"
		prompt, err := service.Create(ctx, CreatePromptParams{
			Category:    "testing",
			Title:       "Hash Test",
			Description: "Testing content hash",
			Content:     content,
		})

		require.NoError(t, err)
		require.True(t, prompt.ContentHash.Valid)

		expectedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
		require.Equal(t, expectedHash, prompt.ContentHash.String)
	})

	t.Run("creates prompt with tags", func(t *testing.T) {
		prompt, err := service.Create(ctx, CreatePromptParams{
			Category:    "development",
			Title:       "Tagged Prompt",
			Description: "Testing tags",
			Content:     "Content with tags",
			Tags:        "golang,testing,automation",
		})

		require.NoError(t, err)
		require.True(t, prompt.Tags.Valid)
		require.Equal(t, "golang,testing,automation", prompt.Tags.String)
	})
}

func TestService_Get(t *testing.T) {
	t.Parallel()

	database, service := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	t.Run("retrieves existing prompt by ID", func(t *testing.T) {
		created, err := service.Create(ctx, CreatePromptParams{
			Category:    "development",
			Title:       "Retrievable Prompt",
			Description: "Can be retrieved",
			Content:     "Test content",
		})
		require.NoError(t, err)

		retrieved, err := service.Get(ctx, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, retrieved.ID)
		require.Equal(t, created.Title, retrieved.Title)
		require.Equal(t, created.Content, retrieved.Content)
	})

	t.Run("returns error for non-existent prompt", func(t *testing.T) {
		_, err := service.Get(ctx, "non-existent-id")
		require.Error(t, err)
		require.Equal(t, sql.ErrNoRows, err)
	})
}

func TestService_List(t *testing.T) {
	t.Parallel()

	database, service := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	// Create test prompts
	for i := 0; i < 5; i++ {
		_, err := service.Create(ctx, CreatePromptParams{
			Category:    "test",
			Title:       fmt.Sprintf("Prompt %d", i),
			Description: fmt.Sprintf("Description %d", i),
			Content:     fmt.Sprintf("Content %d", i),
		})
		require.NoError(t, err)
	}

	t.Run("lists prompts with pagination", func(t *testing.T) {
		prompts, err := service.List(ctx, ListOptions{
			Limit:  3,
			Offset: 0,
		})
		require.NoError(t, err)
		require.Len(t, prompts, 3)
	})

	t.Run("lists prompts with offset", func(t *testing.T) {
		prompts, err := service.List(ctx, ListOptions{
			Limit:  10,
			Offset: 2,
		})
		require.NoError(t, err)
		require.Len(t, prompts, 3)
	})

	t.Run("returns empty list when offset exceeds count", func(t *testing.T) {
		prompts, err := service.List(ctx, ListOptions{
			Limit:  10,
			Offset: 100,
		})
		require.NoError(t, err)
		require.Empty(t, prompts)
	})
}

func TestService_ListByCategory(t *testing.T) {
	t.Parallel()

	database, service := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	// Create prompts in different categories
	categories := []string{"development", "testing", "documentation"}
	for _, cat := range categories {
		for i := 0; i < 3; i++ {
			_, err := service.Create(ctx, CreatePromptParams{
				Category:    cat,
				Title:       fmt.Sprintf("%s Prompt %d", cat, i),
				Description: "Description",
				Content:     fmt.Sprintf("%s Content %d", cat, i), // Unique content per prompt
			})
			require.NoError(t, err)
		}
	}

	t.Run("lists prompts by category", func(t *testing.T) {
		prompts, err := service.ListByCategory(ctx, "development", 10)
		require.NoError(t, err)
		require.Len(t, prompts, 3)
		for _, p := range prompts {
			require.Equal(t, "development", p.Category)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		prompts, err := service.ListByCategory(ctx, "testing", 2)
		require.NoError(t, err)
		require.Len(t, prompts, 2)
	})
}

func TestService_Search(t *testing.T) {
	t.Parallel()

	database, service := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	// Create searchable prompts
	testCases := []struct {
		title       string
		description string
		content     string
		tags        string
	}{
		{"Code Review", "Review Go code", "Please review this Go code for best practices", "golang,code-review"},
		{"API Design", "Design REST API", "Design a RESTful API for user management", "api,rest,design"},
		{"Database Schema", "Create DB schema", "Create a database schema for e-commerce", "database,sql,schema"},
	}

	for _, tc := range testCases {
		_, err := service.Create(ctx, CreatePromptParams{
			Category:    "development",
			Title:       tc.title,
			Description: tc.description,
			Content:     tc.content,
			Tags:        tc.tags,
		})
		require.NoError(t, err)
	}

	t.Run("searches prompts by content", func(t *testing.T) {
		prompts, err := service.Search(ctx, SearchOptions{
			Query: "API",
			Limit: 10,
		})
		require.NoError(t, err)
		require.NotEmpty(t, prompts)

		// Should find the API Design prompt
		found := false
		for _, p := range prompts {
			if p.Title == "API Design" {
				found = true
				break
			}
		}
		require.True(t, found, "Should find API Design prompt")
	})

	t.Run("searches by tags", func(t *testing.T) {
		prompts, err := service.Search(ctx, SearchOptions{
			Query: "golang",
			Limit: 10,
		})
		require.NoError(t, err)
		require.NotEmpty(t, prompts)
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		prompts, err := service.Search(ctx, SearchOptions{
			Query: "xyznonexistentqueryterm",
			Limit: 10,
		})
		require.NoError(t, err)
		require.Empty(t, prompts)
	})
}

func TestService_Update(t *testing.T) {
	t.Parallel()

	database, service := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	t.Run("updates prompt fields", func(t *testing.T) {
		created, err := service.Create(ctx, CreatePromptParams{
			Category:    "development",
			Title:       "Original Title",
			Description: "Original Description",
			Content:     "Original Content",
		})
		require.NoError(t, err)

		err = service.Update(ctx, UpdatePromptParams{
			ID:          created.ID,
			Title:       "Updated Title",
			Description: "Updated Description",
			Content:     "Updated Content",
			Tags:        "new,tags",
		})
		require.NoError(t, err)

		updated, err := service.Get(ctx, created.ID)
		require.NoError(t, err)
		require.Equal(t, "Updated Title", updated.Title)
		require.Equal(t, "Updated Description", updated.Description)
		require.Equal(t, "Updated Content", updated.Content)
		require.True(t, updated.Tags.Valid)
		require.Equal(t, "new,tags", updated.Tags.String)
	})

	t.Run("updates content hash on content change", func(t *testing.T) {
		created, err := service.Create(ctx, CreatePromptParams{
			Category:    "development",
			Title:       "Hash Update Test",
			Description: "Testing",
			Content:     "Original Content",
		})
		require.NoError(t, err)
		originalHash := created.ContentHash.String

		newContent := "New Content"
		err = service.Update(ctx, UpdatePromptParams{
			ID:          created.ID,
			Title:       created.Title,
			Description: created.Description,
			Content:     newContent,
		})
		require.NoError(t, err)

		updated, err := service.Get(ctx, created.ID)
		require.NoError(t, err)

		expectedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(newContent)))
		require.Equal(t, expectedHash, updated.ContentHash.String)
		require.NotEqual(t, originalHash, updated.ContentHash.String)
	})
}

func TestService_Delete(t *testing.T) {
	t.Parallel()

	database, service := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	t.Run("deletes existing prompt", func(t *testing.T) {
		created, err := service.Create(ctx, CreatePromptParams{
			Category:    "development",
			Title:       "To Delete",
			Description: "Will be deleted",
			Content:     "Content",
		})
		require.NoError(t, err)

		err = service.Delete(ctx, created.ID)
		require.NoError(t, err)

		_, err = service.Get(ctx, created.ID)
		require.Error(t, err)
		require.Equal(t, sql.ErrNoRows, err)
	})

	t.Run("no error when deleting non-existent prompt", func(t *testing.T) {
		err := service.Delete(ctx, "non-existent-id")
		require.NoError(t, err)
	})
}

func TestService_IncrementUsage(t *testing.T) {
	t.Parallel()

	database, service := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	t.Run("increments usage count and updates timestamp", func(t *testing.T) {
		created, err := service.Create(ctx, CreatePromptParams{
			Category:    "development",
			Title:       "Usage Test",
			Description: "Testing usage tracking",
			Content:     "Content",
		})
		require.NoError(t, err)
		require.True(t, created.UsageCount.Valid)
		require.Equal(t, int64(0), created.UsageCount.Int64)
		require.False(t, created.LastUsedAt.Valid)

		err = service.IncrementUsage(ctx, created.ID)
		require.NoError(t, err)

		updated, err := service.Get(ctx, created.ID)
		require.NoError(t, err)
		require.True(t, updated.UsageCount.Valid)
		require.Equal(t, int64(1), updated.UsageCount.Int64)
		require.True(t, updated.LastUsedAt.Valid)
		require.NotZero(t, updated.LastUsedAt.Int64)
	})

	t.Run("increments multiple times", func(t *testing.T) {
		created, err := service.Create(ctx, CreatePromptParams{
			Category:    "development",
			Title:       "Multi Usage",
			Description: "Testing",
			Content:     "Multi usage test content",
		})
		require.NoError(t, err)

		for i := 0; i < 5; i++ {
			err = service.IncrementUsage(ctx, created.ID)
			require.NoError(t, err)
		}

		updated, err := service.Get(ctx, created.ID)
		require.NoError(t, err)
		require.True(t, updated.UsageCount.Valid)
		require.Equal(t, int64(5), updated.UsageCount.Int64)
	})
}

func TestService_UpdateRating(t *testing.T) {
	t.Parallel()

	database, service := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	t.Run("updates rating and increments vote count", func(t *testing.T) {
		created, err := service.Create(ctx, CreatePromptParams{
			Category:    "development",
			Title:       "Rating Test",
			Description: "Testing ratings",
			Content:     "Content",
		})
		require.NoError(t, err)
		require.True(t, created.Rating.Valid)
		require.Equal(t, 0.0, created.Rating.Float64)
		require.True(t, created.Votes.Valid)
		require.Equal(t, int64(0), created.Votes.Int64)

		err = service.UpdateRating(ctx, created.ID, 4.5)
		require.NoError(t, err)

		updated, err := service.Get(ctx, created.ID)
		require.NoError(t, err)
		require.True(t, updated.Rating.Valid)
		require.Equal(t, 4.5, updated.Rating.Float64)
		require.True(t, updated.Votes.Valid)
		require.Equal(t, int64(1), updated.Votes.Int64)
	})
}

func TestService_GetByTag(t *testing.T) {
	t.Parallel()

	database, service := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	// Create prompts with tags
	_, err := service.Create(ctx, CreatePromptParams{
		Category:    "development",
		Title:       "Go Prompt",
		Description: "Golang related",
		Content:     "Go programming language content",
		Tags:        "golang,backend,api",
	})
	require.NoError(t, err)

	_, err = service.Create(ctx, CreatePromptParams{
		Category:    "development",
		Title:       "Python Prompt",
		Description: "Python related",
		Content:     "Python programming language content",
		Tags:        "python,backend,ml",
	})
	require.NoError(t, err)

	t.Run("finds prompts by tag", func(t *testing.T) {
		prompts, err := service.GetByTag(ctx, "golang", 10)
		require.NoError(t, err)
		require.Len(t, prompts, 1)
		require.Equal(t, "Go Prompt", prompts[0].Title)
	})

	t.Run("finds multiple prompts with shared tag", func(t *testing.T) {
		prompts, err := service.GetByTag(ctx, "backend", 10)
		require.NoError(t, err)
		require.Len(t, prompts, 2)
	})
}

func TestService_ListTop(t *testing.T) {
	t.Parallel()

	database, service := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	// Create prompts with different ratings
	ratings := []float64{4.5, 3.0, 5.0, 2.5, 4.0}
	for i, rating := range ratings {
		prompt, err := service.Create(ctx, CreatePromptParams{
			Category:    "development",
			Title:       fmt.Sprintf("Prompt %d", i),
			Description: "Description",
			Content:     fmt.Sprintf("Content %d", i), // Unique content for each prompt
		})
		require.NoError(t, err)

		err = service.UpdateRating(ctx, prompt.ID, rating)
		require.NoError(t, err)
	}

	t.Run("returns prompts sorted by rating", func(t *testing.T) {
		prompts, err := service.ListTop(ctx, 3)
		require.NoError(t, err)
		require.Len(t, prompts, 3)

		// Should be sorted by rating descending
		require.True(t, prompts[0].Rating.Valid)
		require.Equal(t, 5.0, prompts[0].Rating.Float64)
		require.True(t, prompts[1].Rating.Valid)
		require.Equal(t, 4.5, prompts[1].Rating.Float64)
		require.True(t, prompts[2].Rating.Valid)
		require.Equal(t, 4.0, prompts[2].Rating.Float64)
	})
}
