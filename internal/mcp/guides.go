package mcp

func openclawGuide(baseURL string) string {
	return "# OpenClaw MCP Connection Guide\n\n" +
		"## Streamable HTTP (Recommended)\n\n" +
		"Add to your OpenClaw configuration:\n\n" +
		"```json\n" +
		"{\n" +
		`  "mcpServers": {` + "\n" +
		`    "ds2api": {` + "\n" +
		`      "type": "streamable_http",` + "\n" +
		"      \"url\": \"" + baseURL + `/mcp",` + "\n" +
		`      "headers": {` + "\n" +
		`        "Authorization": "Bearer YOUR_API_KEY"` + "\n" +
		"      }\n" +
		"    }\n" +
		"  }\n" +
		"}\n" +
		"```\n\n" +
		"## Stdio Mode\n\n" +
		"Run ds2api with --mcp-transport stdio flag:\n\n" +
		"```bash\n" +
		"ds2api --mcp-transport stdio\n" +
		"# Then configure:\n" +
		"{\n" +
		`  "mcpServers": {` + "\n" +
		`    "ds2api": {` + "\n" +
		`      "command": "ds2api",` + "\n" +
		`      "args": ["--mcp-mode", "stdio"],` + "\n" +
		`      "type": "stdio"` + "\n" +
		"    }\n" +
		"  }\n" +
		"}\n" +
		"```\n\n" +
		"## Available Tools\n\n" +
		"| Tool | Description |\n|------|-------------|\n" +
		"| chat | Send messages to DeepSeek/Qwen models |\n" +
		"| list_models | List all available models |\n" +
		"| get_status | Service health check |\n" +
		"| get_pool_status | Account pool monitoring |\n" +
		"| embeddings | Generate text embeddings |\n\n" +
		"## Model Support\n\n" +
		"**DeepSeek**: deepseek-chat, deepseek-reasoner, deepseek-chat-search, deepseek-reasoner-search\n" +
		"**Qwen (通义千问)**: qwen-plus, qwen-max, qwen-coder, qwen-flash, qwen3.5-plus, qwen3.5-flash\n"
}

func claudeCodeGuide(baseURL string) string {
	return "# Claude Code MCP Connection Guide\n\n" +
		"## Method 1: Stdio Mode (Recommended for Claude Code)\n\n" +
		"Edit or create ~/.claude/settings.json:\n\n" +
		"```json\n" +
		"{\n" +
		`  "mcpServers": {` + "\n" +
		`    "ds2api": {` + "\n" +
		`      "command": "ds2api",` + "\n" +
		`      "args": ["--mcp-mode", "stdio"],` + "\n" +
		`      "env": {}` + "\n" +
		"    }\n" +
		"  }\n" +
		"}\n" +
		"```\n\n" +
		"Then restart Claude Code or run `claude`.\n\n" +
		"## Method 2: Remote Server (SSE/Streamable HTTP)\n\n" +
		"For connecting to a running ds2api instance:\n\n" +
		"```json\n" +
		"{\n" +
		`  "mcpServers": {` + "\n" +
		`    "ds2api": {` + "\n" +
		`      "type": "streamable_http",` + "\n" +
		"      \"url\": \"" + baseURL + `/mcp"` + "\n" +
		"    }\n" +
		"  }\n" +
		"}\n" +
		"```\n\n" +
		"## Usage Examples in Claude Code\n\n" +
		"After connecting, ask Claude Code things like:\n\n" +
		"- Use the chat tool to ask deepseek-reasoner about Rust async patterns\n" +
		"- List all available models via the list_models tool\n" +
		"- Check account pool status with get_pool_status\n" +
		"- Generate embeddings for these documents\n\n" +
		"## Tips for Claude Code\n\n" +
		"1. **Model Selection**: Use deepseek-reasoner for complex reasoning tasks\n" +
		"2. **Qwen Models**: Use qwen-max or qwen3.5-plus for Chinese language tasks\n" +
		"3. **Streaming**: The chat tool supports streaming for long responses\n" +
		"4. **Pool Monitoring**: Check pool status before heavy workloads\n"
}

func jetbrainsGuide(baseURL string) string {
	return "# JetBrains IDEs MCP Connection Guide\n\n" +
		"## Supported IDEs\n\n" +
		"- IntelliJ IDEA (2024.3+)\n" +
		"- PyCharm (2024.3+)\n" +
		"- WebStorm (2024.3+)\n" +
		"- GoLand (2024.3+)\n" +
		"- PhpStorm (2024.3+)\n\n" +
		"## Configuration Steps\n\n" +
		"### Step 1: Install MCP Plugin\n\n" +
		"In your JetBrains IDE:\n" +
		"1. Go to Settings -> Plugins\n" +
		`2. Search for "MCP" or "Model Context Protocol"` + "\n" +
		"3. Install the official MCP plugin\n" +
		"4. Restart IDE\n\n" +
		"### Step 2: Add ds2api Server\n\n" +
		"In Settings -> Tools -> Model Context Protocol -> Servers:\n\n" +
		"Click + to add a new server:\n\n" +
		"```\n" +
		"Name: ds2api\n" +
		"Type: Streamable HTTP\n" +
		"URL: " + baseURL + "/mcp\n" +
		"Headers: Authorization: Bearer YOUR_API_KEY\n" +
		"```\n\n" +
		"### Step 3: Configure AI Assistant\n\n" +
		"Go to Settings -> Tools -> AI Assistant:\n" +
		"1. Select Custom OpenAI-Compatible Provider\n" +
		"2. Base URL: " + baseURL + "\n" +
		"3. API Key: YOUR_API_KEY\n" +
		"4. Model: deepseek-chat (or any supported model)\n\n" +
		"### Step 4: Verify\n\n" +
		"Open AI Assistant panel (Cmd/Ctrl + Shift + A):\n" +
		"- Type: \"What models are available?\"\n" +
		"Claude Code should call list_models tool automatically\n\n" +
		"## Using Tools in JetBrains\n\n" +
		"The MCP tools appear in the AI Assistant context:\n\n" +
		"| When to Use | Tool | Example Prompt |\n" +
		"|-------------|------|---------------|\n" +
		"| Code generation | chat | Write a Go HTTP handler using deepseek-coder |\n" +
		"| Research | chat | Explain this algorithm using qwen-max |\n" +
		"| Debug help | get_status | Is the API service healthy? |\n" +
		"| Monitor | get_pool_status | How many accounts are available? |\n\n" +
		"## Keyboard Shortcuts\n\n" +
		"- AI Assistant: Cmd/Ctrl + Shift + A\n" +
		"- Tool Results: Shown inline in chat panel\n"
}

func opencodeGuide(baseURL string) string {
	return "# OpenCode MCP Connection Guide\n\n" +
		"## Prerequisites\n\n" +
		"- [OpenCode installed](https://github.com/opencode-ai/opencode)\n" +
		"- ds2api running locally or accessible via network\n\n" +
		"## Configuration\n\n" +
		"### Option 1: Stdio (Recommended)\n\n" +
		"Add to your opencode config file (~/.config/opencode/config.json):\n\n" +
		"```json\n" +
		"{\n" +
		`  "provider": "openai-compatible",` + "\n" +
		`  "model": "deepseek-chat",` + "\n" +
		`  "mcp": {` + "\n" +
		`    "servers": [` + "\n" +
		"      {\n" +
		`        "name": "ds2api",` + "\n" +
		`        "command": "ds2api",` + "\n" +
		`        "args": ["--mcp-mode", "stdio"]` + "\n" +
		"      }\n" +
		"    ]\n" +
		"  },\n" +
		`  "openai_compat": {` + "\n" +
		`    "base_url": "` + baseURL + `",` + "\n" +
		`    "api_key": "YOUR_API_KEY"` + "\n" +
		"  }\n" +
		"}\n" +
		"```\n\n" +
		"### Option 2: HTTP Fallback\n\n" +
		"If you prefer HTTP transport:\n\n" +
		"```json\n" +
		"{\n" +
		`  "mcp": {` + "\n" +
		`    "servers": [` + "\n" +
		"      {\n" +
		`        "name": "ds2api",` + "\n" +
		`        "type": "http",` + "\n" +
		"        \"url\": \"" + baseURL + `/mcp"` + "\n" +
		"      }\n" +
		"    ]\n" +
		"  }\n" +
		"}\n" +
		"```\n\n" +
		"## Running\n\n" +
		"```bash\n" +
		"# Terminal 1: Start ds2api\n" +
		"ds2api\n\n" +
		"# Terminal 2: Start opencode\n" +
		"opencode\n" +
		"```\n\n" +
		"## Workflow Examples\n\n" +
		"Inside OpenCode, you can:\n\n" +
		"1. Chat with DeepSeek/Qwen directly:\n" +
		"> /tool chat model=deepseek-reasoner messages=[{\"role\":\"user\",\"content\":\"Explain this code\"}]\n\n" +
		"2. Check status before heavy operations:\n" +
		"> /tool get_status\n\n" +
		"3. Monitor pool utilization:\n" +
		"> /tool get_pool_status pool_type=all\n\n" +
		"4. Generate embeddings:\n" +
		"> /tool embeddings input=[\"hello world\",\"hello world\"]\n\n" +
		"## Supported Models Quick Reference\n\n" +
		"| Use Case | Model | Engine |\n" +
		"|----------|-------|--------|\n" +
		"| General coding | deepseek-chat | DeepSeek |\n" +
		"| Complex reasoning | deepseek-reasoner | DeepSeek |\n" +
		"| Code specialized | qwen-coder | Qwen (通义千问) |\n" +
		"| Chinese content | qwen-max | Qwen (通义千问) |\n" +
		"| Fast responses | qwen-flash | Qwen (通义千问) |\n" +
		"| Latest and best | qwen3.5-plus | Qwen (通义千问) |\n"
}

func vscodeGuide(baseURL string) string {
	return "# Visual Studio Code MCP Connection Guide\n\n" +
		"## Prerequisites\n\n" +
		"- VS Code 1.90+ (with built-in MCP support)\n" +
		"- OR install the MCP extension from marketplace\n" +
		"- ds2api running and accessible\n\n" +
		"## Method 1: Built-in MCP Support (VS Code 1.97+)\n\n" +
		"Create/edit .vscode/mcp.json in your project root:\n\n" +
		"```json\n" +
		"{\n" +
		`  "servers": {` + "\n" +
		`    "ds2api": {` + "\n" +
		`      "type": "streamable-http",` + "\n" +
		"      \"url\": \"" + baseURL + `/mcp",` + "\n" +
		`      "headers": {` + "\n" +
		`        "Authorization": "Bearer YOUR_API_KEY"` + "\n" +
		"      }\n" +
		"    }\n" +
		"  }\n" +
		"}\n" +
		"```\n\n" +
		"VS Code will auto-connect when you open the project.\n\n" +
		"## Method 2: Extension-Based (older versions)\n\n" +
		"Install MCP Client extension, then add to settings.json:\n\n" +
		"```json\n" +
		"{\n" +
		`  "mcp.servers": {` + "\n" +
		`    "ds2api": {` + "\n" +
		`      "type": "streamable-http",` + "\n" +
		"      \"url\": \"" + baseURL + `/mcp",` + "\n" +
		`      "headers": {` + "\n" +
		`        "Authorization": "Bearer YOUR_API_KEY"` + "\n" +
		"      }\n" +
		"    }\n" +
		"  }\n" +
		"}\n" +
		"```\n\n" +
		"## Method 3: Claude Code Extension\n\n" +
		"If using Claude Code inside VS Code:\n\n" +
		"```json\n" +
		"{\n" +
		`  "mcpServers": {` + "\n" +
		`    "ds2api": {` + "\n" +
		`      "type": "streamable-http",` + "\n" +
		"      \"url\": \"" + baseURL + `/mcp"` + "\n" +
		"    }\n" +
		"  }\n" +
		"}\n" +
		"```\n\n" +
		"## Verification\n\n" +
		"1. Open Command Palette (Ctrl+Shift+P)\n" +
		`2. Run "MCP: Show Connected Servers"` + "\n" +
		`3. You should see "ds2api" listed with its tools` + "\n\n" +
		"## Using in VS Code\n\n" +
		"### Chat Panel Integration\n\n" +
		"In the built-in AI chat (Ctrl+Alt+I):\n" +
		"- Ask: What models does ds2api provide? -> calls list_models\n" +
		"- Ask: Check if the API is healthy -> calls get_status\n\n" +
		"### Inline Tool Usage\n\n" +
		"Select code, then use commands:\n" +
		"- Ask ds2api about selection -> uses chat tool with selected text\n" +
		"- Explain with Qwen (通义千问) -> uses chat with qwen-max model\n\n" +
		"### Status Bar Integration\n\n" +
		"The MCP extension shows connection status in the status bar:\n" +
		"- Connected: ds2api reachable\n" +
		"- Disconnected: check ds2api is running\n\n" +
		"## Debugging\n\n" +
		"Enable debug mode in config:\n\n" +
		"```json\n" +
		"{\n" +
		`  "servers": {` + "\n" +
		`    "ds2api": {` + "\n" +
		`      "type": "streamable-http",` + "\n" +
		"      \"url\": \"" + baseURL + `/mcp",` + "\n" +
		`      "logLevel": "debug"` + "\n" +
		"    }\n" +
		"  }\n" +
		"}\n" +
		"```\n\n" +
		"View logs: Output Panel -> select MCP channel\n"
}
