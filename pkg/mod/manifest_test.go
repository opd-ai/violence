package mod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManifest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		manifest  Manifest
		wantError bool
	}{
		{
			name: "valid minimal manifest",
			manifest: Manifest{
				Name:    "test-mod",
				Version: "1.0.0",
				Author:  "Test Author",
			},
			wantError: false,
		},
		{
			name: "valid full manifest",
			manifest: Manifest{
				Name:           "advanced-mod",
				Version:        "2.1.3",
				Description:    "An advanced mod with all features",
				Author:         "Mod Developer",
				License:        "MIT",
				Homepage:       "https://example.com",
				Tags:           []string{"gameplay", "weapons"},
				GenreOverrides: []string{"fps", "horror"},
				Dependencies: []Dependency{
					{Name: "base-mod", Version: "^1.0.0"},
					{Name: "util-mod", Version: ">=2.0.0", Optional: true},
				},
				Conflicts:      []string{"old-mod"},
				MinGameVersion: "1.0.0",
				MaxGameVersion: "2.0.0",
				EntryPoint:     "mod.wasm",
				Permissions: PermissionSet{
					FileRead:  true,
					AssetLoad: true,
				},
			},
			wantError: false,
		},
		{
			name: "missing name",
			manifest: Manifest{
				Version: "1.0.0",
				Author:  "Test",
			},
			wantError: true,
		},
		{
			name: "missing version",
			manifest: Manifest{
				Name:   "test-mod",
				Author: "Test",
			},
			wantError: true,
		},
		{
			name: "missing author",
			manifest: Manifest{
				Name:    "test-mod",
				Version: "1.0.0",
			},
			wantError: true,
		},
		{
			name: "invalid name - uppercase",
			manifest: Manifest{
				Name:    "TestMod",
				Version: "1.0.0",
				Author:  "Test",
			},
			wantError: true,
		},
		{
			name: "invalid name - spaces",
			manifest: Manifest{
				Name:    "test mod",
				Version: "1.0.0",
				Author:  "Test",
			},
			wantError: true,
		},
		{
			name: "invalid name - underscore",
			manifest: Manifest{
				Name:    "test_mod",
				Version: "1.0.0",
				Author:  "Test",
			},
			wantError: true,
		},
		{
			name: "invalid name - starts with hyphen",
			manifest: Manifest{
				Name:    "-test-mod",
				Version: "1.0.0",
				Author:  "Test",
			},
			wantError: true,
		},
		{
			name: "invalid name - ends with hyphen",
			manifest: Manifest{
				Name:    "test-mod-",
				Version: "1.0.0",
				Author:  "Test",
			},
			wantError: true,
		},
		{
			name: "invalid name - too long",
			manifest: Manifest{
				Name:    "this-is-a-very-long-mod-name-that-exceeds-the-maximum-allowed-length-of-sixty-four-characters",
				Version: "1.0.0",
				Author:  "Test",
			},
			wantError: true,
		},
		{
			name: "invalid version - not semver",
			manifest: Manifest{
				Name:    "test-mod",
				Version: "1.0",
				Author:  "Test",
			},
			wantError: true,
		},
		{
			name: "invalid version - non-numeric",
			manifest: Manifest{
				Name:    "test-mod",
				Version: "abc",
				Author:  "Test",
			},
			wantError: true,
		},
		{
			name: "author too long",
			manifest: Manifest{
				Name:    "test-mod",
				Version: "1.0.0",
				Author:  string(make([]byte, 200)),
			},
			wantError: true,
		},
		{
			name: "description too long",
			manifest: Manifest{
				Name:        "test-mod",
				Version:     "1.0.0",
				Author:      "Test",
				Description: string(make([]byte, 600)),
			},
			wantError: true,
		},
		{
			name: "invalid min game version",
			manifest: Manifest{
				Name:           "test-mod",
				Version:        "1.0.0",
				Author:         "Test",
				MinGameVersion: "1.0",
			},
			wantError: true,
		},
		{
			name: "invalid max game version",
			manifest: Manifest{
				Name:           "test-mod",
				Version:        "1.0.0",
				Author:         "Test",
				MaxGameVersion: "abc",
			},
			wantError: true,
		},
		{
			name: "too many tags",
			manifest: Manifest{
				Name:    "test-mod",
				Version: "1.0.0",
				Author:  "Test",
				Tags:    []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
			},
			wantError: true,
		},
		{
			name: "tag too long",
			manifest: Manifest{
				Name:    "test-mod",
				Version: "1.0.0",
				Author:  "Test",
				Tags:    []string{"this-is-a-very-long-tag-that-exceeds-limit"},
			},
			wantError: true,
		},
		{
			name: "empty tag",
			manifest: Manifest{
				Name:    "test-mod",
				Version: "1.0.0",
				Author:  "Test",
				Tags:    []string{"valid", ""},
			},
			wantError: true,
		},
		{
			name: "invalid dependency",
			manifest: Manifest{
				Name:    "test-mod",
				Version: "1.0.0",
				Author:  "Test",
				Dependencies: []Dependency{
					{Name: "", Version: "1.0.0"},
				},
			},
			wantError: true,
		},
		{
			name: "semver with v prefix",
			manifest: Manifest{
				Name:    "test-mod",
				Version: "v1.0.0",
				Author:  "Test",
			},
			wantError: false,
		},
		{
			name: "semver with prerelease",
			manifest: Manifest{
				Name:    "test-mod",
				Version: "1.0.0-alpha.1",
				Author:  "Test",
			},
			wantError: false,
		},
		{
			name: "semver with build metadata",
			manifest: Manifest{
				Name:    "test-mod",
				Version: "1.0.0+build.123",
				Author:  "Test",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.Validate()
			if tt.wantError && err == nil {
				t.Errorf("Validate() expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestDependency_Validate(t *testing.T) {
	tests := []struct {
		name       string
		dependency Dependency
		wantError  bool
	}{
		{
			name:       "valid exact version",
			dependency: Dependency{Name: "base-mod", Version: "1.0.0"},
			wantError:  false,
		},
		{
			name:       "valid caret constraint",
			dependency: Dependency{Name: "util-mod", Version: "^1.2.3"},
			wantError:  false,
		},
		{
			name:       "valid tilde constraint",
			dependency: Dependency{Name: "lib-mod", Version: "~2.0.0"},
			wantError:  false,
		},
		{
			name:       "valid greater than constraint",
			dependency: Dependency{Name: "api-mod", Version: ">=1.0.0"},
			wantError:  false,
		},
		{
			name:       "valid less than constraint",
			dependency: Dependency{Name: "compat-mod", Version: "<2.0.0"},
			wantError:  false,
		},
		{
			name:       "optional dependency",
			dependency: Dependency{Name: "opt-mod", Version: "1.0.0", Optional: true},
			wantError:  false,
		},
		{
			name:       "missing name",
			dependency: Dependency{Version: "1.0.0"},
			wantError:  true,
		},
		{
			name:       "missing version",
			dependency: Dependency{Name: "test-mod"},
			wantError:  true,
		},
		{
			name:       "invalid name format",
			dependency: Dependency{Name: "Invalid_Mod", Version: "1.0.0"},
			wantError:  true,
		},
		{
			name:       "invalid version constraint",
			dependency: Dependency{Name: "test-mod", Version: "abc"},
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dependency.Validate()
			if tt.wantError && err == nil {
				t.Errorf("Validate() expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestIsValidSemver(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"1.0.0", true},
		{"v1.0.0", true},
		{"0.0.1", true},
		{"10.20.30", true},
		{"1.0.0-alpha", true},
		{"1.0.0-alpha.1", true},
		{"1.0.0-beta.2", true},
		{"1.0.0+build.123", true},
		{"1.0.0-rc.1+build.456", true},
		{"1.0", false},
		{"1", false},
		{"abc", false},
		{"1.0.x", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := IsValidSemver(tt.version)
			if got != tt.want {
				t.Errorf("IsValidSemver(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestLoadManifest(t *testing.T) {
	tmpDir := t.TempDir()

	validManifest := `{
  "name": "test-mod",
  "version": "1.0.0",
  "description": "Test mod",
  "author": "Tester"
}`

	invalidJSON := `{invalid json`

	invalidManifest := `{
  "name": "Invalid Name",
  "version": "1.0.0",
  "author": "Tester"
}`

	tests := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name:      "valid manifest",
			content:   validManifest,
			wantError: false,
		},
		{
			name:      "invalid json",
			content:   invalidJSON,
			wantError: true,
		},
		{
			name:      "invalid manifest",
			content:   invalidManifest,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.name+".json")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			manifest, err := LoadManifest(path)
			if tt.wantError {
				if err == nil {
					t.Errorf("LoadManifest() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("LoadManifest() unexpected error: %v", err)
				}
				if manifest == nil {
					t.Errorf("LoadManifest() returned nil manifest")
				}
			}
		})
	}

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := LoadManifest(filepath.Join(tmpDir, "nonexistent.json"))
		if err == nil {
			t.Errorf("LoadManifest() expected error for nonexistent file")
		}
	})
}

func TestLoadManifestFromDir(t *testing.T) {
	tmpDir := t.TempDir()

	validManifest := `{
  "name": "dir-mod",
  "version": "1.0.0",
  "description": "Directory test mod",
  "author": "Tester"
}`

	modDir := filepath.Join(tmpDir, "test-mod")
	if err := os.Mkdir(modDir, 0o755); err != nil {
		t.Fatalf("Failed to create mod directory: %v", err)
	}

	manifestPath := filepath.Join(modDir, "mod.json")
	if err := os.WriteFile(manifestPath, []byte(validManifest), 0o644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	manifest, err := LoadManifestFromDir(modDir)
	if err != nil {
		t.Errorf("LoadManifestFromDir() unexpected error: %v", err)
	}
	if manifest == nil {
		t.Errorf("LoadManifestFromDir() returned nil manifest")
	}
	if manifest != nil && manifest.Name != "dir-mod" {
		t.Errorf("LoadManifestFromDir() name = %q, want %q", manifest.Name, "dir-mod")
	}
}

func TestManifest_Save(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		manifest  Manifest
		wantError bool
	}{
		{
			name: "valid manifest",
			manifest: Manifest{
				Name:        "save-test",
				Version:     "1.0.0",
				Description: "Test saving",
				Author:      "Tester",
			},
			wantError: false,
		},
		{
			name: "invalid manifest",
			manifest: Manifest{
				Name:    "Invalid Name",
				Version: "1.0.0",
				Author:  "Tester",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.name+".json")
			err := tt.manifest.Save(path)

			if tt.wantError {
				if err == nil {
					t.Errorf("Save() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Save() unexpected error: %v", err)
				}

				// Verify file was created and can be loaded
				loaded, err := LoadManifest(path)
				if err != nil {
					t.Errorf("Failed to load saved manifest: %v", err)
				}
				if loaded.Name != tt.manifest.Name {
					t.Errorf("Loaded name = %q, want %q", loaded.Name, tt.manifest.Name)
				}
			}
		})
	}
}

func TestManifest_HasDependency(t *testing.T) {
	manifest := Manifest{
		Name:    "test-mod",
		Version: "1.0.0",
		Author:  "Tester",
		Dependencies: []Dependency{
			{Name: "dep1", Version: "1.0.0"},
			{Name: "dep2", Version: "2.0.0"},
		},
	}

	tests := []struct {
		modName string
		want    bool
	}{
		{"dep1", true},
		{"dep2", true},
		{"dep3", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.modName, func(t *testing.T) {
			got := manifest.HasDependency(tt.modName)
			if got != tt.want {
				t.Errorf("HasDependency(%q) = %v, want %v", tt.modName, got, tt.want)
			}
		})
	}
}

func TestManifest_HasConflict(t *testing.T) {
	manifest := Manifest{
		Name:      "test-mod",
		Version:   "1.0.0",
		Author:    "Tester",
		Conflicts: []string{"old-mod", "incompatible-mod"},
	}

	tests := []struct {
		modName string
		want    bool
	}{
		{"old-mod", true},
		{"incompatible-mod", true},
		{"compatible-mod", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.modName, func(t *testing.T) {
			got := manifest.HasConflict(tt.modName)
			if got != tt.want {
				t.Errorf("HasConflict(%q) = %v, want %v", tt.modName, got, tt.want)
			}
		})
	}
}

func TestManifest_IsCompatibleWithGenre(t *testing.T) {
	tests := []struct {
		name           string
		genreOverrides []string
		testGenre      string
		want           bool
	}{
		{
			name:           "no overrides - compatible with all",
			genreOverrides: nil,
			testGenre:      "fps",
			want:           true,
		},
		{
			name:           "empty overrides - compatible with all",
			genreOverrides: []string{},
			testGenre:      "horror",
			want:           true,
		},
		{
			name:           "specific genre - matching",
			genreOverrides: []string{"fps", "horror"},
			testGenre:      "fps",
			want:           true,
		},
		{
			name:           "specific genre - not matching",
			genreOverrides: []string{"fps", "horror"},
			testGenre:      "rpg",
			want:           false,
		},
		{
			name:           "case insensitive matching",
			genreOverrides: []string{"FPS", "Horror"},
			testGenre:      "fps",
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := Manifest{
				Name:           "test-mod",
				Version:        "1.0.0",
				Author:         "Tester",
				GenreOverrides: tt.genreOverrides,
			}

			got := manifest.IsCompatibleWithGenre(tt.testGenre)
			if got != tt.want {
				t.Errorf("IsCompatibleWithGenre(%q) = %v, want %v", tt.testGenre, got, tt.want)
			}
		})
	}
}

func TestPermissionSet_ToModPermissions(t *testing.T) {
	tests := []struct {
		name string
		ps   PermissionSet
		want ModPermissions
	}{
		{
			name: "all permissions",
			ps: PermissionSet{
				FileRead:    true,
				FileWrite:   true,
				EntitySpawn: true,
				AssetLoad:   true,
				UIModify:    true,
				Network:     true,
			},
			want: ModPermissions{
				AllowFileRead:    true,
				AllowFileWrite:   true,
				AllowEntitySpawn: true,
				AllowAssetLoad:   true,
				AllowUIModify:    true,
			},
		},
		{
			name: "no permissions",
			ps:   PermissionSet{},
			want: ModPermissions{},
		},
		{
			name: "selective permissions",
			ps: PermissionSet{
				FileRead:  true,
				AssetLoad: true,
			},
			want: ModPermissions{
				AllowFileRead:  true,
				AllowAssetLoad: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ps.ToModPermissions()
			if got != tt.want {
				t.Errorf("ToModPermissions() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
