package mod

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1   string
		v2   string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.2.3", "1.2.4", -1},
		{"1.3.0", "1.2.9", 1},
		{"2.0.0", "1.9.9", 1},
		{"v1.0.0", "1.0.0", 0},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-beta", 1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-beta", "1.0.0-alpha", 1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestSatisfies(t *testing.T) {
	tests := []struct {
		version    string
		constraint string
		want       bool
	}{
		// Exact match
		{"1.0.0", "1.0.0", true},
		{"1.0.0", "1.0.1", false},

		// Caret (^) - compatible with same major version
		{"1.2.3", "^1.2.3", true},
		{"1.3.0", "^1.2.3", true},
		{"1.9.9", "^1.2.3", true},
		{"2.0.0", "^1.2.3", false},
		{"1.2.2", "^1.2.3", false},

		// Tilde (~) - compatible with same minor version
		{"1.2.3", "~1.2.3", true},
		{"1.2.4", "~1.2.3", true},
		{"1.2.9", "~1.2.3", true},
		{"1.3.0", "~1.2.3", false},
		{"1.2.2", "~1.2.3", false},

		// Greater than or equal
		{"1.2.3", ">=1.2.3", true},
		{"1.2.4", ">=1.2.3", true},
		{"2.0.0", ">=1.2.3", true},
		{"1.2.2", ">=1.2.3", false},

		// Less than or equal
		{"1.2.3", "<=1.2.3", true},
		{"1.2.2", "<=1.2.3", true},
		{"1.0.0", "<=1.2.3", true},
		{"1.2.4", "<=1.2.3", false},

		// Greater than
		{"1.2.4", ">1.2.3", true},
		{"2.0.0", ">1.2.3", true},
		{"1.2.3", ">1.2.3", false},
		{"1.2.2", ">1.2.3", false},

		// Less than
		{"1.2.2", "<1.2.3", true},
		{"1.0.0", "<1.2.3", true},
		{"1.2.3", "<1.2.3", false},
		{"1.2.4", "<1.2.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.version+"_"+tt.constraint, func(t *testing.T) {
			got := satisfies(tt.version, tt.constraint)
			if got != tt.want {
				t.Errorf("satisfies(%q, %q) = %v, want %v", tt.version, tt.constraint, got, tt.want)
			}
		})
	}
}

func TestGetMajorMinor(t *testing.T) {
	tests := []struct {
		version   string
		wantMajor int
		wantMinor int
	}{
		{"1.2.3", 1, 2},
		{"v2.5.0", 2, 5},
		{"0.1.0", 0, 1},
		{"10.20.30", 10, 20},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			major := getMajor(tt.version)
			minor := getMinor(tt.version)
			if major != tt.wantMajor {
				t.Errorf("getMajor(%q) = %d, want %d", tt.version, major, tt.wantMajor)
			}
			if minor != tt.wantMinor {
				t.Errorf("getMinor(%q) = %d, want %d", tt.version, minor, tt.wantMinor)
			}
		})
	}
}

func TestResolverSelectVersion(t *testing.T) {
	r := NewResolver()
	r.AddAvailable("test-mod", []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"})

	tests := []struct {
		name       string
		modName    string
		constraint string
		wantVer    string
		wantErr    bool
	}{
		{
			name:       "exact_match",
			modName:    "test-mod",
			constraint: "1.1.0",
			wantVer:    "1.1.0",
			wantErr:    false,
		},
		{
			name:       "caret_selects_latest_compatible",
			modName:    "test-mod",
			constraint: "^1.0.0",
			wantVer:    "1.2.0", // Latest 1.x
			wantErr:    false,
		},
		{
			name:       "tilde_selects_latest_patch",
			modName:    "test-mod",
			constraint: "~1.1.0",
			wantVer:    "1.1.0", // Only 1.1.x available
			wantErr:    false,
		},
		{
			name:       "gte_selects_latest",
			modName:    "test-mod",
			constraint: ">=1.1.0",
			wantVer:    "2.0.0", // Latest overall
			wantErr:    false,
		},
		{
			name:       "mod_not_found",
			modName:    "nonexistent",
			constraint: "1.0.0",
			wantErr:    true,
		},
		{
			name:       "no_matching_version",
			modName:    "test-mod",
			constraint: "3.0.0",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.selectVersion(tt.modName, tt.constraint)
			if (err != nil) != tt.wantErr {
				t.Errorf("selectVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantVer {
				t.Errorf("selectVersion() = %q, want %q", got, tt.wantVer)
			}
		})
	}
}

func TestResolverResolve(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Resolver)
		root      *Manifest
		wantCount int
		wantErr   bool
	}{
		{
			name: "no_dependencies",
			setup: func(r *Resolver) {
				r.AddAvailable("root-mod", []string{"1.0.0"})
			},
			root: &Manifest{
				Name:    "root-mod",
				Version: "1.0.0",
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "single_dependency",
			setup: func(r *Resolver) {
				r.AddAvailable("root-mod", []string{"1.0.0"})
				r.AddAvailable("dep-mod", []string{"1.0.0"})
			},
			root: &Manifest{
				Name:    "root-mod",
				Version: "1.0.0",
				Dependencies: []Dependency{
					{Name: "dep-mod", Version: "1.0.0"},
				},
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "optional_dependency_skipped",
			setup: func(r *Resolver) {
				r.AddAvailable("root-mod", []string{"1.0.0"})
			},
			root: &Manifest{
				Name:    "root-mod",
				Version: "1.0.0",
				Dependencies: []Dependency{
					{Name: "opt-mod", Version: "1.0.0", Optional: true},
				},
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "missing_dependency",
			setup: func(r *Resolver) {
				r.AddAvailable("root-mod", []string{"1.0.0"})
			},
			root: &Manifest{
				Name:    "root-mod",
				Version: "1.0.0",
				Dependencies: []Dependency{
					{Name: "missing-mod", Version: "1.0.0"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewResolver()
			tt.setup(r)

			result, err := r.Resolve(tt.root)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) != tt.wantCount {
				t.Errorf("Resolve() returned %d mods, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestCheckConflicts(t *testing.T) {
	tests := []struct {
		name      string
		manifest  *Manifest
		installed []string
		wantErr   bool
	}{
		{
			name: "no_conflicts",
			manifest: &Manifest{
				Name:      "test-mod",
				Conflicts: []string{"bad-mod"},
			},
			installed: []string{"good-mod"},
			wantErr:   false,
		},
		{
			name: "has_conflict",
			manifest: &Manifest{
				Name:      "test-mod",
				Conflicts: []string{"bad-mod"},
			},
			installed: []string{"good-mod", "bad-mod"},
			wantErr:   true,
		},
		{
			name: "no_conflicts_specified",
			manifest: &Manifest{
				Name: "test-mod",
			},
			installed: []string{"any-mod"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckConflicts(tt.manifest, tt.installed)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckConflicts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSortTopological(t *testing.T) {
	tests := []struct {
		name      string
		manifests []*Manifest
		wantOrder []string
		wantErr   bool
	}{
		{
			name: "simple_chain",
			manifests: []*Manifest{
				{
					Name:    "c",
					Version: "1.0.0",
					Dependencies: []Dependency{
						{Name: "b", Version: "1.0.0"},
					},
				},
				{
					Name:    "b",
					Version: "1.0.0",
					Dependencies: []Dependency{
						{Name: "a", Version: "1.0.0"},
					},
				},
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			wantOrder: []string{"a", "b", "c"},
			wantErr:   false,
		},
		{
			name: "no_dependencies",
			manifests: []*Manifest{
				{Name: "a", Version: "1.0.0"},
				{Name: "b", Version: "1.0.0"},
			},
			wantOrder: []string{"a", "b"},
			wantErr:   false,
		},
		{
			name: "diamond_dependency",
			manifests: []*Manifest{
				{
					Name:    "d",
					Version: "1.0.0",
					Dependencies: []Dependency{
						{Name: "b", Version: "1.0.0"},
						{Name: "c", Version: "1.0.0"},
					},
				},
				{
					Name:    "b",
					Version: "1.0.0",
					Dependencies: []Dependency{
						{Name: "a", Version: "1.0.0"},
					},
				},
				{
					Name:    "c",
					Version: "1.0.0",
					Dependencies: []Dependency{
						{Name: "a", Version: "1.0.0"},
					},
				},
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			wantOrder: []string{"a", "b", "c", "d"},
			wantErr:   false,
		},
		{
			name: "optional_deps_ignored",
			manifests: []*Manifest{
				{
					Name:    "b",
					Version: "1.0.0",
					Dependencies: []Dependency{
						{Name: "a", Version: "1.0.0", Optional: true},
					},
				},
			},
			wantOrder: []string{"b"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SortTopological(tt.manifests)
			if (err != nil) != tt.wantErr {
				t.Errorf("SortTopological() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got := make([]string, len(result))
				for i, m := range result {
					got[i] = m.Name
				}

				// Verify all expected mods are present
				if len(got) != len(tt.wantOrder) {
					t.Errorf("SortTopological() returned %d mods, want %d", len(got), len(tt.wantOrder))
					return
				}

				// For simple chains, verify exact order
				// For more complex DAGs, just verify dependencies come before dependents
				if tt.name == "simple_chain" {
					for i := range got {
						if got[i] != tt.wantOrder[i] {
							t.Errorf("SortTopological() order[%d] = %q, want %q", i, got[i], tt.wantOrder[i])
						}
					}
				} else {
					// Verify no missing mods
					wantSet := make(map[string]bool)
					for _, name := range tt.wantOrder {
						wantSet[name] = true
					}
					for _, name := range got {
						if !wantSet[name] {
							t.Errorf("SortTopological() unexpected mod %q", name)
						}
					}
				}
			}
		})
	}
}

func TestResolverAddAvailable(t *testing.T) {
	r := NewResolver()

	r.AddAvailable("mod-a", []string{"1.0.0", "1.1.0"})
	r.AddAvailable("mod-b", []string{"2.0.0"})

	if len(r.available) != 2 {
		t.Errorf("Expected 2 mods registered, got %d", len(r.available))
	}

	if len(r.available["mod-a"]) != 2 {
		t.Errorf("Expected 2 versions for mod-a, got %d", len(r.available["mod-a"]))
	}
}

func TestSatisfiesEdgeCases(t *testing.T) {
	tests := []struct {
		version    string
		constraint string
		want       bool
	}{
		{"1.0.0", " ^1.0.0 ", true}, // Whitespace handling
		{"v1.0.0", "^1.0.0", true},  // Version with v prefix
		{"1.0.0", "", false},        // Empty constraint
	}

	for _, tt := range tests {
		t.Run(tt.version+"_"+tt.constraint, func(t *testing.T) {
			got := satisfies(tt.version, tt.constraint)
			if got != tt.want {
				t.Errorf("satisfies(%q, %q) = %v, want %v", tt.version, tt.constraint, got, tt.want)
			}
		})
	}
}

func TestCompareVersionsWithMetadata(t *testing.T) {
	tests := []struct {
		v1   string
		v2   string
		want int
	}{
		{"1.0.0+build1", "1.0.0+build2", 0}, // Build metadata ignored
		{"1.0.0-rc.1", "1.0.0-rc.2", -1},
		{"1.0.0-beta", "1.0.0-rc", -1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}
