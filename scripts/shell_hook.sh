#!/usr/bin/env bash

# ====================================================================
# Friction Watchdog Terminal Hook
# Source this file in your ~/.zshrc or ~/.bashrc
# ====================================================================

# Track the last command typed (populated by preexec/DEBUG before it runs)
_fw_last_cmd=""

mcp_terminal_notify() {
    local EXIT_CODE=$?
    # Send the exit code, command, and timestamp to the Friction Watchdog server
    # Run in the background and discard output to prevent slowing down the prompt
    if command -v curl >/dev/null 2>&1; then
        (curl -s --max-time 0.5 -X POST http://localhost:8080/terminal-event \
            -d "exit_code=${EXIT_CODE}" \
            -d "timestamp=$(date +%s)" \
            --data-urlencode "command=${_fw_last_cmd}" &) >/dev/null 2>&1
    fi
}

# ---------------------------
# Shell Detection and Hooking
# ---------------------------

if [ -n "$ZSH_VERSION" ]; then
    # Zsh: capture command in preexec, report result in precmd
    autoload -Uz add-zsh-hook
    fw_preexec() { _fw_last_cmd="$1"; }
    add-zsh-hook preexec fw_preexec
    add-zsh-hook precmd mcp_terminal_notify
elif [ -n "$BASH_VERSION" ]; then
    # Bash: DEBUG trap fires before each command, PROMPT_COMMAND fires after
    trap 'export _fw_last_cmd="$BASH_COMMAND"' DEBUG
    if [[ ! "$PROMPT_COMMAND" == *mcp_terminal_notify* ]]; then
        if [ -z "$PROMPT_COMMAND" ]; then
            PROMPT_COMMAND="mcp_terminal_notify"
        else
            PROMPT_COMMAND="${PROMPT_COMMAND}; mcp_terminal_notify"
        fi
    fi
fi
