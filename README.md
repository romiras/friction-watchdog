# Friction Watchdog MCP Server

**Friction Watchdog** is an automated "friction detector" that monitors your development workflow and alerts your AI agent (via the Model Context Protocol) when you are stuck. It helps prevent "task-locking" and cognitive burnout by triggering a coaching intervention when it detects long periods of inactivity or repetitive command failures.

## 🚀 How It Works

The system operates as an **Active State Observer** with three main layers:

1.  **Sensors (The Eyes):**
    *   **FS Monitor:** Watches your project directory for file writes (using `fsnotify`). It intelligently ignores noise via `.gitignore` parsing and debounces rapid events.
    *   **Terminal Hook:** A lightweight shell integration that reports the exit codes of your commands to the watchdog.
2.  **Logic (The Brain):**
    *   The watchdog maintains a state machine. It tracks the time since your last successful action and the count of consecutive terminal errors.
    *   If it detects "friction" (e.g., 15 minutes of silence or 3+ failed test runs), it sends a notification to your AI agent.
3.  **Intervention (The Coaching):**
    *   Your AI agent (Gemini CLI, Claude Code, etc.) receives a `notifications/resources/updated` signal.
    *   The agent then reads the friction report and switches from "Coder" mode to "Coaching" mode, asking you diagnostic questions to help you decompress, delegate, or pivot.

## 🧰 MCP Capabilities

*   **Resources:**
    *   `productivity://friction-report`: A JSON report containing idle time, recent files, and error count. (Push notifications via `notifications/resources/updated`).
*   **Tools:**
    *   `reset_timer`: Manually resets the inactivity timer and error count.
*   **Prompts:**
    *   `coaching_mode`: Returns instructions for the AI to act as a productivity coach based on the current friction report.

## 🛠 Installation

### 1. Build the Server
Ensure you have Go installed (1.26+ recommended).

```bash
go mod tidy
go build -o friction-watchdog ./cmd/server
```

### 2. Shell Integration
To enable the Terminal Sensor, source the provided hook script in your shell configuration.

**For Zsh:**
```bash
echo "source $(pwd)/scripts/shell_hook.sh" >> ~/.zshrc
source ~/.zshrc
```

**For Bash:**
```bash
echo "source $(pwd)/scripts/shell_hook.sh" >> ~/.bashrc
source ~/.bashrc
```

## 🔍 Debugging & Verification

You can verify the server using the **MCP Inspector**:

```bash
# From the project root
npx @modelcontextprotocol/inspector ./friction-watchdog
```

Once the Inspector is running:
1. Click **List Resources** to see `productivity://friction-report`.
2. Click **Read Resource** to see the current JSON state.
3. Observe **Notifications** in the console when you save a file or run a command.

## 🤖 MCP Configuration

Add the Friction Watchdog to your MCP client configuration (e.g., `mcp_config.json`).

```json
{
  "mcpServers": {
    "friction-watchdog": {
      "command": "/path/to/friction-watchdog/friction-watchdog",
      "env": {
        "WATCHDOG_THRESHOLD_MIN": "15"
      }
    }
  }
}
```

### System Instruction for your Agent
To make the "coaching" effective, add these instructions to your agent's System Prompt:

> You are connected to the `friction-watchdog` MCP server.
> 1. If you receive a `notifications/resources/updated` for `productivity://friction-report`, immediately call `resources/read`.
> 2. If the report indicates a high `idle_time_min` or `error_count`, stop your current task.
> 3. Switch to "Coaching Mode": Ask the user if they are stuck, suggest decomposing the task, or offer to perform a deep research task to unblock them.

## ⚙️ Configuration

| Environment Variable | Default | Description |
| :--- | :--- | :--- |
| `WATCHDOG_THRESHOLD_MIN` | `15` | Minutes of inactivity before triggering an alert. |
| (Hardcoded) Port | `8080` | The local port the Terminal Sensor sends data to. |

## 📂 Project Structure

- `cmd/server/`: Main entry point for the MCP server.
- `internal/mcp/`: MCP protocol handlers and JSON-RPC types.
- `internal/state/`: Central state manager for tracking productivity metrics.
- `internal/terminal/`: HTTP receiver for terminal exit code hooks.
- `internal/watcher/`: File system monitor with `.gitignore` and debounce support.
- `scripts/`: Shell integration scripts.

## 📄 License
MIT
