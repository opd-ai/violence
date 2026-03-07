//go:build js

package config

import "errors"

// Watch is a no-op on WASM since filesystem watching is not supported in browsers.
// viper.WatchConfig() calls fsnotify.NewWatcher() which fatally exits on WASM.
func Watch(callback ReloadCallback) (stop func(), err error) {
	return nil, errors.New("config watching not supported on WASM")
}
