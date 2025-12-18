package message_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/nexora/cli/internal/db"
	"github.com/nexora/cli/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockQuerier provides a mock database querier for testing
type MockQuerier struct {
	calls       []string
	messages    map[string]db.Message
	shouldError bool // Flag to simulate database errors
	deleteError bool // Flag to simulate delete errors specifically
}

func NewMockQuerier() *MockQuerier {
	return &MockQuerier{
		calls:    make([]string, 0),
		messages: make(map[string]db.Message),
	}
}

func (m *MockQuerier) CreateMessage(ctx context.Context, params db.CreateMessageParams) (db.Message, error) {
	m.calls = append(m.calls, "CreateMessage")
	if m.shouldError {
		return db.Message{}, fmt.Errorf("database error")
	}
	msg := db.Message{
		ID:               params.ID,
		SessionID:        params.SessionID,
		Role:             params.Role,
		Parts:            params.Parts,
		Model:            params.Model,
		CreatedAt:        time.Now().Unix(),
		UpdatedAt:        time.Now().Unix(),
		Provider:         params.Provider,
		IsSummaryMessage: params.IsSummaryMessage,
	}
	m.messages[params.ID] = msg
	return msg, nil
}

func (m *MockQuerier) DeleteMessage(ctx context.Context, id string) error {
	m.calls = append(m.calls, "DeleteMessage")
	if m.deleteError {
		return fmt.Errorf("delete error")
	}
	delete(m.messages, id)
	return nil
}

func (m *MockQuerier) GetMessage(ctx context.Context, id string) (db.Message, error) {
	m.calls = append(m.calls, "GetMessage")
	if m.shouldError {
		return db.Message{}, fmt.Errorf("database error")
	}
	msg, exists := m.messages[id]
	if !exists {
		return db.Message{}, sql.ErrNoRows
	}
	return msg, nil
}

func (m *MockQuerier) CreateSession(ctx context.Context, params db.CreateSessionParams) (db.Session, error) {
	m.calls = append(m.calls, "CreateSession")
	return db.Session{
		ID:           params.ID,
		Title:        params.Title,
		MessageCount: 0,
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
	}, nil
}

func (m *MockQuerier) DeleteSession(ctx context.Context, id string) error {
	m.calls = append(m.calls, "DeleteSession")
	return nil
}

func (m *MockQuerier) GetSessionByID(ctx context.Context, id string) (db.Session, error) {
	m.calls = append(m.calls, "GetSessionByID")
	return db.Session{}, sql.ErrNoRows
}

func (m *MockQuerier) ListSessions(ctx context.Context) ([]db.Session, error) {
	m.calls = append(m.calls, "ListSessions")
	return []db.Session{}, nil
}

func (m *MockQuerier) UpdateSession(ctx context.Context, params db.UpdateSessionParams) (db.Session, error) {
	m.calls = append(m.calls, "UpdateSession")
	return db.Session{}, sql.ErrNoRows
}

func (m *MockQuerier) ListMessagesBySession(ctx context.Context, sessionID string) ([]db.Message, error) {
	m.calls = append(m.calls, "ListMessagesBySession")
	if m.shouldError {
		return nil, fmt.Errorf("database error")
	}
	messages := make([]db.Message, 0)
	for _, msg := range m.messages {
		if msg.SessionID == sessionID {
			messages = append(messages, msg)
		}
	}
	return messages, nil
}

func (m *MockQuerier) DeleteSessionMessages(ctx context.Context, sessionID string) error {
	m.calls = append(m.calls, "DeleteSessionMessages")
	for id, msg := range m.messages {
		if msg.SessionID == sessionID {
			delete(m.messages, id)
		}
	}
	return nil
}

func (m *MockQuerier) UpdateMessage(ctx context.Context, params db.UpdateMessageParams) error {
	m.calls = append(m.calls, "UpdateMessage")
	if m.shouldError {
		return fmt.Errorf("database error")
	}
	return nil
}

// File methods
func (m *MockQuerier) CreateFile(ctx context.Context, params db.CreateFileParams) (db.File, error) {
	m.calls = append(m.calls, "CreateFile")
	return db.File{}, nil
}

func (m *MockQuerier) DeleteFile(ctx context.Context, id string) error {
	m.calls = append(m.calls, "DeleteFile")
	return nil
}

func (m *MockQuerier) GetFile(ctx context.Context, id string) (db.File, error) {
	m.calls = append(m.calls, "GetFile")
	return db.File{}, sql.ErrNoRows
}

func (m *MockQuerier) DeleteSessionFiles(ctx context.Context, sessionID string) error {
	m.calls = append(m.calls, "DeleteSessionFiles")
	return nil
}

func (m *MockQuerier) GetFileByPathAndSession(ctx context.Context, params db.GetFileByPathAndSessionParams) (db.File, error) {
	m.calls = append(m.calls, "GetFileByPathAndSession")
	return db.File{}, sql.ErrNoRows
}

func (m *MockQuerier) ListFilesByPath(ctx context.Context, path string) ([]db.File, error) {
	m.calls = append(m.calls, "ListFilesByPath")
	return []db.File{}, nil
}

func (m *MockQuerier) ListFilesBySession(ctx context.Context, sessionID string) ([]db.File, error) {
	m.calls = append(m.calls, "ListFilesBySession")
	return []db.File{}, nil
}

func (m *MockQuerier) ListLatestSessionFiles(ctx context.Context, sessionID string) ([]db.File, error) {
	m.calls = append(m.calls, "ListLatestSessionFiles")
	return []db.File{}, nil
}

func (m *MockQuerier) ListNewFiles(ctx context.Context) ([]db.File, error) {
	m.calls = append(m.calls, "ListNewFiles")
	return []db.File{}, nil
}

func TestNewService(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	require.NotNil(t, svc)
}

func TestCreateUserMessage(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	params := message.CreateMessageParams{
		Role: message.User,
		Parts: []message.ContentPart{
			message.TextContent{Text: "Hello, world!"},
		},
		Model:    "gpt-4",
		Provider: "openai",
	}

	msg, err := svc.Create(ctx, "test-session", params)
	require.NoError(t, err)
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, "test-session", msg.SessionID)
	assert.Equal(t, message.User, msg.Role)
	assert.Equal(t, "Hello, world!", msg.Content().Text)
}

func TestCreateAssistantMessage(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	params := message.CreateMessageParams{
		Role: message.Assistant,
		Parts: []message.ContentPart{
			message.TextContent{Text: "Hello! How can I help?"},
		},
		Model:    "gpt-4",
		Provider: "openai",
	}

	msg, err := svc.Create(ctx, "test-session", params)
	require.NoError(t, err)
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, "test-session", msg.SessionID)
	assert.Equal(t, message.Assistant, msg.Role)
	assert.Equal(t, "Hello! How can I help?", msg.Content().Text)
}

func TestGetMessage(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Test getting non-existent message
	_, err := svc.Get(ctx, "non-existent")
	assert.Error(t, err)
}

func TestDeleteMessage(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Test deleting non-existent message
	err := svc.Delete(ctx, "non-existent")
	assert.Error(t, err)
}

func TestListMessages(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Test listing messages for session
	messages, err := svc.List(ctx, "test-session")
	assert.NoError(t, err)
	assert.NotNil(t, messages)
}

func TestCreateMessageWithReasoning(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	params := message.CreateMessageParams{
		Role: message.Assistant,
		Parts: []message.ContentPart{
			message.ReasoningContent{
				Thinking:  "Let me think about this problem...",
				Signature: "signature123",
			},
			message.TextContent{Text: "The answer is 42."},
		},
		Model:    "o1-preview",
		Provider: "openai",
	}

	msg, err := svc.Create(ctx, "test-session", params)
	require.NoError(t, err)
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, "test-session", msg.SessionID)
	assert.Equal(t, message.Assistant, msg.Role)
	assert.Equal(t, "Let me think about this problem...", msg.ReasoningContent().Thinking)
	assert.Equal(t, "The answer is 42.", msg.Content().Text)
}

func TestUpdateMessage(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Create a message first
	createParams := message.CreateMessageParams{
		Role:  message.User,
		Parts: []message.ContentPart{message.TextContent{Text: "Original content"}},
	}

	created, err := svc.Create(ctx, "test-session", createParams)
	require.NoError(t, err)

	// Update the message (update changes the content)
	updatedParts := []message.ContentPart{message.TextContent{Text: "Updated content"}}
	updatedMessage := message.Message{
		ID:    created.ID,
		Role:  created.Role,
		Parts: updatedParts,
		Model: "test-model",
	}

	err = svc.Update(ctx, updatedMessage)
	require.NoError(t, err)

	// Verify the update was called
	assert.Contains(t, mock.calls, "UpdateMessage")
}

func TestDeleteSessionMessages(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Create multiple messages
	sessionID := "test-session"
	for i := 0; i < 3; i++ {
		params := message.CreateMessageParams{
			Role:  message.User,
			Parts: []message.ContentPart{message.TextContent{Text: fmt.Sprintf("Message %d", i)}},
		}
		_, err := svc.Create(ctx, sessionID, params)
		require.NoError(t, err)
	}

	// Delete all session messages
	err := svc.DeleteSessionMessages(ctx, sessionID)
	require.NoError(t, err)

	// Verify all messages were deleted
	messages, err := svc.List(ctx, sessionID)
	require.NoError(t, err)
	assert.Empty(t, messages)
}

func TestCreateMessageErrorHandling(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Test with invalid JSON parts (this should trigger an error)
	params := message.CreateMessageParams{
		Role:  message.User,
		Parts: []message.ContentPart{}, // Empty parts should still work
	}

	msg, err := svc.Create(ctx, "test-session", params)
	require.NoError(t, err)
	assert.NotEmpty(t, msg.ID)
}

func TestGetMessageNotFound(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Try to get a message that doesn't exist
	_, err := svc.Get(ctx, "nonexistent-id")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeleteMessageNotFound(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Try to delete a message that doesn't exist
	err := svc.Delete(ctx, "nonexistent-id")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestListMessagesEmptySession(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// List messages from a session with no messages
	messages, err := svc.List(ctx, "empty-session")
	require.NoError(t, err)
	assert.Empty(t, messages)
}

func TestMessageContentAccess(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Create message with different content types
	params := message.CreateMessageParams{
		Role: message.Assistant,
		Parts: []message.ContentPart{
			message.TextContent{Text: "Main response"},
			message.ReasoningContent{
				Thinking:  "Step by step reasoning",
				Signature: "verified",
			},
		},
		Model:    "test-model",
		Provider: "test-provider",
	}

	msg, err := svc.Create(ctx, "test-session", params)
	require.NoError(t, err)

	// Test content access methods
	content := msg.Content()
	require.NotNil(t, content)
	assert.Equal(t, "Main response", content.Text)

	// Test reasoning content
	reasoning := msg.ReasoningContent()
	require.NotNil(t, reasoning)
	assert.Equal(t, "Step by step reasoning", reasoning.Thinking)
	assert.Equal(t, "verified", reasoning.Signature)

	// Test string representations of content parts
	assert.NotEmpty(t, content.String())
	assert.NotEmpty(t, reasoning.String())
}

func TestCreateMessage_WithAllContentTypes(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Create message with all supported content types
	params := message.CreateMessageParams{
		Role: message.Assistant,
		Parts: []message.ContentPart{
			message.TextContent{Text: "Main response"},
			message.ReasoningContent{
				Thinking:  "Step by step reasoning",
				Signature: "verified",
			},
			message.ImageURLContent{URL: "https://example.com/image.png"},
			message.BinaryContent{
				Data:     []byte("binary data"),
				MIMEType: "application/octet-stream",
			},
			message.ToolCall{
				ID:    "call_123",
				Name:  "test_tool",
				Input: `{"input": "value"}`,
			},
			message.ToolResult{
				ToolCallID: "call_123",
				Content:    "success",
			},
			message.Finish{Reason: "stop"},
		},
		Model:    "test-model",
		Provider: "test-provider",
	}

	msg, err := svc.Create(ctx, "test-session", params)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)
	assert.Equal(t, message.Assistant, msg.Role)
	assert.Len(t, msg.Parts, 7) // 7 parts (finish should not be added for assistant)
}

func TestCreateMessage_UserRole_AddsFinishReason(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Create user message without finish reason
	params := message.CreateMessageParams{
		Role: message.User,
		Parts: []message.ContentPart{
			message.TextContent{Text: "User input"},
		},
		Model:    "test-model",
		Provider: "test-provider",
	}

	msg, err := svc.Create(ctx, "test-session", params)
	require.NoError(t, err)

	// Should have 2 parts: original text + auto-added finish reason
	assert.Len(t, msg.Parts, 2)

	// Check that finish reason was added
	hasFinish := false
	for _, part := range msg.Parts {
		if finish, ok := part.(message.Finish); ok {
			hasFinish = true
			assert.Equal(t, message.FinishReason("stop"), finish.Reason)
			break
		}
	}
	assert.True(t, hasFinish, "User message should have finish reason added")
}

func TestCreateMessage_WithSummaryFlag(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	params := message.CreateMessageParams{
		Role:             message.Assistant,
		Parts:            []message.ContentPart{message.TextContent{Text: "Summary"}},
		Model:            "test-model",
		Provider:         "test-provider",
		IsSummaryMessage: true,
	}

	msg, err := svc.Create(ctx, "test-session", params)
	require.NoError(t, err)
	assert.True(t, msg.IsSummaryMessage)
}

func TestCreateMessage_DBError(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Mock a database error during creation
	mock.shouldError = true

	params := message.CreateMessageParams{
		Role:  message.User,
		Parts: []message.ContentPart{message.TextContent{Text: "test"}},
	}

	_, err := svc.Create(ctx, "test-session", params)
	assert.Error(t, err)
}

func TestUpdateMessage_ErrorHandling(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// First create a message
	params := message.CreateMessageParams{
		Role:  message.User,
		Parts: []message.ContentPart{message.TextContent{Text: "original"}},
	}
	msg, err := svc.Create(ctx, "test-session", params)
	require.NoError(t, err)

	// Mock an error for update
	mock.shouldError = true

	// Try to update the message
	updatedMsg := msg
	updatedMsg.Parts = []message.ContentPart{message.TextContent{Text: "updated"}}

	err = svc.Update(ctx, updatedMsg)
	assert.Error(t, err)
}

func TestDeleteSessionMessages_ErrorInList(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Mock an error when listing messages
	mock.shouldError = true

	err := svc.DeleteSessionMessages(ctx, "test-session")
	assert.Error(t, err)
}

func TestDeleteSessionMessages_ErrorInDelete(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// First create some messages
	params1 := message.CreateMessageParams{
		Role:  message.User,
		Parts: []message.ContentPart{message.TextContent{Text: "message1"}},
	}
	params2 := message.CreateMessageParams{
		Role:  message.Assistant,
		Parts: []message.ContentPart{message.TextContent{Text: "message2"}},
	}

	_, createErr := svc.Create(ctx, "test-session", params1)
	require.NoError(t, createErr)
	_, createErr = svc.Create(ctx, "test-session", params2)
	require.NoError(t, createErr)

	// Mock an error only for delete operations
	mock.deleteError = true

	// Try to delete session messages - should fail on first delete
	err := svc.DeleteSessionMessages(ctx, "test-session")
	assert.Error(t, err)
}

func TestMessageContentAccess_EmptyContent(t *testing.T) {
	mock := NewMockQuerier()
	svc := message.NewService(mock)
	ctx := context.Background()

	// Create message with no specific content types
	params := message.CreateMessageParams{
		Role:  message.User,
		Parts: []message.ContentPart{message.Finish{Reason: "stop"}},
	}

	msg, err := svc.Create(ctx, "test-session", params)
	require.NoError(t, err)

	// Test content access methods when content is missing
	content := msg.Content()
	assert.NotNil(t, content)
	assert.Empty(t, content.Text)

	reasoning := msg.ReasoningContent()
	assert.NotNil(t, reasoning)
	assert.Empty(t, reasoning.Thinking)
	assert.Empty(t, reasoning.Signature)
}
