package cmd

import (
	"testing"

	"github.com/nexora/nexora/internal/indexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseQueryType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected indexer.QueryType
		wantErr  bool
	}{
		{
			name:     "All type",
			input:    "all",
			expected: indexer.QueryTypeAll,
			wantErr:  false,
		},
		{
			name:     "Semantic type",
			input:    "semantic",
			expected: indexer.QueryTypeSemantic,
			wantErr:  false,
		},
		{
			name:     "Text type",
			input:    "text",
			expected: indexer.QueryTypeText,
			wantErr:  false,
		},
		{
			name:     "Graph type",
			input:    "graph",
			expected: indexer.QueryTypeGraph,
			wantErr:  false,
		},
		{
			name:     "Case insensitive - ALL",
			input:    "ALL",
			expected: indexer.QueryTypeAll,
			wantErr:  false,
		},
		{
			name:     "Case insensitive - Semantic",
			input:    "Semantic",
			expected: indexer.QueryTypeSemantic,
			wantErr:  false,
		},
		{
			name:    "Unknown type",
			input:   "unknown",
			wantErr: true,
		},
		{
			name:    "Invalid type",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseQueryType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unknown query type")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestQueryCmdMetadata(t *testing.T) {
	t.Run("Command has correct metadata", func(t *testing.T) {
		assert.Equal(t, "query [search terms]", queryCmd.Use)
		assert.Equal(t, "Search indexed code using AI-accelerated search", queryCmd.Short)
		assert.Contains(t, queryCmd.Long, "Semantic similarity")
		assert.NotEmpty(t, queryCmd.Long)
	})

	t.Run("Has correct flags", func(t *testing.T) {
		flags := queryCmd.Flags()

		dbFlag := flags.Lookup("database")
		assert.NotNil(t, dbFlag)
		assert.Equal(t, "b", dbFlag.Shorthand)

		typeFlag := flags.Lookup("type")
		assert.NotNil(t, typeFlag)
		assert.Equal(t, "t", typeFlag.Shorthand)
		assert.Equal(t, "all", typeFlag.DefValue)

		limitFlag := flags.Lookup("limit")
		assert.NotNil(t, limitFlag)
		assert.Equal(t, "l", limitFlag.Shorthand)
		assert.Equal(t, "20", limitFlag.DefValue)

		explainFlag := flags.Lookup("explain")
		assert.NotNil(t, explainFlag)
		assert.Equal(t, "e", explainFlag.Shorthand)

		contextFlag := flags.Lookup("context")
		assert.NotNil(t, contextFlag)
		assert.Equal(t, "x", contextFlag.Shorthand)
	})
}
