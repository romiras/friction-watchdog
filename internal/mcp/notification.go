package mcp

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"friction-watchdog/internal/state"
)

func SendUpdateNotification() {
	msg := NotificationRPC{
		JSONRPC: "2.0",
		Method:  "notifications/resources/updated",
	}
	msg.Params.URI = "productivity://friction-report"

	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(msg); err != nil {
		log.Printf("Error sending notification: %v", err)
	}
}

func StartIdleChecker(stateManager *state.Manager) {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for range ticker.C {
			timeSinceLast, errors, threshold := stateManager.GetStatus()

			// Trigger 1: Idle. Trigger 2: Error loop (>= 3)
			if timeSinceLast >= threshold || errors >= 3 {
				log.Println("[Watchdog] FRICTION detected. Sending MCP notification to agent...")

				var reason string
				if errors >= 3 {
					reason = "error_loop"
				} else {
					reason = "idle_timeout"
				}
				stateManager.SetTriggerReason(reason)

				SendUpdateNotification()

				// Reset counters to avoid spamming the agent every second
				stateManager.RecordActivity("")
			}
		}
	}()
}
