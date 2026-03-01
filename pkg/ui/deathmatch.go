package ui

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	KillFeedMaxEntries = 5
	KillFeedDuration   = 5 * time.Second
	ScoreboardColumns  = 4
)

// KillFeedEntry represents a single kill notification.
type KillFeedEntry struct {
	KillerID   uint64
	VictimID   uint64
	KillerName string
	VictimName string
	Timestamp  time.Time
	Suicide    bool
	TeamKill   bool
}

// KillFeed displays real-time kill notifications.
type KillFeed struct {
	Entries []KillFeedEntry
	X       int
	Y       int
}

// NewKillFeed creates a new kill feed display.
func NewKillFeed(x, y int) *KillFeed {
	return &KillFeed{
		Entries: make([]KillFeedEntry, 0, KillFeedMaxEntries),
		X:       x,
		Y:       y,
	}
}

// AddKill adds a kill notification to the feed.
func (kf *KillFeed) AddKill(killerID, victimID uint64, killerName, victimName string, suicide, teamKill bool) {
	entry := KillFeedEntry{
		KillerID:   killerID,
		VictimID:   victimID,
		KillerName: killerName,
		VictimName: victimName,
		Timestamp:  time.Now(),
		Suicide:    suicide,
		TeamKill:   teamKill,
	}

	kf.Entries = append(kf.Entries, entry)

	// Keep only most recent entries
	if len(kf.Entries) > KillFeedMaxEntries {
		kf.Entries = kf.Entries[1:]
	}
}

// Update removes expired kill feed entries.
func (kf *KillFeed) Update() {
	now := time.Now()
	validEntries := make([]KillFeedEntry, 0, len(kf.Entries))

	for _, entry := range kf.Entries {
		if now.Sub(entry.Timestamp) < KillFeedDuration {
			validEntries = append(validEntries, entry)
		}
	}

	kf.Entries = validEntries
}

// Draw renders the kill feed to the screen.
func (kf *KillFeed) Draw(screen *ebiten.Image) {
	y := kf.Y

	for _, entry := range kf.Entries {
		var msg string
		var textColor color.RGBA

		if entry.Suicide {
			msg = fmt.Sprintf("%s [SUICIDE]", entry.VictimName)
			textColor = color.RGBA{R: 255, G: 100, B: 100, A: 255}
		} else if entry.TeamKill {
			msg = fmt.Sprintf("%s [TK] %s", entry.KillerName, entry.VictimName)
			textColor = color.RGBA{R: 255, G: 200, B: 0, A: 255}
		} else {
			msg = fmt.Sprintf("%s killed %s", entry.KillerName, entry.VictimName)
			textColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
		}

		text.Draw(screen, msg, basicfont.Face7x13, kf.X, y, textColor)
		y += 15
	}
}

// ScoreboardEntry represents a player's stats in the scoreboard.
// Fields are exported to allow struct literal construction in rendering contexts.
// Use NewScoreboardEntry for canonical construction.
type ScoreboardEntry struct {
	PlayerID   uint64
	PlayerName string
	Team       int
	Frags      int
	Deaths     int
	Assists    int
}

// NewScoreboardEntry creates a ScoreboardEntry with the given stats.
func NewScoreboardEntry(id uint64, name string, team, frags, deaths, assists int) ScoreboardEntry {
	return ScoreboardEntry{
		PlayerID:   id,
		PlayerName: name,
		Team:       team,
		Frags:      frags,
		Deaths:     deaths,
		Assists:    assists,
	}
}

// Scoreboard displays end-of-match or in-game statistics.
type Scoreboard struct {
	Title      string
	Entries    []ScoreboardEntry
	Visible    bool
	ShowTeams  bool
	WinnerText string
}

// NewScoreboard creates a new scoreboard display.
func NewScoreboard(title string, showTeams bool) *Scoreboard {
	return &Scoreboard{
		Title:     title,
		Entries:   make([]ScoreboardEntry, 0),
		ShowTeams: showTeams,
	}
}

// SetEntries updates the scoreboard with player stats.
func (sb *Scoreboard) SetEntries(entries []ScoreboardEntry) {
	sb.Entries = entries
}

// SetWinner sets the winner text displayed at the top.
func (sb *Scoreboard) SetWinner(winnerText string) {
	sb.WinnerText = winnerText
}

// Show makes the scoreboard visible.
func (sb *Scoreboard) Show() {
	sb.Visible = true
}

// Hide makes the scoreboard invisible.
func (sb *Scoreboard) Hide() {
	sb.Visible = false
}

// Toggle toggles scoreboard visibility.
func (sb *Scoreboard) Toggle() {
	sb.Visible = !sb.Visible
}

// Draw renders the scoreboard to the screen.
func (sb *Scoreboard) Draw(screen *ebiten.Image) {
	if !sb.Visible {
		return
	}

	screenWidth := screen.Bounds().Dx()
	screenHeight := screen.Bounds().Dy()

	// Semi-transparent background
	bgColor := color.RGBA{R: 0, G: 0, B: 0, A: 200}
	vector.DrawFilledRect(screen, 50, 50, float32(screenWidth-100), float32(screenHeight-100), bgColor, false)

	// Border
	borderColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	vector.StrokeRect(screen, 50, 50, float32(screenWidth-100), float32(screenHeight-100), 2, borderColor, false)

	y := 80

	// Title
	titleColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	titleX := (screenWidth / 2) - (len(sb.Title) * 7 / 2)
	text.Draw(screen, sb.Title, basicfont.Face7x13, titleX, y, titleColor)
	y += 30

	// Winner text
	if sb.WinnerText != "" {
		winnerColor := color.RGBA{R: 255, G: 215, B: 0, A: 255}
		winnerX := (screenWidth / 2) - (len(sb.WinnerText) * 7 / 2)
		text.Draw(screen, sb.WinnerText, basicfont.Face7x13, winnerX, y, winnerColor)
		y += 25
	}

	// Column headers
	headerColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	headerY := y

	if sb.ShowTeams {
		text.Draw(screen, "Player", basicfont.Face7x13, 100, headerY, headerColor)
		text.Draw(screen, "Team", basicfont.Face7x13, 250, headerY, headerColor)
		text.Draw(screen, "K", basicfont.Face7x13, 350, headerY, headerColor)
		text.Draw(screen, "D", basicfont.Face7x13, 400, headerY, headerColor)
		text.Draw(screen, "A", basicfont.Face7x13, 450, headerY, headerColor)
		text.Draw(screen, "K/D", basicfont.Face7x13, 500, headerY, headerColor)
	} else {
		text.Draw(screen, "Player", basicfont.Face7x13, 100, headerY, headerColor)
		text.Draw(screen, "Frags", basicfont.Face7x13, 300, headerY, headerColor)
		text.Draw(screen, "Deaths", basicfont.Face7x13, 400, headerY, headerColor)
		text.Draw(screen, "K/D", basicfont.Face7x13, 500, headerY, headerColor)
	}
	y += 25

	// Entries
	entryColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	redTeamColor := color.RGBA{R: 255, G: 100, B: 100, A: 255}
	blueTeamColor := color.RGBA{R: 100, G: 100, B: 255, A: 255}

	for i, entry := range sb.Entries {
		if i >= 16 {
			break // Max 16 entries visible
		}

		playerColor := entryColor
		if sb.ShowTeams {
			if entry.Team == 0 {
				playerColor = redTeamColor
			} else if entry.Team == 1 {
				playerColor = blueTeamColor
			}
		}

		// Player name
		text.Draw(screen, entry.PlayerName, basicfont.Face7x13, 100, y, playerColor)

		if sb.ShowTeams {
			// Team
			teamName := "Red"
			if entry.Team == 1 {
				teamName = "Blue"
			}
			text.Draw(screen, teamName, basicfont.Face7x13, 250, y, playerColor)

			// K/D/A
			text.Draw(screen, fmt.Sprintf("%d", entry.Frags), basicfont.Face7x13, 350, y, entryColor)
			text.Draw(screen, fmt.Sprintf("%d", entry.Deaths), basicfont.Face7x13, 400, y, entryColor)
			text.Draw(screen, fmt.Sprintf("%d", entry.Assists), basicfont.Face7x13, 450, y, entryColor)
		} else {
			// Frags/Deaths
			text.Draw(screen, fmt.Sprintf("%d", entry.Frags), basicfont.Face7x13, 300, y, entryColor)
			text.Draw(screen, fmt.Sprintf("%d", entry.Deaths), basicfont.Face7x13, 400, y, entryColor)
		}

		// K/D ratio
		kdRatio := 0.0
		if entry.Deaths > 0 {
			kdRatio = float64(entry.Frags) / float64(entry.Deaths)
		} else if entry.Frags > 0 {
			kdRatio = float64(entry.Frags)
		}
		text.Draw(screen, fmt.Sprintf("%.2f", kdRatio), basicfont.Face7x13, 500, y, entryColor)

		y += 18
	}
}
