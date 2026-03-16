// Package mod provides mod loading and management.
package mod

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	logrus "github.com/sirupsen/logrus"
)

// Mod represents a loaded game modification.
type Mod struct {
	// Manifest contains the mod's metadata
	Manifest *Manifest `json:"manifest"`

	// Path is the filesystem location of the mod
	Path string `json:"-"`

	// Enabled indicates if the mod is active
	Enabled bool `json:"-"`

	// Legacy fields for backward compatibility (deprecated)
	Name        string            `json:"name,omitempty"`
	Version     string            `json:"version,omitempty"`
	Description string            `json:"description,omitempty"`
	Author      string            `json:"author,omitempty"`
	Config      map[string]string `json:"config,omitempty"`
}

// Loader manages loading and listing of mods.
type Loader struct {
	mods          []Mod
	modsDir       string
	mu            sync.RWMutex
	conflicts     map[string][]string // mod name -> conflicting mods
	warnings      []string
	pluginManager *PluginManager
	wasmLoader    *WASMLoader

	// EnableUnsafePlugins allows loading Go plugins (DEPRECATED).
	// This is unsafe for untrusted mods. Use WASM mods instead.
	EnableUnsafePlugins bool
}

// NewLoader creates a new mod loader.
func NewLoader() *Loader {
	return &Loader{
		mods:                make([]Mod, 0),
		modsDir:             "mods",
		conflicts:           make(map[string][]string),
		warnings:            make([]string, 0),
		pluginManager:       NewPluginManager(),
		wasmLoader:          NewWASMLoader(),
		EnableUnsafePlugins: false,
	}
}

// NewLoaderWithDir creates a mod loader with a custom directory.
func NewLoaderWithDir(dir string) *Loader {
	return &Loader{
		mods:                make([]Mod, 0),
		modsDir:             dir,
		conflicts:           make(map[string][]string),
		warnings:            make([]string, 0),
		pluginManager:       NewPluginManager(),
		wasmLoader:          NewWASMLoader(),
		EnableUnsafePlugins: false,
	}
}

// LoadMod loads a mod from the given path.
// The path should point to a directory containing a mod.json manifest file.
func (l *Loader) LoadMod(path string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.checkModAlreadyLoaded(path); err != nil {
		return err
	}

	mod, err := l.readAndParseManifest(path)
	if err != nil {
		return err
	}

	if err := validateModFields(&mod); err != nil {
		return err
	}

	if err := l.checkModConflicts(&mod); err != nil {
		return err
	}

	l.mods = append(l.mods, mod)
	return nil
}

// checkModAlreadyLoaded verifies that a mod from the given path is not already loaded.
func (l *Loader) checkModAlreadyLoaded(path string) error {
	for _, mod := range l.mods {
		if mod.Path == path {
			return fmt.Errorf("mod already loaded from %s", path)
		}
	}
	return nil
}

// readAndParseManifest reads and parses the mod.json manifest from the given path.
func (l *Loader) readAndParseManifest(path string) (Mod, error) {
	manifest, err := LoadManifestFromDir(path)
	if err != nil {
		return Mod{}, fmt.Errorf("failed to load manifest: %w", err)
	}

	mod := Mod{
		Manifest: manifest,
		Path:     path,
		Enabled:  true,
		// Populate legacy fields for backward compatibility
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
	}

	// Convert manifest config to legacy format if present
	if manifest.Config != nil {
		mod.Config = make(map[string]string)
		for k, v := range manifest.Config {
			mod.Config[k] = fmt.Sprintf("%v", v)
		}
	}

	return mod, nil
}

// validateModFields ensures that required mod fields are present.
func validateModFields(mod *Mod) error {
	// Validation is now handled by Manifest.Validate()
	// This function remains for backward compatibility
	if mod.Manifest != nil {
		return mod.Manifest.Validate()
	}

	// Fallback to legacy validation
	if mod.Name == "" {
		return fmt.Errorf("mod.json missing required field: name")
	}
	if mod.Version == "" {
		return fmt.Errorf("mod.json missing required field: version")
	}
	return nil
}

// checkModConflicts verifies that the mod does not conflict with any enabled mods.
func (l *Loader) checkModConflicts(mod *Mod) error {
	modName := getModName(mod)

	if err := l.checkLegacyConflicts(modName); err != nil {
		return err
	}

	if err := l.checkManifestConflicts(mod, modName); err != nil {
		return err
	}

	return nil
}

// getModName retrieves the effective name of a mod.
func getModName(mod *Mod) string {
	if mod.Manifest != nil {
		return mod.Manifest.Name
	}
	return mod.Name
}

// checkLegacyConflicts verifies against the legacy conflicts registry.
func (l *Loader) checkLegacyConflicts(modName string) error {
	conflicts, ok := l.conflicts[modName]
	if !ok {
		return nil
	}

	for _, conflict := range conflicts {
		if err := l.checkConflictWithEnabled(modName, conflict, ""); err != nil {
			return err
		}
	}
	return nil
}

// checkManifestConflicts verifies conflicts declared in the mod manifest.
func (l *Loader) checkManifestConflicts(mod *Mod, modName string) error {
	if mod.Manifest == nil {
		return nil
	}

	for _, conflict := range mod.Manifest.Conflicts {
		if err := l.checkConflictWithEnabled(modName, conflict, " (declared in manifest)"); err != nil {
			return err
		}
	}
	return nil
}

// checkConflictWithEnabled checks if a conflicting mod is enabled.
func (l *Loader) checkConflictWithEnabled(modName, conflict, suffix string) error {
	for _, existing := range l.mods {
		if !existing.Enabled {
			continue
		}
		existingName := getModName(&existing)
		if existingName == conflict {
			return fmt.Errorf("mod %s conflicts with %s%s", modName, conflict, suffix)
		}
	}
	return nil
}

// UnloadMod unloads a mod by name.
func (l *Loader) UnloadMod(name string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, mod := range l.mods {
		if mod.Name == name {
			l.mods = append(l.mods[:i], l.mods[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("mod not found: %s", name)
}

// ListMods returns all loaded mods.
func (l *Loader) ListMods() []Mod {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]Mod, len(l.mods))
	copy(result, l.mods)
	return result
}

// GetMod returns a mod by name.
func (l *Loader) GetMod(name string) (*Mod, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for i := range l.mods {
		if l.mods[i].Name == name {
			return &l.mods[i], nil
		}
	}
	return nil, fmt.Errorf("mod not found: %s", name)
}

// EnableMod enables a mod by name.
func (l *Loader) EnableMod(name string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i := range l.mods {
		if l.mods[i].Name == name {
			l.mods[i].Enabled = true
			return nil
		}
	}
	return fmt.Errorf("mod not found: %s", name)
}

// DisableMod disables a mod by name.
func (l *Loader) DisableMod(name string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i := range l.mods {
		if l.mods[i].Name == name {
			l.mods[i].Enabled = false
			return nil
		}
	}
	return fmt.Errorf("mod not found: %s", name)
}

// AddConflict registers a conflict between two mods.
func (l *Loader) AddConflict(mod1, mod2 string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.conflicts[mod1] = append(l.conflicts[mod1], mod2)
	l.conflicts[mod2] = append(l.conflicts[mod2], mod1)
}

// SetModsDir sets the directory to search for mods.
func (l *Loader) SetModsDir(dir string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.modsDir = dir
}

// GetModsDir returns the current mods directory.
func (l *Loader) GetModsDir() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.modsDir
}

// PluginManager returns the plugin manager for this loader.
// This provides access to hooks and generators registered by plugins.
//
// Deprecated: Use WASMLoader() for safe mod execution in sandboxed environments.
func (l *Loader) PluginManager() *PluginManager {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.pluginManager
}

// WASMLoader returns the WASM loader for this loader.
// This is the recommended way to load untrusted mods safely.
func (l *Loader) WASMLoader() *WASMLoader {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.wasmLoader
}

// RegisterPlugin registers a plugin with the loader's plugin manager.
// The plugin is loaded immediately and its lifecycle managed by the loader.
// Requires EnableUnsafePlugins flag to be set.
//
// Deprecated: Unsafe for untrusted mods. Use WASM mods via WASMLoader() instead.
func (l *Loader) RegisterPlugin(p Plugin) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.EnableUnsafePlugins {
		return fmt.Errorf("unsafe plugins are disabled; set EnableUnsafePlugins=true or use WASM mods")
	}

	logrus.WithFields(logrus.Fields{
		"system_name": "mod_loader",
		"plugin_name": p.Name(),
	}).Warn("Loading unsafe plugin - this is not recommended for untrusted mods")

	return l.pluginManager.LoadPlugin(p)
}

// UnloadPlugin unloads a plugin by name.
func (l *Loader) UnloadPlugin(name string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.pluginManager.UnloadPlugin(name)
}

// LoadAllMods scans the mods directory and loads all valid mods found.
// Each subdirectory containing a mod.json is treated as a mod.
// Mods are loaded in alphabetical order for deterministic behavior.
// Returns the number of mods successfully loaded and any error from scanning.
func (l *Loader) LoadAllMods() (int, error) {
	l.mu.RLock()
	modsDir := l.modsDir
	l.mu.RUnlock()

	entries, err := os.ReadDir(modsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // No mods directory is not an error
		}
		return 0, fmt.Errorf("failed to read mods directory %s: %w", modsDir, err)
	}

	// Collect directories and sort for deterministic order
	dirs := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)

	loaded := 0
	for _, dir := range dirs {
		modPath := filepath.Join(modsDir, dir)
		manifestPath := filepath.Join(modPath, "mod.json")

		// Skip directories without mod.json
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		if err := l.LoadMod(modPath); err != nil {
			l.mu.Lock()
			warning := fmt.Sprintf("failed to load mod from %s: %s", dir, err.Error())
			l.warnings = append(l.warnings, warning)
			l.mu.Unlock()
			logrus.WithFields(logrus.Fields{
				"system_name": "mod_loader",
				"mod_dir":     dir,
			}).Warn(warning)
			continue
		}
		loaded++
	}

	return loaded, nil
}

// GetWarnings returns all accumulated warning messages from the loader.
func (l *Loader) GetWarnings() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Combine loader warnings with plugin manager warnings
	pmWarnings := l.pluginManager.Warnings()
	all := make([]string, 0, len(l.warnings)+len(pmWarnings))
	all = append(all, l.warnings...)
	all = append(all, pmWarnings...)
	return all
}

// ClearWarnings removes all warnings.
func (l *Loader) ClearWarnings() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.warnings = l.warnings[:0]
	l.pluginManager.ClearWarnings()
}

// RegisterOverride registers a generation parameter override.
func (l *Loader) RegisterOverride(override ParamOverride) {
	l.pluginManager.Overrides().Register(override)
}

// GetOverrides returns the effective parameter overrides for a generator type.
func (l *Loader) GetOverrides(generatorType string) map[string]interface{} {
	return l.pluginManager.Overrides().GetAll(generatorType)
}

// ComputeSHA256 computes the SHA256 checksum of data and returns it as hex string.
func ComputeSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// LoadWASMModule loads a WASM module from raw bytes.
func LoadWASMModule(data []byte) error {
	// Global WASM loader instance for convenience
	// In production, use loader.WASMLoader() for better control
	loader := NewWASMLoader()
	// Generate a unique name based on data hash
	name := ComputeSHA256(data)[:16] // Use first 16 chars of hash as name
	_, err := loader.LoadWASMFromBytes(name, data)
	return err
}
