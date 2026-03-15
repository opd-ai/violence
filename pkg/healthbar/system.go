package healthbar

import (
	"image"
	"image/color"
	"math"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/ui"
	"github.com/sirupsen/logrus"
)

// System manages health bar rendering for entities.
type System struct {
	genre           string
	baseColor       color.RGBA
	damageColor     color.RGBA
	criticalColor   color.RGBA
	backgroundColor color.RGBA
	borderColor     color.RGBA
	iconSize        float32
	fadeDelay       float64
	logger          *logrus.Entry
	statusIconCache map[StatusIconType]*ebiten.Image
	barImagePool    []*ebiten.Image
	poolIndex       int
	maxPoolSize     int
}

// NewSystem creates a health bar rendering system.
func NewSystem(genre string) *System {
	sys := &System{
		genre:           genre,
		baseColor:       color.RGBA{0, 255, 0, 255},
		damageColor:     color.RGBA{255, 255, 0, 255},
		criticalColor:   color.RGBA{255, 0, 0, 255},
		backgroundColor: color.RGBA{40, 40, 40, 200},
		borderColor:     color.RGBA{200, 200, 200, 255},
		iconSize:        12,
		fadeDelay:       3.0,
		statusIconCache: make(map[StatusIconType]*ebiten.Image),
		barImagePool:    make([]*ebiten.Image, 0, 50),
		maxPoolSize:     50,
		logger: logrus.WithFields(logrus.Fields{
			"system":  "healthbar",
			"package": "healthbar",
		}),
	}
	sys.applyGenreTheme()
	return sys
}

// SetGenre updates genre-specific theming.
func (s *System) SetGenre(genre string) {
	s.genre = genre
	s.applyGenreTheme()
	s.statusIconCache = make(map[StatusIconType]*ebiten.Image)
}

// applyGenreTheme sets colors based on game genre.
func (s *System) applyGenreTheme() {
	switch s.genre {
	case "cyberpunk":
		s.baseColor = color.RGBA{0, 255, 255, 255}
		s.damageColor = color.RGBA{255, 0, 255, 255}
		s.criticalColor = color.RGBA{255, 0, 100, 255}
		s.borderColor = color.RGBA{0, 255, 255, 255}
	case "horror":
		s.baseColor = color.RGBA{100, 255, 100, 255}
		s.damageColor = color.RGBA{200, 200, 0, 255}
		s.criticalColor = color.RGBA{200, 0, 0, 255}
		s.borderColor = color.RGBA{150, 150, 150, 255}
		s.backgroundColor = color.RGBA{20, 20, 20, 220}
	case "scifi":
		s.baseColor = color.RGBA{0, 200, 255, 255}
		s.damageColor = color.RGBA{255, 200, 0, 255}
		s.criticalColor = color.RGBA{255, 50, 0, 255}
		s.borderColor = color.RGBA{100, 200, 255, 255}
	case "postapoc":
		s.baseColor = color.RGBA{150, 200, 50, 255}
		s.damageColor = color.RGBA{200, 150, 0, 255}
		s.criticalColor = color.RGBA{180, 50, 0, 255}
		s.borderColor = color.RGBA{120, 120, 100, 255}
		s.backgroundColor = color.RGBA{50, 45, 40, 200}
	default: // fantasy
		s.baseColor = color.RGBA{0, 255, 0, 255}
		s.damageColor = color.RGBA{255, 255, 0, 255}
		s.criticalColor = color.RGBA{255, 0, 0, 255}
		s.borderColor = color.RGBA{200, 200, 200, 255}
	}
}

// Update processes all entities with health bars.
func (s *System) Update(w *engine.World) {
	deltaTime := 1.0 / 60.0

	healthType := reflect.TypeOf(&engine.Health{})
	barType := reflect.TypeOf(&Component{})
	statusIconType := reflect.TypeOf(&StatusIconsComponent{})

	entities := w.Query(healthType)
	for _, eid := range entities {
		barComp, hasBar := w.GetComponent(eid, barType)
		if hasBar {
			bar := barComp.(*Component)
			if bar.LastDamageAge < s.fadeDelay*2 {
				bar.LastDamageAge += deltaTime
			}
		}

		statusIconComp, hasIcons := w.GetComponent(eid, statusIconType)
		if hasIcons {
			icons := statusIconComp.(*StatusIconsComponent)
			icons.UpdateDurations(deltaTime)
		}
	}
}

// RenderHealthBars draws health bars for visible entities.
func (s *System) RenderHealthBars(screen *ebiten.Image, w *engine.World, cameraX, cameraY, cameraDirX, cameraDirY float64, screenWidth, screenHeight int) {
	healthType := reflect.TypeOf(&engine.Health{})
	barType := reflect.TypeOf(&Component{})
	posType := reflect.TypeOf(&engine.Position{})

	entities := w.Query(healthType)

	for _, eid := range entities {
		healthComp, ok := w.GetComponent(eid, healthType)
		if !ok {
			continue
		}
		health := healthComp.(*engine.Health)

		barComp, hasBar := w.GetComponent(eid, barType)
		if !hasBar {
			continue
		}
		bar := barComp.(*Component)

		if !bar.Visible {
			continue
		}

		if health.Max == 0 {
			continue
		}

		healthPct := float64(health.Current) / float64(health.Max)
		if healthPct >= 1.0 && !bar.ShowWhenFull && bar.LastDamageAge > s.fadeDelay {
			continue
		}

		posComp, hasPos := w.GetComponent(eid, posType)
		if !hasPos {
			continue
		}
		pos := posComp.(*engine.Position)

		screenX, screenY, visible := s.worldToScreen(pos.X, pos.Y, cameraX, cameraY, cameraDirX, cameraDirY, screenWidth, screenHeight)
		if !visible {
			continue
		}

		s.drawHealthBar(screen, screenX, screenY-bar.OffsetY, bar.Width, bar.Height, healthPct, bar)

		s.drawStatusIcons(screen, w, eid, screenX, screenY-bar.OffsetY-bar.Height-4)
	}
}

// worldToScreen converts world coordinates to screen space.
func (s *System) worldToScreen(worldX, worldY, camX, camY, camDirX, camDirY float64, screenWidth, screenHeight int) (float32, float32, bool) {
	relX := worldX - camX
	relY := worldY - camY

	perpX := -camDirY
	perpY := camDirX

	viewDist := relX*camDirX + relY*camDirY
	if viewDist < 0.5 || viewDist > 20.0 {
		return 0, 0, false
	}

	viewPerp := relX*perpX + relY*perpY

	screenX := float32(screenWidth)/2 + float32(viewPerp/viewDist*float64(screenWidth)/2)
	screenY := float32(screenHeight) / 2

	if screenX < -50 || screenX > float32(screenWidth)+50 {
		return 0, 0, false
	}

	return screenX, screenY, true
}

// drawHealthBar renders a single health bar.
func (s *System) drawHealthBar(screen *ebiten.Image, x, y, width, height float32, healthPct float64, bar *Component) {
	alpha := uint8(255)
	if bar.LastDamageAge > s.fadeDelay {
		fadePct := (bar.LastDamageAge - s.fadeDelay) / s.fadeDelay
		if fadePct > 1.0 {
			fadePct = 1.0
		}
		alpha = uint8(255.0 * (1.0 - fadePct))
	}

	bg := s.backgroundColor
	bg.A = alpha
	border := s.borderColor
	border.A = alpha

	vector.DrawFilledRect(screen, x-1, y-1, width+2, height+2, border, false)
	vector.DrawFilledRect(screen, x, y, width, height, bg, false)

	fillWidth := float32(healthPct) * width
	if fillWidth < 0 {
		fillWidth = 0
	}
	if fillWidth > width {
		fillWidth = width
	}

	fillColor := s.getHealthColor(healthPct, bar)
	fillColor.A = alpha

	if fillWidth > 0 {
		vector.DrawFilledRect(screen, x, y, fillWidth, height, fillColor, false)
	}
}

// getHealthColor determines bar color based on health percentage.
func (s *System) getHealthColor(healthPct float64, bar *Component) color.RGBA {
	if bar.CustomColor != nil {
		return *bar.CustomColor
	}

	if healthPct < 0.25 {
		return s.criticalColor
	} else if healthPct < 0.5 {
		return s.damageColor
	}
	return s.baseColor
}

// drawStatusIcons renders status effect icons above the health bar.
func (s *System) drawStatusIcons(screen *ebiten.Image, w *engine.World, eid engine.Entity, x, y float32) {
	statusType := reflect.TypeOf(&StatusIconsComponent{})
	comp, ok := w.GetComponent(eid, statusType)
	if !ok {
		return
	}

	statusComp := comp.(*StatusIconsComponent)
	if len(statusComp.Icons) == 0 {
		return
	}

	iconSpacing := s.iconSize + 2
	totalWidth := float32(len(statusComp.Icons))*iconSpacing - 2
	startX := x - totalWidth/2

	for i, icon := range statusComp.Icons {
		iconX := startX + float32(i)*iconSpacing
		s.drawStatusIcon(screen, iconX, y, icon)
	}
}

// drawStatusIcon renders a single status effect icon.
func (s *System) drawStatusIcon(screen *ebiten.Image, x, y float32, icon StatusIcon) {
	img := s.getStatusIconImage(icon.Type)
	if img == nil {
		return
	}

	opts := &ebiten.DrawImageOptions{}
	bounds := img.Bounds()
	scale := float64(s.iconSize) / float64(bounds.Dx())
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(float64(x), float64(y))

	if icon.Stacks > 1 {
		opts.ColorScale.ScaleAlpha(1.0)
	} else if icon.Duration < 2.0 {
		pulse := math.Sin(icon.Duration * math.Pi * 4)
		opts.ColorScale.ScaleAlpha(0.6 + 0.4*float32(pulse))
	}

	screen.DrawImage(img, opts)
}

// getStatusIconImage retrieves or generates a status icon.
func (s *System) getStatusIconImage(iconType StatusIconType) *ebiten.Image {
	if cached, exists := s.statusIconCache[iconType]; exists {
		return cached
	}

	size := int(s.iconSize)
	img := ebiten.NewImage(size, size)

	iconColor := s.getIconColor(iconType)
	s.drawIconShape(img, iconType, iconColor, size)

	s.statusIconCache[iconType] = img
	return img
}

// getIconColor returns the color for a status icon type.
func (s *System) getIconColor(iconType StatusIconType) color.RGBA {
	switch iconType {
	case IconPoison:
		return color.RGBA{100, 255, 100, 255}
	case IconBurn:
		return color.RGBA{255, 100, 0, 255}
	case IconFreeze:
		return color.RGBA{100, 200, 255, 255}
	case IconStun:
		return color.RGBA{255, 255, 100, 255}
	case IconBleed:
		return color.RGBA{200, 0, 0, 255}
	case IconRegen:
		return color.RGBA{0, 255, 150, 255}
	case IconShield:
		return color.RGBA{150, 150, 255, 255}
	case IconHaste:
		return color.RGBA{255, 200, 100, 255}
	case IconSlow:
		return color.RGBA{100, 100, 150, 255}
	case IconWeak:
		return color.RGBA{150, 100, 100, 255}
	case IconBerserk:
		return color.RGBA{255, 0, 100, 255}
	case IconInvisible:
		return color.RGBA{200, 200, 255, 128}
	default:
		return color.RGBA{200, 200, 200, 255}
	}
}

// drawIconShape draws the icon shape based on type.
func (s *System) drawIconShape(img *ebiten.Image, iconType StatusIconType, col color.RGBA, size int) {
	center := float32(size) / 2

	switch iconType {
	case IconPoison:
		s.drawDroplet(img, center, center, float32(size)*0.4, col)
	case IconBurn:
		s.drawFlame(img, center, center, float32(size)*0.4, col)
	case IconFreeze:
		s.drawSnowflake(img, center, center, float32(size)*0.4, col)
	case IconStun:
		s.drawStar(img, center, center, float32(size)*0.4, col)
	case IconBleed:
		s.drawDroplet(img, center, center, float32(size)*0.4, col)
	case IconRegen:
		s.drawPlus(img, center, center, float32(size)*0.35, col)
	case IconShield:
		s.drawShield(img, center, center, float32(size)*0.4, col)
	case IconHaste:
		s.drawArrow(img, center, center, float32(size)*0.4, col, 1)
	case IconSlow:
		s.drawArrow(img, center, center, float32(size)*0.4, col, -1)
	case IconWeak:
		s.drawCross(img, center, center, float32(size)*0.35, col)
	case IconBerserk:
		s.drawSword(img, center, center, float32(size)*0.4, col)
	case IconInvisible:
		s.drawEye(img, center, center, float32(size)*0.4, col)
	default:
		vector.DrawFilledCircle(img, center, center, float32(size)*0.4, col, false)
	}
}

// Icon shape drawing helpers
func (s *System) drawDroplet(img *ebiten.Image, cx, cy, r float32, col color.RGBA) {
	vector.DrawFilledCircle(img, cx, cy+r*0.2, r*0.6, col, false)
	vector.DrawFilledCircle(img, cx, cy-r*0.3, r*0.3, col, false)
}

func (s *System) drawFlame(img *ebiten.Image, cx, cy, r float32, col color.RGBA) {
	points := []float32{
		cx, cy - r,
		cx + r*0.4, cy + r*0.5,
		cx, cy,
		cx - r*0.4, cy + r*0.5,
	}
	vector.DrawFilledRect(img, cx-r*0.15, cy-r*0.3, r*0.3, r*1.3, col, false)
	for i := 0; i < len(points)-2; i += 2 {
		vector.StrokeLine(img, points[i], points[i+1], points[i+2], points[i+3], 2, col, false)
	}
}

func (s *System) drawSnowflake(img *ebiten.Image, cx, cy, r float32, col color.RGBA) {
	vector.StrokeLine(img, cx, cy-r, cx, cy+r, 2, col, false)
	vector.StrokeLine(img, cx-r, cy, cx+r, cy, 2, col, false)
	vector.StrokeLine(img, cx-r*0.7, cy-r*0.7, cx+r*0.7, cy+r*0.7, 1.5, col, false)
	vector.StrokeLine(img, cx+r*0.7, cy-r*0.7, cx-r*0.7, cy+r*0.7, 1.5, col, false)
}

func (s *System) drawStar(img *ebiten.Image, cx, cy, r float32, col color.RGBA) {
	vector.DrawFilledCircle(img, cx, cy, r*0.3, col, false)
	for i := 0; i < 8; i++ {
		angle := float64(i) * math.Pi / 4
		x1 := cx + float32(math.Cos(angle))*r*0.35
		y1 := cy + float32(math.Sin(angle))*r*0.35
		x2 := cx + float32(math.Cos(angle))*r
		y2 := cy + float32(math.Sin(angle))*r
		vector.StrokeLine(img, x1, y1, x2, y2, 2, col, false)
	}
}

func (s *System) drawPlus(img *ebiten.Image, cx, cy, r float32, col color.RGBA) {
	vector.DrawFilledRect(img, cx-r*0.25, cy-r, r*0.5, r*2, col, false)
	vector.DrawFilledRect(img, cx-r, cy-r*0.25, r*2, r*0.5, col, false)
}

func (s *System) drawShield(img *ebiten.Image, cx, cy, r float32, col color.RGBA) {
	vector.DrawFilledCircle(img, cx, cy-r*0.2, r*0.7, col, false)
	points := []float32{
		cx - r*0.7, cy - r*0.2,
		cx, cy + r,
		cx + r*0.7, cy - r*0.2,
	}
	for i := 0; i < len(points)-2; i += 2 {
		vector.StrokeLine(img, points[i], points[i+1], points[i+2], points[i+3], 2, col, false)
	}
}

func (s *System) drawArrow(img *ebiten.Image, cx, cy, r float32, col color.RGBA, dir int) {
	dirF := float32(dir)
	vector.DrawFilledRect(img, cx-r*0.15, cy-r*0.5*dirF, r*0.3, r*dirF, col, false)
	vector.StrokeLine(img, cx-r*0.5, cy-r*0.3*dirF, cx, cy-r*0.8*dirF, 2, col, false)
	vector.StrokeLine(img, cx+r*0.5, cy-r*0.3*dirF, cx, cy-r*0.8*dirF, 2, col, false)
}

func (s *System) drawCross(img *ebiten.Image, cx, cy, r float32, col color.RGBA) {
	vector.StrokeLine(img, cx-r, cy-r, cx+r, cy+r, 2.5, col, false)
	vector.StrokeLine(img, cx+r, cy-r, cx-r, cy+r, 2.5, col, false)
}

func (s *System) drawSword(img *ebiten.Image, cx, cy, r float32, col color.RGBA) {
	vector.DrawFilledRect(img, cx-r*0.1, cy-r*0.8, r*0.2, r*1.3, col, false)
	vector.DrawFilledRect(img, cx-r*0.4, cy-r*0.5, r*0.8, r*0.15, col, false)
	vector.DrawFilledRect(img, cx-r*0.15, cy+r*0.5, r*0.3, r*0.3, col, false)
}

func (s *System) drawEye(img *ebiten.Image, cx, cy, r float32, col color.RGBA) {
	vector.DrawFilledCircle(img, cx, cy, r*0.6, col, false)
	vector.DrawFilledCircle(img, cx, cy, r*0.3, color.RGBA{0, 0, 0, col.A}, false)
	vector.StrokeLine(img, cx-r, cy, cx+r, cy, 2, col, false)
}

// StatusIconsComponent stores status icons for an entity.
type StatusIconsComponent struct {
	Icons []StatusIcon
}

// Type implements engine.Component interface.
func (s *StatusIconsComponent) Type() string {
	return "statusicons"
}

// AddIcon adds a status icon to the component.
func (s *StatusIconsComponent) AddIcon(iconType StatusIconType, duration float64, stacks int) {
	for i := range s.Icons {
		if s.Icons[i].Type == iconType {
			s.Icons[i].Duration = duration
			s.Icons[i].Stacks = stacks
			return
		}
	}

	col := color.RGBA{255, 255, 255, 255}
	s.Icons = append(s.Icons, StatusIcon{
		Type:     iconType,
		Duration: duration,
		Stacks:   stacks,
		Color:    col,
	})
}

// RemoveIcon removes a status icon by type.
func (s *StatusIconsComponent) RemoveIcon(iconType StatusIconType) {
	for i := range s.Icons {
		if s.Icons[i].Type == iconType {
			s.Icons = append(s.Icons[:i], s.Icons[i+1:]...)
			return
		}
	}
}

// UpdateDurations decreases all icon durations and removes expired icons.
func (s *StatusIconsComponent) UpdateDurations(deltaTime float64) {
	i := 0
	for i < len(s.Icons) {
		s.Icons[i].Duration -= deltaTime
		if s.Icons[i].Duration <= 0 {
			s.Icons = append(s.Icons[:i], s.Icons[i+1:]...)
		} else {
			i++
		}
	}
}

// Clear removes all status icons.
func (s *StatusIconsComponent) Clear() {
	s.Icons = s.Icons[:0]
}

// poolImage gets a reusable image from the pool.
func (s *System) poolImage(width, height int) *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, width, height))
}

// RenderHealthBarsWithLayout draws health bars using layout manager to prevent overlap.
func (s *System) RenderHealthBarsWithLayout(screen *ebiten.Image, w *engine.World, cameraX, cameraY, cameraDirX, cameraDirY float64, screenWidth, screenHeight int, layoutMgr *ui.LayoutManager) {
	healthType := reflect.TypeOf(&engine.Health{})
	barType := reflect.TypeOf(&Component{})
	posType := reflect.TypeOf(&engine.Position{})

	entities := w.Query(healthType)

	// First pass: collect visible health bars and prioritize by distance
	type barRenderInfo struct {
		eid       engine.Entity
		screenX   float32
		screenY   float32
		health    *engine.Health
		bar       *Component
		healthPct float64
		distance  float64
	}

	visibleBars := make([]barRenderInfo, 0, len(entities))

	for _, eid := range entities {
		healthComp, ok := w.GetComponent(eid, healthType)
		if !ok {
			continue
		}
		health := healthComp.(*engine.Health)

		barComp, hasBar := w.GetComponent(eid, barType)
		if !hasBar {
			continue
		}
		bar := barComp.(*Component)

		if !bar.Visible {
			continue
		}

		if health.Max == 0 {
			continue
		}

		healthPct := float64(health.Current) / float64(health.Max)
		if healthPct >= 1.0 && !bar.ShowWhenFull && bar.LastDamageAge > s.fadeDelay {
			continue
		}

		posComp, hasPos := w.GetComponent(eid, posType)
		if !hasPos {
			continue
		}
		pos := posComp.(*engine.Position)

		screenX, screenY, visible := s.worldToScreen(pos.X, pos.Y, cameraX, cameraY, cameraDirX, cameraDirY, screenWidth, screenHeight)
		if !visible {
			continue
		}

		// Calculate distance from camera for priority
		dx := pos.X - cameraX
		dy := pos.Y - cameraY
		distance := math.Sqrt(dx*dx + dy*dy)

		visibleBars = append(visibleBars, barRenderInfo{
			eid:       eid,
			screenX:   screenX,
			screenY:   screenY,
			health:    health,
			bar:       bar,
			healthPct: healthPct,
			distance:  distance,
		})
	}

	// Second pass: render with layout manager, closer entities get higher priority
	for _, info := range visibleBars {
		barY := info.screenY - info.bar.OffsetY
		barX := info.screenX - info.bar.Width/2

		// Determine priority based on distance
		priority := ui.PrioritySecondary
		if info.distance < 3.0 {
			priority = ui.PriorityCritical
		} else if info.distance < 8.0 {
			priority = ui.PriorityImportant
		}

		// Reserve space for health bar
		adjustedX, adjustedY, visible := layoutMgr.Reserve(
			"healthbar",
			barX,
			barY,
			info.bar.Width,
			info.bar.Height+8, // Extra space for status icons
			priority,
			true, // Health bars can move to avoid overlap
		)

		if !visible {
			// Too much clutter, skip this health bar
			continue
		}

		// Draw health bar at adjusted position
		s.drawHealthBar(screen, adjustedX, adjustedY, info.bar.Width, info.bar.Height, info.healthPct, info.bar)

		// Draw status icons above health bar
		s.drawStatusIcons(screen, w, info.eid, adjustedX+info.bar.Width/2, adjustedY-4)
	}
}
