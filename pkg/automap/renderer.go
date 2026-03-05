package automap

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// RenderConfig holds visual settings for minimap rendering.
type RenderConfig struct {
	X, Y          float32
	Width, Height float32
	CellSize      float32
	PlayerX       float64
	PlayerY       float64
	PlayerAngle   float64
	Walls         [][]bool
	Enemies       []EnemyMarker
	Items         []ItemMarker
	Opacity       float32
	ShowFogOfWar  bool
}

// EnemyMarker represents an enemy on the minimap.
type EnemyMarker struct {
	X, Y      float64
	IsBoss    bool
	IsHostile bool
	HealthPct float64
}

// ItemMarker represents an item on the minimap.
type ItemMarker struct {
	X, Y    float64
	IsQuest bool
	IsRare  bool
}

// GenreTheme holds colors for genre-specific minimap rendering.
type GenreTheme struct {
	Background color.RGBA
	Wall       color.RGBA
	Floor      color.RGBA
	Player     color.RGBA
	Enemy      color.RGBA
	EnemyBoss  color.RGBA
	Item       color.RGBA
	ItemRare   color.RGBA
	Secret     color.RGBA
	Objective  color.RGBA
	FogOfWar   color.RGBA
	Border     color.RGBA
}

// GetGenreTheme returns colors for a genre.
func GetGenreTheme(genreID string) GenreTheme {
	switch genreID {
	case "scifi":
		return GenreTheme{
			Background: color.RGBA{R: 10, G: 20, B: 30, A: 200},
			Wall:       color.RGBA{R: 60, G: 90, B: 120, A: 255},
			Floor:      color.RGBA{R: 20, G: 40, B: 60, A: 255},
			Player:     color.RGBA{R: 0, G: 200, B: 255, A: 255},
			Enemy:      color.RGBA{R: 255, G: 100, B: 100, A: 255},
			EnemyBoss:  color.RGBA{R: 255, G: 50, B: 0, A: 255},
			Item:       color.RGBA{R: 200, G: 200, B: 255, A: 255},
			ItemRare:   color.RGBA{R: 255, G: 255, B: 100, A: 255},
			Secret:     color.RGBA{R: 100, G: 255, B: 200, A: 255},
			Objective:  color.RGBA{R: 255, G: 200, B: 0, A: 255},
			FogOfWar:   color.RGBA{R: 5, G: 10, B: 15, A: 180},
			Border:     color.RGBA{R: 80, G: 120, B: 160, A: 255},
		}
	case "horror":
		return GenreTheme{
			Background: color.RGBA{R: 15, G: 5, B: 5, A: 200},
			Wall:       color.RGBA{R: 60, G: 30, B: 30, A: 255},
			Floor:      color.RGBA{R: 30, G: 20, B: 20, A: 255},
			Player:     color.RGBA{R: 200, G: 200, B: 200, A: 255},
			Enemy:      color.RGBA{R: 255, G: 50, B: 50, A: 255},
			EnemyBoss:  color.RGBA{R: 200, G: 0, B: 0, A: 255},
			Item:       color.RGBA{R: 180, G: 180, B: 160, A: 255},
			ItemRare:   color.RGBA{R: 255, G: 220, B: 180, A: 255},
			Secret:     color.RGBA{R: 100, G: 100, B: 180, A: 255},
			Objective:  color.RGBA{R: 255, G: 200, B: 100, A: 255},
			FogOfWar:   color.RGBA{R: 10, G: 5, B: 5, A: 200},
			Border:     color.RGBA{R: 80, G: 40, B: 40, A: 255},
		}
	case "cyberpunk":
		return GenreTheme{
			Background: color.RGBA{R: 5, G: 5, B: 15, A: 200},
			Wall:       color.RGBA{R: 40, G: 40, B: 60, A: 255},
			Floor:      color.RGBA{R: 15, G: 15, B: 25, A: 255},
			Player:     color.RGBA{R: 0, G: 255, B: 255, A: 255},
			Enemy:      color.RGBA{R: 255, G: 0, B: 128, A: 255},
			EnemyBoss:  color.RGBA{R: 255, G: 0, B: 255, A: 255},
			Item:       color.RGBA{R: 200, G: 200, B: 255, A: 255},
			ItemRare:   color.RGBA{R: 255, G: 255, B: 0, A: 255},
			Secret:     color.RGBA{R: 0, G: 255, B: 128, A: 255},
			Objective:  color.RGBA{R: 255, G: 128, B: 0, A: 255},
			FogOfWar:   color.RGBA{R: 5, G: 5, B: 10, A: 190},
			Border:     color.RGBA{R: 100, G: 0, B: 200, A: 255},
		}
	case "postapoc":
		return GenreTheme{
			Background: color.RGBA{R: 20, G: 15, B: 10, A: 200},
			Wall:       color.RGBA{R: 80, G: 70, B: 50, A: 255},
			Floor:      color.RGBA{R: 40, G: 35, B: 25, A: 255},
			Player:     color.RGBA{R: 200, G: 180, B: 140, A: 255},
			Enemy:      color.RGBA{R: 200, G: 100, B: 80, A: 255},
			EnemyBoss:  color.RGBA{R: 180, G: 60, B: 40, A: 255},
			Item:       color.RGBA{R: 160, G: 160, B: 140, A: 255},
			ItemRare:   color.RGBA{R: 220, G: 200, B: 140, A: 255},
			Secret:     color.RGBA{R: 100, G: 140, B: 120, A: 255},
			Objective:  color.RGBA{R: 220, G: 180, B: 100, A: 255},
			FogOfWar:   color.RGBA{R: 15, G: 12, B: 8, A: 190},
			Border:     color.RGBA{R: 100, G: 90, B: 70, A: 255},
		}
	default: // fantasy
		return GenreTheme{
			Background: color.RGBA{R: 20, G: 15, B: 30, A: 200},
			Wall:       color.RGBA{R: 100, G: 80, B: 60, A: 255},
			Floor:      color.RGBA{R: 50, G: 45, B: 40, A: 255},
			Player:     color.RGBA{R: 100, G: 200, B: 255, A: 255},
			Enemy:      color.RGBA{R: 255, G: 100, B: 100, A: 255},
			EnemyBoss:  color.RGBA{R: 200, G: 50, B: 200, A: 255},
			Item:       color.RGBA{R: 255, G: 255, B: 180, A: 255},
			ItemRare:   color.RGBA{R: 255, G: 215, B: 0, A: 255},
			Secret:     color.RGBA{R: 150, G: 255, B: 150, A: 255},
			Objective:  color.RGBA{R: 255, G: 180, B: 80, A: 255},
			FogOfWar:   color.RGBA{R: 10, G: 8, B: 15, A: 190},
			Border:     color.RGBA{R: 120, G: 100, B: 80, A: 255},
		}
	}
}

// RenderMinimap draws the minimap to the screen.
func (m *Map) RenderMinimap(screen *ebiten.Image, cfg RenderConfig) {
	theme := GetGenreTheme(currentGenre)
	cfg = applyDefaultRenderConfig(cfg)

	centerX := cfg.X + cfg.Width/2
	centerY := cfg.Y + cfg.Height/2
	playerGridX := int(cfg.PlayerX)
	playerGridY := int(cfg.PlayerY)
	visibleRadius := int(math.Min(float64(cfg.Width), float64(cfg.Height)) / float64(cfg.CellSize) / 2)

	drawMinimapBackground(screen, cfg, theme)
	m.renderTerrainCells(screen, cfg, theme, centerX, centerY, playerGridX, playerGridY, visibleRadius)
	m.renderAnnotations(screen, cfg, theme, centerX, centerY, playerGridX, playerGridY, visibleRadius)
	m.renderItems(screen, cfg, theme, centerX, centerY, playerGridX, playerGridY, visibleRadius)
	m.renderEnemies(screen, cfg, theme, centerX, centerY, playerGridX, playerGridY, visibleRadius)
	drawPlayerIndicator(screen, cfg, theme, centerX, centerY)
}

// applyDefaultRenderConfig sets default values for missing config parameters.
func applyDefaultRenderConfig(cfg RenderConfig) RenderConfig {
	if cfg.CellSize == 0 {
		cfg.CellSize = 3.0
	}
	if cfg.Opacity == 0 {
		cfg.Opacity = 0.85
	}
	return cfg
}

// drawMinimapBackground renders background and border for the minimap.
func drawMinimapBackground(screen *ebiten.Image, cfg RenderConfig, theme GenreTheme) {
	bgColor := theme.Background
	bgColor.A = uint8(float32(bgColor.A) * cfg.Opacity)
	vector.DrawFilledRect(screen, cfg.X, cfg.Y, cfg.Width, cfg.Height, bgColor, false)

	borderThickness := float32(2.0)
	vector.StrokeRect(screen, cfg.X, cfg.Y, cfg.Width, cfg.Height, borderThickness, theme.Border, false)
}

// renderTerrainCells draws floor and wall cells within visible radius.
func (m *Map) renderTerrainCells(screen *ebiten.Image, cfg RenderConfig, theme GenreTheme, centerX, centerY float32, playerGridX, playerGridY, visibleRadius int) {
	for dy := -visibleRadius; dy <= visibleRadius; dy++ {
		for dx := -visibleRadius; dx <= visibleRadius; dx++ {
			gridX := playerGridX + dx
			gridY := playerGridY + dy

			if gridX < 0 || gridX >= m.Width || gridY < 0 || gridY >= m.Height {
				continue
			}

			screenX := centerX + float32(dx)*cfg.CellSize
			screenY := centerY + float32(dy)*cfg.CellSize

			if screenX < cfg.X || screenX+cfg.CellSize > cfg.X+cfg.Width ||
				screenY < cfg.Y || screenY+cfg.CellSize > cfg.Y+cfg.Height {
				continue
			}

			if !m.Revealed[gridY][gridX] {
				if cfg.ShowFogOfWar {
					vector.DrawFilledRect(screen, screenX, screenY, cfg.CellSize, cfg.CellSize, theme.FogOfWar, false)
				}
				continue
			}

			isWall := false
			if cfg.Walls != nil && gridY < len(cfg.Walls) && gridX < len(cfg.Walls[gridY]) {
				isWall = cfg.Walls[gridY][gridX]
			}

			cellColor := theme.Floor
			if isWall {
				cellColor = theme.Wall
			}

			vector.DrawFilledRect(screen, screenX, screenY, cfg.CellSize, cfg.CellSize, cellColor, false)
		}
	}
}

// renderAnnotations draws map annotations like secrets and objectives.
func (m *Map) renderAnnotations(screen *ebiten.Image, cfg RenderConfig, theme GenreTheme, centerX, centerY float32, playerGridX, playerGridY, visibleRadius int) {
	for _, ann := range m.Annotations {
		dx := ann.X - playerGridX
		dy := ann.Y - playerGridY

		if math.Abs(float64(dx)) > float64(visibleRadius) || math.Abs(float64(dy)) > float64(visibleRadius) {
			continue
		}

		if ann.X < 0 || ann.X >= m.Width || ann.Y < 0 || ann.Y >= m.Height {
			continue
		}
		if !m.Revealed[ann.Y][ann.X] {
			continue
		}

		screenX := centerX + float32(dx)*cfg.CellSize
		screenY := centerY + float32(dy)*cfg.CellSize

		var markerColor color.RGBA
		switch ann.Type {
		case AnnotationSecret:
			markerColor = theme.Secret
		case AnnotationObjective:
			markerColor = theme.Objective
		case AnnotationItem:
			markerColor = theme.Item
		default:
			continue
		}

		markerSize := cfg.CellSize * 1.2
		vector.DrawFilledCircle(screen, screenX+cfg.CellSize/2, screenY+cfg.CellSize/2, markerSize, markerColor, false)
	}
}

// renderItems draws items on the minimap with rarity distinction.
func (m *Map) renderItems(screen *ebiten.Image, cfg RenderConfig, theme GenreTheme, centerX, centerY float32, playerGridX, playerGridY, visibleRadius int) {
	for _, item := range cfg.Items {
		dx := int(item.X) - playerGridX
		dy := int(item.Y) - playerGridY

		if math.Abs(float64(dx)) > float64(visibleRadius) || math.Abs(float64(dy)) > float64(visibleRadius) {
			continue
		}

		gridX := int(item.X)
		gridY := int(item.Y)
		if gridX < 0 || gridX >= m.Width || gridY < 0 || gridY >= m.Height {
			continue
		}
		if !m.Revealed[gridY][gridX] {
			continue
		}

		screenX := centerX + float32(dx)*cfg.CellSize
		screenY := centerY + float32(dy)*cfg.CellSize

		itemColor := theme.Item
		if item.IsRare {
			itemColor = theme.ItemRare
		}

		itemSize := cfg.CellSize * 0.8
		vector.DrawFilledRect(screen, screenX+cfg.CellSize/2-itemSize/2, screenY+cfg.CellSize/2-itemSize/2,
			itemSize, itemSize, itemColor, false)
	}
}

// renderEnemies draws enemies with boss health bars.
func (m *Map) renderEnemies(screen *ebiten.Image, cfg RenderConfig, theme GenreTheme, centerX, centerY float32, playerGridX, playerGridY, visibleRadius int) {
	for _, enemy := range cfg.Enemies {
		dx := int(enemy.X) - playerGridX
		dy := int(enemy.Y) - playerGridY

		if math.Abs(float64(dx)) > float64(visibleRadius) || math.Abs(float64(dy)) > float64(visibleRadius) {
			continue
		}

		gridX := int(enemy.X)
		gridY := int(enemy.Y)
		if gridX < 0 || gridX >= m.Width || gridY < 0 || gridY >= m.Height {
			continue
		}
		if !m.Revealed[gridY][gridX] {
			continue
		}

		screenX := centerX + float32(dx)*cfg.CellSize
		screenY := centerY + float32(dy)*cfg.CellSize

		enemyColor := theme.Enemy
		if enemy.IsBoss {
			enemyColor = theme.EnemyBoss
		}

		enemySize := cfg.CellSize * 0.9
		if enemy.IsBoss {
			enemySize = cfg.CellSize * 1.3
		}

		vector.DrawFilledCircle(screen, screenX+cfg.CellSize/2, screenY+cfg.CellSize/2, enemySize/2, enemyColor, false)

		if enemy.IsBoss && enemy.HealthPct < 1.0 {
			drawBossHealthBar(screen, cfg, screenX, screenY, enemy.HealthPct)
		}
	}
}

// drawBossHealthBar renders a health bar for boss enemies.
func drawBossHealthBar(screen *ebiten.Image, cfg RenderConfig, screenX, screenY float32, healthPct float64) {
	barWidth := cfg.CellSize * 1.5
	barHeight := cfg.CellSize * 0.3
	barX := screenX + cfg.CellSize/2 - barWidth/2
	barY := screenY - cfg.CellSize*0.5

	vector.DrawFilledRect(screen, barX, barY, barWidth, barHeight, color.RGBA{R: 40, G: 40, B: 40, A: 200}, false)
	vector.DrawFilledRect(screen, barX, barY, barWidth*float32(healthPct), barHeight,
		color.RGBA{R: 200, G: 50, B: 50, A: 255}, false)
}

// drawPlayerIndicator renders the player marker and direction arrow.
func drawPlayerIndicator(screen *ebiten.Image, cfg RenderConfig, theme GenreTheme, centerX, centerY float32) {
	playerSize := cfg.CellSize * 1.5
	vector.DrawFilledCircle(screen, centerX, centerY, playerSize/2, theme.Player, false)

	dirLen := cfg.CellSize * 1.2
	dirEndX := centerX + float32(math.Cos(cfg.PlayerAngle))*dirLen
	dirEndY := centerY + float32(math.Sin(cfg.PlayerAngle))*dirLen
	vector.StrokeLine(screen, centerX, centerY, dirEndX, dirEndY, 2.0, theme.Player, false)
}
