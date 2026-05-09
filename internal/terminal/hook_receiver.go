package terminal

import (
	"log"
	"net/http"

	"friction-watchdog/internal/state"
)

func StartReceiver(stateManager *state.Manager) {
	http.HandleFunc("/terminal-event", func(w http.ResponseWriter, r *http.Request) {
		exitCode := r.FormValue("exit_code")
		if exitCode != "0" && exitCode != "" {
			log.Printf("[Terminal Sensor] Ошибка выполнения команды (code: %s)", exitCode)
			stateManager.RecordError()
		} else {
			log.Printf("[Terminal Sensor] Успешная команда")
			stateManager.RecordActivity("")
		}
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		log.Println("[Terminal Sensor] Listening on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Printf("Terminal receiver error: %v", err)
		}
	}()
}
