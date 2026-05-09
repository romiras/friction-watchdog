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
		log.Printf("Ошибка отправки уведомления: %v", err)
	}
}

func StartIdleChecker(stateManager *state.Manager) {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for range ticker.C {
			timeSinceLast, errors, threshold := stateManager.GetStatus()

			// Триггер 1: Простой. Триггер 2: Цикл ошибок (>= 3)
			if timeSinceLast >= threshold || errors >= 3 {
				log.Println("[Watchdog] Обнаружено ТРЕНИЕ. Отправка MCP-уведомления агенту...")

				SendUpdateNotification()

				// Сбрасываем счетчики, чтобы не спамить агента каждую секунду
				stateManager.RecordActivity("")
			}
		}
	}()
}
