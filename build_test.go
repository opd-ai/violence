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

// TestBuildWorkflowBinarySigning validates GPG signing is configured for Linux and Windows builds
func TestBuildWorkflowBinarySigning(t *testing.T) {
	workflowPath := filepath.Join(".github", "workflows", "build.yml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to read build workflow: %v", err)
	}

	content := string(data)

	// Verify GPG signing step exists for Linux
	if !strings.Contains(content, "GPG sign binary") {
		t.Error("Build workflow missing GPG signing step")
	}

	// Verify GPG secrets are referenced
	if !strings.Contains(content, "GPG_PRIVATE_KEY") {
		t.Error("Build workflow missing GPG_PRIVATE_KEY secret reference")
	}
	if !strings.Contains(content, "GPG_PASSPHRASE") {
		t.Error("Build workflow missing GPG_PASSPHRASE secret reference")
	}

	// Verify signing only runs on tags
	if !strings.Contains(content, "startsWith(github.ref, 'refs/tags/v')") {
		t.Error("GPG signing should only run on version tags")
	}

	// Verify detached signatures are generated
	if !strings.Contains(content, "--detach-sign") {
		t.Error("GPG signing should produce detached signatures")
	}
	if !strings.Contains(content, "--armor") {
		t.Error("GPG signatures should be ASCII-armored (.asc)")
	}

	// Verify SHA256 checksums are generated
	if !strings.Contains(content, "sha256sum") || !strings.Contains(content, ".sha256") {
		t.Error("Build workflow should generate SHA256 checksums")
	}

	// Verify .asc signature files are included in artifacts
	if !strings.Contains(content, "violence-linux-${{ matrix.arch }}.asc") {
		t.Error("Linux artifact should include .asc signature file")
	}
	if !strings.Contains(content, "violence-windows-amd64.exe.asc") {
		t.Error("Windows artifact should include .asc signature file")
	}
}

// TestBuildWorkflowMacOSNotarization validates macOS notarization is configured
func TestBuildWorkflowMacOSNotarization(t *testing.T) {
	workflowPath := filepath.Join(".github", "workflows", "build.yml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to read build workflow: %v", err)
	}

	content := string(data)

	// Verify notarization step exists
	if !strings.Contains(content, "Notarize macOS binary") {
		t.Error("Build workflow missing macOS notarization step")
	}

	// Verify Apple secrets are referenced
	if !strings.Contains(content, "APPLE_ID") {
		t.Error("Build workflow missing APPLE_ID secret reference")
	}
	if !strings.Contains(content, "APPLE_PASSWORD") {
		t.Error("Build workflow missing APPLE_PASSWORD secret reference")
	}
	if !strings.Contains(content, "APPLE_TEAM_ID") {
		t.Error("Build workflow missing APPLE_TEAM_ID secret reference")
	}

	// Verify notarytool is used
	if !strings.Contains(content, "notarytool") {
		t.Error("macOS notarization should use notarytool")
	}

	// Verify codesign is used
	if !strings.Contains(content, "codesign") {
		t.Error("macOS binary should be signed with codesign before notarization")
	}

	// Verify notarization only runs on tags
	// Find the notarize step and check the if condition follows it
	notarizeIdx := strings.Index(content, "Notarize macOS binary")
	if notarizeIdx < 0 {
		t.Fatal("Could not find notarization section")
	}
	// Check nearby lines for the tag condition
	nearbyContent := content[notarizeIdx : notarizeIdx+200]
	if !strings.Contains(nearbyContent, "refs/tags/v") {
		t.Error("macOS notarization should only run on version tags")
	}
}

// TestReleaseWorkflowValid validates the release workflow structure
func TestReleaseWorkflowValid(t *testing.T) {
	workflowPath := filepath.Join(".github", "workflows", "release.yml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to read release workflow: %v", err)
	}

	var workflow struct {
		Name string `yaml:"name"`
		On   struct {
			Push struct {
				Tags []string `yaml:"tags"`
			} `yaml:"push"`
		} `yaml:"on"`
		Jobs map[string]struct {
			Name string `yaml:"name"`
		} `yaml:"jobs"`
	}

	if err := yaml.Unmarshal(data, &workflow); err != nil {
		t.Fatalf("Failed to parse workflow YAML: %v", err)
	}

	// Verify workflow name
	if workflow.Name != "Release" {
		t.Errorf("Expected workflow name 'Release', got '%s'", workflow.Name)
	}

	// Verify release trigger is on tags
	if len(workflow.On.Push.Tags) == 0 {
		t.Error("Release workflow should trigger on tag push")
	}
	foundVTag := false
	for _, tag := range workflow.On.Push.Tags {
		if tag == "v*" {
			foundVTag = true
			break
		}
	}
	if !foundVTag {
		t.Error("Release workflow should trigger on 'v*' tags")
	}

	// Expected jobs
	expectedJobs := []string{"build", "release"}
	for _, jobID := range expectedJobs {
		if _, exists := workflow.Jobs[jobID]; !exists {
			t.Errorf("Missing expected job: %s", jobID)
		}
	}
}

// TestReleaseWorkflowContent validates release workflow details
func TestReleaseWorkflowContent(t *testing.T) {
	workflowPath := filepath.Join(".github", "workflows", "release.yml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to read release workflow: %v", err)
	}

	content := string(data)

	// Verify it downloads all artifacts
	if !strings.Contains(content, "download-artifact") {
		t.Error("Release workflow should download build artifacts")
	}

	// Verify draft release creation
	if !strings.Contains(content, "draft: true") {
		t.Error("Release should be created as draft")
	}

	// Verify it uses action-gh-release
	if !strings.Contains(content, "softprops/action-gh-release") {
		t.Error("Release workflow should use softprops/action-gh-release")
	}

	// Verify it references GITHUB_TOKEN
	if !strings.Contains(content, "GITHUB_TOKEN") {
		t.Error("Release workflow should use GITHUB_TOKEN for authentication")
	}

	// Verify all platform artifacts are collected
	platformArtifacts := []string{
		"violence-linux-amd64",
		"violence-linux-arm64",
		"violence-darwin-universal",
		"violence-windows-amd64",
		"violence.wasm",
		"violence-ios",
		"violence-android",
	}
	for _, artifact := range platformArtifacts {
		if !strings.Contains(content, artifact) {
			t.Errorf("Release workflow missing reference to artifact: %s", artifact)
		}
	}

	// Verify signature files are included
	if !strings.Contains(content, ".asc") {
		t.Error("Release workflow should include GPG signature files")
	}

	// Verify checksums are included
	if !strings.Contains(content, "CHECKSUMS-SHA256") {
		t.Error("Release workflow should generate combined checksums file")
	}

	// Verify release notes generation
	if !strings.Contains(content, "release_notes") {
		t.Error("Release workflow should generate release notes")
	}

	// Verify reusable workflow call to build
	if !strings.Contains(content, "uses: ./.github/workflows/build.yml") {
		t.Error("Release workflow should call build workflow as reusable workflow")
	}

	// Verify secrets inheritance
	if !strings.Contains(content, "secrets: inherit") {
		t.Error("Release workflow should inherit secrets for signing")
	}
}
