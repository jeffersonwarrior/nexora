#!/bin/bash

# Apply conversation threading fix to prevent infinite loops
# This script modifies agent.go to use state-based continuation instead of phrase-based

set -e

echo "Applying conversation threading fix..."

# Backup the original file
cp internal/agent/agent.go internal/agent/agent.go.backup

# Add conversation manager to sessionAgent struct
echo "Adding conversation manager to struct..."

# Find the struct definition and add the field
awk '
/aiops\s+aiops\.Ops \/\/ AIOPS client for operational support/ {
    print
    print "	convoMgr            *ConversationManager // Manages conversation state"
    next
}
{ print }
' internal/agent/agent.go > internal/agent/agent_go.tmp

mv internal/agent/agent_go.tmp internal/agent/agent.go

# Add initialization in constructor (find NewSessionAgent function)
echo "Adding initialization in constructor..."

awk '
/return &sessionAgent\{/ {
    print
    print "		convoMgr:            NewConversationManager(),"
    next
}
{ print }
' internal/agent/agent.go > internal/agent/agent_go.tmp

mv internal/agent/agent_go.tmp internal/agent/agent.go

# Replace the phrase-based shouldContinueAfterTool function
echo "Replacing phrase-based continuation with state-based..."

# Create the new function
cat > /tmp/new_should_continue.txt << 'EOF'
// shouldContinueAfterTool checks if the conversation should continue based on state
func (a *sessionAgent) shouldContinueAfterTool(ctx context.Context, sessionID string, currentAssistant *message.Message) bool {
	// Use the conversation manager to determine continuation based on state
	return a.convoMgr.ShouldContinue(sessionID)
}
EOF

# Find and replace the function
awk '
/^\/\/ shouldContinueAfterTool checks if the last AI response suggests unfinished work$/ {
    # Skip old function
    skip = 1
    next
}
skip && /^}$/ {
    skip = 0
    # Insert new function
    while ((getline line < "/tmp/new_should_continue.txt") > 0) {
        print line
    }
    close("/tmp/new_should_continue.txt")
    next
}
skip { next }
{ print }
' internal/agent/agent.go > internal/agent/agent_go.tmp

mv internal/agent/agent_go.tmp internal/agent/agent.go

# Update the Run method to use the conversation manager
echo "Updating Run method to use conversation manager..."

# Find the section where shouldContinue is determined
awk '
/shouldContinue := hasToolResults && currentAssistant \!= nil && len\(currentAssistant\.ToolCalls\(\)\) > 0/ {
    print
    print "	if !shouldContinue && a.convoMgr.IsConversationCompleted(call.SessionID) {"
    print "		// Don/'"'"'t continue - conversation is marked as completed"
    print "		cancel()"
    print "		return result, err"
    print "	}"
    
    # Add message recording before this section
    print ""
    print "	// Record the assistant message and let the manager handle state"
    print "	if currentAssistant != nil {"
    print "		a.convoMgr.RecordMessage(call.SessionID, *currentAssistant)"
    print "	}"
    print ""
    
    next
}
{ print }
' internal/agent/agent.go > internal/agent/agent_go.tmp

mv internal/agent/agent_go.tmp internal/agent/agent.go

# Update conversation_loop.go to disable phrase-based auto-continue
echo "Updating conversation_loop.go..."

cp internal/agent/conversation_loop.go internal/agent/conversation_loop.go.backup

# Replace ShouldAutoContinue to return false
awk '
/SouldAutoContinue determines if the conversation should automatically continue/ {
    print
    print "func (c *ConversationLoopManager) ShouldAutoContinue(sessionID string, msg message.Message) bool {"
    print "	// Don'"'"'t use auto-continue based on tool results alone"
    print "	// Let the conversation manager handle state-based continuation"
    print "	return false"
    print "}"
    next
}
{
    # Skip the original function
    if (skip && /^}/) {
        skip = 0
        next
    }
    if (skip) next
    if (/func \(c \*ConversationLoopManager\) ShouldAutoContinue/) {
        skip = 1
        next
    }
    print
}
' internal/agent/conversation_loop.go > internal/agent/conversation_loop_go.tmp

mv internal/agent/conversation_loop_go.tmp internal/agent/conversation_loop.go

rm -f /tmp/new_should_continue.txt

echo "Fix applied successfully!"
echo ""
echo "Changes made:"
echo "1. Added ConversationManager to sessionAgent struct"
echo "2. Initialized ConversationManager in constructor"
echo "3. Replaced phrase-based continuation with state-based continuation"
echo "4. Added message recording to track conversation state"
echo "5. Disabled phrase-based auto-continue in conversation loop"
echo ""
echo "To test the fix:"
echo "  go test ./internal/agent/..."
echo "  go build ./cmd/nexora"
echo ""
echo "To rollback if needed:"
echo "  cp internal/agent/agent.go.backup internal/agent/agent.go"
echo "  cp internal/agent/conversation_loop.go.backup internal/agent/conversation_loop.go"