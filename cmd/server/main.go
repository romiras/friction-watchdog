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
	// ВАЖНО: Защищаем stdout от логов. Все логи идут в stderr.
	log.SetOutput(os.Stderr)
	log.Println("Starting Friction Watchdog MCP Server...")

	// Для тестирования ставим 15 минут.
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

	// Блокирующий цикл обработки MCP (через stdin)
	mcp.StartHandler(ctx, stateManager)

	log.Println("Friction Watchdog MCP Server stopped.")
}
