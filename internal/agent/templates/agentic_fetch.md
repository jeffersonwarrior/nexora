Fetches content from a URL or searches the web, then processes it using AI to extract information or answer questions.

<when_to_use>
**Use for**: Web search, extract info from URLs, analyze/summarize web pages, research topics
**DON'T use**: Raw content (use fetch tool - faster), API/JSON responses, non-analyzed content
</when_to_use>

<usage>
- Provide prompt describing what information to find/extract (required)
- Optional URL to fetch and analyze specific content
- No URL = agent searches web for relevant information
- Spawns sub-agent with web_search, web_fetch, analysis tools
</usage>

<parameters>
- prompt: What information to find/extract (required)
- url: URL to fetch from (optional - if omitted, searches web)
</parameters>

<usage_notes>
- IMPORTANT: Prefer MCP-provided web tools (start with "mcp_") if available - fewer restrictions
- URL mode: Must be fully-formed valid URL (HTTP → HTTPS auto-upgrade)
- Search mode: Just provide prompt - agent searches and fetches pages
- Read-only, does not modify files
- Uses AI processing = more tokens than fetch tool
</usage_notes>

<limitations>
Max 5MB per page; HTTP/HTTPS only; no auth/cookies; sites may block; depends on DuckDuckGo
</limitations>

<tips>
- Be specific in prompt
- Omit URL for research tasks (agent searches)
- Ask agent to focus on specific sections for complex pages
- Use fetch tool instead if you just need raw content
</tips>

<examples>
Search: prompt="What are the main new features in the latest Python release?"
Analyze URL: url="https://docs.python.org/3/whatsnew/3.12.html" prompt="Summarize key changes"
</examples>
