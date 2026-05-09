#!/usr/bin/env bash

# ====================================================================
# Friction Watchdog Terminal Hook
# Source this file in your ~/.zshrc or ~/.bashrc
# ====================================================================

mcp_terminal_notify() {
    local EXIT_CODE=$?
    # Send the exit code and timestamp to the Friction Watchdog server
    # Run in the background and discard output to prevent slowing down the prompt
    if command -v curl >/dev/null 2>&1; then
        (curl -s --max-time 0.5 -X POST http://localhost:8080/terminal-event \
            -d "exit_code=${EXIT_CODE}" \
            -d "timestamp=$(date +%s)" &) > /dev/null 2>&1
    fi
}

# ---------------------------
# Shell Detection and Hooking
# ---------------------------

if [ -n "$ZSH_VERSION" ]; then
    # Zsh: add to precmd hooks array
    autoload -Uz add-zsh-hook
    add-zsh-hook precmd mcp_terminal_notify
elif [ -n "$BASH_VERSION" ]; then
    # Bash: add to PROMPT_COMMAND array
    # If PROMPT_COMMAND is just a string, append to it.
    # Otherwise, bash 5.1+ supports an array. For compatibility we use string append:
    if [[ ! "$PROMPT_COMMAND" == *mcp_terminal_notify* ]]; then
        if [ -z "$PROMPT_COMMAND" ]; then
            PROMPT_COMMAND="mcp_terminal_notify"
        else
            PROMPT_COMMAND="${PROMPT_COMMAND}; mcp_terminal_notify"
        fi
    fi
fi
