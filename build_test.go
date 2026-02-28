package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestBuildWorkflowValid validates the GitHub Actions build workflow
func TestBuildWorkflowValid(t *testing.T) {
	workflowPath := filepath.Join(".github", "workflows", "build.yml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to read build workflow: %v", err)
	}

	var workflow struct {
		Name string `yaml:"name"`
		Jobs map[string]struct {
			Name string `yaml:"name"`
		} `yaml:"jobs"`
	}

	if err := yaml.Unmarshal(data, &workflow); err != nil {
		t.Fatalf("Failed to parse workflow YAML: %v", err)
	}

	// Verify workflow name
	if workflow.Name != "Multi-Platform Build" {
		t.Errorf("Expected workflow name 'Multi-Platform Build', got '%s'", workflow.Name)
	}

	// Expected build jobs
	expectedJobs := []string{
		"build-linux",
		"build-macos",
		"build-windows",
		"build-wasm",
		"build-ios",
		"build-android",
		"summary",
	}

	// Check all expected jobs exist
	for _, jobID := range expectedJobs {
		if _, exists := workflow.Jobs[jobID]; !exists {
			t.Errorf("Missing expected job: %s", jobID)
		}
	}

	// Verify job count (should have exactly the expected jobs)
	if len(workflow.Jobs) != len(expectedJobs) {
		t.Errorf("Expected %d jobs, found %d", len(expectedJobs), len(workflow.Jobs))
	}
}

// TestBuildWorkflowMobileJobs validates mobile-specific job configuration
func TestBuildWorkflowMobileJobs(t *testing.T) {
	workflowPath := filepath.Join(".github", "workflows", "build.yml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to read build workflow: %v", err)
	}

	content := string(data)

	// Check iOS job has required steps
	if !strings.Contains(content, "build-ios:") {
		t.Error("Missing build-ios job")
	}
	if !strings.Contains(content, "gomobile bind -target=ios") {
		t.Error("iOS job missing gomobile bind command")
	}
	if !strings.Contains(content, "Violence.xcframework") {
		t.Error("iOS job not producing .xcframework")
	}

	// Check Android job has required steps
	if !strings.Contains(content, "build-android:") {
		t.Error("Missing build-android job")
	}
	if !strings.Contains(content, "gomobile bind -target=android") {
		t.Error("Android job missing gomobile bind command")
	}
	if !strings.Contains(content, "violence.aar") {
		t.Error("Android job not producing .aar")
	}
	if !strings.Contains(content, "setup-java@v4") {
		t.Error("Android job missing Java setup")
	}

	// Check summary job depends on mobile builds
	if !strings.Contains(content, "needs: [build-linux, build-macos, build-windows, build-wasm, build-ios, build-android]") {
		t.Error("Summary job not depending on all build jobs including mobile")
	}
}

// TestBuildDocumentation validates BUILD_MATRIX.md mentions mobile platforms
func TestBuildDocumentation(t *testing.T) {
	docPath := filepath.Join("docs", "BUILD_MATRIX.md")
	data, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("Failed to read BUILD_MATRIX.md: %v", err)
	}

	content := string(data)

	// Check mobile platforms are documented
	requiredSections := []string{
		"### iOS",
		"### Android",
		"violence-ios",
		"violence-android",
		".xcframework",
		".aar",
		"gomobile",
	}

	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			t.Errorf("BUILD_MATRIX.md missing required section: %s", section)
		}
	}

	// Check platforms count
	if !strings.Contains(content, "6 platforms") && !strings.Contains(content, "8 platform") {
		t.Error("BUILD_MATRIX.md should mention updated platform count")
	}
}
