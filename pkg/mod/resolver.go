package mod

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Resolver handles dependency resolution and version constraints.
type Resolver struct {
	available map[string][]string // map[modName][]versions
}

// NewResolver creates a dependency resolver.
func NewResolver() *Resolver {
	return &Resolver{
		available: make(map[string][]string),
	}
}

// AddAvailable registers available mod versions.
func (r *Resolver) AddAvailable(name string, versions []string) {
	r.available[name] = versions
}

// Resolve computes installation order for a mod and its dependencies.
// Returns ordered list of (name, version) tuples or error if unresolvable.
func (r *Resolver) Resolve(root *Manifest) ([]ModVersion, error) {
	resolved := make(map[string]string) // map[modName]selectedVersion
	visiting := make(map[string]bool)   // cycle detection
	visited := make(map[string]bool)    // completion tracking

	if err := r.visit(root.Name, root.Version, root, resolved, visiting, visited); err != nil {
		return nil, err
	}

	// Convert to ordered list
	result := make([]ModVersion, 0, len(resolved))
	for name, version := range resolved {
		result = append(result, ModVersion{Name: name, Version: version})
	}

	// Topological sort (dependencies before dependents)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// ModVersion represents a resolved mod with specific version.
type ModVersion struct {
	Name    string
	Version string
}

// visit performs DFS to resolve dependencies.
func (r *Resolver) visit(name, version string, manifest *Manifest, resolved map[string]string, visiting, visited map[string]bool) error {
	if visiting[name] {
		return fmt.Errorf("circular dependency detected: %s", name)
	}
	if visited[name] {
		// Already resolved, check version compatibility
		if resolved[name] != version {
			return fmt.Errorf("version conflict for %s: need %s but already resolved to %s", name, version, resolved[name])
		}
		return nil
	}

	visiting[name] = true
	defer func() { visiting[name] = false }()

	// Resolve dependencies first
	for _, dep := range manifest.Dependencies {
		if dep.Optional {
			continue // Skip optional dependencies
		}

		// Find compatible version
		selectedVersion, err := r.selectVersion(dep.Name, dep.Version)
		if err != nil {
			return fmt.Errorf("cannot resolve %s dependency %s: %w", name, dep.Name, err)
		}

		// Load dependency manifest (stub - would fetch from registry)
		depManifest := &Manifest{
			Name:         dep.Name,
			Version:      selectedVersion,
			Dependencies: nil, // Would load actual deps from registry
		}

		if err := r.visit(dep.Name, selectedVersion, depManifest, resolved, visiting, visited); err != nil {
			return err
		}
	}

	resolved[name] = version
	visited[name] = true
	return nil
}

// selectVersion picks the best matching version for a constraint.
func (r *Resolver) selectVersion(modName, constraint string) (string, error) {
	available, ok := r.available[modName]
	if !ok || len(available) == 0 {
		return "", fmt.Errorf("mod %s not found in registry", modName)
	}

	// Find all matching versions
	var matches []string
	for _, version := range available {
		if satisfies(version, constraint) {
			matches = append(matches, version)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no version of %s satisfies constraint %s", modName, constraint)
	}

	// Return highest matching version
	sort.Slice(matches, func(i, j int) bool {
		return compareVersions(matches[i], matches[j]) > 0
	})

	return matches[0], nil
}

// satisfies checks if a version satisfies a constraint.
// Supports: ^1.2.3 (caret), ~1.2.3 (tilde), >=1.2.3, <=1.2.3, >1.2.3, <1.2.3, 1.2.3 (exact)
func satisfies(version, constraint string) bool {
	constraint = strings.TrimSpace(constraint)

	// Exact match
	if !strings.ContainsAny(constraint, "^~>=<") {
		return version == constraint
	}

	// Caret range (^1.2.3 allows >=1.2.3 and <2.0.0)
	if strings.HasPrefix(constraint, "^") {
		base := strings.TrimPrefix(constraint, "^")
		if compareVersions(version, base) < 0 {
			return false
		}
		major := getMajor(base)
		nextMajor := fmt.Sprintf("%d.0.0", major+1)
		return compareVersions(version, nextMajor) < 0
	}

	// Tilde range (~1.2.3 allows >=1.2.3 and <1.3.0)
	if strings.HasPrefix(constraint, "~") {
		base := strings.TrimPrefix(constraint, "~")
		if compareVersions(version, base) < 0 {
			return false
		}
		major := getMajor(base)
		minor := getMinor(base)
		nextMinor := fmt.Sprintf("%d.%d.0", major, minor+1)
		return compareVersions(version, nextMinor) < 0
	}

	// Comparison operators
	if strings.HasPrefix(constraint, ">=") {
		base := strings.TrimPrefix(constraint, ">=")
		return compareVersions(version, base) >= 0
	}
	if strings.HasPrefix(constraint, "<=") {
		base := strings.TrimPrefix(constraint, "<=")
		return compareVersions(version, base) <= 0
	}
	if strings.HasPrefix(constraint, ">") {
		base := strings.TrimPrefix(constraint, ">")
		return compareVersions(version, base) > 0
	}
	if strings.HasPrefix(constraint, "<") {
		base := strings.TrimPrefix(constraint, "<")
		return compareVersions(version, base) < 0
	}

	return false
}

// compareVersions returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
func compareVersions(v1, v2 string) int {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.SplitN(v1, "-", 2)
	parts2 := strings.SplitN(v2, "-", 2)

	core1 := parts1[0]
	core2 := parts2[0]

	// Compare core version (major.minor.patch)
	nums1 := strings.Split(core1, ".")
	nums2 := strings.Split(core2, ".")

	for i := 0; i < 3; i++ {
		n1, n2 := 0, 0
		if i < len(nums1) {
			n1, _ = strconv.Atoi(nums1[i])
		}
		if i < len(nums2) {
			n2, _ = strconv.Atoi(nums2[i])
		}
		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	// If core versions equal, compare prerelease
	if len(parts1) > 1 && len(parts2) == 1 {
		return -1 // v1 is prerelease, v2 is stable
	}
	if len(parts1) == 1 && len(parts2) > 1 {
		return 1 // v1 is stable, v2 is prerelease
	}
	if len(parts1) > 1 && len(parts2) > 1 {
		pre1 := parts1[1]
		pre2 := parts2[1]
		if pre1 < pre2 {
			return -1
		}
		if pre1 > pre2 {
			return 1
		}
	}

	return 0
}

// getMajor extracts major version number.
func getMajor(version string) int {
	version = strings.TrimPrefix(version, "v")
	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		major, _ := strconv.Atoi(parts[0])
		return major
	}
	return 0
}

// getMinor extracts minor version number.
func getMinor(version string) int {
	version = strings.TrimPrefix(version, "v")
	parts := strings.Split(version, ".")
	if len(parts) > 1 {
		minor, _ := strconv.Atoi(parts[1])
		return minor
	}
	return 0
}

// CheckConflicts verifies no conflicting mods are installed.
func CheckConflicts(manifest *Manifest, installed []string) error {
	for _, conflict := range manifest.Conflicts {
		for _, installedMod := range installed {
			if installedMod == conflict {
				return fmt.Errorf("mod %s conflicts with installed mod %s", manifest.Name, conflict)
			}
		}
	}
	return nil
}

// SortTopological orders mods by dependency (dependencies before dependents).
func SortTopological(manifests []*Manifest) ([]*Manifest, error) {
	// Build dependency graph
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	modMap := make(map[string]*Manifest)

	for _, m := range manifests {
		modMap[m.Name] = m
		if _, exists := inDegree[m.Name]; !exists {
			inDegree[m.Name] = 0
		}
		for _, dep := range m.Dependencies {
			if dep.Optional {
				continue
			}
			graph[dep.Name] = append(graph[dep.Name], m.Name)
			inDegree[m.Name]++
		}
	}

	// Kahn's algorithm
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	var result []*Manifest
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if manifest, ok := modMap[current]; ok {
			result = append(result, manifest)
		}

		for _, dependent := range graph[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	if len(result) != len(manifests) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}
