package watcher

import (
	"bufio"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"friction-watchdog/internal/state"

	"github.com/fsnotify/fsnotify"
)

var (
	debounceMap = make(map[string]time.Time)
	debounceMu  sync.Mutex
	debounceDur = 500 * time.Millisecond
)

func StartFSWatcher(ctx context.Context, dir string, stateManager *state.Manager) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("FS Watcher error: %v", err)
	}

	ignoreList := loadGitignore(dir)

	go func() {
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
				log.Println("[FS Watcher] Stopping...")
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Handle new directory creation to watch it recursively
				if event.Op&fsnotify.Create == fsnotify.Create {
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() && !isIgnored(event.Name, ignoreList) {
						watcher.Add(event.Name)
					}
				}

				if event.Has(fsnotify.Write) && !isIgnored(event.Name, ignoreList) {
					if shouldDebounce(event.Name) {
						continue
					}
					log.Printf("[FS Sensor] Activity: %s", filepath.Base(event.Name))
					stateManager.RecordActivity(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				if err != nil {
					log.Printf("Watcher error: %v", err)
				}
			}
		}
	}()

	// Add the current directory and all subdirectories, respecting ignores
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path != dir && isIgnored(path, ignoreList) {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking directory to add watcher: %v", err)
	}
}

func shouldDebounce(path string) bool {
	debounceMu.Lock()
	defer debounceMu.Unlock()
	lastTrigger, exists := debounceMap[path]
	if exists && time.Since(lastTrigger) < debounceDur {
		return true
	}
	debounceMap[path] = time.Now()
	return false
}

func loadGitignore(dir string) []string {
	var ignoreList []string
	// Basic hardcoded ignores
	ignoreList = append(ignoreList, ".git", ".idea", ".vscode", "node_modules", "vendor")

	file, err := os.Open(filepath.Join(dir, ".gitignore"))
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				line = strings.TrimSuffix(line, "/")
				ignoreList = append(ignoreList, line)
			}
		}
	}
	return ignoreList
}

func isIgnored(path string, ignoreList []string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, part := range parts {
		for _, ignore := range ignoreList {
			if part == ignore {
				return true
			}
		}
	}
	return false
}
