// Package ui provides mod browser UI and auto-update mechanisms.
package ui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/input"
	"github.com/opd-ai/violence/pkg/mod"
	"github.com/opd-ai/violence/pkg/mod/registry"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font/basicfont"
)

// ModBrowserState represents the current view in the mod browser.
type ModBrowserState int

const (
	ModBrowserStateBrowse ModBrowserState = iota
	ModBrowserStateDetails
	ModBrowserStateInstalling
	ModBrowserStateUpdating
)

// ModBrowser manages mod browsing, installation, and auto-updates.
type ModBrowser struct {
	registryURL     string
	state           ModBrowserState
	mods            []registry.ModRecord
	selectedIndex   int
	scrollOffset    int
	visible         bool
	searchQuery     string
	installedMods   map[string]string // name -> version
	updateAvailable map[string]string // name -> new version
	mu              sync.RWMutex
	httpClient      *http.Client
	installing      bool
	installProgress string
	errorMessage    string
	errorTime       time.Time
	autoUpdateCheck time.Time
}

// NewModBrowser creates a new mod browser UI.
func NewModBrowser(registryURL string) *ModBrowser {
	return &ModBrowser{
		registryURL:     registryURL,
		state:           ModBrowserStateBrowse,
		mods:            []registry.ModRecord{},
		installedMods:   make(map[string]string),
		updateAvailable: make(map[string]string),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetVisible toggles mod browser visibility.
func (mb *ModBrowser) SetVisible(visible bool) {
	mb.visible = visible
	if visible && len(mb.mods) == 0 {
		mb.RefreshModList()
	}
}

// IsVisible returns whether the mod browser is visible.
func (mb *ModBrowser) IsVisible() bool {
	return mb.visible
}

// RefreshModList fetches latest mods from registry.
func (mb *ModBrowser) RefreshModList() error {
	url := mb.registryURL + "/search"
	if mb.searchQuery != "" {
		url += "?name=" + mb.searchQuery
	}

	resp, err := mb.httpClient.Get(url)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch mod list")
		mb.setError("Failed to connect to mod registry")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("registry returned status %d", resp.StatusCode)
		logrus.WithError(err).Error("Mod list fetch failed")
		mb.setError(fmt.Sprintf("Registry error: %d", resp.StatusCode))
		return err
	}

	var result struct {
		Results []registry.ModRecord `json:"results"`
		Count   int                  `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logrus.WithError(err).Error("Failed to parse mod list")
		mb.setError("Failed to parse registry response")
		return err
	}

	mb.mu.Lock()
	mb.mods = result.Results
	mb.selectedIndex = 0
	mb.scrollOffset = 0
	mb.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"system_name": "mod_browser",
		"count":       len(mb.mods),
	}).Info("Mod list refreshed")

	return nil
}

// CheckForUpdates compares installed mods with registry versions.
func (mb *ModBrowser) CheckForUpdates() error {
	mb.mu.Lock()
	installedNames := make([]string, 0, len(mb.installedMods))
	for name := range mb.installedMods {
		installedNames = append(installedNames, name)
	}
	mb.mu.Unlock()

	if len(installedNames) == 0 {
		return nil
	}

	// Fetch latest version for each installed mod
	updates := make(map[string]string)
	for _, name := range installedNames {
		url := fmt.Sprintf("%s/search?name=%s", mb.registryURL, name)
		resp, err := mb.httpClient.Get(url)
		if err != nil {
			continue
		}

		var result struct {
			Results []registry.ModRecord `json:"results"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Find exact match and compare version
		for _, mod := range result.Results {
			if mod.Name == name {
				mb.mu.RLock()
				installedVer := mb.installedMods[name]
				mb.mu.RUnlock()

				if mod.Version != installedVer {
					updates[name] = mod.Version
					logrus.WithFields(logrus.Fields{
						"system_name":       "mod_browser",
						"mod_name":          name,
						"installed_version": installedVer,
						"new_version":       mod.Version,
					}).Info("Update available")
				}
				break
			}
		}
	}

	mb.mu.Lock()
	mb.updateAvailable = updates
	mb.autoUpdateCheck = time.Now()
	mb.mu.Unlock()

	return nil
}

// DownloadMod fetches and verifies a mod from the registry.
func (mb *ModBrowser) DownloadMod(name, version string) ([]byte, string, error) {
	url := fmt.Sprintf("%s/download/%s/%s", mb.registryURL, name, version)

	logrus.WithFields(logrus.Fields{
		"system_name": "mod_browser",
		"mod_name":    name,
		"version":     version,
		"url":         url,
	}).Info("Downloading mod")

	resp, err := mb.httpClient.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	expectedChecksum := resp.Header.Get("X-SHA256")
	if expectedChecksum == "" {
		return nil, "", fmt.Errorf("no checksum in response headers")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read mod data: %w", err)
	}

	return data, expectedChecksum, nil
}

// InstallMod downloads and installs a mod with checksum verification.
func (mb *ModBrowser) InstallMod(name, version string) error {
	mb.mu.Lock()
	if mb.installing {
		mb.mu.Unlock()
		return fmt.Errorf("installation already in progress")
	}
	mb.installing = true
	mb.state = ModBrowserStateInstalling
	mb.installProgress = fmt.Sprintf("Downloading %s v%s...", name, version)
	mb.mu.Unlock()

	defer func() {
		mb.mu.Lock()
		mb.installing = false
		mb.state = ModBrowserStateBrowse
		mb.mu.Unlock()
	}()

	// Download with checksum verification
	data, expectedChecksum, err := mb.DownloadMod(name, version)
	if err != nil {
		mb.setError(fmt.Sprintf("Download failed: %v", err))
		return err
	}

	mb.mu.Lock()
	mb.installProgress = "Verifying checksum..."
	mb.mu.Unlock()

	// Verify checksum
	actualChecksum := mod.ComputeSHA256(data)
	if actualChecksum != expectedChecksum {
		err := fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
		logrus.WithError(err).WithFields(logrus.Fields{
			"system_name": "mod_browser",
			"mod_name":    name,
		}).Error("Checksum verification failed")
		mb.setError("Checksum verification failed!")
		return err
	}

	mb.mu.Lock()
	mb.installProgress = "Installing..."
	mb.mu.Unlock()

	// Load the mod
	if err := mod.LoadWASMModule(data); err != nil {
		mb.setError(fmt.Sprintf("Installation failed: %v", err))
		return err
	}

	// Mark as installed
	mb.mu.Lock()
	mb.installedMods[name] = version
	delete(mb.updateAvailable, name)
	mb.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"system_name": "mod_browser",
		"mod_name":    name,
		"version":     version,
	}).Info("Mod installed successfully")

	return nil
}

// AutoUpdate checks for and installs updates for all installed mods.
func (mb *ModBrowser) AutoUpdate() error {
	mb.mu.Lock()
	mb.state = ModBrowserStateUpdating
	updatesToInstall := make(map[string]string)
	for name, newVersion := range mb.updateAvailable {
		updatesToInstall[name] = newVersion
	}
	mb.mu.Unlock()

	if len(updatesToInstall) == 0 {
		return nil
	}

	// Sort for deterministic order
	names := make([]string, 0, len(updatesToInstall))
	for name := range updatesToInstall {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		version := updatesToInstall[name]
		logrus.WithFields(logrus.Fields{
			"system_name": "mod_browser",
			"mod_name":    name,
			"version":     version,
		}).Info("Auto-updating mod")

		if err := mb.InstallMod(name, version); err != nil {
			logrus.WithError(err).WithField("mod_name", name).Error("Auto-update failed")
			continue
		}
	}

	mb.mu.Lock()
	mb.state = ModBrowserStateBrowse
	mb.mu.Unlock()

	return nil
}

// Update processes input and updates browser state.
func (mb *ModBrowser) Update(im *input.Manager) error {
	if !mb.visible {
		return nil
	}

	// Auto-check for updates every 30 minutes
	mb.mu.RLock()
	timeSinceCheck := time.Since(mb.autoUpdateCheck)
	mb.mu.RUnlock()

	if timeSinceCheck > 30*time.Minute {
		go mb.CheckForUpdates()
	}

	// Input handling is managed externally via navigation methods
	return nil
}

// NavigateDown moves selection down in the list.
func (mb *ModBrowser) NavigateDown() {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.state == ModBrowserStateBrowse && mb.selectedIndex < len(mb.mods)-1 {
		mb.selectedIndex++
		if mb.selectedIndex >= mb.scrollOffset+10 {
			mb.scrollOffset = mb.selectedIndex - 9
		}
	}
}

// NavigateUp moves selection up in the list.
func (mb *ModBrowser) NavigateUp() {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.state == ModBrowserStateBrowse && mb.selectedIndex > 0 {
		mb.selectedIndex--
		if mb.selectedIndex < mb.scrollOffset {
			mb.scrollOffset = mb.selectedIndex
		}
	}
}

// Confirm performs the action for the selected item.
func (mb *ModBrowser) Confirm() {
	mb.mu.Lock()
	state := mb.state
	selectedIndex := mb.selectedIndex
	modsLen := len(mb.mods)
	mb.mu.Unlock()

	switch state {
	case ModBrowserStateBrowse:
		if selectedIndex < modsLen {
			mb.mu.Lock()
			mb.state = ModBrowserStateDetails
			mb.mu.Unlock()
		}
	case ModBrowserStateDetails:
		if selectedIndex < modsLen {
			mb.mu.RLock()
			selectedMod := mb.mods[selectedIndex]
			mb.mu.RUnlock()
			go mb.InstallMod(selectedMod.Name, selectedMod.Version)
		}
	}
}

// Cancel goes back or closes the browser.
func (mb *ModBrowser) Cancel() {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	switch mb.state {
	case ModBrowserStateBrowse:
		mb.visible = false
	case ModBrowserStateDetails:
		mb.state = ModBrowserStateBrowse
	}
}

// Refresh triggers a mod list refresh.
func (mb *ModBrowser) Refresh() {
	go mb.RefreshModList()
}

// Draw renders the mod browser UI.
func (mb *ModBrowser) Draw(screen *ebiten.Image) {
	if !mb.visible {
		return
	}

	mb.mu.RLock()
	defer mb.mu.RUnlock()

	// Semi-transparent background
	vector.DrawFilledRect(screen, 50, 50, 700, 500, color.RGBA{0, 0, 0, 200}, false)
	vector.StrokeRect(screen, 50, 50, 700, 500, 2, color.RGBA{100, 100, 100, 255}, false)

	// Title
	title := "Mod Browser"
	text.Draw(screen, title, basicfont.Face7x13, 60, 70, color.White)

	// Update count display
	if len(mb.updateAvailable) > 0 {
		updateText := fmt.Sprintf("(%d updates available)", len(mb.updateAvailable))
		text.Draw(screen, updateText, basicfont.Face7x13, 200, 70, color.RGBA{255, 200, 0, 255})
	}

	// Error message
	if mb.errorMessage != "" && time.Since(mb.errorTime) < 5*time.Second {
		text.Draw(screen, mb.errorMessage, basicfont.Face7x13, 60, 90, color.RGBA{255, 0, 0, 255})
	}

	switch mb.state {
	case ModBrowserStateBrowse:
		mb.drawBrowse(screen)
	case ModBrowserStateDetails:
		mb.drawDetails(screen)
	case ModBrowserStateInstalling, ModBrowserStateUpdating:
		mb.drawProgress(screen)
	}
}

func (mb *ModBrowser) drawBrowse(screen *ebiten.Image) {
	y := 110
	visibleCount := 10
	for i := mb.scrollOffset; i < mb.scrollOffset+visibleCount && i < len(mb.mods); i++ {
		mod := mb.mods[i]

		fg := color.RGBA{255, 255, 255, 255}
		if i == mb.selectedIndex {
			// Highlight selected
			vector.DrawFilledRect(screen, 60, float32(y-10), 680, 20, color.RGBA{50, 50, 150, 255}, false)
			fg = color.RGBA{255, 255, 0, 255}
		}

		// Mod name and version
		line := fmt.Sprintf("%s v%s - %s", mod.Name, mod.Version, mod.Author)
		text.Draw(screen, line, basicfont.Face7x13, 65, y, fg)

		// Update indicator
		if newVer, hasUpdate := mb.updateAvailable[mod.Name]; hasUpdate {
			updateLabel := fmt.Sprintf(" [UPDATE: v%s]", newVer)
			text.Draw(screen, updateLabel, basicfont.Face7x13, 400, y, color.RGBA{0, 255, 0, 255})
		}

		// Installed indicator
		if installedVer, installed := mb.installedMods[mod.Name]; installed && installedVer == mod.Version {
			text.Draw(screen, " [INSTALLED]", basicfont.Face7x13, 550, y, color.RGBA{100, 255, 100, 255})
		}

		y += 22
	}

	// Controls hint
	controls := "Up/Down: Navigate | Enter: Details | R: Refresh | Esc: Close"
	text.Draw(screen, controls, basicfont.Face7x13, 60, 530, color.RGBA{150, 150, 150, 255})
}

func (mb *ModBrowser) drawDetails(screen *ebiten.Image) {
	if mb.selectedIndex >= len(mb.mods) {
		return
	}

	mod := mb.mods[mb.selectedIndex]
	y := 110

	// Details
	details := []string{
		fmt.Sprintf("Name: %s", mod.Name),
		fmt.Sprintf("Version: %s", mod.Version),
		fmt.Sprintf("Author: %s", mod.Author),
		fmt.Sprintf("Size: %.2f MB", float64(mod.Size)/1024/1024),
		fmt.Sprintf("Downloads: %d", mod.Downloads),
		fmt.Sprintf("Description: %s", mod.Description),
	}

	for _, line := range details {
		text.Draw(screen, line, basicfont.Face7x13, 65, y, color.White)
		y += 20
	}

	// Install button hint
	installHint := "Enter: Install | Esc: Back"
	if installedVer, installed := mb.installedMods[mod.Name]; installed && installedVer == mod.Version {
		installHint = "Already installed | Esc: Back"
	}
	text.Draw(screen, installHint, basicfont.Face7x13, 60, 530, color.RGBA{150, 150, 150, 255})
}

func (mb *ModBrowser) drawProgress(screen *ebiten.Image) {
	text.Draw(screen, mb.installProgress, basicfont.Face7x13, 65, 300, color.White)

	// Simple progress animation
	dots := "..."
	text.Draw(screen, dots, basicfont.Face7x13, 300, 300, color.RGBA{100, 100, 100, 255})
}

func (mb *ModBrowser) setError(msg string) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.errorMessage = msg
	mb.errorTime = time.Now()
}

// SetInstalledMods updates the list of installed mods.
func (mb *ModBrowser) SetInstalledMods(installed map[string]string) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.installedMods = installed
}

// GetUpdateCount returns the number of available updates.
func (mb *ModBrowser) GetUpdateCount() int {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return len(mb.updateAvailable)
}
