//go:build windows

// Package mod provides WASM-based mod runtime with sandboxing.
// This file is the Windows stub: wasmer-go depends on open_memstream, a POSIX
// function that is not available on Windows. Mod loading is therefore
// unsupported on the Windows target; all methods return ErrNotSupported.
package mod

import "errors"

// ErrNotSupported is returned by all WASMLoader methods on the Windows target.
var ErrNotSupported = errors.New("WASM mod loading is not supported on this platform")

// WASMConfig defines security constraints for WASM module execution.
type WASMConfig struct {
	MemoryLimitBytes uint32
	FuelLimit        uint64
	AllowFileRead    bool
	AllowFileWrite   bool
	AllowedPaths     []string
}

// DefaultWASMConfig returns the default (secure) configuration.
func DefaultWASMConfig() WASMConfig {
	return WASMConfig{
		MemoryLimitBytes: 64 * 1024 * 1024,
		FuelLimit:        1_000_000_000,
		AllowFileRead:    true,
		AllowFileWrite:   false,
		AllowedPaths:     []string{"mods/"},
	}
}

// WASMLoader is a no-op loader for the Windows target.
type WASMLoader struct {
	config  WASMConfig
	modules map[string]*WASMModule
}

// NewWASMLoader creates a WASMLoader with default configuration.
func NewWASMLoader() *WASMLoader {
	return &WASMLoader{config: DefaultWASMConfig(), modules: make(map[string]*WASMModule)}
}

// NewWASMLoaderWithConfig creates a WASMLoader with the given configuration.
func NewWASMLoaderWithConfig(config WASMConfig) *WASMLoader {
	return &WASMLoader{config: config, modules: make(map[string]*WASMModule)}
}

// WASMModule represents a loaded WASM module (stub on Windows target).
type WASMModule struct {
	Name string
	Path string
}

// LoadWASM always returns ErrNotSupported on the Windows target.
func (wl *WASMLoader) LoadWASM(_ string) (*WASMModule, error) {
	return nil, ErrNotSupported
}

// LoadWASMFromBytes always returns ErrNotSupported on the Windows target.
func (wl *WASMLoader) LoadWASMFromBytes(_ string, _ []byte) (*WASMModule, error) {
	return nil, ErrNotSupported
}

// UnloadWASM always returns ErrNotSupported on the Windows target.
func (wl *WASMLoader) UnloadWASM(_ string) error {
	return ErrNotSupported
}

// GetModule always returns ErrNotSupported on the Windows target.
func (wl *WASMLoader) GetModule(_ string) (*WASMModule, error) {
	return nil, ErrNotSupported
}

// ListModules always returns an empty slice on the Windows target (no modules can be loaded).
func (wl *WASMLoader) ListModules() []string {
	return []string{}
}

// isPathAllowed always returns false on the Windows target.
func (wl *WASMLoader) isPathAllowed(_ string) bool {
	return false
}

// Call always returns ErrNotSupported on the Windows target.
func (wm *WASMModule) Call(_ string, _ ...interface{}) (interface{}, error) {
	return nil, ErrNotSupported
}

// HasExport always returns false on the Windows target.
func (wm *WASMModule) HasExport(_ string) bool {
	return false
}
