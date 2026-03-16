//go:build !js

package config

import (
	"context"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// watchState holds the single active file watcher and associated state.
var watchState struct {
	mu       sync.Mutex
	active   bool
	cancel   context.CancelFunc
	callback ReloadCallback
	watcher  *fsnotify.Watcher
	done     chan struct{} // closed when the watch goroutine exits
}

// Watch starts watching the config file for changes and calls the callback on
// reload. Returns a stop function to cancel watching. Only one underlying
// fsnotify watcher is kept at a time; calling Watch while a watcher is active
// replaces the callback but reuses the existing watcher.
func Watch(callback ReloadCallback) (stop func(), err error) {
	watchState.mu.Lock()
	defer watchState.mu.Unlock()

	if !watchState.active {
		if err := startWatcher(callback); err != nil {
			return func() {}, err
		}
	} else {
		watchState.callback = callback
	}

	return makeStopFunc(), nil
}

// startWatcher creates a new fsnotify watcher and begins the watch loop.
// Must be called with watchState.mu held.
func startWatcher(callback ReloadCallback) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	configFile := viper.ConfigFileUsed()
	if configFile != "" {
		if err := w.Add(configFile); err != nil {
			w.Close()
			return err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	watchState.watcher = w
	watchState.cancel = cancel
	watchState.callback = callback
	watchState.active = true
	watchState.done = done

	go func() {
		defer close(done)
		watchLoop(ctx, w, configFile)
	}()
	return nil
}

// makeStopFunc returns a stop function that tears down the active watcher and
// blocks until the goroutine has fully exited.
// Must be called with watchState.mu held (captures state at call time).
func makeStopFunc() func() {
	cancel := watchState.cancel
	w := watchState.watcher
	done := watchState.done
	return func() {
		if cancel != nil {
			cancel()
		}
		if w != nil {
			w.Close()
		}
		if done != nil {
			<-done // wait for the goroutine to finish
		}
		watchState.mu.Lock()
		watchState.active = false
		watchState.callback = nil
		watchState.cancel = nil
		watchState.watcher = nil
		watchState.done = nil
		watchState.mu.Unlock()
	}
}

// watchLoop processes fsnotify events until ctx is cancelled or the watcher is closed.
func watchLoop(ctx context.Context, w *fsnotify.Watcher, configFile string) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				applyReload(ctx, configFile, w)
			}
		case _, ok := <-w.Errors:
			if !ok {
				return
			}
		}
	}
}

// applyReload re-reads viper config and invokes the current callback.
func applyReload(ctx context.Context, configFile string, w *fsnotify.Watcher) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	if err := viper.ReadInConfig(); err != nil {
		return
	}

	// Re-add the file in case it was recreated by an editor rename-swap.
	if configFile != "" {
		_ = w.Add(configFile)
	}

	watchState.mu.Lock()
	cb := watchState.callback
	watchState.mu.Unlock()

	reloadConfiguration(cb)
}

// reloadConfiguration unmarshals viper state into C and invokes cb.
func reloadConfiguration(cb ReloadCallback) {
	mu.Lock()
	old := C
	var newCfg Config
	if err := viper.Unmarshal(&newCfg); err == nil {
		C = newCfg
		mu.Unlock()
		if cb != nil {
			cb(old, newCfg)
		}
	} else {
		mu.Unlock()
	}
}
