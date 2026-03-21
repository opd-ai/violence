package statusbar

import (
	"image"
	"image/color"
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/status"
	"github.com/sirupsen/logrus"
)

// System manages status bar rendering and updates.
type System struct {
	genreID string
	logger  *logrus.Entry

	// Genre-specific colors
	buffColor   color.RGBA
	debuffColor color.RGBA
	bgColor     color.RGBA
	borderColor color.RGBA

	// Icon cache to avoid per-frame allocations
	iconCache     map[string]*ebiten.Image
	iconCacheMu   sync.RWMutex
	maxCacheSize  int
	cacheAccess   map[string]int64 // LRU tracking
	accessCounter int64
}

// NewSystem creates a status bar system for the given genre.
func NewSystem(genreID string) *System {
	s := &System{
		genreID:      genreID,
		iconCache:    make(map[string]*ebiten.Image),
		maxCacheSize: 32,
		cacheAccess:  make(map[string]int64),
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "statusbar",
			"genre":       genreID,
		}),
	}
	s.SetGenre(genreID)
	return s
}

// SetGenre updates genre-specific visual settings.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID

	switch genreID {
	case "fantasy":
		s.buffColor = color.RGBA{255, 215, 0, 255}   // Gold
		s.debuffColor = color.RGBA{128, 0, 128, 255} // Purple
		s.bgColor = color.RGBA{20, 15, 10, 200}
		s.borderColor = color.RGBA{139, 119, 101, 255}
	case "scifi":
		s.buffColor = color.RGBA{0, 255, 255, 255}   // Cyan
		s.debuffColor = color.RGBA{255, 140, 0, 255} // Orange
		s.bgColor = color.RGBA{10, 20, 30, 200}
		s.borderColor = color.RGBA{100, 150, 200, 255}
	case "horror":
		s.buffColor = color.RGBA{144, 238, 144, 255} // Sickly green
		s.debuffColor = color.RGBA{139, 0, 0, 255}   // Blood red
		s.bgColor = color.RGBA{15, 10, 10, 220}
		s.borderColor = color.RGBA{80, 60, 60, 255}
	case "cyberpunk":
		s.buffColor = color.RGBA{255, 0, 255, 255}   // Neon pink
		s.debuffColor = color.RGBA{0, 255, 128, 255} // Toxic green
		s.bgColor = color.RGBA{10, 5, 20, 200}
		s.borderColor = color.RGBA{200, 0, 200, 255}
	case "postapoc":
		s.buffColor = color.RGBA{210, 105, 30, 255}    // Rust orange
		s.debuffColor = color.RGBA{189, 183, 107, 255} // Khaki/sickly yellow
		s.bgColor = color.RGBA{25, 20, 15, 200}
		s.borderColor = color.RGBA{139, 90, 43, 255}
	default:
		// Default to fantasy
		s.buffColor = color.RGBA{255, 215, 0, 255}
		s.debuffColor = color.RGBA{128, 0, 128, 255}
		s.bgColor = color.RGBA{20, 15, 10, 200}
		s.borderColor = color.RGBA{139, 119, 101, 255}
	}

	// Clear icon cache on genre change
	s.iconCacheMu.Lock()
	s.iconCache = make(map[string]*ebiten.Image)
	s.cacheAccess = make(map[string]int64)
	s.iconCacheMu.Unlock()
}

// Update synchronizes the status bar component with the entity's active status effects.
func (s *System) Update(w *engine.World) {
	statusBarType := reflect.TypeOf((*Component)(nil))
	statusCompType := reflect.TypeOf((*status.StatusComponent)(nil))

	entities := w.Query(statusBarType)

	for _, ent := range entities {
		barComp, found := w.GetComponent(ent, statusBarType)
		if !found {
			continue
		}
		bar, ok := barComp.(*Component)
		if !ok {
			continue
		}

		// Get status component from same entity
		statusComp, found := w.GetComponent(ent, statusCompType)
		if !found {
			bar.ClearIcons()
			continue
		}
		sc, ok := statusComp.(*status.StatusComponent)
		if !ok {
			bar.ClearIcons()
			continue
		}

		// Update icons from active effects
		s.updateIconsFromEffects(bar, sc)
	}
}

// updateIconsFromEffects syncs icon states with active status effects.
func (s *System) updateIconsFromEffects(bar *Component, sc *status.StatusComponent) {
	bar.ClearIcons()

	for _, effect := range sc.ActiveEffects {
		iconType := s.effectTypeToIconType(effect.EffectName)
		effectColor := s.uint32ToColor(effect.VisualColor)

		// Determine if expiring
		isExpiring := effect.TimeRemaining < 3*time.Second

		icon := IconState{
			EffectName:        effect.EffectName,
			DisplayName:       formatEffectName(effect.EffectName),
			IconType:          iconType,
			Color:             effectColor,
			DurationRemaining: effect.TimeRemaining,
			TotalDuration:     effect.TimeRemaining + effect.TickInterval, // Approximate
			StackCount:        1,
			IsExpiring:        isExpiring,
		}

		// Check for existing icon to stack
		existingIdx := -1
		for i, existing := range bar.Icons {
			if existing.EffectName == effect.EffectName {
				existingIdx = i
				break
			}
		}

		if existingIdx >= 0 {
			bar.Icons[existingIdx].StackCount++
			// Use longest remaining duration
			if effect.TimeRemaining > bar.Icons[existingIdx].DurationRemaining {
				bar.Icons[existingIdx].DurationRemaining = effect.TimeRemaining
			}
		} else {
			bar.AddIcon(icon)
		}
	}
}

// effectTypeToIconType maps effect names to icon types.
func (s *System) effectTypeToIconType(effectName string) IconType {
	switch effectName {
	case "burning", "bleeding", "poisoned", "irradiated", "infected", "corroded":
		return IconDamage
	case "regeneration", "nanoheal":
		return IconHeal
	case "blessed", "overcharged":
		return IconBuff
	case "cursed", "terrified":
		return IconDebuff
	case "stunned", "emp_stunned":
		return IconStun
	case "slowed":
		return IconSlow
	default:
		return IconDebuff // Default to debuff for unknown effects
	}
}

// uint32ToColor converts a packed RGBA uint32 to color.RGBA.
func (s *System) uint32ToColor(c uint32) color.RGBA {
	return color.RGBA{
		R: uint8((c >> 24) & 0xFF),
		G: uint8((c >> 16) & 0xFF),
		B: uint8((c >> 8) & 0xFF),
		A: uint8(c & 0xFF),
	}
}

// formatEffectName converts internal effect names to display names.
func formatEffectName(name string) string {
	switch name {
	case "burning":
		return "Burning"
	case "bleeding":
		return "Bleeding"
	case "poisoned":
		return "Poisoned"
	case "stunned":
		return "Stunned"
	case "slowed":
		return "Slowed"
	case "regeneration":
		return "Regen"
	case "blessed":
		return "Blessed"
	case "cursed":
		return "Cursed"
	case "irradiated":
		return "Irradiated"
	case "emp_stunned":
		return "EMP"
	case "nanoheal":
		return "Nanoheal"
	case "overcharged":
		return "Overcharge"
	case "corroded":
		return "Corroded"
	case "terrified":
		return "Terrified"
	case "infected":
		return "Infected"
	default:
		return name
	}
}

// Render draws the status bar to the screen.
func (s *System) Render(screen *ebiten.Image, w *engine.World, playerEntity engine.Entity) {
	statusBarType := reflect.TypeOf((*Component)(nil))

	barComp, found := w.GetComponent(playerEntity, statusBarType)
	if !found {
		return
	}
	bar, ok := barComp.(*Component)
	if !ok {
		return
	}

	if !bar.Visible || len(bar.Icons) == 0 {
		return
	}

	// Render each icon
	for i, icon := range bar.Icons {
		x := bar.X + float32(i*(bar.IconSize+bar.IconSpacing))
		y := bar.Y

		s.renderIcon(screen, x, y, float32(bar.IconSize), &icon)
	}
}

// renderIcon draws a single status effect icon.
func (s *System) renderIcon(screen *ebiten.Image, x, y, size float32, icon *IconState) {
	// Background
	vector.DrawFilledRect(screen, x, y, size, size, s.bgColor, false)

	// Radial cooldown indicator (if duration known)
	if icon.TotalDuration > 0 && icon.DurationRemaining > 0 {
		progress := float64(icon.DurationRemaining) / float64(icon.TotalDuration)
		s.drawRadialProgress(screen, x, y, size, progress, icon.Color)
	}

	// Icon symbol
	centerX := x + size/2
	centerY := y + size/2
	s.drawIconSymbol(screen, centerX, centerY, size*0.6, icon.IconType, icon.Color)

	// Stack count (if > 1)
	if icon.StackCount > 1 {
		s.drawStackCount(screen, x+size-4, y+size-4, icon.StackCount)
	}

	// Expiring pulse effect
	if icon.IsExpiring {
		pulseAlpha := uint8(128 + 64*math.Sin(float64(time.Now().UnixMilli())/200.0))
		pulseColor := color.RGBA{255, 255, 255, pulseAlpha}
		vector.StrokeRect(screen, x, y, size, size, 2, pulseColor, false)
	} else {
		// Normal border
		vector.StrokeRect(screen, x, y, size, size, 1, s.borderColor, false)
	}
}

// drawRadialProgress draws a pie-chart style progress indicator.
func (s *System) drawRadialProgress(screen *ebiten.Image, x, y, size float32, progress float64, c color.RGBA) {
	if progress <= 0 || progress > 1 {
		return
	}

	centerX := x + size/2
	centerY := y + size/2
	radius := size / 2.5

	// Draw filled arc using small line segments
	segments := 16
	angleEnd := -math.Pi/2 + 2*math.Pi*progress
	angleStart := -math.Pi / 2

	// Dim the color for the progress fill
	fillColor := color.RGBA{c.R / 2, c.G / 2, c.B / 2, 128}

	// Draw triangles to fill the arc
	for i := 0; i < segments; i++ {
		segStart := angleStart + float64(i)*2*math.Pi/float64(segments)
		segEnd := angleStart + float64(i+1)*2*math.Pi/float64(segments)

		// Only draw segments within the progress range
		if segStart > angleEnd {
			break
		}
		if segEnd > angleEnd {
			segEnd = angleEnd
		}

		// Calculate vertices
		x1 := centerX + float32(math.Cos(segStart))*radius
		y1 := centerY + float32(math.Sin(segStart))*radius
		x2 := centerX + float32(math.Cos(segEnd))*radius
		y2 := centerY + float32(math.Sin(segEnd))*radius

		// Draw triangle fan segment
		vector.StrokeLine(screen, centerX, centerY, x1, y1, 2, fillColor, false)
		vector.StrokeLine(screen, x1, y1, x2, y2, 2, fillColor, false)
	}
}

// drawIconSymbol draws the status effect symbol.
func (s *System) drawIconSymbol(screen *ebiten.Image, cx, cy, size float32, iconType IconType, c color.RGBA) {
	halfSize := size / 2

	switch iconType {
	case IconDamage:
		// Downward arrow (damage)
		vector.StrokeLine(screen, cx, cy-halfSize, cx, cy+halfSize, 2, c, false)
		vector.StrokeLine(screen, cx-halfSize*0.6, cy+halfSize*0.4, cx, cy+halfSize, 2, c, false)
		vector.StrokeLine(screen, cx+halfSize*0.6, cy+halfSize*0.4, cx, cy+halfSize, 2, c, false)

	case IconHeal:
		// Upward arrow (heal)
		vector.StrokeLine(screen, cx, cy+halfSize, cx, cy-halfSize, 2, c, false)
		vector.StrokeLine(screen, cx-halfSize*0.6, cy-halfSize*0.4, cx, cy-halfSize, 2, c, false)
		vector.StrokeLine(screen, cx+halfSize*0.6, cy-halfSize*0.4, cx, cy-halfSize, 2, c, false)

	case IconBuff:
		// Upward chevron (buff)
		vector.StrokeLine(screen, cx-halfSize, cy+halfSize*0.5, cx, cy-halfSize*0.5, 2, c, false)
		vector.StrokeLine(screen, cx, cy-halfSize*0.5, cx+halfSize, cy+halfSize*0.5, 2, c, false)
		vector.StrokeLine(screen, cx-halfSize*0.6, cy+halfSize, cx, cy, 2, c, false)
		vector.StrokeLine(screen, cx, cy, cx+halfSize*0.6, cy+halfSize, 2, c, false)

	case IconDebuff:
		// Downward chevron (debuff)
		vector.StrokeLine(screen, cx-halfSize, cy-halfSize*0.5, cx, cy+halfSize*0.5, 2, c, false)
		vector.StrokeLine(screen, cx, cy+halfSize*0.5, cx+halfSize, cy-halfSize*0.5, 2, c, false)
		vector.StrokeLine(screen, cx-halfSize*0.6, cy-halfSize, cx, cy, 2, c, false)
		vector.StrokeLine(screen, cx, cy, cx+halfSize*0.6, cy-halfSize, 2, c, false)

	case IconStun:
		// Star/spiral (stun)
		for i := 0; i < 6; i++ {
			angle := float64(i) * math.Pi / 3
			x1 := cx + float32(math.Cos(angle))*halfSize*0.4
			y1 := cy + float32(math.Sin(angle))*halfSize*0.4
			x2 := cx + float32(math.Cos(angle))*halfSize
			y2 := cy + float32(math.Sin(angle))*halfSize
			vector.StrokeLine(screen, x1, y1, x2, y2, 1.5, c, false)
		}

	case IconSlow:
		// Horizontal lines (slow)
		vector.StrokeLine(screen, cx-halfSize, cy-halfSize*0.4, cx+halfSize, cy-halfSize*0.4, 2, c, false)
		vector.StrokeLine(screen, cx-halfSize*0.7, cy, cx+halfSize*0.7, cy, 2, c, false)
		vector.StrokeLine(screen, cx-halfSize, cy+halfSize*0.4, cx+halfSize, cy+halfSize*0.4, 2, c, false)
	}
}

// drawStackCount draws a small number for stacked effects.
func (s *System) drawStackCount(screen *ebiten.Image, x, y float32, count int) {
	if count < 2 || count > 9 {
		return
	}

	// Background circle
	vector.DrawFilledCircle(screen, x, y, 4, color.RGBA{0, 0, 0, 200}, false)

	// Number using simple pixel art (for counts 2-9)
	digitColor := color.RGBA{255, 255, 255, 255}
	s.drawDigit(screen, x-2, y-3, count, digitColor)
}

// drawDigit draws a single digit (2-9) as a tiny 3x5 bitmap.
func (s *System) drawDigit(screen *ebiten.Image, x, y float32, digit int, c color.RGBA) {
	// Simplified digit rendering - just draw small rectangles
	// This is a minimal implementation for stack counts
	if digit < 2 || digit > 9 {
		return
	}

	// Draw a simple representation
	vector.DrawFilledRect(screen, x, y, 4, 5, c, false)
}

// GetIcon returns a cached or generated icon image.
func (s *System) GetIcon(iconType IconType, effectColor color.RGBA) *ebiten.Image {
	key := s.iconCacheKey(iconType, effectColor)

	s.iconCacheMu.RLock()
	if img, exists := s.iconCache[key]; exists {
		s.iconCacheMu.RUnlock()
		s.iconCacheMu.Lock()
		s.accessCounter++
		s.cacheAccess[key] = s.accessCounter
		s.iconCacheMu.Unlock()
		return img
	}
	s.iconCacheMu.RUnlock()

	// Generate new icon
	img := s.generateIcon(iconType, effectColor)

	s.iconCacheMu.Lock()
	defer s.iconCacheMu.Unlock()

	// Evict if at capacity
	if len(s.iconCache) >= s.maxCacheSize {
		s.evictLRU()
	}

	s.accessCounter++
	s.iconCache[key] = img
	s.cacheAccess[key] = s.accessCounter

	return img
}

// iconCacheKey generates a unique key for icon caching.
func (s *System) iconCacheKey(iconType IconType, c color.RGBA) string {
	return string(rune('A'+iconType)) + string([]byte{c.R, c.G, c.B, c.A})
}

// generateIcon creates a procedural icon image.
func (s *System) generateIcon(iconType IconType, c color.RGBA) *ebiten.Image {
	const size = 16
	img := ebiten.NewImage(size, size)

	// Fill background
	img.Fill(s.bgColor)

	// Draw symbol centered
	s.drawIconSymbol(img, size/2, size/2, 10, iconType, c)

	return img
}

// evictLRU removes the least recently used cache entry.
func (s *System) evictLRU() {
	var oldestKey string
	var oldestAccess int64 = math.MaxInt64

	for key, access := range s.cacheAccess {
		if access < oldestAccess {
			oldestAccess = access
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(s.iconCache, oldestKey)
		delete(s.cacheAccess, oldestKey)
	}
}

// UpdatePlayerStatusBar is a convenience method to update a specific player's status bar.
func (s *System) UpdatePlayerStatusBar(w *engine.World, playerEntity engine.Entity) {
	statusBarType := reflect.TypeOf((*Component)(nil))
	statusCompType := reflect.TypeOf((*status.StatusComponent)(nil))

	barComp, found := w.GetComponent(playerEntity, statusBarType)
	if !found {
		return
	}
	bar, ok := barComp.(*Component)
	if !ok {
		return
	}

	statusComp, found := w.GetComponent(playerEntity, statusCompType)
	if !found {
		bar.ClearIcons()
		return
	}
	sc, ok := statusComp.(*status.StatusComponent)
	if !ok {
		bar.ClearIcons()
		return
	}

	s.updateIconsFromEffects(bar, sc)
}

// RenderDirect renders the status bar without querying ECS (for standalone use).
func (s *System) RenderDirect(screen *ebiten.Image, bar *Component) {
	if bar == nil || !bar.Visible || len(bar.Icons) == 0 {
		return
	}

	for i, icon := range bar.Icons {
		x := bar.X + float32(i*(bar.IconSize+bar.IconSpacing))
		y := bar.Y

		s.renderIcon(screen, x, y, float32(bar.IconSize), &icon)
	}
}

// GetBounds returns the screen rectangle occupied by the status bar.
func (s *System) GetBounds(bar *Component) image.Rectangle {
	if bar == nil || len(bar.Icons) == 0 {
		return image.Rectangle{}
	}

	width := len(bar.Icons)*(bar.IconSize+bar.IconSpacing) - bar.IconSpacing
	return image.Rect(int(bar.X), int(bar.Y), int(bar.X)+width, int(bar.Y)+bar.IconSize)
}
