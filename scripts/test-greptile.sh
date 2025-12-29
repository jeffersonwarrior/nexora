#!/bin/bash

# Greptile API Test Script
# This script helps test the Greptile API integration locally

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
API_KEY=""
GITHUB_TOKEN=""
REPO=""
BRANCH="main"
SHA=""
HELP=false

# Function to print colored output
print_color() {
    printf "${1}${2}${NC}\n"
}

# Function to print usage
usage() {
    echo "Greptile API Test Script"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -k, --api-key        Greptile API key (or set GREPTILE_API_KEY env var)"
    echo "  -t, --github-token   GitHub token (or set GITHUB_TOKEN env var)"
    echo "  -r, --repo           Repository in format owner/repo (default: auto-detect)"
    echo "  -b, --branch         Branch to analyze (default: main)"
    echo "  -s, --sha            Specific commit SHA (default: latest)"
    echo "  -h, --help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 -k AVudwmE0XdbfQHob -t ghp_xxx -r nexora/nexora"
    echo "  $0 --api-key=AVudwmE0XdbfQHob --github-token=ghp_xxx"
    echo ""
    echo "Environment Variables:"
    echo "  GREPTILE_API_KEY     Greptile API key"
    echo "  GITHUB_TOKEN         GitHub token with repo access"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -k|--api-key)
            API_KEY="$2"
            shift 2
            ;;
        --api-key=*)
            API_KEY="${1#*=}"
            shift
            ;;
        -t|--github-token)
            GITHUB_TOKEN="$2"
            shift 2
            ;;
        --github-token=*)
            GITHUB_TOKEN="${1#*=}"
            shift
            ;;
        -r|--repo)
            REPO="$2"
            shift 2
            ;;
        --repo=*)
            REPO="${1#*=}"
            shift
            ;;
        -b|--branch)
            BRANCH="$2"
            shift 2
            ;;
        --branch=*)
            BRANCH="${1#*=}"
            shift
            ;;
        -s|--sha)
            SHA="$2"
            shift 2
            ;;
        --sha=*)
            SHA="${1#=*}"
            shift
            ;;
        -h|--help)
            HELP=true
            shift
            ;;
        *)
            print_color $RED "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Show help if requested
if [ "$HELP" = true ]; then
    usage
    exit 0
fi

# Get API key from environment if not provided
if [ -z "$API_KEY" ]; then
    API_KEY="${GREPTILE_API_KEY:-}"
fi

# Get GitHub token from environment if not provided
if [ -z "$GITHUB_TOKEN" ]; then
    GITHUB_TOKEN="${GITHUB_TOKEN:-}"
fi

# Auto-detect repository if not provided
if [ -z "$REPO" ]; then
    if git remote get-url origin &>/dev/null; then
        REPO=$(git remote get-url origin | sed 's|https://github.com/||' | sed 's|.git$||' | sed 's|git@github.com:||')
        print_color $BLUE "Auto-detected repository: $REPO"
    else
        print_color $RED "Could not auto-detect repository. Please specify with -r option."
        exit 1
    fi
fi

# Validate required parameters
if [ -z "$API_KEY" ]; then
    print_color $RED "Error: Greptile API key is required"
    print_color $YELLOW "Set GREPTILE_API_KEY environment variable or use -k option"
    echo ""
    usage
    exit 1
fi

if [ -z "$GITHUB_TOKEN" ]; then
    print_color $RED "Error: GitHub token is required"
    print_color $YELLOW "Set GITHUB_TOKEN environment variable or use -t option"
    echo ""
    usage
    exit 1
fi

# Get latest SHA if not provided
if [ -z "$SHA" ]; then
    SHA=$(git rev-parse "$BRANCH" 2>/dev/null || git rev-parse HEAD)
    print_color $BLUE "Using SHA: $SHA"
fi

# Print configuration
print_color $BLUE "\n=== Configuration ==="
echo "Repository: $REPO"
echo "Branch: $BRANCH"
echo "SHA: $SHA"
echo "API Key: ${API_KEY:0:10}..."
echo "GitHub Token: ${GITHUB_TOKEN:0:10}..."
echo ""

# Create request JSON
cat > greptile-request.json << EOF
{
    "remote": "github",
    "repository": "$REPO",
    "branch": "$BRANCH",
    "reload": true,
    "notify": true
}
EOF

print_color $BLUE "=== Request JSON ==="
cat greptile-request.json | jq .
echo ""

# Submit repository for indexing
print_color $BLUE "=== Submitting Repository for Indexing ==="

RESPONSE=$(curl -s -X POST \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -H "X-GitHub-Token: $GITHUB_TOKEN" \
    -d @greptile-request.json \
    https://api.greptile.com/v2/repositories)

print_color $BLUE "Greptile API Response:"
echo "$RESPONSE" | jq .

# Extract message and status endpoint
MESSAGE=$(echo "$RESPONSE" | jq -r '.message // empty')
STATUS_ENDPOINT=$(echo "$RESPONSE" | jq -r '.statusEndpoint // empty')

if [ -z "$MESSAGE" ] || [ "$MESSAGE" = "null" ]; then
    print_color $RED "❌ Failed to submit repository for indexing"
    print_color $YELLOW "Check your API key and repository access"
    exit 1
fi

print_color $GREEN "✅ Repository submitted for indexing successfully!"
print_color $GREEN "Message: $MESSAGE"
print_color $GREEN "Status Endpoint: $STATUS_ENDPOINT"

# Wait for completion
print_color $BLUE "\n=== Waiting for Indexing Completion ==="

MAX_WAIT=300
WAIT_INTERVAL=10
ELAPSED=0

while [ $ELAPSED -lt $MAX_WAIT ]; do
    # Check indexing status
    STATUS_RESPONSE=$(curl -s -X GET \
        -H "Authorization: Bearer $API_KEY" \
        "$STATUS_ENDPOINT")
    
    # Extract status from the response
    STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.status // "unknown"')
    echo "[$(date +%H:%M:%S)] Status: $STATUS (elapsed: ${ELAPSED}s)"
    
    if [ "$STATUS" = "completed" ]; then
        print_color $GREEN "✅ Indexing completed!"
        break
    elif [ "$STATUS" = "failed" ] || [ "$STATUS" = "error" ]; then
        print_color $RED "❌ Indexing failed with status: $STATUS"
        echo "$STATUS_RESPONSE" | jq .
        exit 1
    fi
    
    sleep $WAIT_INTERVAL
    ELAPSED=$((ELAPSED + WAIT_INTERVAL))
done

if [ $ELAPSED -ge $MAX_WAIT ]; then
    print_color $YELLOW "⚠️ Indexing timed out after $MAX_WAIT seconds"
    print_color $BLUE "The indexing may still be processing. Check the Greptile dashboard."
    exit 1
fi

# Show results
print_color $BLUE "\n=== Review Results ==="

# Extract and display results
SUMMARY=$(echo "$STATUS_RESPONSE" | jq -r '.summary // "No summary available"')
SCORE=$(echo "$STATUS_RESPONSE" | jq -r '.score // null')
CONFIDENCE=$(echo "$STATUS_RESPONSE" | jq -r '.confidence // null')
ISSUES=$(echo "$STATUS_RESPONSE" | jq -r '.issues // []')
ISSUES_COUNT=$(echo "$ISSUES" | jq 'length')

echo ""
print_color $BLUE "Summary:"
echo "$SUMMARY"

if [ "$SCORE" != "null" ] && [ -n "$SCORE" ]; then
    echo ""
    print_color $BLUE "Score: $SCORE/100"
fi

if [ "$CONFIDENCE" != "null" ] && [ -n "$CONFIDENCE" ]; then
    echo ""
    print_color $BLUE "Confidence: $CONFIDENCE"
fi

echo ""
if [ "$ISSUES_COUNT" -gt 0 ]; then
    print_color $YELLOW "Issues found: $ISSUES_COUNT"
    echo ""
    
    echo "$ISSUES" | jq -r '.[] | @base64' | while read -r issue; do
        issue_json=$(echo "$issue" | base64 -d)
        
        severity=$(echo "$issue_json" | jq -r '.severity // "info"')
        title=$(echo "$issue_json" | jq -r '.title // "Untitled Issue"')
        description=$(echo "$issue_json" | jq -r '.description // "No description"')
        file=$(echo "$issue_json" | jq -r '.file // null')
        line=$(echo "$issue_json" | jq -r '.line // null')
        
        # Add severity emoji
        case "$severity" in
            "critical"|"high") emoji="🚨" ;;
            "medium") emoji="⚠️" ;;
            "low") emoji="💡" ;;
            *) emoji="ℹ️" ;;
        esac
        
        print_color $YELLOW "$emoji $title"
        echo "$description"
        
        if [ "$file" != "null" ] && [ -n "$file" ]; then
            if [ "$line" != "null" ] && [ -n "$line" ]; then
                echo "Location: $file:$line"
            else
                echo "Location: $file"
            fi
        fi
        echo "---"
        echo ""
    done
else
    print_color $GREEN "✅ No issues found!"
fi

# Cleanup
rm -f greptile-request.json

print_color $GREEN "\n✅ Indexing completed successfully!"
echo "Status Endpoint: $STATUS_ENDPOINT"
echo "You can now query this repository using the Greptile API"