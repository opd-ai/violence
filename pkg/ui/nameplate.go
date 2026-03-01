// Package ui provides nameplate rendering for multiplayer players.
package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// NameplatePlayer represents a player's info for nameplate display.
// Fields are exported to allow struct literal construction in rendering contexts.
// Use NewNameplatePlayer for canonical construction with validation.
type NameplatePlayer struct {
	PlayerID   string
	PlayerName string
	SquadTag   string // Up to 4 characters
	ScreenX    float32
	ScreenY    float32
	IsTeammate bool
	IsSelf     bool
}

// NewNameplatePlayer creates a NameplatePlayer with validation.
func NewNameplatePlayer(id, name, tag string, x, y float32, isTeammate, isSelf bool) NameplatePlayer {
	// Truncate squad tag to max length
	if len(tag) > 4 {
		tag = tag[:4]
	}
	return NameplatePlayer{
		PlayerID:   id,
		PlayerName: name,
		SquadTag:   tag,
		ScreenX:    x,
		ScreenY:    y,
		IsTeammate: isTeammate,
		IsSelf:     isSelf,
	}
}

// Nameplate manages display of player nameplates in multiplayer.
type Nameplate struct {
	players         []NameplatePlayer
	maxTagLength    int
	teammateColor   color.RGBA
	enemyColor      color.RGBA
	selfColor       color.RGBA
	tagBGColor      color.RGBA
	tagBorderColor  color.RGBA
	nameBGColor     color.RGBA
	nameBorderColor color.RGBA
}

// NewNameplate creates a new nameplate manager with default colors.
func NewNameplate() *Nameplate {
	return &Nameplate{
		players:         make([]NameplatePlayer, 0, 16),
		maxTagLength:    4,
		teammateColor:   color.RGBA{0, 255, 0, 255},     // Green
		enemyColor:      color.RGBA{255, 0, 0, 255},     // Red
		selfColor:       color.RGBA{255, 255, 0, 255},   // Yellow
		tagBGColor:      color.RGBA{40, 40, 40, 200},    // Dark semi-transparent
		tagBorderColor:  color.RGBA{255, 255, 255, 255}, // White
		nameBGColor:     color.RGBA{30, 30, 30, 180},    // Darker semi-transparent
		nameBorderColor: color.RGBA{200, 200, 200, 255}, // Light gray
	}
}

// SetPlayers updates the list of players to display nameplates for.
func (n *Nameplate) SetPlayers(players []NameplatePlayer) {
	n.players = make([]NameplatePlayer, len(players))
	copy(n.players, players)
}

// AddPlayer adds a single player to the nameplate display.
func (n *Nameplate) AddPlayer(player NameplatePlayer) {
	// Truncate squad tag to max length
	if len(player.SquadTag) > n.maxTagLength {
		player.SquadTag = player.SquadTag[:n.maxTagLength]
	}
	n.players = append(n.players, player)
}

// ClearPlayers removes all players from the nameplate display.
func (n *Nameplate) ClearPlayers() {
	n.players = n.players[:0]
}

// GetPlayerCount returns the number of players being displayed.
func (n *Nameplate) GetPlayerCount() int {
	return len(n.players)
}

// SetTeammateColor sets the color for teammate nameplates.
func (n *Nameplate) SetTeammateColor(c color.RGBA) {
	n.teammateColor = c
}

// SetEnemyColor sets the color for enemy nameplates.
func (n *Nameplate) SetEnemyColor(c color.RGBA) {
	n.enemyColor = c
}

// SetSelfColor sets the color for the local player's nameplate.
func (n *Nameplate) SetSelfColor(c color.RGBA) {
	n.selfColor = c
}

// Draw renders all nameplates onto the screen.
func (n *Nameplate) Draw(screen *ebiten.Image) {
	for _, player := range n.players {
		n.drawPlayerNameplate(screen, player)
	}
}

// drawPlayerNameplate renders a single player's nameplate at their screen position.
func (n *Nameplate) drawPlayerNameplate(screen *ebiten.Image, player NameplatePlayer) {
	// Determine text color based on relationship
	textColor := n.enemyColor
	if player.IsSelf {
		textColor = n.selfColor
	} else if player.IsTeammate {
		textColor = n.teammateColor
	}

	x := player.ScreenX
	y := player.ScreenY

	// Draw squad tag above name if present
	if player.SquadTag != "" {
		tagWidth := float32(len(player.SquadTag)*7 + 6) // 7px per char + padding
		tagHeight := float32(14)
		tagX := x - tagWidth/2
		tagY := y - 30 // Above name

		// Tag background
		vector.DrawFilledRect(screen, tagX, tagY, tagWidth, tagHeight, n.tagBGColor, false)

		// Tag border
		vector.StrokeRect(screen, tagX, tagY, tagWidth, tagHeight, 1, n.tagBorderColor, false)

		// Tag text
		text.Draw(screen, player.SquadTag, basicfont.Face7x13, int(tagX+3), int(tagY+11), textColor)

		// Adjust name position to be below tag
		y += 2
	}

	// Draw player name
	nameWidth := float32(len(player.PlayerName)*7 + 6) // 7px per char + padding
	nameHeight := float32(14)
	nameX := x - nameWidth/2
	nameY := y - 15 // Below tag or at default position

	// Name background
	vector.DrawFilledRect(screen, nameX, nameY, nameWidth, nameHeight, n.nameBGColor, false)

	// Name border
	vector.StrokeRect(screen, nameX, nameY, nameWidth, nameHeight, 1, n.nameBorderColor, false)

	// Name text
	text.Draw(screen, player.PlayerName, basicfont.Face7x13, int(nameX+3), int(nameY+11), textColor)
}
