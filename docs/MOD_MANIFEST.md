# Mod Manifest Schema

## Overview

The Violence mod system uses a `mod.json` manifest file to define mod metadata, dependencies, permissions, and configuration. This document describes the complete schema.

## Manifest Structure

### Required Fields

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| `name` | string | Unique mod identifier | Lowercase alphanumeric + hyphens only. Must start and end with alphanumeric. Max 64 characters. Example: `"advanced-weapons-mod"` |
| `version` | string | Mod version | Must be valid semver (e.g., `"1.2.3"`, `"v2.0.0-beta.1"`). Supports prerelease and build metadata. |
| `author` | string | Creator name or identifier | Max 128 characters |

### Optional Fields

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| `description` | string | Human-readable mod description | Max 500 characters |
| `license` | string | License identifier | Examples: `"MIT"`, `"GPL-3.0"`, `"Proprietary"` |
| `homepage` | string | URL to mod website/repository | Valid URL string |
| `tags` | string[] | Keywords for discovery | Max 10 tags. Each tag max 32 characters. |
| `genre_overrides` | string[] | Target genres | Empty = compatible with all genres. Case-insensitive matching. |
| `dependencies` | Dependency[] | Required mods | See Dependency schema below |
| `conflicts` | string[] | Incompatible mod names | Mod names that conflict with this mod |
| `min_game_version` | string | Minimum game version | Must be valid semver |
| `max_game_version` | string | Maximum game version | Must be valid semver |
| `entry_point` | string | WASM binary path | Relative to mod directory |
| `permissions` | PermissionSet | Requested capabilities | See Permissions schema below |
| `config` | object | Mod-specific config | Key-value pairs (any JSON types) |

## Dependency Schema

Dependencies specify other mods required by this mod.

```json
{
  "name": "base-mod",
  "version": "^1.0.0",
  "optional": false
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Dependency mod name (same format as mod name) |
| `version` | string | Yes | Version constraint (semver with operators) |
| `optional` | boolean | No | If true, dependency is recommended but not required |

### Version Constraint Operators

- **Exact**: `"1.0.0"` - Exact version match
- **Caret**: `"^1.2.3"` - Compatible with version (>=1.2.3 <2.0.0)
- **Tilde**: `"~1.2.3"` - Approximate version (>=1.2.3 <1.3.0)
- **Range**: `">=1.0.0"`, `"<2.0.0"` - Comparison operators

## Permissions Schema

Permissions define what capabilities the mod requires.

```json
{
  "file_read": true,
  "file_write": false,
  "entity_spawn": true,
  "asset_load": true,
  "ui_modify": false,
  "network": false
}
```

| Permission | Description | Default |
|------------|-------------|---------|
| `file_read` | Read files in allowed directories | `false` |
| `file_write` | Write files in allowed directories | `false` |
| `entity_spawn` | Spawn game entities | `false` |
| `asset_load` | Load textures and sounds | `false` |
| `ui_modify` | Modify UI elements | `false` |
| `network` | Network access (reserved for future use) | `false` |

## Complete Example

See `docs/MOD_MANIFEST_EXAMPLE.json` for a full example manifest.

```json
{
  "name": "weapon-enhancer",
  "version": "1.0.0",
  "description": "Enhances weapon damage and effects",
  "author": "YourName",
  "license": "MIT",
  "homepage": "https://github.com/example/weapon-enhancer",
  "tags": ["weapons", "gameplay", "balance"],
  "genre_overrides": ["fps", "action"],
  "dependencies": [
    {
      "name": "core-lib",
      "version": "^2.0.0"
    }
  ],
  "conflicts": ["old-weapon-mod"],
  "min_game_version": "1.0.0",
  "entry_point": "enhancer.wasm",
  "permissions": {
    "entity_spawn": true,
    "asset_load": true
  },
  "config": {
    "damage_multiplier": 1.5
  }
}
```

## Validation Rules

The manifest is validated on load with the following rules:

1. **Name**: Must match pattern `^[a-z0-9][a-z0-9-]*[a-z0-9]$`
2. **Version**: Must be valid semver (supports v prefix, prerelease, build metadata)
3. **Author**: Required, max 128 characters
4. **Description**: Optional, max 500 characters
5. **Tags**: Max 10 tags, each max 32 characters, no empty tags
6. **Dependencies**: Each must have valid name and version constraint
7. **Game Versions**: If specified, must be valid semver

## Backward Compatibility

The mod loader maintains backward compatibility with legacy manifest formats:

- Old `mod.json` files without the new fields will continue to work
- Legacy fields are populated from the new Manifest structure
- Existing mods don't need immediate migration

## API Usage

### Loading a Manifest

```go
import "github.com/opd-ai/violence/pkg/mod"

// From file
manifest, err := mod.LoadManifest("path/to/mod.json")

// From directory (looks for mod.json inside)
manifest, err := mod.LoadManifestFromDir("path/to/mod/")

// Validate
if err := manifest.Validate(); err != nil {
    log.Fatalf("Invalid manifest: %v", err)
}
```

### Creating a Manifest

```go
manifest := &mod.Manifest{
    Name:        "my-mod",
    Version:     "1.0.0",
    Author:      "Your Name",
    Description: "My awesome mod",
    Tags:        []string{"gameplay", "weapons"},
    Dependencies: []mod.Dependency{
        {Name: "base-lib", Version: "^1.0.0"},
    },
    Permissions: mod.PermissionSet{
        FileRead:  true,
        AssetLoad: true,
    },
}

// Save to file
if err := manifest.Save("mod.json"); err != nil {
    log.Fatalf("Failed to save: %v", err)
}
```

### Checking Compatibility

```go
// Check if mod is compatible with a genre
if manifest.IsCompatibleWithGenre("fps") {
    // Load mod
}

// Check for dependencies
if manifest.HasDependency("required-lib") {
    // Verify required-lib is loaded
}

// Check for conflicts
if manifest.HasConflict("incompatible-mod") {
    // Don't load both mods
}
```

## Future Enhancements

Planned additions to the manifest schema:

- **Screenshots**: Array of screenshot URLs for mod marketplace
- **Download Stats**: Automatically populated by registry
- **Rating**: User rating metadata
- **Changelog**: Version history
- **Compatibility Matrix**: Tested game version combinations
- **Asset Hashes**: Integrity verification for downloaded files
