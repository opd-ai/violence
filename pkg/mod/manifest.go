// Package mod provides mod loading and management.
package mod

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Manifest represents a mod's metadata and configuration.
// This is typically stored in a mod.json file at the root of the mod directory.
type Manifest struct {
	// Name is the unique identifier for the mod (lowercase, alphanumeric + hyphens)
	Name string `json:"name"`

	// Version is the semver-compatible version string (e.g., "1.0.0")
	Version string `json:"version"`

	// Description is a short human-readable description of the mod
	Description string `json:"description"`

	// Author is the creator's name or identifier
	Author string `json:"author"`

	// License specifies the mod's license (e.g., "MIT", "GPL-3.0", "Proprietary")
	License string `json:"license,omitempty"`

	// Homepage is the URL to the mod's website or repository
	Homepage string `json:"homepage,omitempty"`

	// Tags are keywords for discovery and categorization
	Tags []string `json:"tags,omitempty"`

	// GenreOverrides specify which genres this mod targets or modifies
	// Empty means compatible with all genres
	GenreOverrides []string `json:"genre_overrides,omitempty"`

	// Dependencies lists other mods this mod requires
	Dependencies []Dependency `json:"dependencies,omitempty"`

	// Conflicts lists mods incompatible with this mod
	Conflicts []string `json:"conflicts,omitempty"`

	// MinGameVersion is the minimum game version required (semver)
	MinGameVersion string `json:"min_game_version,omitempty"`

	// MaxGameVersion is the maximum compatible game version (semver)
	MaxGameVersion string `json:"max_game_version,omitempty"`

	// EntryPoint is the path to the mod's WASM binary (relative to mod directory)
	EntryPoint string `json:"entry_point,omitempty"`

	// Permissions specify what capabilities the mod requires
	Permissions PermissionSet `json:"permissions,omitempty"`

	// Config contains mod-specific configuration values
	Config map[string]interface{} `json:"config,omitempty"`
}

// Dependency represents a required mod and its version constraint.
type Dependency struct {
	// Name is the mod identifier
	Name string `json:"name"`

	// Version is the version constraint (semver, e.g., "^1.0.0", ">=2.0.0")
	Version string `json:"version"`

	// Optional indicates this dependency is recommended but not required
	Optional bool `json:"optional,omitempty"`
}

// PermissionSet defines requested mod capabilities.
type PermissionSet struct {
	FileRead    bool `json:"file_read,omitempty"`
	FileWrite   bool `json:"file_write,omitempty"`
	EntitySpawn bool `json:"entity_spawn,omitempty"`
	AssetLoad   bool `json:"asset_load,omitempty"`
	UIModify    bool `json:"ui_modify,omitempty"`
	Network     bool `json:"network,omitempty"`
}

// ToModPermissions converts PermissionSet to ModPermissions (API type).
func (p PermissionSet) ToModPermissions() ModPermissions {
	return ModPermissions{
		AllowFileRead:    p.FileRead,
		AllowFileWrite:   p.FileWrite,
		AllowEntitySpawn: p.EntitySpawn,
		AllowAssetLoad:   p.AssetLoad,
		AllowUIModify:    p.UIModify,
	}
}

var (
	// validNamePattern enforces lowercase alphanumeric + hyphens for mod names
	validNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

	// semverPattern validates semantic version strings
	semverPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9.-]+))?(?:\+([a-zA-Z0-9.-]+))?$`)
)

// Validate checks that the manifest has valid fields.
// Returns error if any required field is missing or invalid.
func (m *Manifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !validNamePattern.MatchString(m.Name) {
		return fmt.Errorf("name must be lowercase alphanumeric + hyphens (got %q)", m.Name)
	}
	if len(m.Name) > 64 {
		return fmt.Errorf("name too long (max 64 characters)")
	}

	if m.Version == "" {
		return fmt.Errorf("version is required")
	}
	if !IsValidSemver(m.Version) {
		return fmt.Errorf("version must be valid semver (got %q)", m.Version)
	}

	if m.Author == "" {
		return fmt.Errorf("author is required")
	}
	if len(m.Author) > 128 {
		return fmt.Errorf("author name too long (max 128 characters)")
	}

	if len(m.Description) > 500 {
		return fmt.Errorf("description too long (max 500 characters)")
	}

	if m.MinGameVersion != "" && !IsValidSemver(m.MinGameVersion) {
		return fmt.Errorf("min_game_version must be valid semver (got %q)", m.MinGameVersion)
	}

	if m.MaxGameVersion != "" && !IsValidSemver(m.MaxGameVersion) {
		return fmt.Errorf("max_game_version must be valid semver (got %q)", m.MaxGameVersion)
	}

	// Validate dependencies
	for i, dep := range m.Dependencies {
		if err := dep.Validate(); err != nil {
			return fmt.Errorf("dependency %d invalid: %w", i, err)
		}
	}

	// Validate tags
	for i, tag := range m.Tags {
		if len(tag) == 0 {
			return fmt.Errorf("tag %d is empty", i)
		}
		if len(tag) > 32 {
			return fmt.Errorf("tag %d too long (max 32 characters)", i)
		}
	}

	if len(m.Tags) > 10 {
		return fmt.Errorf("too many tags (max 10)")
	}

	return nil
}

// Validate checks that a dependency has valid fields.
func (d *Dependency) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("dependency name is required")
	}
	if !validNamePattern.MatchString(d.Name) {
		return fmt.Errorf("dependency name must be lowercase alphanumeric + hyphens (got %q)", d.Name)
	}
	if d.Version == "" {
		return fmt.Errorf("dependency version constraint is required")
	}
	// Version constraints can include operators like ^, >=, etc.
	// We validate the core semver part exists
	if !hasValidSemverCore(d.Version) {
		return fmt.Errorf("dependency version must contain valid semver (got %q)", d.Version)
	}
	return nil
}

// IsValidSemver checks if a string is a valid semantic version.
func IsValidSemver(version string) bool {
	return semverPattern.MatchString(version)
}

// hasValidSemverCore checks if a version constraint contains a valid semver.
// Handles constraints like "^1.0.0", ">=2.0.0", "~1.2.3"
func hasValidSemverCore(constraint string) bool {
	// Remove common constraint operators
	cleaned := strings.TrimLeft(constraint, "^~>=<!")
	cleaned = strings.TrimSpace(cleaned)
	return IsValidSemver(cleaned)
}

// LoadManifest reads and parses a manifest file.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	return &manifest, nil
}

// LoadManifestFromDir reads mod.json from a directory.
func LoadManifestFromDir(dir string) (*Manifest, error) {
	manifestPath := filepath.Join(dir, "mod.json")
	return LoadManifest(manifestPath)
}

// Save writes the manifest to a file as formatted JSON.
func (m *Manifest) Save(path string) error {
	if err := m.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid manifest: %w", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// HasDependency checks if a mod depends on another mod by name.
func (m *Manifest) HasDependency(modName string) bool {
	for _, dep := range m.Dependencies {
		if dep.Name == modName {
			return true
		}
	}
	return false
}

// HasConflict checks if a mod conflicts with another mod by name.
func (m *Manifest) HasConflict(modName string) bool {
	for _, conflict := range m.Conflicts {
		if conflict == modName {
			return true
		}
	}
	return false
}

// IsCompatibleWithGenre checks if the mod is compatible with a genre.
// If GenreOverrides is empty, the mod is compatible with all genres.
func (m *Manifest) IsCompatibleWithGenre(genre string) bool {
	if len(m.GenreOverrides) == 0 {
		return true // Compatible with all genres
	}
	for _, g := range m.GenreOverrides {
		if strings.EqualFold(g, genre) {
			return true
		}
	}
	return false
}
