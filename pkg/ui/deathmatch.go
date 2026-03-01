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

	drawScoreboardBackground(screen, screenWidth, screenHeight)
	y := drawScoreboardHeader(screen, sb.Title, sb.WinnerText, screenWidth)
	y = drawScoreboardColumnHeaders(screen, y, sb.ShowTeams)
	drawScoreboardEntries(screen, sb.Entries, y, sb.ShowTeams)
}

// drawScoreboardBackground renders the background and border.
func drawScoreboardBackground(screen *ebiten.Image, screenWidth, screenHeight int) {
	bgColor := color.RGBA{R: 0, G: 0, B: 0, A: 200}
	vector.DrawFilledRect(screen, 50, 50, float32(screenWidth-100), float32(screenHeight-100), bgColor, false)

	borderColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	vector.StrokeRect(screen, 50, 50, float32(screenWidth-100), float32(screenHeight-100), 2, borderColor, false)
}

// drawScoreboardHeader renders the title and winner text, returns updated Y position.
func drawScoreboardHeader(screen *ebiten.Image, title, winnerText string, screenWidth int) int {
	y := 80

	titleColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	titleX := (screenWidth / 2) - (len(title) * 7 / 2)
	text.Draw(screen, title, basicfont.Face7x13, titleX, y, titleColor)
	y += 30

	if winnerText != "" {
		winnerColor := color.RGBA{R: 255, G: 215, B: 0, A: 255}
		winnerX := (screenWidth / 2) - (len(winnerText) * 7 / 2)
		text.Draw(screen, winnerText, basicfont.Face7x13, winnerX, y, winnerColor)
		y += 25
	}

	return y
}

// drawScoreboardColumnHeaders renders column headers, returns updated Y position.
func drawScoreboardColumnHeaders(screen *ebiten.Image, y int, showTeams bool) int {
	headerColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}

	if showTeams {
		text.Draw(screen, "Player", basicfont.Face7x13, 100, y, headerColor)
		text.Draw(screen, "Team", basicfont.Face7x13, 250, y, headerColor)
		text.Draw(screen, "K", basicfont.Face7x13, 350, y, headerColor)
		text.Draw(screen, "D", basicfont.Face7x13, 400, y, headerColor)
		text.Draw(screen, "A", basicfont.Face7x13, 450, y, headerColor)
		text.Draw(screen, "K/D", basicfont.Face7x13, 500, y, headerColor)
	} else {
		text.Draw(screen, "Player", basicfont.Face7x13, 100, y, headerColor)
		text.Draw(screen, "Frags", basicfont.Face7x13, 300, y, headerColor)
		text.Draw(screen, "Deaths", basicfont.Face7x13, 400, y, headerColor)
		text.Draw(screen, "K/D", basicfont.Face7x13, 500, y, headerColor)
	}

	return y + 25
}

// drawScoreboardEntries renders player entries.
func drawScoreboardEntries(screen *ebiten.Image, entries []ScoreboardEntry, y int, showTeams bool) {
	entryColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	for i, entry := range entries {
		if i >= 16 {
			break
		}
		playerColor := selectPlayerColor(entry.Team, showTeams)
		y = drawScoreboardEntry(screen, entry, y, playerColor, entryColor, showTeams)
	}
}

// selectPlayerColor determines the color for a player based on team.
func selectPlayerColor(team int, showTeams bool) color.RGBA {
	if !showTeams {
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	if team == 0 {
		return color.RGBA{R: 255, G: 100, B: 100, A: 255}
	}
	if team == 1 {
		return color.RGBA{R: 100, G: 100, B: 255, A: 255}
	}
	return color.RGBA{R: 255, G: 255, B: 255, A: 255}
}

// drawScoreboardEntry renders a single player entry, returns updated Y position.
func drawScoreboardEntry(screen *ebiten.Image, entry ScoreboardEntry, y int, playerColor, entryColor color.RGBA, showTeams bool) int {
	text.Draw(screen, entry.PlayerName, basicfont.Face7x13, 100, y, playerColor)

	if showTeams {
		drawTeamEntry(screen, entry, y, playerColor, entryColor)
	} else {
		drawDeathmatchEntry(screen, entry, y, entryColor)
	}

	kdRatio := calculateKDRatio(entry.Frags, entry.Deaths)
	text.Draw(screen, fmt.Sprintf("%.2f", kdRatio), basicfont.Face7x13, 500, y, entryColor)

	return y + 18
}

// drawTeamEntry renders team-specific stats for an entry.
func drawTeamEntry(screen *ebiten.Image, entry ScoreboardEntry, y int, playerColor, entryColor color.RGBA) {
	teamName := "Red"
	if entry.Team == 1 {
		teamName = "Blue"
	}
	text.Draw(screen, teamName, basicfont.Face7x13, 250, y, playerColor)
	text.Draw(screen, fmt.Sprintf("%d", entry.Frags), basicfont.Face7x13, 350, y, entryColor)
	text.Draw(screen, fmt.Sprintf("%d", entry.Deaths), basicfont.Face7x13, 400, y, entryColor)
	text.Draw(screen, fmt.Sprintf("%d", entry.Assists), basicfont.Face7x13, 450, y, entryColor)
}

// drawDeathmatchEntry renders deathmatch-specific stats for an entry.
func drawDeathmatchEntry(screen *ebiten.Image, entry ScoreboardEntry, y int, entryColor color.RGBA) {
	text.Draw(screen, fmt.Sprintf("%d", entry.Frags), basicfont.Face7x13, 300, y, entryColor)
	text.Draw(screen, fmt.Sprintf("%d", entry.Deaths), basicfont.Face7x13, 400, y, entryColor)
}

// calculateKDRatio computes the kill/death ratio.
func calculateKDRatio(frags, deaths int) float64 {
	if deaths > 0 {
		return float64(frags) / float64(deaths)
	}
	if frags > 0 {
		return float64(frags)
	}
	return 0.0
}
