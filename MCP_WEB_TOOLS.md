# Web Search and Web Reader MCP Tools

This configuration adds web search and web reader capabilities to Nexora through Z.AI's MCP servers.

## Configuration

The `nexora.json` file now includes two new MCP servers:

### Web Reader MCP
- **Name**: `web-reader`
- **Type**: HTTP
- **URL**: `https://api.z.ai/api/mcp/web_reader/mcp`
- **Purpose**: Reads and processes web content from URLs

### Web Search Prime MCP
- **Name**: `web-search-prime`
- **Type**: HTTP
- **URL**: `https://api.z.ai/api/mcp/web_search_prime/mcp`
- **Purpose**: Performs web searches and returns results

## Usage

Once configured, these tools will be available to agents in Nexora. The tools can be used to:

1. **Web Search**: Search for information on the web
2. **Web Reading**: Fetch and process content from specific URLs

## API Key

The configuration uses the provided API key: `85c99bec0fa64a0d8a4a01463868667a.RsDzW0iuxtgvYqd2`

## Tool Integration

These MCP tools integrate with Nexora's existing tool system and will be available to agents when:
- The MCP servers are properly configured and accessible
- The tools are listed in the agent's allowed tools (if restricted)
- The MCP servers are not disabled in the configuration

## Testing

To test if the MCP tools are working:

1. Start Nexora: `go run -tags dev .`
2. The MCP servers should initialize during startup
3. Check the logs for MCP connection status
4. Use the tools through agent interactions

## Troubleshooting

If the MCP tools are not working:

1. Check network connectivity to `api.z.ai`
2. Verify the API key is correct and active
3. Check Nexora logs for MCP connection errors
4. Ensure the MCP servers are not disabled in the configuration

## Security

The API key is stored in the configuration file. Ensure this file has appropriate permissions and is not committed to version control if it contains sensitive information.