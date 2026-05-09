package state

import (
	"encoding/json"
	"path/filepath"
	"sync"
	"time"
)

type Manager struct {
	mu                sync.Mutex
	lastActivity      time.Time
	lastFiles         []string
	errorCount        int
	idleThreshold     time.Duration
	lastFailedCommand string
	triggerReason     string
}

func NewManager(idleThreshold time.Duration) *Manager {
	return &Manager{
		lastActivity:  time.Now(),
		lastFiles:     make([]string, 0),
		idleThreshold: idleThreshold,
	}
}

func (s *Manager) RecordActivity(file string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActivity = time.Now()
	s.errorCount = 0 // Successful activity resets error count

	if file != "" {
		s.lastFiles = append([]string{filepath.Base(file)}, s.lastFiles...)
		if len(s.lastFiles) > 3 {
			s.lastFiles = s.lastFiles[:3]
		}
	}
}

func (s *Manager) RecordError(command string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActivity = time.Now() // Error is also activity
	s.errorCount++
	if command != "" {
		s.lastFailedCommand = command
	}
}

func (s *Manager) SetTriggerReason(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.triggerReason = reason
}

func (s *Manager) GetReport() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	report := map[string]interface{}{
		"idle_time_min":       int(time.Since(s.lastActivity).Minutes()),
		"last_files":          s.lastFiles,
		"error_count":         s.errorCount,
		"last_failed_command": s.lastFailedCommand,
		"trigger_reason":      s.triggerReason,
	}
	bytes, _ := json.Marshal(report)
	return string(bytes)
}

func (s *Manager) GetStatus() (time.Duration, int, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Since(s.lastActivity), s.errorCount, s.idleThreshold
}

func (s *Manager) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActivity = time.Now()
	s.errorCount = 0
}
