package terminal

import (
	"log"
	"net/http"

	"friction-watchdog/internal/state"
)

func StartReceiver(stateManager *state.Manager) {
	http.HandleFunc("/terminal-event", func(w http.ResponseWriter, r *http.Request) {
		exitCode := r.FormValue("exit_code")
		command := r.FormValue("command")
		if exitCode != "0" && exitCode != "" {
			log.Printf("[Terminal Sensor] Command execution error (code: %s, cmd: %s)", exitCode, command)
			stateManager.RecordError(command)
		} else {
			log.Printf("[Terminal Sensor] Successful command")
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
