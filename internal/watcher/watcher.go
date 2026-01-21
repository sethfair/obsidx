package watcher

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher watches a directory for markdown file changes
type FileWatcher struct {
	watcher  *fsnotify.Watcher
	onChange func(path string)
	debounce time.Duration
}

// New creates a new file watcher
func New(onChange func(path string), debounce time.Duration) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		watcher:  watcher,
		onChange: onChange,
		debounce: debounce,
	}, nil
}

// Watch starts watching a directory recursively
func (fw *FileWatcher) Watch(ctx context.Context, rootDir string) error {
	// Recursively add all directories
	if err := fw.addRecursive(rootDir); err != nil {
		return err
	}

	// Debounce map: path -> timer
	pending := make(map[string]*time.Timer)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event, ok := <-fw.watcher.Events:
			if !ok {
				return nil
			}

			// Handle directory creation (need to add to watcher)
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					if err := fw.addRecursive(event.Name); err != nil {
						log.Printf("Error adding new directory to watch: %v\n", err)
					}
				}
			}

			// Only care about markdown files
			if !strings.HasSuffix(event.Name, ".md") {
				continue
			}

			// Only care about write and create events
			if event.Op&fsnotify.Write != fsnotify.Write &&
				event.Op&fsnotify.Create != fsnotify.Create {
				continue
			}

			path := event.Name

			// Cancel existing timer if any
			if timer, exists := pending[path]; exists {
				timer.Stop()
			}

			// Set new debounce timer
			pending[path] = time.AfterFunc(fw.debounce, func() {
				delete(pending, path)
				fw.onChange(path)
			})

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watcher error: %v\n", err)
		}
	}
}

// addRecursive recursively adds a directory and all subdirectories to the watcher
func (fw *FileWatcher) addRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories (starting with .)
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != root {
			return filepath.SkipDir
		}

		if info.IsDir() {
			if err := fw.watcher.Add(path); err != nil {
				return err
			}
		}
		return nil
	})
}

// Close closes the watcher
func (fw *FileWatcher) Close() error {
	return fw.watcher.Close()
}

// IsMarkdownFile checks if a file is markdown
func IsMarkdownFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".md" || ext == ".markdown"
}
