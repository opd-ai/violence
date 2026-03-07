//go:build !js

package config

import (
	"context"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// watcherMu protects the watcher state
var (
	watcherMu       sync.Mutex
	watcherActive   bool
	watcherCtx      context.Context
	watcherCancel   context.CancelFunc
	currentCallback ReloadCallback
)

// Watch starts watching the config file for changes and calls the callback on reload.
// Returns a stop function to cancel watching.
// Only one watcher can be active at a time. Calling Watch when a watcher is active
// will replace the callback but keep the same underlying file watcher (to avoid
// viper race conditions).
func Watch(callback ReloadCallback) (stop func(), err error) {
	watcherMu.Lock()
	defer watcherMu.Unlock()

	if !watcherActive {
		startFileWatcher(callback)
	} else {
		currentCallback = callback
	}

	return createStopWatcherFunc(), nil
}

// startFileWatcher initializes and starts the configuration file watcher.
func startFileWatcher(callback ReloadCallback) {
	ctx, cancel := context.WithCancel(context.Background())
	watcherCtx = ctx
	watcherCancel = cancel
	currentCallback = callback
	watcherActive = true

	viper.WatchConfig()
	viper.OnConfigChange(handleConfigChange)
}

// handleConfigChange processes configuration file change events.
func handleConfigChange(e fsnotify.Event) {
	watcherMu.Lock()
	cb := currentCallback
	ctx := watcherCtx
	watcherMu.Unlock()

	if ctx != nil {
		select {
		case <-ctx.Done():
			return
		default:
		}
	}

	reloadConfiguration(cb)
}

// reloadConfiguration loads the new configuration and invokes the callback.
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

// createStopWatcherFunc returns a function that stops the configuration watcher.
func createStopWatcherFunc() func() {
	return func() {
		watcherMu.Lock()
		defer watcherMu.Unlock()
		if watcherCancel != nil {
			watcherCancel()
			watcherCancel = nil
			watcherCtx = nil
		}
		watcherActive = false
		currentCallback = nil
	}
}
