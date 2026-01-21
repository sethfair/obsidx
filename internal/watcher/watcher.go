package watcher

import (
	"context"
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
	// Add root directory
	if err := fw.watcher.Add(rootDir); err != nil {
		return err
	}

	// Add all subdirectories
	// TODO: walk subdirectories and add them

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
			// Log error but continue watching
			_ = err
		}
	}
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
