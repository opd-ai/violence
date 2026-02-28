// Package config handles loading and storing game configuration.
package config

import (
	"context"
	"errors"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config holds all game configuration values.
type Config struct {
	WindowWidth      int            `mapstructure:"WindowWidth"`
	WindowHeight     int            `mapstructure:"WindowHeight"`
	InternalWidth    int            `mapstructure:"InternalWidth"`
	InternalHeight   int            `mapstructure:"InternalHeight"`
	FOV              float64        `mapstructure:"FOV"`
	MouseSensitivity float64        `mapstructure:"MouseSensitivity"`
	MasterVolume     float64        `mapstructure:"MasterVolume"`
	MusicVolume      float64        `mapstructure:"MusicVolume"`
	SFXVolume        float64        `mapstructure:"SFXVolume"`
	DefaultGenre     string         `mapstructure:"DefaultGenre"`
	VSync            bool           `mapstructure:"VSync"`
	FullScreen       bool           `mapstructure:"FullScreen"`
	MaxTPS           int            `mapstructure:"MaxTPS"` // Maximum ticks per second (0 = unlimited)
	KeyBindings      map[string]int `mapstructure:"KeyBindings"`
	ProfanityFilter  bool           `mapstructure:"ProfanityFilter"` // Client-side profanity filter toggle
}

// C is the global configuration instance.
var C Config

// mu protects concurrent access to C during hot-reload.
var mu sync.RWMutex

// watcherMu protects the watcher state
var (
	watcherMu       sync.Mutex
	watcherActive   bool
	watcherCtx      context.Context
	watcherCancel   context.CancelFunc
	currentCallback ReloadCallback
)

// ReloadCallback is called when the configuration is hot-reloaded.
type ReloadCallback func(old, new Config)

// Load reads configuration from file and environment, populating C.
func Load() error {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.violence")

	viper.SetDefault("WindowWidth", 960)
	viper.SetDefault("WindowHeight", 600)
	viper.SetDefault("InternalWidth", 320)
	viper.SetDefault("InternalHeight", 200)
	viper.SetDefault("FOV", 66.0)
	viper.SetDefault("MouseSensitivity", 1.0)
	viper.SetDefault("MasterVolume", 0.8)
	viper.SetDefault("MusicVolume", 0.7)
	viper.SetDefault("SFXVolume", 0.8)
	viper.SetDefault("DefaultGenre", "fantasy")
	viper.SetDefault("VSync", true)
	viper.SetDefault("FullScreen", false)
	viper.SetDefault("MaxTPS", 60)
	viper.SetDefault("KeyBindings", map[string]int{})
	viper.SetDefault("ProfanityFilter", true)

	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return err
		}
	}

	return viper.Unmarshal(&C)
}

// Save writes the current configuration to file.
func Save() error {
	mu.RLock()
	defer mu.RUnlock()

	viper.Set("WindowWidth", C.WindowWidth)
	viper.Set("WindowHeight", C.WindowHeight)
	viper.Set("InternalWidth", C.InternalWidth)
	viper.Set("InternalHeight", C.InternalHeight)
	viper.Set("FOV", C.FOV)
	viper.Set("MouseSensitivity", C.MouseSensitivity)
	viper.Set("MasterVolume", C.MasterVolume)
	viper.Set("MusicVolume", C.MusicVolume)
	viper.Set("SFXVolume", C.SFXVolume)
	viper.Set("DefaultGenre", C.DefaultGenre)
	viper.Set("VSync", C.VSync)
	viper.Set("FullScreen", C.FullScreen)
	viper.Set("MaxTPS", C.MaxTPS)
	viper.Set("KeyBindings", C.KeyBindings)
	viper.Set("ProfanityFilter", C.ProfanityFilter)

	return viper.WriteConfig()
}

// Watch starts watching the config file for changes and calls the callback on reload.
// Returns a stop function to cancel watching.
// Only one watcher can be active at a time. Calling Watch when a watcher is active
// will replace the callback but keep the same underlying file watcher (to avoid
// viper race conditions).
func Watch(callback ReloadCallback) (stop func(), err error) {
	watcherMu.Lock()
	defer watcherMu.Unlock()

	// If no watcher is active, start one
	if !watcherActive {
		ctx, cancel := context.WithCancel(context.Background())
		watcherCtx = ctx
		watcherCancel = cancel
		currentCallback = callback
		watcherActive = true

		// Start viper's file watcher (only once)
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			watcherMu.Lock()
			cb := currentCallback
			ctx := watcherCtx
			watcherMu.Unlock()

			// Check if watcher has been stopped
			if ctx != nil {
				select {
				case <-ctx.Done():
					return
				default:
				}
			}

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
		})
	} else {
		// Watcher already active, just replace the callback
		currentCallback = callback
	}

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
	}, nil
}

// Get returns a copy of the current config safely.
func Get() Config {
	mu.RLock()
	defer mu.RUnlock()
	return C
}

// Set updates the config safely.
func Set(cfg Config) {
	mu.Lock()
	C = cfg
	mu.Unlock()
}
