// Package mod provides WASM-based mod runtime with sandboxing.
package mod

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/wasmerio/wasmer-go/wasmer"
)

// WASMConfig defines security constraints for WASM module execution.
type WASMConfig struct {
	// MemoryLimitBytes sets maximum memory per module (default 64MB).
	MemoryLimitBytes uint32

	// FuelLimit sets maximum instructions per call (default 1 billion).
	FuelLimit uint64

	// AllowFileRead enables file read access within mods/ directory.
	AllowFileRead bool

	// AllowFileWrite enables file write access within mods/ directory.
	AllowFileWrite bool

	// AllowedPaths restricts file access to specific directories.
	AllowedPaths []string
}

// DefaultWASMConfig returns secure default configuration.
func DefaultWASMConfig() WASMConfig {
	return WASMConfig{
		MemoryLimitBytes: 64 * 1024 * 1024, // 64MB
		FuelLimit:        1_000_000_000,    // 1 billion instructions
		AllowFileRead:    true,
		AllowFileWrite:   false,
		AllowedPaths:     []string{"mods/"},
	}
}

// WASMLoader manages WASM module loading and execution.
type WASMLoader struct {
	config  WASMConfig
	modules map[string]*WASMModule
	mu      sync.RWMutex
}

// NewWASMLoader creates a WASM loader with default configuration.
func NewWASMLoader() *WASMLoader {
	return &WASMLoader{
		config:  DefaultWASMConfig(),
		modules: make(map[string]*WASMModule),
	}
}

// NewWASMLoaderWithConfig creates a WASM loader with custom configuration.
func NewWASMLoaderWithConfig(config WASMConfig) *WASMLoader {
	return &WASMLoader{
		config:  config,
		modules: make(map[string]*WASMModule),
	}
}

// WASMModule represents a loaded WASM module instance.
type WASMModule struct {
	Name     string
	Path     string
	instance *wasmer.Instance
	store    *wasmer.Store
	memory   *wasmer.Memory
}

// LoadWASM loads a WASM module from the given path.
// The path must point to a .wasm file.
func (wl *WASMLoader) LoadWASM(path string) (*WASMModule, error) {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	// Validate path is within allowed directories
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	if !wl.isPathAllowed(absPath) {
		return nil, fmt.Errorf("access denied: path outside allowed directories")
	}

	// Check if already loaded
	modName := filepath.Base(path)
	if _, exists := wl.modules[modName]; exists {
		return nil, fmt.Errorf("module %s already loaded", modName)
	}

	// Read WASM binary
	wasmBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read WASM file: %w", err)
	}

	// Create engine and store
	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	// Compile module
	module, err := wasmer.NewModule(store, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}

	// Create import object with host functions
	importObject := wl.createImportObject(store, modName)

	// Instantiate module
	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
	}

	// Get memory export (if present)
	var memory *wasmer.Memory
	if memExport, err := instance.Exports.GetMemory("memory"); err == nil {
		memory = memExport
	}

	wasmMod := &WASMModule{
		Name:     modName,
		Path:     path,
		instance: instance,
		store:    store,
		memory:   memory,
	}

	wl.modules[modName] = wasmMod

	logrus.WithFields(logrus.Fields{
		"system_name": "wasm_loader",
		"mod_name":    modName,
		"path":        path,
	}).Info("WASM module loaded successfully")

	return wasmMod, nil
}

// UnloadWASM unloads a WASM module by name.
func (wl *WASMLoader) UnloadWASM(name string) error {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	if _, exists := wl.modules[name]; !exists {
		return fmt.Errorf("module %s not loaded", name)
	}

	delete(wl.modules, name)

	logrus.WithFields(logrus.Fields{
		"system_name": "wasm_loader",
		"mod_name":    name,
	}).Info("WASM module unloaded")

	return nil
}

// GetModule retrieves a loaded WASM module by name.
func (wl *WASMLoader) GetModule(name string) (*WASMModule, error) {
	wl.mu.RLock()
	defer wl.mu.RUnlock()

	mod, exists := wl.modules[name]
	if !exists {
		return nil, fmt.Errorf("module %s not found", name)
	}

	return mod, nil
}

// ListModules returns all loaded WASM module names.
func (wl *WASMLoader) ListModules() []string {
	wl.mu.RLock()
	defer wl.mu.RUnlock()

	names := make([]string, 0, len(wl.modules))
	for name := range wl.modules {
		names = append(names, name)
	}
	return names
}

// isPathAllowed checks if a path is within allowed directories.
func (wl *WASMLoader) isPathAllowed(path string) bool {
	for _, allowed := range wl.config.AllowedPaths {
		absAllowed, err := filepath.Abs(allowed)
		if err != nil {
			continue
		}

		// Check if path is within allowed directory
		rel, err := filepath.Rel(absAllowed, path)
		if err == nil && !filepath.IsAbs(rel) && len(rel) > 0 && rel[0] != '.' {
			return true
		}
	}
	return false
}

// createImportObject creates the import object with host functions.
// This defines the API that WASM modules can call.
func (wl *WASMLoader) createImportObject(store *wasmer.Store, modName string) *wasmer.ImportObject {
	importObject := wasmer.NewImportObject()

	// Register host functions under "env" namespace
	env := make(map[string]wasmer.IntoExtern)

	// log_message(ptr, len) - Log a message from WASM
	env["log_message"] = wasmer.NewFunction(
		store,
		wasmer.NewFunctionType(
			wasmer.NewValueTypes(wasmer.I32, wasmer.I32),
			wasmer.NewValueTypes(),
		),
		func(args []wasmer.Value) ([]wasmer.Value, error) {
			// For now, this is a stub - full implementation requires memory access
			logrus.WithFields(logrus.Fields{
				"system_name": "wasm_mod",
				"mod_name":    modName,
			}).Debug("WASM log called")
			return []wasmer.Value{}, nil
		},
	)

	importObject.Register("env", env)
	return importObject
}

// Call invokes an exported function in a WASM module.
func (wm *WASMModule) Call(funcName string, args ...interface{}) (interface{}, error) {
	fn, err := wm.instance.Exports.GetFunction(funcName)
	if err != nil {
		return nil, fmt.Errorf("function %s not found: %w", funcName, err)
	}

	// Call function directly - NativeFunction is already a Go function
	result, err := fn(args...)
	if err != nil {
		return nil, fmt.Errorf("WASM function call failed: %w", err)
	}

	return result, nil
}

// HasExport checks if a WASM module exports a given function.
func (wm *WASMModule) HasExport(name string) bool {
	_, err := wm.instance.Exports.GetFunction(name)
	return err == nil
}
