package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"friction-watchdog/internal/mcp"
	"friction-watchdog/internal/state"
	"friction-watchdog/internal/terminal"
	"friction-watchdog/internal/watcher"
)

func main() {
	// IMPORTANT: Protect stdout from logs. All logs go to stderr.
	log.SetOutput(os.Stderr)
	log.Println("Starting Friction Watchdog MCP Server...")

	// Set 15 minutes for testing.
	idleThreshold := 15 * time.Minute
	if val := os.Getenv("WATCHDOG_THRESHOLD_MIN"); val != "" {
		if dur, err := time.ParseDuration(val + "m"); err == nil {
			idleThreshold = dur
		}
	}

	stateManager := state.NewManager(idleThreshold)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	watcher.StartFSWatcher(ctx, ".", stateManager)
	terminal.StartReceiver(stateManager)
	mcp.StartIdleChecker(stateManager)

	// Blocking loop for MCP processing (via stdin)
	mcp.StartHandler(ctx, stateManager)

	log.Println("Friction Watchdog MCP Server stopped.")
}
