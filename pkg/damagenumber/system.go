// Package damagenumber provides floating combat text for damage feedback.
package damagenumber

import (
	"fmt"
	"image/color"
	"math"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/ui"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// System handles floating damage number animation and rendering.
type System struct {
	genreID string
	logger  *logrus.Entry
	font    font.Face
}

// NewSystem creates a floating damage number system.
func NewSystem(genreID string) *System {
	return &System{
		genreID: genreID,
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "damagenumber",
			"genre":       genreID,
		}),
		font: basicfont.Face7x13,
	}
}

// Update animates damage numbers (rise, fade, scale).
func (s *System) Update(w *engine.World) {
	compType := reflect.TypeOf((*Component)(nil))
	entities := w.Query(compType)

	const deltaTime = 1.0 / 60.0

	for _, ent := range entities {
		comp, found := w.GetComponent(ent, compType)
		if !found {
			continue
		}

		dmg, ok := comp.(*Component)
		if !ok {
			continue
		}

		dmg.Age += deltaTime

		if dmg.Age >= dmg.Lifetime {
			w.RemoveEntity(ent)
			continue
		}

		progress := dmg.Age / dmg.Lifetime

		dmg.Y -= dmg.VelocityY * deltaTime
		dmg.VelocityY *= 0.95

		if progress < 0.2 {
			dmg.Scale = 0.5 + (progress / 0.2 * 0.5)
		} else if progress > 0.7 {
			fadeProgress := (progress - 0.7) / 0.3
			dmg.Alpha = 1.0 - fadeProgress
		} else {
			dmg.Scale = 1.0
			dmg.Alpha = 1.0
		}
	}
}

// Render draws damage numbers to screen.
func (s *System) Render(w *engine.World, screen *ebiten.Image, cameraX, cameraY float64) {
	compType := reflect.TypeOf((*Component)(nil))
	entities := w.Query(compType)

	for _, ent := range entities {
		comp, found := w.GetComponent(ent, compType)
		if !found {
			continue
		}

		dmg, ok := comp.(*Component)
		if !ok {
			continue
		}

		screenX := int(dmg.X - cameraX)
		screenY := int(dmg.Y - cameraY)

		textStr := fmt.Sprintf("%d", dmg.Value)
		if dmg.IsCritical {
			textStr += "!"
		}

		bounds := text.BoundString(s.font, textStr)
		textWidth := bounds.Dx()
		textHeight := bounds.Dy()

		drawX := screenX - textWidth/2
		drawY := screenY - textHeight/2

		renderColor := dmg.Color
		renderColor.A = uint8(dmg.Alpha * 255)

		if dmg.IsCritical {
			scaleModulation := 1.0 + 0.15*math.Sin(dmg.Age*15.0)
			dmg.Scale *= scaleModulation
		}

		if math.Abs(dmg.Scale-1.0) > 0.01 {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(-textWidth/2), float64(-textHeight/2))
			opts.GeoM.Scale(dmg.Scale, dmg.Scale)
			opts.GeoM.Translate(float64(screenX), float64(screenY))

			tmpImg := ebiten.NewImage(textWidth+4, textHeight+4)
			text.Draw(tmpImg, textStr, s.font, 2, textHeight, renderColor)

			opts.ColorScale.ScaleAlpha(float32(dmg.Alpha))
			screen.DrawImage(tmpImg, opts)
		} else {
			text.Draw(screen, textStr, s.font, drawX, drawY+textHeight, renderColor)
		}
	}
}

// RenderWithLayout draws damage numbers using layout manager to prevent overlap.
func (s *System) RenderWithLayout(w *engine.World, screen *ebiten.Image, cameraX, cameraY float64, layoutMgr *ui.LayoutManager) {
	compType := reflect.TypeOf((*Component)(nil))
	entities := w.Query(compType)

	for _, ent := range entities {
		comp, found := w.GetComponent(ent, compType)
		if !found {
			continue
		}

		dmg, ok := comp.(*Component)
		if !ok {
			continue
		}

		screenX := float32(dmg.X - cameraX)
		screenY := float32(dmg.Y - cameraY)

		textStr := fmt.Sprintf("%d", dmg.Value)
		if dmg.IsCritical {
			textStr += "!"
		}

		bounds := text.BoundString(s.font, textStr)
		textWidth := float32(bounds.Dx())
		textHeight := float32(bounds.Dy())

		// Reserve space with layout manager - damage numbers stack vertically
		priority := ui.PriorityImportant
		if dmg.IsCritical {
			priority = ui.PriorityCritical
		}

		adjustedX, adjustedY, visible := layoutMgr.ReserveDamageNumber(
			fmt.Sprintf("dmg_%d", ent),
			screenX-textWidth/2,
			screenY-textHeight/2,
			textWidth*2,  // Extra width for scaling
			textHeight*2, // Extra height for scaling
			priority,
		)

		if !visible {
			continue
		}

		drawX := int(adjustedX + textWidth/2)
		drawY := int(adjustedY + textHeight/2)

		renderColor := dmg.Color
		renderColor.A = uint8(dmg.Alpha * 255)

		if dmg.IsCritical {
			scaleModulation := 1.0 + 0.15*math.Sin(dmg.Age*15.0)
			dmg.Scale *= scaleModulation
		}

		if math.Abs(dmg.Scale-1.0) > 0.01 {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(-textWidth/2), float64(-textHeight/2))
			opts.GeoM.Scale(dmg.Scale, dmg.Scale)
			opts.GeoM.Translate(float64(drawX), float64(drawY))

			tmpImg := ebiten.NewImage(int(textWidth)+4, int(textHeight)+4)
			text.Draw(tmpImg, textStr, s.font, 2, int(textHeight), renderColor)

			opts.ColorScale.ScaleAlpha(float32(dmg.Alpha))
			screen.DrawImage(tmpImg, opts)
		} else {
			text.Draw(screen, textStr, s.font, drawX-int(textWidth/2), drawY+int(textHeight/2), renderColor)
		}
	}
}

// Spawn creates a new damage number entity.
func Spawn(w *engine.World, value int, x, y float64, damageType string, isCritical, isHeal bool) engine.Entity {
	ent := w.AddEntity()

	baseColor := getDamageColor(damageType, isHeal)
	lifetime := 1.5
	velocityY := 40.0

	if isCritical {
		lifetime = 2.0
		velocityY = 60.0
		baseColor = color.RGBA{255, 255, 100, 255}
	}

	if isHeal {
		baseColor = color.RGBA{100, 255, 100, 255}
		velocityY = 30.0
	}

	comp := &Component{
		Value:      value,
		DamageType: damageType,
		IsCritical: isCritical,
		IsHeal:     isHeal,
		X:          x,
		Y:          y,
		VelocityY:  velocityY,
		Lifetime:   lifetime,
		Age:        0,
		Scale:      0.5,
		Alpha:      1.0,
		Color:      baseColor,
	}

	w.AddComponent(ent, comp)

	return ent
}

// getDamageColor returns color based on damage type.
func getDamageColor(damageType string, isHeal bool) color.RGBA {
	if isHeal {
		return color.RGBA{100, 255, 100, 255}
	}

	switch damageType {
	case "physical", "kinetic", "slash", "pierce", "blunt":
		return color.RGBA{255, 200, 200, 255}
	case "fire", "burn", "heat":
		return color.RGBA{255, 100, 50, 255}
	case "ice", "cold", "frost":
		return color.RGBA{100, 200, 255, 255}
	case "lightning", "electric", "shock":
		return color.RGBA{255, 255, 100, 255}
	case "poison", "toxic", "acid":
		return color.RGBA{150, 255, 100, 255}
	case "dark", "shadow", "void":
		return color.RGBA{150, 100, 200, 255}
	case "holy", "light", "radiant":
		return color.RGBA{255, 255, 200, 255}
	case "arcane", "magic", "mystic":
		return color.RGBA{200, 150, 255, 255}
	default:
		return color.RGBA{255, 255, 255, 255}
	}
}
