package main

import (
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/ai"
	"github.com/opd-ai/violence/pkg/ammo"
	"github.com/opd-ai/violence/pkg/audio"
	"github.com/opd-ai/violence/pkg/automap"
	"github.com/opd-ai/violence/pkg/bsp"
	"github.com/opd-ai/violence/pkg/camera"
	"github.com/opd-ai/violence/pkg/class"
	"github.com/opd-ai/violence/pkg/combat"
	"github.com/opd-ai/violence/pkg/config"
	"github.com/opd-ai/violence/pkg/door"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/input"
	"github.com/opd-ai/violence/pkg/loot"
	"github.com/opd-ai/violence/pkg/progression"
	"github.com/opd-ai/violence/pkg/raycaster"
	"github.com/opd-ai/violence/pkg/render"
	"github.com/opd-ai/violence/pkg/rng"
	"github.com/opd-ai/violence/pkg/save"
	"github.com/opd-ai/violence/pkg/status"
	"github.com/opd-ai/violence/pkg/tutorial"
	"github.com/opd-ai/violence/pkg/ui"
	"github.com/opd-ai/violence/pkg/weapon"
)

// GameState represents the current game state.
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateLoading
)

// Game implements ebiten.Game for the VIOLENCE raycasting FPS.
type Game struct {
	state          GameState
	world          *engine.World
	camera         *camera.Camera
	raycaster      *raycaster.Raycaster
	renderer       *render.Renderer
	input          *input.Manager
	audioEngine    *audio.Engine
	hud            *ui.HUD
	menuManager    *ui.MenuManager
	loadingScreen  *ui.LoadingScreen
	tutorialSystem *tutorial.Tutorial
	rng            *rng.RNG
	bspGenerator   *bsp.Generator
	currentMap     [][]int
	genreID        string
	seed           uint64
	automap        *automap.Map
	keycards       map[string]bool
	automapVisible bool

	// v2.0 systems
	arsenal      *weapon.Arsenal
	ammoPool     *ammo.Pool
	combatSystem *combat.System
	statusReg    *status.Registry
	lootTable    *loot.LootTable
	progression  *progression.Progression
	aiAgents     []*ai.Agent
	playerClass  string
}

// NewGame creates and initializes a new game instance.
func NewGame() *Game {
	// Initialize RNG with time-based seed
	seed := uint64(time.Now().UnixNano())
	gameRNG := rng.NewRNG(seed)

	// Initialize camera
	cam := camera.NewCamera(config.C.FOV)
	cam.X = 5.0
	cam.Y = 5.0
	cam.DirX = 1.0
	cam.DirY = 0.0

	// Initialize raycaster and renderer
	rc := raycaster.NewRaycaster(config.C.FOV, config.C.InternalWidth, config.C.InternalHeight)
	rend := render.NewRenderer(config.C.InternalWidth, config.C.InternalHeight, rc)

	g := &Game{
		state:          StateMenu,
		world:          engine.NewWorld(),
		camera:         cam,
		raycaster:      rc,
		renderer:       rend,
		input:          input.NewManager(),
		audioEngine:    audio.NewEngine(),
		hud:            ui.NewHUD(),
		menuManager:    ui.NewMenuManager(),
		loadingScreen:  ui.NewLoadingScreen(),
		tutorialSystem: tutorial.NewTutorial(),
		rng:            gameRNG,
		genreID:        "fantasy",
		seed:           seed,
		keycards:       make(map[string]bool),
		automapVisible: false,
		arsenal:        weapon.NewArsenal(),
		ammoPool:       ammo.NewPool(),
		combatSystem:   combat.NewSystem(),
		statusReg:      status.NewRegistry(),
		lootTable:      loot.NewLootTable(),
		progression:    progression.NewProgression(),
		aiAgents:       make([]*ai.Agent, 0),
		playerClass:    class.Grunt,
	}

	// Initialize BSP generator
	g.bspGenerator = bsp.NewGenerator(64, 64, g.rng)
	g.bspGenerator.SetGenre(g.genreID)

	// Show main menu
	g.menuManager.Show(ui.MenuTypeMain)

	return g
}

// Update handles game logic updates.
func (g *Game) Update() error {
	// Update input manager
	g.input.Update()

	switch g.state {
	case StateMenu:
		return g.updateMenu()
	case StatePlaying:
		return g.updatePlaying()
	case StatePaused:
		return g.updatePaused()
	case StateLoading:
		return g.updateLoading()
	}

	return nil
}

// updateMenu handles menu navigation and actions.
func (g *Game) updateMenu() error {
	if g.input.IsJustPressed(input.ActionMoveForward) {
		g.menuManager.MoveUp()
	}
	if g.input.IsJustPressed(input.ActionMoveBackward) {
		g.menuManager.MoveDown()
	}
	if g.input.IsJustPressed(input.ActionFire) || g.input.IsJustPressed(input.ActionInteract) {
		action := g.menuManager.Select()
		g.handleMenuAction(action)
	}
	if g.input.IsJustPressed(input.ActionPause) {
		g.menuManager.Back()
	}
	return nil
}

// handleMenuAction processes menu selections.
func (g *Game) handleMenuAction(action string) {
	switch action {
	case "new_game":
		g.menuManager.Show(ui.MenuTypeDifficulty)
	case "difficulty_selected":
		g.menuManager.Show(ui.MenuTypeGenre)
	case "genre_selected":
		// Genre was already set by MenuManager.Select() which calls SelectGenre()
		g.genreID = g.menuManager.GetSelectedGenre()
		g.startNewGame()
	case "load_game":
		// Load from slot 1 (first manual save)
		g.loadGame(1)
	case "settings":
		g.menuManager.Show(ui.MenuTypeSettings)
	case "quit":
		// Exit game
	}
}

// startNewGame initializes a new game session.
func (g *Game) startNewGame() {
	// Show loading screen
	g.state = StateLoading
	g.loadingScreen.Show(g.seed, "Generating level...")

	// Generate level
	g.bspGenerator.SetGenre(g.genreID)
	_, tiles := g.bspGenerator.Generate()
	g.currentMap = tiles
	g.raycaster.SetMap(tiles)

	// Initialize automap
	if len(tiles) > 0 && len(tiles[0]) > 0 {
		g.automap = automap.NewMap(len(tiles[0]), len(tiles))
	}

	// Set genre for all systems
	g.world.SetGenre(g.genreID)
	g.renderer.SetGenre(g.genreID)
	g.raycaster.SetGenre(g.genreID)
	camera.SetGenre(g.genreID)
	g.audioEngine.SetGenre(g.genreID)
	tutorial.SetGenre(g.genreID)
	automap.SetGenre(g.genreID)
	door.SetGenre(g.genreID)
	g.arsenal.SetGenre(g.genreID)
	ammo.SetGenre(g.genreID)
	g.combatSystem.SetGenre(g.genreID)
	status.SetGenre(g.genreID)
	loot.SetGenre(g.genreID)
	progression.SetGenre(g.genreID)
	class.SetGenre(g.genreID)
	ai.SetGenre(g.genreID)
	// Reset player position to a safe starting location
	g.camera.X = 5.0
	g.camera.Y = 5.0
	g.camera.DirX = 1.0
	g.camera.DirY = 0.0
	g.camera.Pitch = 0.0

	// Initialize player stats
	g.hud.Health = 100
	g.hud.Armor = 0

	// Initialize starting ammo
	g.ammoPool.Add("bullets", 50)
	g.ammoPool.Add("shells", 8)
	g.ammoPool.Add("cells", 20)
	g.ammoPool.Add("rockets", 0)

	// Set initial ammo display
	currentWeapon := g.arsenal.GetCurrentWeapon()
	g.hud.Ammo = g.ammoPool.Get(currentWeapon.AmmoType)

	// Reset keycards
	g.keycards = make(map[string]bool)
	g.automapVisible = false

	// Reset v2.0 systems
	g.progression = progression.NewProgression()
	progression.SetGenre(g.genreID)
	g.statusReg = status.NewRegistry()
	status.SetGenre(g.genreID)

	// Spawn AI enemies (simple placement for now)
	g.aiAgents = make([]*ai.Agent, 0)
	ai.SetGenre(g.genreID)
	for i := 0; i < 3; i++ {
		agent := ai.NewAgent("enemy_"+string(rune(i+'0')), float64(10+i*5), float64(10+i*3))
		g.aiAgents = append(g.aiAgents, agent)
	}

	// Play music
	g.audioEngine.PlayMusic("theme", 0.5)

	// Hide loading screen and start playing
	g.loadingScreen.Hide()
	g.state = StatePlaying

	// Show movement tutorial
	g.tutorialSystem.ShowPrompt(tutorial.PromptMovement, tutorial.GetMessage(tutorial.PromptMovement))
}

// loadGame loads a saved game state.
func (g *Game) loadGame(slot int) {
	state, err := save.Load(slot)
	if err != nil {
		return
	}

	g.genreID = state.Genre
	g.seed = uint64(state.Seed)
	g.rng.Seed(g.seed)

	// Restore map
	g.currentMap = state.Map.Tiles
	g.raycaster.SetMap(g.currentMap)

	// Restore camera/player
	g.camera.X = state.Player.X
	g.camera.Y = state.Player.Y
	g.camera.DirX = state.Player.DirX
	g.camera.DirY = state.Player.DirY
	g.camera.Pitch = state.Player.Pitch

	// Restore HUD
	g.hud.Health = state.Player.Health
	g.hud.Armor = state.Player.Armor
	g.hud.Ammo = state.Player.Ammo

	// Set genre for all systems
	g.world.SetGenre(g.genreID)
	g.renderer.SetGenre(g.genreID)
	g.raycaster.SetGenre(g.genreID)

	g.state = StatePlaying
	g.menuManager.Hide()
}

// updatePlaying handles gameplay updates.
func (g *Game) updatePlaying() error {
	// Check for pause
	if g.input.IsJustPressed(input.ActionPause) {
		g.state = StatePaused
		g.menuManager.Show(ui.MenuTypePause)
		return nil
	}

	// Toggle automap
	if g.input.IsJustPressed(input.ActionAutomap) {
		g.automapVisible = !g.automapVisible
	}

	// Check for door interaction
	if g.input.IsJustPressed(input.ActionInteract) {
		g.tryInteractDoor()
	}

	// Weapon firing
	if g.input.IsJustPressed(input.ActionFire) {
		currentWeapon := g.arsenal.GetCurrentWeapon()
		if currentWeapon.Name != "" { // Check if weapon is valid
			ammoType := currentWeapon.AmmoType
			availableAmmo := g.ammoPool.Get(ammoType)

			if currentWeapon.Type == weapon.TypeMelee || availableAmmo > 0 {
				// Create raycast function wrapper
				raycastFn := func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64) {
					// Simple raycast against enemies
					for i, agent := range g.aiAgents {
						if agent.Health <= 0 {
							continue
						}
						// Check if ray hits this agent (simplified sphere collision)
						agentDist := (agent.X-x)*(agent.X-x) + (agent.Y-y)*(agent.Y-y)
						if agentDist < maxDist*maxDist {
							// Check if agent is in ray direction
							toAgentX := agent.X - x
							toAgentY := agent.Y - y
							dot := toAgentX*dx + toAgentY*dy
							if dot > 0 { // Agent is in front
								return true, agentDist, agent.X, agent.Y, uint64(i + 1)
							}
						}
					}
					return false, 0, 0, 0, 0
				}

				// Fire weapon and get hit results
				hitResults := g.arsenal.Fire(g.camera.X, g.camera.Y, g.camera.DirX, g.camera.DirY, raycastFn)

				// Consume ammo for non-melee weapons
				if currentWeapon.Type != weapon.TypeMelee {
					g.ammoPool.Consume(ammoType, 1)
					g.hud.Ammo = g.ammoPool.Get(ammoType)
				}

				// Apply damage to hit enemies
				for _, hitResult := range hitResults {
					if hitResult.Hit && hitResult.EntityID > 0 {
						agentIdx := int(hitResult.EntityID - 1)
						if agentIdx >= 0 && agentIdx < len(g.aiAgents) {
							agent := g.aiAgents[agentIdx]
							if agent.Health > 0 {
								// Apply damage
								agent.Health -= currentWeapon.Damage

								if agent.Health <= 0 {
									// Enemy died - award XP
									g.progression.AddXP(50)
									// TODO: spawn loot drops
								}
							}
						}
					}
				}

				// Play weapon sound
				g.audioEngine.PlaySFX("weapon_fire", g.camera.X, g.camera.Y)
			}
		}
	}

	// Update weapon animations
	g.arsenal.Update()

	// Update AI agents (simplified for initial integration)
	for _, agent := range g.aiAgents {
		if agent.Health > 0 {
			// Simple AI: attack player if in range
			dx := g.camera.X - agent.X
			dy := g.camera.Y - agent.Y
			distSq := dx*dx + dy*dy

			if distSq < 100 && agent.Cooldown <= 0 { // Within attack range and cooled down
				// Enemy attacks player
				damage := agent.Damage
				healthDamage := damage

				// Apply damage to player (simplified)
				if g.hud.Armor > 0 {
					armorDamage := damage * 0.5
					g.hud.Armor -= int(armorDamage)
					if g.hud.Armor < 0 {
						healthDamage = -float64(g.hud.Armor)
						g.hud.Armor = 0
					} else {
						healthDamage = damage * 0.5
					}
				}

				g.hud.Health -= int(healthDamage)
				agent.Cooldown = 60 // 1 second at 60 TPS
				g.audioEngine.PlaySFX("enemy_attack", agent.X, agent.Y)

				// Show damage indicator
				g.hud.ShowMessage("Taking damage!")

				// Check for player death
				if g.hud.Health <= 0 {
					g.hud.Health = 0
					// TODO: handle player death
				}
			}

			// Decrement cooldowns
			if agent.Cooldown > 0 {
				agent.Cooldown--
			}
		}
	}

	// Update status effects
	g.statusReg.Tick()

	// Progression updates happen automatically
	// TODO: Add Level and XP to HUD display

	// Movement speed (units per frame at 60 TPS)
	moveSpeed := 0.05
	rotSpeed := 0.03

	deltaX := 0.0
	deltaY := 0.0
	deltaDirX := 0.0
	deltaDirY := 0.0
	deltaPitch := 0.0

	// Keyboard: Forward/backward movement
	if g.input.IsPressed(input.ActionMoveForward) {
		deltaX += g.camera.DirX * moveSpeed
		deltaY += g.camera.DirY * moveSpeed
	}
	if g.input.IsPressed(input.ActionMoveBackward) {
		deltaX -= g.camera.DirX * moveSpeed
		deltaY -= g.camera.DirY * moveSpeed
	}

	// Keyboard: Strafing
	if g.input.IsPressed(input.ActionStrafeLeft) {
		deltaX += g.camera.DirY * moveSpeed
		deltaY -= g.camera.DirX * moveSpeed
	}
	if g.input.IsPressed(input.ActionStrafeRight) {
		deltaX -= g.camera.DirY * moveSpeed
		deltaY += g.camera.DirX * moveSpeed
	}

	// Gamepad: Left stick movement
	leftX, leftY := g.input.GamepadLeftStick()
	deadzone := 0.15
	if leftX*leftX+leftY*leftY > deadzone*deadzone {
		deltaX += (g.camera.DirX*leftY - g.camera.DirY*leftX) * moveSpeed
		deltaY += (g.camera.DirY*leftY + g.camera.DirX*leftX) * moveSpeed
	}

	// Keyboard: Rotation
	if g.input.IsPressed(input.ActionTurnLeft) {
		g.camera.Rotate(-rotSpeed)
	}
	if g.input.IsPressed(input.ActionTurnRight) {
		g.camera.Rotate(rotSpeed)
	}

	// Mouse look
	mouseDX, mouseDY := g.input.MouseDelta()
	if mouseDX != 0 || mouseDY != 0 {
		sensitivity := config.C.MouseSensitivity * 0.002
		g.camera.Rotate(mouseDX * sensitivity)
		deltaPitch = -mouseDY * sensitivity * 10.0
	}

	// Gamepad: Right stick camera
	rightX, rightY := g.input.GamepadRightStick()
	if rightX*rightX+rightY*rightY > deadzone*deadzone {
		g.camera.Rotate(rightX * rotSpeed * 1.5)
		deltaPitch = -rightY * rotSpeed * 15.0
	}

	// Collision detection (simple)
	newX := g.camera.X + deltaX
	newY := g.camera.Y + deltaY
	if g.isWalkable(newX, newY) {
		g.camera.Update(deltaX, deltaY, deltaDirX, deltaDirY, deltaPitch)
		if g.automap != nil {
			g.automap.Reveal(int(newX), int(newY))
		}
	}

	// Update ECS world
	g.world.Update()

	// Update audio listener position
	g.audioEngine.SetListenerPosition(g.camera.X, g.camera.Y)

	// Tutorial completion checks
	if deltaX != 0 || deltaY != 0 {
		if g.tutorialSystem.Active && g.tutorialSystem.Type == tutorial.PromptMovement {
			g.tutorialSystem.Complete()
		}
	}

	return nil
}

// isWalkable checks if a position is walkable (no collision).
func (g *Game) isWalkable(x, y float64) bool {
	if g.currentMap == nil || len(g.currentMap) == 0 {
		return true
	}
	mapX := int(x)
	mapY := int(y)
	if mapY < 0 || mapY >= len(g.currentMap) || mapX < 0 || mapX >= len(g.currentMap[0]) {
		return false
	}
	tile := g.currentMap[mapY][mapX]
	return tile == bsp.TileFloor || tile == bsp.TileEmpty
}

// tryInteractDoor checks if player is facing a door and attempts to open it.
func (g *Game) tryInteractDoor() {
	checkDist := 1.5
	checkX := g.camera.X + g.camera.DirX*checkDist
	checkY := g.camera.Y + g.camera.DirY*checkDist
	mapX := int(checkX)
	mapY := int(checkY)

	if g.currentMap == nil || len(g.currentMap) == 0 {
		return
	}
	if mapY < 0 || mapY >= len(g.currentMap) || mapX < 0 || mapX >= len(g.currentMap[0]) {
		return
	}

	tile := g.currentMap[mapY][mapX]
	if tile == bsp.TileDoor {
		requiredColor := g.getDoorColor(mapX, mapY)
		if requiredColor == "" || g.keycards[requiredColor] {
			g.currentMap[mapY][mapX] = bsp.TileFloor
			g.raycaster.SetMap(g.currentMap)
			g.audioEngine.PlaySFX("door_open", float64(mapX), float64(mapY))
		} else {
			g.hud.ShowMessage("Need " + requiredColor + " keycard")
		}
	}
}

// getDoorColor returns the keycard color required for a door (stub - would be from door metadata).
func (g *Game) getDoorColor(x, y int) string {
	return ""
}

// updatePaused handles pause menu updates.
func (g *Game) updatePaused() error {
	if g.input.IsJustPressed(input.ActionPause) {
		g.state = StatePlaying
		g.menuManager.Hide()
		return nil
	}

	if g.input.IsJustPressed(input.ActionMoveForward) {
		g.menuManager.MoveUp()
	}
	if g.input.IsJustPressed(input.ActionMoveBackward) {
		g.menuManager.MoveDown()
	}
	if g.input.IsJustPressed(input.ActionFire) || g.input.IsJustPressed(input.ActionInteract) {
		action := g.menuManager.Select()
		g.handlePauseAction(action)
	}

	return nil
}

// handlePauseAction processes pause menu selections.
func (g *Game) handlePauseAction(action string) {
	switch action {
	case "resume":
		g.state = StatePlaying
		g.menuManager.Hide()
	case "save":
		// Save to slot 1
		g.saveGame(1)
	case "load":
		g.loadGame(1)
	case "settings":
		g.menuManager.Show(ui.MenuTypeSettings)
	case "quit_to_menu":
		g.state = StateMenu
		g.menuManager.Show(ui.MenuTypeMain)
	}
}

// saveGame saves the current game state.
func (g *Game) saveGame(slot int) {
	state := &save.GameState{
		Version:   "1.0.0",
		Seed:      int64(g.seed),
		Timestamp: time.Now(),
		Genre:     g.genreID,
		Player: save.Player{
			X:      g.camera.X,
			Y:      g.camera.Y,
			DirX:   g.camera.DirX,
			DirY:   g.camera.DirY,
			Pitch:  g.camera.Pitch,
			Health: g.hud.Health,
			Armor:  g.hud.Armor,
			Ammo:   g.hud.Ammo,
		},
		Map: save.Map{
			Width:  len(g.currentMap[0]),
			Height: len(g.currentMap),
			Tiles:  g.currentMap,
		},
		Inventory: save.Inventory{Items: []save.Item{}},
	}
	save.Save(slot, state)
}

// updateLoading handles loading screen updates.
func (g *Game) updateLoading() error {
	// Loading is instantaneous for now, but this could be async
	return nil
}

// Draw renders the game to the screen.
func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying:
		g.drawPlaying(screen)
	case StatePaused:
		g.drawPaused(screen)
	case StateLoading:
		g.drawLoading(screen)
	}
}

// drawMenu renders the menu screen.
func (g *Game) drawMenu(screen *ebiten.Image) {
	ui.DrawMenu(screen, g.menuManager)
}

// drawPlaying renders the game world and HUD.
func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Render 3D world
	g.renderer.Render(screen, g.camera.X, g.camera.Y, g.camera.DirX, g.camera.DirY, g.camera.Pitch)

	// Render automap overlay if visible
	if g.automapVisible && g.automap != nil {
		g.drawAutomap(screen)
	}

	// Render HUD
	g.hud.Update()
	ui.DrawHUD(screen, g.hud)

	// Render tutorial prompts
	if g.tutorialSystem.Active {
		ui.DrawTutorial(screen, g.tutorialSystem.Current)
	}
}

// drawAutomap renders the automap overlay.
func (g *Game) drawAutomap(screen *ebiten.Image) {
	bounds := screen.Bounds()
	w := float32(bounds.Dx())

	overlayX := w*0.75 - 80
	overlayY := float32(20.0)
	overlayW := float32(150.0)
	overlayH := float32(150.0)

	vector.DrawFilledRect(screen, overlayX, overlayY, overlayW, overlayH, color.RGBA{0, 0, 0, 180}, false)
	vector.StrokeRect(screen, overlayX, overlayY, overlayW, overlayH, 1, color.RGBA{100, 100, 100, 255}, false)

	if g.automap == nil || g.currentMap == nil {
		return
	}

	scaleX := overlayW / float32(g.automap.Width)
	scaleY := overlayH / float32(g.automap.Height)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	for y := 0; y < g.automap.Height && y < len(g.currentMap); y++ {
		for x := 0; x < g.automap.Width && x < len(g.currentMap[0]); x++ {
			if !g.automap.Revealed[y][x] {
				continue
			}
			tile := g.currentMap[y][x]
			tileX := overlayX + float32(x)*scale
			tileY := overlayY + float32(y)*scale

			var tileColor color.RGBA
			switch tile {
			case bsp.TileWall:
				tileColor = color.RGBA{150, 150, 150, 255}
			case bsp.TileFloor, bsp.TileEmpty:
				tileColor = color.RGBA{50, 50, 50, 255}
			case bsp.TileDoor:
				tileColor = color.RGBA{100, 100, 200, 255}
			case bsp.TileSecret:
				tileColor = color.RGBA{200, 200, 50, 255}
			default:
				tileColor = color.RGBA{80, 80, 80, 255}
			}
			vector.DrawFilledRect(screen, tileX, tileY, scale, scale, tileColor, false)
		}
	}

	playerX := overlayX + float32(g.camera.X)*scale
	playerY := overlayY + float32(g.camera.Y)*scale
	vector.DrawFilledCircle(screen, playerX, playerY, 2, color.RGBA{255, 255, 0, 255}, false)

	dirLen := scale * 3
	vector.StrokeLine(screen, playerX, playerY, playerX+float32(g.camera.DirX)*dirLen, playerY+float32(g.camera.DirY)*dirLen, 1, color.RGBA{255, 255, 0, 255}, false)
}

// drawPaused renders the paused game state.
func (g *Game) drawPaused(screen *ebiten.Image) {
	// Draw frozen game world
	g.renderer.Render(screen, g.camera.X, g.camera.Y, g.camera.DirX, g.camera.DirY, g.camera.Pitch)

	// Draw pause menu overlay
	ui.DrawMenu(screen, g.menuManager)
}

// drawLoading renders the loading screen.
func (g *Game) drawLoading(screen *ebiten.Image) {
	ui.DrawLoadingScreen(screen, g.loadingScreen)
}

// Layout returns the game's internal resolution.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.C.InternalWidth, config.C.InternalHeight
}

func main() {
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(config.C.WindowWidth, config.C.WindowHeight)
	ebiten.SetVsyncEnabled(config.C.VSync)
	ebiten.SetFullscreen(config.C.FullScreen)
	ebiten.SetWindowTitle("VIOLENCE")

	// Set TPS cap (0 = unlimited, 60 = default)
	if config.C.MaxTPS > 0 {
		ebiten.SetTPS(config.C.MaxTPS)
	}

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
