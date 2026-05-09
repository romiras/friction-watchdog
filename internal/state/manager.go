package state

import (
	"encoding/json"
	"path/filepath"
	"sync"
	"time"
)

type Manager struct {
	mu            sync.Mutex
	lastActivity  time.Time
	lastFiles     []string
	errorCount    int
	idleThreshold time.Duration
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
	s.errorCount = 0 // Успешная активность сбрасывает счетчик ошибок

	if file != "" {
		s.lastFiles = append([]string{filepath.Base(file)}, s.lastFiles...)
		if len(s.lastFiles) > 3 {
			s.lastFiles = s.lastFiles[:3]
		}
	}
}

func (s *Manager) RecordError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActivity = time.Now() // Ошибка — это тоже активность
	s.errorCount++
}

func (s *Manager) GetReport() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	report := map[string]interface{}{
		"idle_time_min": int(time.Since(s.lastActivity).Minutes()),
		"last_files":    s.lastFiles,
		"error_count":   s.errorCount,
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
