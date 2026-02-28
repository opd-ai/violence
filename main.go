package main

import (
	"image/color"
	"log"
	"math"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
	"github.com/opd-ai/violence/pkg/crafting"
	"github.com/opd-ai/violence/pkg/destruct"
	"github.com/opd-ai/violence/pkg/door"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/event"
	"github.com/opd-ai/violence/pkg/federation"
	"github.com/opd-ai/violence/pkg/input"
	"github.com/opd-ai/violence/pkg/inventory"
	"github.com/opd-ai/violence/pkg/lighting"
	"github.com/opd-ai/violence/pkg/loot"
	"github.com/opd-ai/violence/pkg/lore"
	"github.com/opd-ai/violence/pkg/minigame"
	"github.com/opd-ai/violence/pkg/mod"
	"github.com/opd-ai/violence/pkg/network"
	"github.com/opd-ai/violence/pkg/particle"
	"github.com/opd-ai/violence/pkg/progression"
	"github.com/opd-ai/violence/pkg/props"
	"github.com/opd-ai/violence/pkg/quest"
	"github.com/opd-ai/violence/pkg/raycaster"
	"github.com/opd-ai/violence/pkg/render"
	"github.com/opd-ai/violence/pkg/rng"
	"github.com/opd-ai/violence/pkg/save"
	"github.com/opd-ai/violence/pkg/secret"
	"github.com/opd-ai/violence/pkg/shop"
	"github.com/opd-ai/violence/pkg/skills"
	"github.com/opd-ai/violence/pkg/squad"
	"github.com/opd-ai/violence/pkg/status"
	"github.com/opd-ai/violence/pkg/texture"
	"github.com/opd-ai/violence/pkg/tutorial"
	"github.com/opd-ai/violence/pkg/ui"
	"github.com/opd-ai/violence/pkg/upgrade"
	"github.com/opd-ai/violence/pkg/weapon"
)

// GameState represents the current game state.
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateLoading
	StateShop
	StateCrafting
	StateSkills
	StateMods
	StateMultiplayer
	StateCodex
	StateMinigame
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

	// v3.0 systems
	textureAtlas    *texture.Atlas
	lightMap        *lighting.SectorLightMap
	particleSystem  *particle.ParticleSystem
	weatherEmitter  *particle.WeatherEmitter
	postProcessor   *render.PostProcessor
	currentBSPTree  *bsp.Node
	animationTicker int

	// v4.0 systems
	destructibleSystem *destruct.System
	squadCompanions    *squad.Squad
	questTracker       *quest.Tracker
	alarmTrigger       *event.AlarmTrigger
	lockdownTrigger    *event.TimedLockdown
	bossArena          *event.BossArenaEvent
	levelStartTime     time.Time

	// v5.0+ systems
	craftingMenu    *crafting.CraftingMenu
	scrapStorage    *crafting.ScrapStorage
	shopCredits     *shop.Credit
	shopInventory   *shop.ShopInventory
	shopArmory      *shop.Shop
	craftingResult  string
	skillManager    *skills.Manager
	modLoader       *mod.Loader
	networkMode     bool
	multiplayerMgr  interface{} // Can be *network.FFAMatch, *network.TeamMatch, etc.
	skillsTreeIdx   int         // Active tree tab in skills UI
	skillsNodeIdx   int         // Selected node in skills UI
	mpStatusMsg     string      // Multiplayer status message
	mpSelectedMode  int         // Selected multiplayer mode
	playerInventory *inventory.Inventory
	propsManager    *props.Manager
	loreCodex       *lore.Codex
	loreGenerator   *lore.Generator
	loreItems       []*lore.LoreItem
	codexScrollIdx  int // Scroll position for codex UI

	// Minigame system
	activeMinigame     minigame.MiniGame
	minigameDoorX      int // Door coordinates for minigame context
	minigameDoorY      int
	minigameType       string // "lockpick", "hack", "circuit", "code"
	previousState      GameState
	minigameInputTimer int // Frame timer for input delay

	// Secret wall system
	secretManager *secret.Manager

	// Weapon upgrade system
	upgradeManager *upgrade.Manager

	// Weapon mastery system
	masteryManager *weapon.MasteryManager

	// Federation system
	federationHub *federation.FederationHub
	serverBrowser []*federation.ServerAnnouncement // Cached server list
	browserIdx    int                              // Selected server in browser
	useFederation bool                             // Whether to use federation matchmaking
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
		// v3.0 systems
		textureAtlas:    texture.NewAtlas(seed),
		lightMap:        lighting.NewSectorLightMap(64, 64, 0.3),
		particleSystem:  particle.NewParticleSystem(1024, int64(seed)),
		postProcessor:   render.NewPostProcessor(config.C.InternalWidth, config.C.InternalHeight, int64(seed)),
		animationTicker: 0,
		// v4.0 systems
		destructibleSystem: destruct.NewSystem(),
		squadCompanions:    squad.NewSquad(3), // Max 3 squad members
		questTracker:       quest.NewTracker(),
		playerInventory:    inventory.NewInventory(),
		propsManager:       props.NewManager(),
		loreCodex:          lore.NewCodex(),
		loreGenerator:      lore.NewGenerator(int64(seed)),
		loreItems:          make([]*lore.LoreItem, 0),
		codexScrollIdx:     0,
		secretManager:      secret.NewManager(64), // Map width for secret key calculation
		upgradeManager:     upgrade.NewManager(),
		masteryManager:     weapon.NewMasteryManager(),
		federationHub:      federation.NewFederationHub(),
		serverBrowser:      make([]*federation.ServerAnnouncement, 0),
		browserIdx:         0,
		useFederation:      false,
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

	// Manage cursor capture: locked during gameplay, visible in menus
	switch g.state {
	case StatePlaying:
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	default:
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
	}

	switch g.state {
	case StateMenu:
		return g.updateMenu()
	case StatePlaying:
		return g.updatePlaying()
	case StatePaused:
		return g.updatePaused()
	case StateLoading:
		return g.updateLoading()
	case StateShop:
		return g.updateShop()
	case StateCrafting:
		return g.updateCrafting()
	case StateSkills:
		return g.updateSkills()
	case StateMods:
		return g.updateMods()
	case StateMultiplayer:
		return g.updateMultiplayer()
	case StateCodex:
		return g.updateCodex()
	case StateMinigame:
		return g.updateMinigame()
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
	bspTree, tiles := g.bspGenerator.Generate()
	g.currentMap = tiles
	g.currentBSPTree = bspTree
	g.raycaster.SetMap(tiles)

	// Initialize automap
	if len(tiles) > 0 && len(tiles[0]) > 0 {
		g.automap = automap.NewMap(len(tiles[0]), len(tiles))
	}

	// Initialize v3.0 systems with correct dimensions
	g.lightMap = lighting.NewSectorLightMap(len(tiles[0]), len(tiles), 0.3)
	g.weatherEmitter = particle.NewWeatherEmitter(g.particleSystem, g.genreID, 0, 0, float64(len(tiles[0])), float64(len(tiles)))

	// Set genre for all systems (Step 29: SetGenre cascade)
	g.setGenre(g.genreID)

	// Place decorative props in rooms (v3.0 enhancement)
	g.propsManager.Clear()
	g.propsManager.SetGenre(g.genreID)
	rooms := bsp.GetRooms(bspTree)
	for _, room := range rooms {
		propRoom := &props.Room{X: room.X, Y: room.Y, W: room.W, H: room.H}
		g.propsManager.PlaceProps(propRoom, 0.2, g.seed+uint64(room.X*1000+room.Y))
	}

	// Generate and place lore items in rooms
	g.loreItems = make([]*lore.LoreItem, 0)
	g.loreGenerator.SetGenre(g.genreID)
	loreItemsPerLevel := 5 + len(rooms)/3 // Scale with level size
	if loreItemsPerLevel > len(rooms) {
		loreItemsPerLevel = len(rooms)
	}
	for i := 0; i < loreItemsPerLevel && i < len(rooms); i++ {
		room := rooms[i]
		itemX := float64(room.X+1) + g.rng.Float64()*float64(room.W-2)
		itemY := float64(room.Y+1) + g.rng.Float64()*float64(room.H-2)
		itemType := lore.LoreItemType(i % 4)
		context := g.getLoreContext(*room)
		itemID := "lore_" + g.genreID + "_" + string(rune(i+'0'))
		loreItem := g.loreGenerator.GenerateLoreItem(itemID, itemType, itemX, itemY, context)
		g.loreItems = append(g.loreItems, &loreItem)
		codexEntry := g.loreGenerator.Generate(loreItem.CodexID)
		g.loreCodex.AddEntry(codexEntry)
	}

	// Scan map for secret walls and register them
	g.secretManager = secret.NewManager(len(tiles[0]))
	for y := 0; y < len(tiles); y++ {
		for x := 0; x < len(tiles[y]); x++ {
			if tiles[y][x] == bsp.TileSecret {
				// Determine slide direction based on neighboring floor tiles
				dir := secret.DirNorth
				if y+1 < len(tiles) && tiles[y+1][x] == bsp.TileFloor {
					dir = secret.DirSouth
				} else if y > 0 && tiles[y-1][x] == bsp.TileFloor {
					dir = secret.DirNorth
				} else if x+1 < len(tiles[y]) && tiles[y][x+1] == bsp.TileFloor {
					dir = secret.DirEast
				} else if x > 0 && tiles[y][x-1] == bsp.TileFloor {
					dir = secret.DirWest
				}
				g.secretManager.Add(x, y, dir)
			}
		}
	}

	// Reset player position to a safe starting location
	// Find the center of the first BSP room as spawn point
	spawnX, spawnY := 5.0, 5.0
	if len(rooms) > 0 {
		spawnX = float64(rooms[0].X + rooms[0].W/2)
		spawnY = float64(rooms[0].Y + rooms[0].H/2)
	} else {
		// Fallback: scan for any walkable tile near (5,5)
		for dy := 0; dy < len(tiles); dy++ {
			for dx := 0; dx < len(tiles[0]); dx++ {
				if isWalkableTile(tiles[dy][dx]) {
					spawnX = float64(dx) + 0.5
					spawnY = float64(dy) + 0.5
					goto foundSpawn
				}
			}
		}
	foundSpawn:
	}
	g.camera.X = spawnX
	g.camera.Y = spawnY
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

	// Spawn destructible objects (barrels, crates)
	g.destructibleSystem = destruct.NewSystem()
	destruct.SetGenre(g.genreID)
	// Spawn some barrels and crates in the level
	for i := 0; i < 5; i++ {
		// Explosive barrels
		barrel := destruct.NewDestructibleObject(
			"barrel_"+string(rune(i+'0')),
			"barrel",
			50.0, // health
			float64(15+i*4),
			float64(8+i*2),
			true, // explosive
		)
		barrel.AddDropItem("ammo_shells")
		g.destructibleSystem.Add(&barrel.Destructible)

		// Non-explosive crates
		crate := destruct.NewDestructibleObject(
			"crate_"+string(rune(i+'0')),
			"crate",
			30.0, // health
			float64(12+i*3),
			float64(12+i*3),
			false, // not explosive
		)
		crate.AddDropItem("health_small")
		g.destructibleSystem.Add(&crate.Destructible)
	}

	// Initialize squad companions
	g.squadCompanions = squad.NewSquad(3)
	squad.SetGenre(g.genreID)
	// Spawn 2 squad companions near player
	g.squadCompanions.AddMember("companion_1", "grunt", "assault_rifle", g.camera.X-2, g.camera.Y+1, g.seed)
	g.squadCompanions.AddMember("companion_2", "medic", "pistol", g.camera.X-2, g.camera.Y-1, g.seed)

	// Initialize quest tracker with level objectives
	g.questTracker = quest.NewTracker()
	g.questTracker.SetGenre(g.genreID)

	// Convert BSP rooms to quest rooms
	questRooms := make([]quest.Room, len(rooms))
	for i, r := range rooms {
		questRooms[i] = quest.Room{X: r.X, Y: r.Y, Width: r.W, Height: r.H}
	}

	// Find exit position (furthest room from player spawn)
	exitPos := g.findExitPosition(rooms, g.camera.X, g.camera.Y)

	layout := quest.LevelLayout{
		Width:       len(tiles[0]),
		Height:      len(tiles),
		ExitPos:     exitPos,
		SecretCount: len(g.secretManager.GetAll()),
		Rooms:       questRooms,
	}
	g.questTracker.GenerateWithLayout(g.seed, layout)

	// Initialize event triggers
	g.alarmTrigger = event.NewAlarmTrigger("alarm_1", 30.0)         // 30 second alarm
	g.lockdownTrigger = event.NewTimedLockdown("lockdown_1", 180.0) // 3 minute escape timer
	// Find a room for boss arena (use center of map for now)
	centerX := len(tiles[0]) / 2
	centerY := len(tiles) / 2
	g.bossArena = event.NewBossArenaEvent("boss_1", "center_room", 3, 5.0) // 3 waves, 5 sec between
	// Trigger boss arena when player enters center region
	if int(g.camera.X) == centerX && int(g.camera.Y) == centerY {
		g.bossArena.Trigger()
	}

	// Set event genre
	event.SetGenre(g.genreID)

	// Initialize shop and crafting systems (v5.0)
	g.shopCredits = shop.NewCredit(100) // Start with 100 credits
	g.shopArmory = shop.NewArmory(g.genreID)
	g.shopInventory = &g.shopArmory.Inventory
	g.scrapStorage = crafting.NewScrapStorage()
	scrapName := crafting.GetScrapNameForGenre(g.genreID)
	g.scrapStorage.Add(scrapName, 10) // Start with some scrap
	g.craftingMenu = crafting.NewCraftingMenu(g.scrapStorage, g.genreID)
	g.craftingResult = ""

	// Initialize skills system (v5.0)
	g.skillManager = skills.NewManager()
	g.skillManager.AddPoints(3) // Start with 3 skill points
	g.skillsTreeIdx = 0
	g.skillsNodeIdx = 0

	// Initialize mod loader (v5.0)
	if g.modLoader == nil {
		g.modLoader = mod.NewLoader()
	}
	// Scan mods directory for available mods
	g.scanMods()

	// Reset inventory
	g.playerInventory = inventory.NewInventory()
	inventory.SetGenre(g.genreID)

	// Track level start time for speedrun objectives
	g.levelStartTime = time.Now()

	// Play music
	g.audioEngine.PlayMusic("theme", 0.5)

	// Hide loading screen and start playing
	g.loadingScreen.Hide()
	g.state = StatePlaying

	// Show movement tutorial
	g.tutorialSystem.ShowPrompt(tutorial.PromptMovement, tutorial.GetMessage(tutorial.PromptMovement))
}

// setGenre propagates genre setting to all v3.0 systems (Step 29).
func (g *Game) setGenre(genreID string) {
	g.genreID = genreID

	// v1.0 systems
	g.world.SetGenre(genreID)
	g.raycaster.SetGenre(genreID)
	camera.SetGenre(genreID)
	g.audioEngine.SetGenre(genreID)
	tutorial.SetGenre(genreID)
	automap.SetGenre(genreID)
	door.SetGenre(genreID)

	// v2.0 systems
	g.arsenal.SetGenre(genreID)
	ammo.SetGenre(genreID)
	g.combatSystem.SetGenre(genreID)
	status.SetGenre(genreID)
	loot.SetGenre(genreID)
	progression.SetGenre(genreID)
	class.SetGenre(genreID)
	ai.SetGenre(genreID)

	// v3.0 systems
	g.textureAtlas.SetGenre(genreID)
	g.lightMap.SetGenre(genreID)
	g.postProcessor.SetGenre(genreID)
	g.renderer.SetGenre(genreID)
	// WeatherEmitter doesn't have SetGenre - recreate it on genre change
	if g.particleSystem != nil && len(g.currentMap) > 0 && len(g.currentMap[0]) > 0 {
		g.weatherEmitter = particle.NewWeatherEmitter(g.particleSystem, genreID, 0, 0, float64(len(g.currentMap[0])), float64(len(g.currentMap)))
	}

	// v4.0 systems
	destruct.SetGenre(genreID)
	squad.SetGenre(genreID)
	if g.questTracker != nil {
		g.questTracker.SetGenre(genreID)
	}
	event.SetGenre(genreID)

	// v5.0 systems
	shop.SetGenre(genreID)
	if g.shopArmory != nil {
		g.shopArmory.SetGenre(genreID)
		g.shopInventory = &g.shopArmory.Inventory
	}
	crafting.SetGenre(genreID)
	skills.SetGenre(genreID)
	network.SetGenre(genreID)
	if g.propsManager != nil {
		g.propsManager.SetGenre(genreID)
	}
	if g.loreGenerator != nil {
		g.loreGenerator.SetGenre(genreID)
	}

	// Generate genre-specific textures
	g.textureAtlas.GenerateWallSet(genreID)
	g.textureAtlas.GenerateGenreAnimations(genreID)
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

	// Restore progression
	if g.progression != nil {
		g.progression.Level = state.Progression.Level
		g.progression.XP = state.Progression.XP
	}

	// Restore keycards
	if state.Keycards != nil {
		g.keycards = state.Keycards
	} else {
		g.keycards = make(map[string]bool)
	}

	// Restore ammo pool
	if state.AmmoPool != nil && g.ammoPool != nil {
		// Clear current ammo and restore from save
		for ammoType, amount := range state.AmmoPool {
			g.ammoPool.Set(ammoType, amount)
		}
	}

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

	// Open shop overlay
	if g.input.IsJustPressed(input.ActionShop) {
		g.openShop()
		return nil
	}

	// Open crafting overlay
	if g.input.IsJustPressed(input.ActionCraft) {
		g.openCrafting()
		return nil
	}

	// Open skills overlay
	if g.input.IsJustPressed(input.ActionSkills) {
		g.openSkills()
		return nil
	}

	// Open multiplayer lobby
	if g.input.IsJustPressed(input.ActionMultiplayer) {
		g.openMultiplayer()
		return nil
	}

	// Open lore codex
	if g.input.IsJustPressed(input.ActionCodex) {
		g.state = StateCodex
		g.codexScrollIdx = 0
		return nil
	}

	// Use quick slot item
	if g.input.IsJustPressed(input.ActionUseItem) {
		g.useQuickSlotItem()
	}

	// Check for door interaction and lore item collection
	if g.input.IsJustPressed(input.ActionInteract) {
		g.tryCollectLore()
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
								// Award mastery XP for successful hit
								if g.masteryManager != nil {
									g.masteryManager.AddMasteryXP(g.arsenal.CurrentSlot, 10)
								}

								// Apply damage with upgrades and mastery bonuses
								upgradedDamage := g.getUpgradedWeaponDamage(currentWeapon)
								agent.Health -= upgradedDamage

								if agent.Health <= 0 {
									// Enemy died - award XP and credits
									oldLevel := g.progression.Level
									g.progression.AddXP(50)
									// Check for level-up (every 100 XP)
									newLevel := g.progression.XP / 100
									if newLevel > oldLevel {
										g.progression.Level = newLevel
										// Award skill point on level-up
										if g.skillManager != nil {
											g.skillManager.AddPoints(1)
										}
										g.hud.ShowMessage("Level Up! Skill point earned!")
									}
									if g.shopCredits != nil {
										g.shopCredits.Add(25) // 25 credits per kill
									}
									// Drop upgrade tokens for weapon upgrades
									if g.upgradeManager != nil {
										g.upgradeManager.GetTokens().Add(1) // 1 token per kill
									}
									// Drop scrap for crafting
									if g.scrapStorage != nil {
										scrapName := crafting.GetScrapNameForGenre(g.genreID)
										g.scrapStorage.Add(scrapName, 3)
									}
									// Update kill count objective
									if g.questTracker != nil {
										g.questTracker.UpdateProgress("bonus_kills", 1)
									}
								}
							}
						}
					}
				}

				// Check for destructible hits (simple raycast)
				if hitResults == nil || len(hitResults) == 0 {
					// No enemy hit, check destructibles
					allDestructibles := g.destructibleSystem.GetAll()
					for _, obj := range allDestructibles {
						if obj.IsDestroyed() {
							continue
						}
						// Check if ray hits this object (simplified sphere collision)
						objDist := (obj.X-g.camera.X)*(obj.X-g.camera.X) + (obj.Y-g.camera.Y)*(obj.Y-g.camera.Y)
						if objDist < 100 { // Max range
							// Check if object is in ray direction
							toObjX := obj.X - g.camera.X
							toObjY := obj.Y - g.camera.Y
							dot := toObjX*g.camera.DirX + toObjY*g.camera.DirY
							if dot > 0 { // Object is in front
								// Apply damage with upgrades
								upgradedDamage := g.getUpgradedWeaponDamage(currentWeapon)
								destroyed := obj.Damage(upgradedDamage)
								if destroyed {
									// Spawn particles for destruction
									if g.particleSystem != nil {
										debrisColor := color.RGBA{R: 100, G: 80, B: 60, A: 255}
										g.particleSystem.SpawnBurst(obj.X, obj.Y, 0, 15, 8.0, 1.0, 1.5, 1.0, debrisColor)
									}
									// Drop scrap from destroyed objects
									if g.scrapStorage != nil {
										scrapName := crafting.GetScrapNameForGenre(g.genreID)
										g.scrapStorage.Add(scrapName, 2)
									}
									// Award credits for destruction
									if g.shopCredits != nil {
										g.shopCredits.Add(10)
									}
									// Play destruction sound
									g.audioEngine.PlaySFX("barrel_explode", obj.X, obj.Y)
								}
								break // Only hit one object
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

	// Update squad companions
	if g.squadCompanions != nil {
		// Update squad AI with player as leader
		g.squadCompanions.Update(g.camera.X, g.camera.Y, g.currentMap, g.camera.X, g.camera.Y, g.seed)
	}

	// Update event triggers
	deltaTime := 1.0 / 60.0 // Assuming 60 TPS
	if g.alarmTrigger != nil && g.alarmTrigger.IsActive() {
		g.alarmTrigger.Update(deltaTime)
		// During alarm, all enemies are on alert (TODO: implement alert state)
	}
	if g.lockdownTrigger != nil && g.lockdownTrigger.IsActive() {
		g.lockdownTrigger.Update(deltaTime)
		// Update speedrun objective with remaining time
		if g.questTracker != nil {
			remainingTime := g.lockdownTrigger.GetRemaining()
			// Check if player is out of time
			if g.lockdownTrigger.IsExpired() {
				// TODO: handle lockdown failure (restart level or game over)
				g.hud.ShowMessage("Lockdown complete - you are trapped!")
			} else if remainingTime < 10 {
				g.hud.ShowMessage("WARNING: 10 seconds remaining!")
			}
		}
	}

	// Check boss arena trigger
	if g.bossArena != nil && !g.bossArena.IsTriggered() {
		// Check if player entered boss arena region (simplified)
		centerX := float64(len(g.currentMap[0]) / 2)
		centerY := float64(len(g.currentMap) / 2)
		distToCenterSq := (g.camera.X-centerX)*(g.camera.X-centerX) + (g.camera.Y-centerY)*(g.camera.Y-centerY)
		if distToCenterSq < 25 { // Within 5 units of center
			g.bossArena.Trigger()
			// Generate and play event audio sting (parameters available for future procedural audio)
			_ = event.GenerateEventAudioSting(g.seed, event.EventBossArena)
			g.audioEngine.PlaySFX("boss_encounter", g.camera.X, g.camera.Y)
			// Show event text
			eventText := event.GenerateEventText(g.seed, event.EventBossArena)
			g.hud.ShowMessage(eventText)
			// TODO: spawn boss waves
		}
	}

	// Update quest objectives
	if g.questTracker != nil {
		// Update speedrun timer
		elapsedTime := time.Since(g.levelStartTime).Seconds()
		for i := range g.questTracker.Objectives {
			obj := &g.questTracker.Objectives[i]
			if obj.ID == "bonus_speed" && !obj.Complete {
				if elapsedTime > float64(obj.Count) {
					// Failed speed run objective
					obj.Complete = false // Already false, but explicit
				}
			}
		}
	}

	// Progression updates happen automatically
	// TODO: Add Level and XP to HUD display

	// Update v3.0 systems (Step 28: Wire to game loop)
	// Update particles (assuming 60 TPS = 1/60 second per frame)
	if g.particleSystem != nil {
		g.particleSystem.Update(deltaTime)
	}

	// Update weather emitter
	if g.weatherEmitter != nil {
		g.weatherEmitter.Update(deltaTime)
	}

	// Update secret wall animations
	if g.secretManager != nil {
		g.secretManager.Update(deltaTime)
	}

	// Recalculate lighting with flashlight
	if g.lightMap != nil {
		// Create genre-specific flashlight
		preset := lighting.GetFlashlightPreset(g.genreID)
		flashlight := lighting.NewConeLight(g.camera.X, g.camera.Y, g.camera.DirX, g.camera.DirY, preset)

		// Clear previous lights and add current flashlight
		g.lightMap.Clear()
		g.lightMap.AddLight(flashlight.GetContributionAsPointLight())
		g.lightMap.Calculate()
	}

	// Update reverb based on current room position
	if g.currentBSPTree != nil {
		g.audioEngine.UpdateReverb(int(g.camera.X), int(g.camera.Y), g.currentBSPTree)
	}

	// Increment animation ticker for animated textures
	g.animationTicker++

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
		deltaPitch = -mouseDY * sensitivity * 3.0 // Balanced pitch sensitivity
	}

	// Gamepad: Right stick camera
	rightX, rightY := g.input.GamepadRightStick()
	if rightX*rightX+rightY*rightY > deadzone*deadzone {
		g.camera.Rotate(rightX * rotSpeed * 1.5)
		deltaPitch = -rightY * rotSpeed * 15.0
	}

	// Collision detection with wall-sliding
	newX := g.camera.X + deltaX
	newY := g.camera.Y + deltaY
	if g.isWalkable(newX, newY) {
		g.camera.Update(deltaX, deltaY, deltaDirX, deltaDirY, deltaPitch)
	} else if g.isWalkable(newX, g.camera.Y) {
		// Slide along X axis only
		g.camera.Update(deltaX, 0, deltaDirX, deltaDirY, deltaPitch)
	} else if g.isWalkable(g.camera.X, newY) {
		// Slide along Y axis only
		g.camera.Update(0, deltaY, deltaDirX, deltaDirY, deltaPitch)
	} else {
		// Completely blocked \u2014 still allow look changes
		g.camera.Update(0, 0, deltaDirX, deltaDirY, deltaPitch)
	}
	if g.automap != nil {
		g.automap.Reveal(int(g.camera.X), int(g.camera.Y))
	}

	// Update ECS world
	g.world.Update()

	// Update audio listener position
	g.audioEngine.SetListenerPosition(g.camera.X, g.camera.Y)

	// Tutorial completion checks — dismiss on movement or any key press
	if g.tutorialSystem.Active {
		if g.tutorialSystem.Type == tutorial.PromptMovement && (deltaX != 0 || deltaY != 0) {
			g.tutorialSystem.Complete()
		}
		// Allow dismissing any tutorial with fire/interact
		if g.input.IsJustPressed(input.ActionFire) || g.input.IsJustPressed(input.ActionInteract) {
			g.tutorialSystem.Complete()
		}
	}

	return nil
}

// isWalkableTile returns true if the tile type permits player movement.
func isWalkableTile(tile int) bool {
	switch {
	case tile == bsp.TileFloor || tile == bsp.TileEmpty:
		return true
	case tile >= 20 && tile <= 29: // Genre-specific floor tiles (TileFloorStone..TileFloorDirt)
		return true
	default:
		return false
	}
}

// playerRadius is the collision bounding radius around the player position.
const playerRadius = 0.25

// isWalkable checks if a position is walkable using radius-based collision.
// Tests four corners of the player's bounding box to prevent wall-corner clipping.
func (g *Game) isWalkable(x, y float64) bool {
	if g.currentMap == nil || len(g.currentMap) == 0 {
		return true
	}
	// Check the four corners of the player's bounding box
	offsets := [4][2]float64{
		{-playerRadius, -playerRadius},
		{playerRadius, -playerRadius},
		{-playerRadius, playerRadius},
		{playerRadius, playerRadius},
	}
	for _, off := range offsets {
		cx := x + off[0]
		cy := y + off[1]
		mapX := int(cx)
		mapY := int(cy)
		if mapY < 0 || mapY >= len(g.currentMap) || mapX < 0 || mapX >= len(g.currentMap[0]) {
			return false
		}
		if !isWalkableTile(g.currentMap[mapY][mapX]) {
			return false
		}
	}
	return true
}

// tryInteractDoor checks if player is facing a door and attempts to open it.
// Also checks for secret walls that can be triggered.
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

	// Check for secret walls first
	if tile == bsp.TileSecret {
		if g.secretManager != nil && g.secretManager.TriggerAt(mapX, mapY, "player") {
			g.audioEngine.PlaySFX("secret_open", float64(mapX), float64(mapY))
			g.hud.ShowMessage("Secret discovered!")
			// Update quest tracker for secret discovery
			if g.questTracker != nil {
				g.questTracker.UpdateProgress("bonus_secrets", 1)
			}
		}
		return
	}

	if tile == bsp.TileDoor {
		requiredColor := g.getDoorColor(mapX, mapY)
		if requiredColor == "" || g.keycards[requiredColor] {
			// Open door immediately if no keycard required or player has keycard
			g.currentMap[mapY][mapX] = bsp.TileFloor
			g.raycaster.SetMap(g.currentMap)
			g.audioEngine.PlaySFX("door_open", float64(mapX), float64(mapY))
		} else {
			// Door is locked - offer lockpicking minigame
			g.startMinigame(mapX, mapY)
		}
	}
}

// startMinigame initiates a minigame for the current genre.
func (g *Game) startMinigame(doorX, doorY int) {
	// Determine difficulty based on progression level
	difficulty := g.progression.Level / 3
	if difficulty > 3 {
		difficulty = 3
	}

	// Use seed based on door position for deterministic generation
	seed := int64(g.seed) + int64(doorX*1000+doorY)

	// Create genre-appropriate minigame
	g.activeMinigame = minigame.GetGenreMiniGame(g.genreID, difficulty, seed)
	g.activeMinigame.Start()
	g.minigameDoorX = doorX
	g.minigameDoorY = doorY
	g.previousState = g.state
	g.state = StateMinigame
	g.minigameInputTimer = 0

	// Determine minigame type for rendering
	switch g.genreID {
	case "fantasy":
		g.minigameType = "lockpick"
	case "cyberpunk":
		g.minigameType = "circuit"
	case "scifi", "postapoc":
		g.minigameType = "code"
	default:
		g.minigameType = "hack"
	}
}

// getDoorColor returns the keycard color required for a door (stub - would be from door metadata).
func (g *Game) getDoorColor(x, y int) string {
	return ""
}

// useQuickSlotItem uses the active item in the inventory quick slot.
func (g *Game) useQuickSlotItem() {
	if g.playerInventory == nil {
		return
	}

	// Check if quick slot has an item
	activeItem := g.playerInventory.GetQuickSlot()
	if activeItem == nil {
		// Try to auto-equip a medkit if available
		if g.playerInventory.Has("medkit") {
			medkit := &inventory.Medkit{
				ID:         "medkit",
				Name:       "Medkit",
				HealAmount: 25,
			}
			g.playerInventory.SetQuickSlot(medkit)
			activeItem = medkit
		} else {
			g.hud.ShowMessage("No item equipped")
			return
		}
	}

	// Create entity wrapper for player
	playerEntity := &inventory.Entity{
		Health:    float64(g.hud.Health),
		MaxHealth: float64(g.hud.MaxHealth),
		X:         g.camera.X,
		Y:         g.camera.Y,
	}

	// Use the quick slot item
	if err := g.playerInventory.UseQuickSlot(playerEntity); err != nil {
		g.hud.ShowMessage(err.Error())
		return
	}

	// Apply health change back to HUD
	g.hud.Health = int(playerEntity.Health)

	// Play sound effect
	g.audioEngine.PlaySFX("item_use", g.camera.X, g.camera.Y)
	g.hud.ShowMessage("Used " + activeItem.GetName())
}

// tryCollectLore checks if player is near a lore item and collects it.
func (g *Game) tryCollectLore() {
	collectDist := 2.0
	for _, loreItem := range g.loreItems {
		if loreItem.Activated {
			continue
		}
		dx := loreItem.PosX - g.camera.X
		dy := loreItem.PosY - g.camera.Y
		dist := dx*dx + dy*dy
		if dist < collectDist*collectDist {
			loreItem.Activated = true
			g.loreCodex.MarkFound(loreItem.CodexID)
			typeName := lore.GetLoreItemTypeName(loreItem.Type, g.genreID)
			g.hud.ShowMessage("Found: " + typeName)
			g.audioEngine.PlaySFX("lore_pickup", g.camera.X, g.camera.Y)
			return
		}
	}
}

// getLoreContext determines appropriate lore context based on room properties.
func (g *Game) getLoreContext(room bsp.Room) lore.ContextType {
	// Simple heuristic based on room position and size
	if room.W*room.H < 50 {
		return lore.ContextStorage
	}
	if room.X < 20 && room.Y < 20 {
		return lore.ContextQuarters
	}
	if room.W*room.H > 200 {
		return lore.ContextCombat
	}
	if room.X > 40 || room.Y > 40 {
		return lore.ContextEscape
	}
	return lore.ContextGeneral
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
	case "shop":
		g.openShop()
	case "skills":
		g.openSkills()
	case "multiplayer":
		g.openMultiplayer()
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

// openShop transitions to the shop state.
func (g *Game) openShop() {
	if g.shopArmory == nil {
		g.shopArmory = shop.NewArmory(g.genreID)
		g.shopInventory = &g.shopArmory.Inventory
	}
	if g.shopCredits == nil {
		g.shopCredits = shop.NewCredit(0)
	}
	g.menuManager.Show(ui.MenuTypeShop)
	g.state = StateShop
}

// openCrafting transitions to the crafting state.
func (g *Game) openCrafting() {
	if g.scrapStorage == nil {
		g.scrapStorage = crafting.NewScrapStorage()
	}
	if g.craftingMenu == nil {
		g.craftingMenu = crafting.NewCraftingMenu(g.scrapStorage, g.genreID)
	}
	g.craftingResult = ""
	g.menuManager.Show(ui.MenuTypeCrafting)
	g.state = StateCrafting
}

// updateShop handles shop screen input.
func (g *Game) updateShop() error {
	// Back to pause/playing
	if g.input.IsJustPressed(input.ActionPause) || g.input.IsJustPressed(input.ActionShop) {
		g.state = StatePlaying
		g.menuManager.Hide()
		return nil
	}

	// Navigate shop items
	if g.input.IsJustPressed(input.ActionMoveForward) {
		g.menuManager.MoveUp()
	}
	if g.input.IsJustPressed(input.ActionMoveBackward) {
		g.menuManager.MoveDown()
	}

	// Purchase selected item
	if g.input.IsJustPressed(input.ActionFire) || g.input.IsJustPressed(input.ActionInteract) {
		g.handleShopPurchase()
	}

	return nil
}

// handleShopPurchase attempts to buy the selected shop item.
func (g *Game) handleShopPurchase() {
	if g.shopArmory == nil || g.shopCredits == nil {
		return
	}

	allItems := g.shopArmory.Inventory.GetAllItems()
	idx := g.menuManager.GetSelectedIndex()
	if idx < 0 || idx >= len(allItems) {
		return
	}

	item := allItems[idx]
	if g.shopArmory.Purchase(item.ID, g.shopCredits) {
		// Apply purchased item effects
		g.applyShopItem(item.ID)
		g.hud.ShowMessage("Purchased: " + item.Name)
		g.audioEngine.PlaySFX("shop_buy", g.camera.X, g.camera.Y)
	} else {
		g.hud.ShowMessage("Cannot afford: " + item.Name)
	}
}

// applyShopItem applies the effects of a purchased shop item.
func (g *Game) applyShopItem(itemID string) {
	switch itemID {
	case "ammo_bullets":
		g.ammoPool.Add("bullets", 20)
	case "ammo_shells":
		g.ammoPool.Add("shells", 10)
	case "ammo_cells":
		g.ammoPool.Add("cells", 15)
	case "ammo_rockets":
		g.ammoPool.Add("rockets", 5)
	case "ammo_arrows":
		g.ammoPool.Add("arrows", 20)
	case "ammo_bolts":
		g.ammoPool.Add("bolts", 10)
	case "medkit":
		g.playerInventory.Add(inventory.Item{ID: "medkit", Name: "Medkit", Qty: 1})
	case "grenade", "plasma_grenade", "emp_grenade", "bomb":
		g.playerInventory.Add(inventory.Item{ID: "grenade", Name: "Grenade", Qty: 1})
	case "proximity_mine":
		g.playerInventory.Add(inventory.Item{ID: "proximity_mine", Name: "Proximity Mine", Qty: 1})
	case "armor_vest":
		g.hud.Armor += 50
		if g.hud.Armor > g.hud.MaxArmor {
			g.hud.Armor = g.hud.MaxArmor
		}
	// Weapon upgrades
	case "upgrade_damage":
		currentWeapon := g.arsenal.GetCurrentWeapon()
		weaponID := currentWeapon.Name
		if g.upgradeManager.ApplyUpgrade(weaponID, upgrade.UpgradeDamage, 2) {
			g.hud.ShowMessage("Damage upgrade applied!")
		}
	case "upgrade_firerate":
		currentWeapon := g.arsenal.GetCurrentWeapon()
		weaponID := currentWeapon.Name
		if g.upgradeManager.ApplyUpgrade(weaponID, upgrade.UpgradeFireRate, 2) {
			g.hud.ShowMessage("Fire rate upgrade applied!")
		}
	case "upgrade_clipsize":
		currentWeapon := g.arsenal.GetCurrentWeapon()
		weaponID := currentWeapon.Name
		if g.upgradeManager.ApplyUpgrade(weaponID, upgrade.UpgradeClipSize, 2) {
			g.hud.ShowMessage("Clip size upgrade applied!")
		}
	case "upgrade_accuracy":
		currentWeapon := g.arsenal.GetCurrentWeapon()
		weaponID := currentWeapon.Name
		if g.upgradeManager.ApplyUpgrade(weaponID, upgrade.UpgradeAccuracy, 2) {
			g.hud.ShowMessage("Accuracy upgrade applied!")
		}
	case "upgrade_range":
		currentWeapon := g.arsenal.GetCurrentWeapon()
		weaponID := currentWeapon.Name
		if g.upgradeManager.ApplyUpgrade(weaponID, upgrade.UpgradeRange, 2) {
			g.hud.ShowMessage("Range upgrade applied!")
		}
	}
	// Update HUD ammo display
	currentWeapon := g.arsenal.GetCurrentWeapon()
	g.hud.Ammo = g.ammoPool.Get(currentWeapon.AmmoType)
}

// updateCrafting handles crafting screen input.
func (g *Game) updateCrafting() error {
	// Back to playing
	if g.input.IsJustPressed(input.ActionPause) || g.input.IsJustPressed(input.ActionCraft) {
		g.state = StatePlaying
		g.menuManager.Hide()
		return nil
	}

	// Navigate recipes
	if g.input.IsJustPressed(input.ActionMoveForward) {
		g.menuManager.MoveUp()
	}
	if g.input.IsJustPressed(input.ActionMoveBackward) {
		g.menuManager.MoveDown()
	}

	// Craft selected recipe
	if g.input.IsJustPressed(input.ActionFire) || g.input.IsJustPressed(input.ActionInteract) {
		g.handleCraftItem()
	}

	return nil
}

// handleCraftItem attempts to craft the selected recipe.
func (g *Game) handleCraftItem() {
	if g.craftingMenu == nil {
		return
	}

	allRecipes := g.craftingMenu.GetAllRecipes()
	idx := g.menuManager.GetSelectedIndex()
	if idx < 0 || idx >= len(allRecipes) {
		return
	}

	recipe := allRecipes[idx]
	outputID, outputQty, err := g.craftingMenu.Craft(recipe.ID)
	if err != nil {
		g.craftingResult = "Not enough materials!"
		return
	}

	// Apply crafted item to player resources
	g.applyCraftedItem(outputID, outputQty)
	g.craftingResult = recipe.Name + " crafted!"
	g.audioEngine.PlaySFX("craft_complete", g.camera.X, g.camera.Y)
}

// applyCraftedItem adds crafted items to the player's inventory.
func (g *Game) applyCraftedItem(outputID string, qty int) {
	switch outputID {
	case "bullets", "shells", "cells", "rockets", "arrows", "bolts", "mana", "explosives":
		g.ammoPool.Add(outputID, qty)
	case "medkit", "potion":
		g.playerInventory.Add(inventory.Item{ID: outputID, Name: "Medkit", Qty: qty})
	}
	// Update HUD ammo display
	currentWeapon := g.arsenal.GetCurrentWeapon()
	g.hud.Ammo = g.ammoPool.Get(currentWeapon.AmmoType)
}

// getUpgradedWeaponDamage returns the weapon damage with all upgrades applied.
func (g *Game) getUpgradedWeaponDamage(baseWeapon weapon.Weapon) float64 {
	damage := baseWeapon.Damage

	// Apply upgrade bonuses
	if g.upgradeManager != nil {
		upgrades := g.upgradeManager.GetUpgrades(baseWeapon.Name)
		for _, upgradeType := range upgrades {
			wu := upgrade.NewWeaponUpgrade(upgradeType)
			damage, _, _, _, _ = wu.ApplyWeaponStats(damage, 0, 0, 0, 0)
		}
	}

	// Apply mastery bonuses
	if g.masteryManager != nil {
		bonuses := g.masteryManager.GetBonus(g.arsenal.CurrentSlot)
		damage *= bonuses.HeadshotDamage // Applies headshot damage bonus to all damage
	}

	return damage
}

// drawShop renders the shop overlay screen.
func (g *Game) drawShop(screen *ebiten.Image) {
	// Draw frozen game world
	g.renderer.Render(screen, g.camera.X, g.camera.Y, g.camera.DirX, g.camera.DirY, g.camera.Pitch)

	// Build shop state for UI
	shopState := g.buildShopState()
	ui.DrawShop(screen, shopState)
}

// buildShopState creates the shop display state from game data.
func (g *Game) buildShopState() *ui.ShopState {
	if g.shopArmory == nil || g.shopCredits == nil {
		return nil
	}

	allItems := g.shopArmory.Inventory.GetAllItems()
	uiItems := make([]ui.ShopItem, len(allItems))
	for i, item := range allItems {
		uiItems[i] = ui.ShopItem{
			ID:    item.ID,
			Name:  item.Name,
			Price: item.Price,
			Stock: item.Stock,
		}
	}

	return &ui.ShopState{
		ShopName: g.shopArmory.GetShopName(),
		Items:    uiItems,
		Credits:  g.shopCredits.Get(),
		Selected: g.menuManager.GetSelectedIndex(),
	}
}

// getUpgradeTokenCount returns the current upgrade token count for display.
func (g *Game) getUpgradeTokenCount() int {
	if g.upgradeManager == nil {
		return 0
	}
	return g.upgradeManager.GetTokens().GetCount()
}

// drawCrafting renders the crafting overlay screen.
func (g *Game) drawCrafting(screen *ebiten.Image) {
	// Draw frozen game world
	g.renderer.Render(screen, g.camera.X, g.camera.Y, g.camera.DirX, g.camera.DirY, g.camera.Pitch)

	// Build crafting state for UI
	craftState := g.buildCraftingState()
	ui.DrawCrafting(screen, craftState)
}

// buildCraftingState creates the crafting display state from game data.
func (g *Game) buildCraftingState() *ui.CraftingState {
	if g.craftingMenu == nil || g.scrapStorage == nil {
		return nil
	}

	allRecipes := g.craftingMenu.GetAllRecipes()
	availableRecipes := g.craftingMenu.GetAvailableRecipes()

	// Build set of available recipe IDs for quick lookup
	availableIDs := make(map[string]bool)
	for _, r := range availableRecipes {
		availableIDs[r.ID] = true
	}

	uiRecipes := make([]ui.CraftingRecipe, len(allRecipes))
	for i, r := range allRecipes {
		uiRecipes[i] = ui.CraftingRecipe{
			ID:        r.ID,
			Name:      r.Name,
			Inputs:    r.Inputs,
			OutputQty: r.OutputQty,
			CanCraft:  availableIDs[r.ID],
		}
	}

	return &ui.CraftingState{
		Recipes:    uiRecipes,
		ScrapName:  crafting.GetScrapNameForGenre(g.genreID),
		ScrapAmts:  g.scrapStorage.GetAll(),
		Selected:   g.menuManager.GetSelectedIndex(),
		LastResult: g.craftingResult,
	}
}

// openSkills transitions to the skills state.
func (g *Game) openSkills() {
	if g.skillManager == nil {
		g.skillManager = skills.NewManager()
	}
	g.skillsTreeIdx = 0
	g.skillsNodeIdx = 0
	g.menuManager.Show(ui.MenuTypeSkills)
	g.state = StateSkills
}

// updateSkills handles skills screen input.
func (g *Game) updateSkills() error {
	// Back to playing
	if g.input.IsJustPressed(input.ActionPause) || g.input.IsJustPressed(input.ActionSkills) {
		g.state = StatePlaying
		g.menuManager.Hide()
		return nil
	}

	// Switch tree tabs with strafe keys (left/right)
	if g.input.IsJustPressed(input.ActionStrafeLeft) {
		if g.skillsTreeIdx > 0 {
			g.skillsTreeIdx--
			g.skillsNodeIdx = 0
		}
	}
	if g.input.IsJustPressed(input.ActionStrafeRight) {
		if g.skillsTreeIdx < 2 { // 3 trees: combat, survival, tech
			g.skillsTreeIdx++
			g.skillsNodeIdx = 0
		}
	}

	// Navigate nodes
	if g.input.IsJustPressed(input.ActionMoveForward) {
		if g.skillsNodeIdx > 0 {
			g.skillsNodeIdx--
		}
	}
	if g.input.IsJustPressed(input.ActionMoveBackward) {
		g.skillsNodeIdx++
	}

	// Allocate skill point
	if g.input.IsJustPressed(input.ActionFire) || g.input.IsJustPressed(input.ActionInteract) {
		g.handleSkillAllocate()
	}

	return nil
}

// handleSkillAllocate attempts to allocate a skill point.
func (g *Game) handleSkillAllocate() {
	if g.skillManager == nil {
		return
	}

	treeIDs := []string{"combat", "survival", "tech"}
	if g.skillsTreeIdx < 0 || g.skillsTreeIdx >= len(treeIDs) {
		return
	}
	treeID := treeIDs[g.skillsTreeIdx]

	tree, err := g.skillManager.GetTree(treeID)
	if err != nil {
		return
	}

	// Get sorted node list for this tree
	nodes := g.getTreeNodeList(tree)
	if g.skillsNodeIdx < 0 || g.skillsNodeIdx >= len(nodes) {
		return
	}

	nodeID := nodes[g.skillsNodeIdx].ID
	if err := g.skillManager.AllocatePoint(treeID, nodeID); err != nil {
		g.hud.ShowMessage("Cannot allocate: " + err.Error())
	} else {
		g.hud.ShowMessage("Skill unlocked!")
		g.audioEngine.PlaySFX("skill_unlock", g.camera.X, g.camera.Y)
	}
}

// getTreeNodeList returns nodes for a tree in a stable order.
func (g *Game) getTreeNodeList(tree *skills.Tree) []skills.Node {
	if tree == nil {
		return nil
	}
	// Return nodes in consistent order based on node IDs
	nodes := make([]skills.Node, 0, len(tree.Nodes))
	for _, node := range tree.Nodes {
		nodes = append(nodes, *node)
	}
	// Sort by ID for stable ordering
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[i].ID > nodes[j].ID {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
	return nodes
}

// drawSkills renders the skills overlay screen.
func (g *Game) drawSkills(screen *ebiten.Image) {
	// Draw frozen game world
	g.renderer.Render(screen, g.camera.X, g.camera.Y, g.camera.DirX, g.camera.DirY, g.camera.Pitch)

	// Build skills state for UI
	skillsState := g.buildSkillsState()
	ui.DrawSkills(screen, skillsState)
}

// buildSkillsState creates the skills display state from game data.
func (g *Game) buildSkillsState() *ui.SkillsState {
	if g.skillManager == nil {
		return nil
	}

	treeIDs := []string{"combat", "survival", "tech"}
	treeNames := []string{"Combat", "Survival", "Tech"}
	trees := make([]ui.SkillTreeState, 0, len(treeIDs))

	for i, treeID := range treeIDs {
		tree, err := g.skillManager.GetTree(treeID)
		if err != nil {
			continue
		}

		nodeList := g.getTreeNodeList(tree)
		uiNodes := make([]ui.SkillNode, len(nodeList))
		for j, node := range nodeList {
			// Check if prerequisites are met
			available := true
			for _, reqID := range node.Requires {
				if !tree.IsAllocated(reqID) {
					available = false
					break
				}
			}
			available = available && tree.GetPoints() >= node.Cost && !tree.IsAllocated(node.ID)

			uiNodes[j] = ui.SkillNode{
				ID:          node.ID,
				Name:        node.Name,
				Description: node.Description,
				Cost:        node.Cost,
				Allocated:   tree.IsAllocated(node.ID),
				Available:   available,
			}
		}

		trees = append(trees, ui.SkillTreeState{
			TreeName: treeNames[i],
			TreeID:   treeID,
			Nodes:    uiNodes,
			Points:   tree.GetPoints(),
			Selected: g.skillsNodeIdx,
		})
	}

	// Total points available across all trees (they share the same pool via AddPoints)
	totalPoints := 0
	if t, err := g.skillManager.GetTree("combat"); err == nil {
		totalPoints = t.GetPoints()
	}

	return &ui.SkillsState{
		Trees:       trees,
		ActiveTree:  g.skillsTreeIdx,
		Selected:    g.skillsNodeIdx,
		TotalPoints: totalPoints,
	}
}

// openMultiplayer transitions to the multiplayer lobby state.
func (g *Game) openMultiplayer() {
	g.mpSelectedMode = 0
	g.mpStatusMsg = ""
	g.useFederation = false
	g.serverBrowser = make([]*federation.ServerAnnouncement, 0)
	g.browserIdx = 0
	g.menuManager.Show(ui.MenuTypeMultiplayer)
	g.state = StateMultiplayer
}

// updateMultiplayer handles multiplayer lobby input.
func (g *Game) updateMultiplayer() error {
	// Back to playing
	if g.input.IsJustPressed(input.ActionPause) || g.input.IsJustPressed(input.ActionMultiplayer) {
		g.state = StatePlaying
		g.menuManager.Hide()
		return nil
	}

	// Toggle between local and federation mode with L key
	if g.input.IsJustPressed(input.ActionCodex) {
		g.useFederation = !g.useFederation
		if g.useFederation {
			g.refreshServerBrowser()
		}
	}

	// Navigate modes or servers
	if g.input.IsJustPressed(input.ActionMoveForward) {
		if g.useFederation {
			if g.browserIdx > 0 {
				g.browserIdx--
			}
		} else {
			if g.mpSelectedMode > 0 {
				g.mpSelectedMode--
			}
		}
	}
	if g.input.IsJustPressed(input.ActionMoveBackward) {
		if g.useFederation {
			if g.browserIdx < len(g.serverBrowser)-1 {
				g.browserIdx++
			}
		} else {
			g.mpSelectedMode++
			modes := g.getMultiplayerModes()
			if g.mpSelectedMode >= len(modes) {
				g.mpSelectedMode = len(modes) - 1
			}
		}
	}

	// Refresh server browser with R key
	if g.useFederation && g.input.IsJustPressed(input.ActionCraft) {
		g.refreshServerBrowser()
	}

	// Select mode or join server
	if g.input.IsJustPressed(input.ActionFire) || g.input.IsJustPressed(input.ActionInteract) {
		if g.useFederation {
			g.handleFederationJoin()
		} else {
			g.handleMultiplayerSelect()
		}
	}

	return nil
}

// updateCodex handles lore codex UI updates.
func (g *Game) updateCodex() error {
	// Close codex
	if g.input.IsJustPressed(input.ActionPause) || g.input.IsJustPressed(input.ActionCodex) {
		g.state = StatePlaying
		return nil
	}

	// Scroll through entries
	foundEntries := g.loreCodex.GetFoundEntries()
	if g.input.IsJustPressed(input.ActionMoveForward) {
		if g.codexScrollIdx > 0 {
			g.codexScrollIdx--
		}
	}
	if g.input.IsJustPressed(input.ActionMoveBackward) {
		if g.codexScrollIdx < len(foundEntries)-1 {
			g.codexScrollIdx++
		}
	}

	return nil
}

// updateMinigame handles minigame input and progression.
func (g *Game) updateMinigame() error {
	if g.activeMinigame == nil {
		g.state = g.previousState
		return nil
	}

	// Input delay to prevent double-inputs
	g.minigameInputTimer++

	// Escape cancels minigame
	if g.input.IsJustPressed(input.ActionPause) {
		g.activeMinigame = nil
		g.state = g.previousState
		g.hud.ShowMessage("Minigame cancelled")
		return nil
	}

	// Update minigame state
	finished := g.activeMinigame.Update()

	// Handle minigame-specific inputs
	switch g.minigameType {
	case "lockpick":
		g.updateLockpickGame()
	case "hack":
		g.updateHackGame()
	case "circuit":
		g.updateCircuitGame()
	case "code":
		g.updateCodeGame()
	}

	// Check if minigame completed
	if finished {
		progress := g.activeMinigame.GetProgress()
		if progress >= 1.0 {
			// Success - open door
			g.currentMap[g.minigameDoorY][g.minigameDoorX] = bsp.TileFloor
			g.raycaster.SetMap(g.currentMap)
			g.audioEngine.PlaySFX("door_open", float64(g.minigameDoorX), float64(g.minigameDoorY))
			g.hud.ShowMessage("Lock bypassed!")
		} else {
			// Failed
			g.hud.ShowMessage("Bypass failed - need keycard")
		}
		g.activeMinigame = nil
		g.state = g.previousState
	}

	return nil
}

// updateLockpickGame handles lockpicking minigame input.
func (g *Game) updateLockpickGame() {
	if g.minigameInputTimer < 3 {
		return
	}

	lpGame, ok := g.activeMinigame.(*minigame.LockpickGame)
	if !ok {
		return
	}

	// Advance lockpick position automatically
	lpGame.Advance()

	// Space/Fire to attempt unlock
	if g.input.IsJustPressed(input.ActionFire) || g.input.IsJustPressed(input.ActionInteract) {
		success := lpGame.Attempt()
		if success {
			g.audioEngine.PlaySFX("lockpick_success", g.camera.X, g.camera.Y)
		} else {
			g.audioEngine.PlaySFX("lockpick_fail", g.camera.X, g.camera.Y)
		}
		g.minigameInputTimer = 0
	}
}

// updateHackGame handles hacking minigame input.
func (g *Game) updateHackGame() {
	if g.minigameInputTimer < 5 {
		return
	}

	hackGame, ok := g.activeMinigame.(*minigame.HackGame)
	if !ok {
		return
	}

	// Number keys 1-6 for node selection
	for i := 0; i < 6; i++ {
		key := ebiten.Key(int(ebiten.Key1) + i)
		if inpututil.IsKeyJustPressed(key) {
			success := hackGame.Input(i)
			if success {
				g.audioEngine.PlaySFX("hack_correct", g.camera.X, g.camera.Y)
			} else {
				g.audioEngine.PlaySFX("hack_wrong", g.camera.X, g.camera.Y)
			}
			g.minigameInputTimer = 0
		}
	}
}

// updateCircuitGame handles circuit trace minigame input.
func (g *Game) updateCircuitGame() {
	if g.minigameInputTimer < 5 {
		return
	}

	circuitGame, ok := g.activeMinigame.(*minigame.CircuitTraceGame)
	if !ok {
		return
	}

	// Arrow keys or WASD for movement
	moved := false
	if g.input.IsJustPressed(input.ActionMoveForward) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		moved = circuitGame.Move(0) // up
	} else if g.input.IsJustPressed(input.ActionStrafeRight) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		moved = circuitGame.Move(1) // right
	} else if g.input.IsJustPressed(input.ActionMoveBackward) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		moved = circuitGame.Move(2) // down
	} else if g.input.IsJustPressed(input.ActionStrafeLeft) || inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		moved = circuitGame.Move(3) // left
	}

	if moved {
		g.audioEngine.PlaySFX("circuit_move", g.camera.X, g.camera.Y)
		g.minigameInputTimer = 0
	}
}

// updateCodeGame handles bypass code minigame input.
func (g *Game) updateCodeGame() {
	if g.minigameInputTimer < 5 {
		return
	}

	codeGame, ok := g.activeMinigame.(*minigame.BypassCodeGame)
	if !ok {
		return
	}

	// Number keys 0-9 for code entry
	for i := 0; i < 10; i++ {
		key := ebiten.Key(int(ebiten.Key0) + i)
		if inpututil.IsKeyJustPressed(key) {
			success := codeGame.InputDigit(i)
			if success {
				g.audioEngine.PlaySFX("code_beep", g.camera.X, g.camera.Y)
			} else {
				g.audioEngine.PlaySFX("code_wrong", g.camera.X, g.camera.Y)
			}
			g.minigameInputTimer = 0
		}
	}

	// Backspace to clear
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		codeGame.Clear()
		g.minigameInputTimer = 0
	}
}

// handleMultiplayerSelect initializes the selected multiplayer mode.
func (g *Game) handleMultiplayerSelect() {
	modes := g.getMultiplayerModes()
	if g.mpSelectedMode < 0 || g.mpSelectedMode >= len(modes) {
		return
	}

	mode := modes[g.mpSelectedMode]
	switch mode.ID {
	case "coop":
		session, err := network.NewCoopSession("local_coop", 4, g.seed)
		if err != nil {
			g.mpStatusMsg = "Failed: " + err.Error()
			return
		}
		g.multiplayerMgr = session
		g.networkMode = true
		g.mpStatusMsg = "Co-op session started! Waiting for players..."
	case "ffa":
		match, err := network.NewFFAMatch("local_ffa", 20, 10*time.Minute, g.seed)
		if err != nil {
			g.mpStatusMsg = "Failed: " + err.Error()
			return
		}
		g.multiplayerMgr = match
		g.networkMode = true
		g.mpStatusMsg = "Free-for-All match started!"
	case "team":
		match, err := network.NewTeamMatch("local_team", 50, 15*time.Minute, g.seed)
		if err != nil {
			g.mpStatusMsg = "Failed: " + err.Error()
			return
		}
		g.multiplayerMgr = match
		g.networkMode = true
		g.mpStatusMsg = "Team Deathmatch started!"
	case "territory":
		match, err := network.NewTerritoryMatch("local_territory", 100, 20*time.Minute, g.seed)
		if err != nil {
			g.mpStatusMsg = "Failed: " + err.Error()
			return
		}
		g.multiplayerMgr = match
		g.networkMode = true
		g.mpStatusMsg = "Territory Control started!"
	default:
		g.mpStatusMsg = "Unknown mode"
	}
	g.hud.ShowMessage(g.mpStatusMsg)
}

// getMultiplayerModes returns the available multiplayer modes.
func (g *Game) getMultiplayerModes() []ui.MultiplayerMode {
	return []ui.MultiplayerMode{
		{ID: "coop", Name: "Cooperative", Description: "2-4 player cooperative campaign", MaxPlayers: 4},
		{ID: "ffa", Name: "Free-for-All", Description: "Every player for themselves", MaxPlayers: 8},
		{ID: "team", Name: "Team Deathmatch", Description: "Red vs Blue team combat", MaxPlayers: 16},
		{ID: "territory", Name: "Territory Control", Description: "Capture and hold strategic points", MaxPlayers: 16},
	}
}

// refreshServerBrowser queries the federation hub for available servers.
func (g *Game) refreshServerBrowser() {
	if g.federationHub == nil {
		g.mpStatusMsg = "Federation not available"
		return
	}

	genre := g.genreID
	query := &federation.ServerQuery{
		Genre: &genre,
	}

	servers := g.federationHub.QueryServers(query)
	g.serverBrowser = servers
	g.browserIdx = 0

	if len(servers) == 0 {
		g.mpStatusMsg = "No servers found. Press R to refresh."
	} else {
		g.mpStatusMsg = "Found " + string(rune(len(servers)+'0')) + " servers. Press L for local mode."
	}
}

// handleFederationJoin connects to a federated server.
func (g *Game) handleFederationJoin() {
	if len(g.serverBrowser) == 0 {
		g.mpStatusMsg = "No servers available. Press R to refresh."
		return
	}

	if g.browserIdx < 0 || g.browserIdx >= len(g.serverBrowser) {
		g.mpStatusMsg = "Invalid server selection"
		return
	}

	server := g.serverBrowser[g.browserIdx]
	g.mpStatusMsg = "Connecting to " + server.Name + "..."
	g.networkMode = true
	g.hud.ShowMessage(g.mpStatusMsg)
}

// drawMultiplayer renders the multiplayer lobby screen.
func (g *Game) drawMultiplayer(screen *ebiten.Image) {
	// Draw frozen game world
	g.renderer.Render(screen, g.camera.X, g.camera.Y, g.camera.DirX, g.camera.DirY, g.camera.Pitch)

	state := &ui.MultiplayerState{
		Modes:      g.getMultiplayerModes(),
		Selected:   g.mpSelectedMode,
		Connected:  g.networkMode,
		ServerAddr: "localhost",
		StatusMsg:  g.mpStatusMsg,
	}
	ui.DrawMultiplayer(screen, state)
}

// updateMods handles mods screen input.
func (g *Game) updateMods() error {
	// Back to playing
	if g.input.IsJustPressed(input.ActionPause) {
		g.state = StatePlaying
		g.menuManager.Hide()
		return nil
	}

	// Navigate mods
	if g.input.IsJustPressed(input.ActionMoveForward) {
		g.menuManager.MoveUp()
	}
	if g.input.IsJustPressed(input.ActionMoveBackward) {
		g.menuManager.MoveDown()
	}

	// Toggle mod enable/disable
	if g.input.IsJustPressed(input.ActionFire) || g.input.IsJustPressed(input.ActionInteract) {
		g.handleModToggle()
	}

	return nil
}

// handleModToggle toggles the selected mod on/off.
func (g *Game) handleModToggle() {
	if g.modLoader == nil {
		return
	}

	mods := g.modLoader.ListMods()
	idx := g.menuManager.GetSelectedIndex()
	if idx < 0 || idx >= len(mods) {
		return
	}

	modEntry := mods[idx]
	if modEntry.Enabled {
		g.modLoader.DisableMod(modEntry.Name)
		g.hud.ShowMessage("Disabled: " + modEntry.Name)
	} else {
		g.modLoader.EnableMod(modEntry.Name)
		g.hud.ShowMessage("Enabled: " + modEntry.Name)
	}
}

// drawMods renders the mods screen.
func (g *Game) drawMods(screen *ebiten.Image) {
	// Draw frozen game world
	g.renderer.Render(screen, g.camera.X, g.camera.Y, g.camera.DirX, g.camera.DirY, g.camera.Pitch)

	state := g.buildModsState()
	ui.DrawMods(screen, state)
}

// buildModsState creates the mods display state from game data.
func (g *Game) buildModsState() *ui.ModsState {
	if g.modLoader == nil {
		return nil
	}

	mods := g.modLoader.ListMods()
	uiMods := make([]ui.ModInfo, len(mods))
	for i, m := range mods {
		uiMods[i] = ui.ModInfo{
			Name:        m.Name,
			Version:     m.Version,
			Description: m.Description,
			Author:      m.Author,
			Enabled:     m.Enabled,
		}
	}

	return &ui.ModsState{
		Mods:     uiMods,
		ModsDir:  g.modLoader.GetModsDir(),
		Selected: g.menuManager.GetSelectedIndex(),
	}
}

// scanMods scans the mods directory for available mods.
func (g *Game) scanMods() {
	if g.modLoader == nil {
		return
	}

	modsDir := g.modLoader.GetModsDir()
	entries, err := os.ReadDir(modsDir)
	if err != nil {
		// Mods directory doesn't exist or can't be read; not an error
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			modPath := modsDir + "/" + entry.Name()
			// Attempt to load; ignore errors for invalid mods
			_ = g.modLoader.LoadMod(modPath)
		}
	}
}

// saveGame saves the current game state.
func (g *Game) saveGame(slot int) {
	// Collect ammo pool state
	ammoPoolState := make(map[string]int)
	if g.ammoPool != nil {
		ammoPoolState["bullets"] = g.ammoPool.Get("bullets")
		ammoPoolState["shells"] = g.ammoPool.Get("shells")
		ammoPoolState["cells"] = g.ammoPool.Get("cells")
		ammoPoolState["rockets"] = g.ammoPool.Get("rockets")
	}

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
		Inventory: save.Inventory{Items: []save.Item{}}, // TODO: populate from inventory system when implemented
		Progression: save.ProgressionState{
			Level: g.progression.Level,
			XP:    g.progression.XP,
		},
		Keycards: g.keycards,
		AmmoPool: ammoPoolState,
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
	case StateShop:
		g.drawShop(screen)
	case StateCrafting:
		g.drawCrafting(screen)
	case StateSkills:
		g.drawSkills(screen)
	case StateMods:
		g.drawMods(screen)
	case StateMultiplayer:
		g.drawMultiplayer(screen)
	case StateCodex:
		g.drawCodex(screen)
	case StateMinigame:
		g.drawMinigame(screen)
	}
}

// drawMenu renders the menu screen.
func (g *Game) drawMenu(screen *ebiten.Image) {
	ui.DrawMenu(screen, g.menuManager)
}

// drawPlaying renders the game world and HUD.
func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Wire v3.0 systems to renderer (Step 28)
	g.renderer.SetTextureAtlas(g.textureAtlas)
	g.renderer.SetLightMap(g.lightMap)
	g.renderer.SetPostProcessor(g.postProcessor)
	g.renderer.Tick() // Increment animation ticker

	// Render 3D world
	g.renderer.Render(screen, g.camera.X, g.camera.Y, g.camera.DirX, g.camera.DirY, g.camera.Pitch)

	// Render props as sprites in world space
	if g.propsManager != nil {
		g.renderProps(screen)
	}

	// Render lore items as sprites in world space
	if len(g.loreItems) > 0 {
		g.renderLoreItems(screen)
	}

	// Render particles on top of 3D world (simple sprite overlay for now)
	// TODO: Add particle rendering to renderer or as separate overlay
	if g.particleSystem != nil {
		g.renderParticles(screen)
	}

	// Render automap overlay if visible
	if g.automapVisible && g.automap != nil {
		g.drawAutomap(screen)
	}

	// Render HUD
	g.hud.Update()
	ui.DrawHUD(screen, g.hud)

	// Render quest objectives (top-right corner)
	if g.questTracker != nil {
		g.drawQuestObjectives(screen)
	}

	// Render tutorial prompts
	if g.tutorialSystem.Active {
		ui.DrawTutorial(screen, g.tutorialSystem.Current)
	}
}

// renderParticles draws particles as simple colored pixels (placeholder implementation).
func (g *Game) renderParticles(screen *ebiten.Image) {
	// Simplified particle rendering - just draw colored points
	// A full implementation would project 3D particles to screen space
	particles := g.particleSystem.GetActiveParticles()
	for _, p := range particles {
		// Simple 2D projection (would need proper 3D-to-2D projection)
		dx := p.X - g.camera.X
		dy := p.Y - g.camera.Y
		dist := dx*dx + dy*dy
		if dist < 400 { // Only render nearby particles
			// Very simplified screen position calculation
			screenX := config.C.InternalWidth/2 + int(dx*10)
			screenY := config.C.InternalHeight/2 + int(dy*10)
			if screenX >= 0 && screenX < config.C.InternalWidth && screenY >= 0 && screenY < config.C.InternalHeight {
				// Draw a small colored rectangle
				particleColor := color.RGBA{R: p.R, G: p.G, B: p.B, A: p.A}
				vector.DrawFilledRect(screen, float32(screenX), float32(screenY), 2, 2, particleColor, false)
			}
		}
	}
}

// renderProps draws decorative props as simple sprites in world space.
func (g *Game) renderProps(screen *ebiten.Image) {
	allProps := g.propsManager.GetProps()

	// Calculate camera plane from direction and FOV
	fov := g.camera.FOV
	planeX := -g.camera.DirY * fov / 66.0 // Standard plane calculation
	planeY := g.camera.DirX * fov / 66.0

	for _, prop := range allProps {
		// Calculate vector from camera to prop
		dx := prop.X - g.camera.X
		dy := prop.Y - g.camera.Y

		// Only render props within visible range
		dist := dx*dx + dy*dy
		if dist > 400 { // Skip distant props
			continue
		}

		// Transform prop position to camera space
		invDet := 1.0 / (planeX*g.camera.DirY - g.camera.DirX*planeY)
		transformX := invDet * (g.camera.DirY*dx - g.camera.DirX*dy)
		transformY := invDet * (-planeY*dx + planeX*dy)

		// Skip props behind camera
		if transformY <= 0.1 {
			continue
		}

		// Calculate screen X position
		spriteScreenX := int((float64(config.C.InternalWidth) / 2.0) * (1.0 + transformX/transformY))

		// Calculate sprite height based on distance
		spriteHeight := int(float64(config.C.InternalHeight) / transformY)
		spriteWidth := spriteHeight // Square sprites for simplicity

		// Draw bounds
		drawStartX := spriteScreenX - spriteWidth/2
		drawEndX := spriteScreenX + spriteWidth/2
		drawStartY := config.C.InternalHeight/2 - spriteHeight/2
		drawEndY := config.C.InternalHeight/2 + spriteHeight/2

		// Clip to screen bounds
		if drawEndX < 0 || drawStartX >= config.C.InternalWidth {
			continue
		}
		if drawStartX < 0 {
			drawStartX = 0
		}
		if drawEndX >= config.C.InternalWidth {
			drawEndX = config.C.InternalWidth - 1
		}

		// Choose color based on prop type (placeholder - will be replaced with actual sprites)
		var propColor color.RGBA
		switch prop.SpriteType {
		case props.PropBarrel:
			propColor = color.RGBA{139, 69, 19, 255} // Brown
		case props.PropCrate:
			propColor = color.RGBA{160, 82, 45, 255} // Saddle brown
		case props.PropTable:
			propColor = color.RGBA{101, 67, 33, 255} // Dark brown
		case props.PropTerminal:
			propColor = color.RGBA{50, 50, 200, 255} // Blue
		case props.PropBones:
			propColor = color.RGBA{220, 220, 200, 255} // Bone white
		case props.PropPlant:
			propColor = color.RGBA{34, 139, 34, 255} // Forest green
		case props.PropPillar:
			propColor = color.RGBA{128, 128, 128, 255} // Gray
		case props.PropTorch:
			propColor = color.RGBA{255, 165, 0, 255} // Orange
		case props.PropDebris:
			propColor = color.RGBA{105, 105, 105, 255} // Dim gray
		case props.PropContainer:
			propColor = color.RGBA{192, 192, 192, 255} // Silver
		default:
			propColor = color.RGBA{128, 128, 128, 255} // Gray default
		}

		// Draw simplified sprite as a colored rectangle
		vector.DrawFilledRect(screen,
			float32(drawStartX), float32(drawStartY),
			float32(drawEndX-drawStartX), float32(drawEndY-drawStartY),
			propColor, false)
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

// drawQuestObjectives renders quest objectives on screen.
func (g *Game) drawQuestObjectives(screen *ebiten.Image) {
	if g.questTracker == nil {
		return
	}

	// Get active objectives
	mainObjs := g.questTracker.GetMainObjectives()
	bonusObjs := g.questTracker.GetBonusObjectives()

	// Position at top-right corner
	bounds := screen.Bounds()
	w := float32(bounds.Dx())
	startX := w - 250
	startY := float32(10)

	// Draw background
	bgHeight := float32(20 + (len(mainObjs)+len(bonusObjs))*15)
	vector.DrawFilledRect(screen, startX, startY, 240, bgHeight, color.RGBA{0, 0, 0, 150}, false)
	vector.StrokeRect(screen, startX, startY, 240, bgHeight, 1, color.RGBA{100, 100, 100, 200}, false)

	// Note: We can't render text without adding a text rendering system
	// For now, this creates the UI box where objectives would be displayed
	// A full implementation would use ebitenutil.DebugPrintAt or a proper text renderer

	// Draw placeholder indicator for each objective (colored dots)
	y := startY + 10
	for range mainObjs {
		objColor := color.RGBA{255, 200, 50, 255} // Yellow for main objectives
		vector.DrawFilledCircle(screen, startX+10, y, 3, objColor, false)
		y += 15
	}
	for range bonusObjs {
		objColor := color.RGBA{100, 150, 255, 255} // Blue for bonus objectives
		vector.DrawFilledCircle(screen, startX+10, y, 3, objColor, false)
		y += 15
	}
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

// renderLoreItems draws lore items as simple sprites in world space.
func (g *Game) renderLoreItems(screen *ebiten.Image) {
	fov := g.camera.FOV
	planeX := -g.camera.DirY * fov / 66.0
	planeY := g.camera.DirX * fov / 66.0

	for _, loreItem := range g.loreItems {
		if loreItem.Activated {
			continue
		}

		dx := loreItem.PosX - g.camera.X
		dy := loreItem.PosY - g.camera.Y
		dist := dx*dx + dy*dy
		if dist > 400 {
			continue
		}

		invDet := 1.0 / (planeX*g.camera.DirY - g.camera.DirX*planeY)
		transformX := invDet * (g.camera.DirY*dx - g.camera.DirX*dy)
		transformY := invDet * (-planeY*dx + planeX*dy)

		if transformY <= 0.1 {
			continue
		}

		spriteScreenX := int((float64(config.C.InternalWidth) / 2.0) * (1.0 + transformX/transformY))
		spriteHeight := int(float64(config.C.InternalHeight) / transformY / 2)
		spriteWidth := spriteHeight

		drawStartX := spriteScreenX - spriteWidth/2
		drawEndX := spriteScreenX + spriteWidth/2
		drawStartY := config.C.InternalHeight/2 - spriteHeight/2
		drawEndY := config.C.InternalHeight/2 + spriteHeight/2

		if drawEndX < 0 || drawStartX >= config.C.InternalWidth {
			continue
		}
		if drawStartX < 0 {
			drawStartX = 0
		}
		if drawEndX >= config.C.InternalWidth {
			drawEndX = config.C.InternalWidth - 1
		}

		// Color based on lore item type - pulsing glow effect
		pulse := float32(0.7 + 0.3*float64(g.animationTicker%60)/60.0)
		var loreColor color.RGBA
		switch loreItem.Type {
		case lore.LoreItemNote:
			loreColor = color.RGBA{uint8(255 * pulse), uint8(255 * pulse), 200, 255}
		case lore.LoreItemAudioLog:
			loreColor = color.RGBA{100, uint8(200 * pulse), uint8(255 * pulse), 255}
		case lore.LoreItemGraffiti:
			loreColor = color.RGBA{uint8(255 * pulse), 100, 100, 255}
		case lore.LoreItemBodyArrangement:
			loreColor = color.RGBA{150, 150, uint8(150 * pulse), 255}
		}

		vector.DrawFilledRect(screen, float32(drawStartX), float32(drawStartY),
			float32(drawEndX-drawStartX), float32(drawEndY-drawStartY), loreColor, false)
	}
}

// drawCodex renders the lore codex UI overlay.
func (g *Game) drawCodex(screen *ebiten.Image) {
	// Draw semi-transparent background
	bgColor := color.RGBA{0, 0, 0, 200}
	vector.DrawFilledRect(screen, 0, 0, float32(config.C.InternalWidth), float32(config.C.InternalHeight), bgColor, false)

	// Draw simple border
	borderColor := color.RGBA{100, 100, 150, 255}
	vector.StrokeRect(screen, 20, 20, float32(config.C.InternalWidth-40), float32(config.C.InternalHeight-40), 2, borderColor, false)

	// Get found entries
	foundEntries := g.loreCodex.GetFoundEntries()
	if len(foundEntries) == 0 {
		// No lore discovered - show centered message using existing HUD methods
		g.hud.ShowMessage("No lore discovered yet. Explore to find lore items!")
		return
	}

	// Show current entry
	if g.codexScrollIdx >= len(foundEntries) {
		g.codexScrollIdx = len(foundEntries) - 1
	}
	if g.codexScrollIdx < 0 {
		g.codexScrollIdx = 0
	}

	entry := foundEntries[g.codexScrollIdx]

	// For now, just display basic info using HUD message system
	// Future: implement proper text rendering
	displayText := entry.Title + " | " + entry.Category + " | Entry " +
		string(rune(g.codexScrollIdx+1+'0')) + "/" + string(rune(len(foundEntries)+'0'))
	g.hud.ShowMessage(displayText)
}

// drawMinigame renders the active minigame interface.
func (g *Game) drawMinigame(screen *ebiten.Image) {
	if g.activeMinigame == nil {
		return
	}

	// Draw dimmed background
	bgColor := color.RGBA{0, 0, 0, 180}
	vector.DrawFilledRect(screen, 0, 0, float32(config.C.InternalWidth), float32(config.C.InternalHeight), bgColor, false)

	centerX := float32(config.C.InternalWidth / 2)
	centerY := float32(config.C.InternalHeight / 2)

	// Draw minigame-specific UI
	switch g.minigameType {
	case "lockpick":
		g.drawLockpickGame(screen, centerX, centerY)
	case "hack":
		g.drawHackGame(screen, centerX, centerY)
	case "circuit":
		g.drawCircuitGame(screen, centerX, centerY)
	case "code":
		g.drawCodeGame(screen, centerX, centerY)
	}

	// Draw progress bar
	progress := g.activeMinigame.GetProgress()
	barWidth := float32(200)
	barHeight := float32(20)
	barX := centerX - barWidth/2
	barY := centerY + 120

	// Background
	vector.DrawFilledRect(screen, barX, barY, barWidth, barHeight, color.RGBA{50, 50, 50, 255}, false)
	// Progress fill
	vector.DrawFilledRect(screen, barX, barY, barWidth*float32(progress), barHeight, color.RGBA{0, 255, 0, 255}, false)
	// Border
	vector.StrokeRect(screen, barX, barY, barWidth, barHeight, 2, color.RGBA{255, 255, 255, 255}, false)

	// Draw attempts remaining
	attempts := g.activeMinigame.GetAttempts()
	// Simple attempt indicators as circles
	for i := 0; i < attempts; i++ {
		circleX := centerX - 30 + float32(i*20)
		circleY := centerY + 150
		vector.DrawFilledCircle(screen, circleX, circleY, 5, color.RGBA{255, 255, 0, 255}, false)
	}
}

// drawLockpickGame renders lockpicking interface.
func (g *Game) drawLockpickGame(screen *ebiten.Image, centerX, centerY float32) {
	lpGame, ok := g.activeMinigame.(*minigame.LockpickGame)
	if !ok {
		return
	}

	// Draw lock cylinder
	cylinderWidth := float32(200)
	cylinderHeight := float32(30)
	cylinderX := centerX - cylinderWidth/2
	cylinderY := centerY - cylinderHeight/2

	vector.DrawFilledRect(screen, cylinderX, cylinderY, cylinderWidth, cylinderHeight, color.RGBA{100, 100, 100, 255}, false)
	vector.StrokeRect(screen, cylinderX, cylinderY, cylinderWidth, cylinderHeight, 2, color.RGBA{200, 200, 200, 255}, false)

	// Draw target zone
	targetX := cylinderX + cylinderWidth*float32(lpGame.Target)
	targetWidth := cylinderWidth * float32(lpGame.Tolerance*2)
	vector.DrawFilledRect(screen, targetX-targetWidth/2, cylinderY, targetWidth, cylinderHeight, color.RGBA{0, 200, 0, 100}, false)

	// Draw lockpick position
	pickX := cylinderX + cylinderWidth*float32(lpGame.Position)
	vector.DrawFilledRect(screen, pickX-2, cylinderY-10, 4, cylinderHeight+20, color.RGBA{255, 255, 0, 255}, false)

	// Draw unlocked pins
	for i := 0; i < lpGame.UnlockedPins; i++ {
		pinX := centerX - 50 + float32(i*25)
		pinY := centerY - 60
		vector.DrawFilledCircle(screen, pinX, pinY, 8, color.RGBA{0, 255, 0, 255}, false)
	}
}

// drawHackGame renders hacking interface.
func (g *Game) drawHackGame(screen *ebiten.Image, centerX, centerY float32) {
	hackGame, ok := g.activeMinigame.(*minigame.HackGame)
	if !ok {
		return
	}

	// Draw node grid (6 nodes in a circle)
	nodeRadius := float32(80)
	for i := 0; i < 6; i++ {
		angle := float32(i) * 3.14159 * 2.0 / 6.0
		nodeX := centerX + nodeRadius*float32(cosf(angle))
		nodeY := centerY + nodeRadius*float32(sinf(angle))

		nodeColor := color.RGBA{100, 100, 200, 255}
		// Highlight nodes in sequence
		for j, node := range hackGame.Sequence {
			if j < len(hackGame.PlayerInput) && node == i {
				nodeColor = color.RGBA{0, 255, 0, 255}
			}
		}

		vector.DrawFilledCircle(screen, nodeX, nodeY, 15, nodeColor, false)
		vector.StrokeCircle(screen, nodeX, nodeY, 15, 2, color.RGBA{255, 255, 255, 255}, false)
	}

	// Draw sequence indicators at top
	for i := range hackGame.Sequence {
		boxX := centerX - 60 + float32(i*20)
		boxY := centerY - 100
		boxColor := color.RGBA{50, 50, 50, 255}
		if i < len(hackGame.PlayerInput) {
			boxColor = color.RGBA{0, 200, 0, 255}
		}
		vector.DrawFilledRect(screen, boxX, boxY, 15, 15, boxColor, false)
		vector.StrokeRect(screen, boxX, boxY, 15, 15, 1, color.RGBA{200, 200, 200, 255}, false)
	}
}

// drawCircuitGame renders circuit trace interface.
func (g *Game) drawCircuitGame(screen *ebiten.Image, centerX, centerY float32) {
	circuitGame, ok := g.activeMinigame.(*minigame.CircuitTraceGame)
	if !ok {
		return
	}

	gridSize := len(circuitGame.Grid)
	cellSize := float32(30)
	startX := centerX - float32(gridSize)*cellSize/2
	startY := centerY - float32(gridSize)*cellSize/2

	// Draw grid
	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			cellX := startX + float32(x)*cellSize
			cellY := startY + float32(y)*cellSize

			cellColor := color.RGBA{50, 50, 50, 255}
			if circuitGame.Grid[y][x] == 2 {
				cellColor = color.RGBA{200, 0, 0, 255} // blocked
			}
			if x == circuitGame.CurrentX && y == circuitGame.CurrentY {
				cellColor = color.RGBA{0, 255, 0, 255} // current
			}
			if x == circuitGame.TargetX && y == circuitGame.TargetY {
				cellColor = color.RGBA{0, 200, 255, 255} // target
			}

			vector.DrawFilledRect(screen, cellX, cellY, cellSize-2, cellSize-2, cellColor, false)
			vector.StrokeRect(screen, cellX, cellY, cellSize-2, cellSize-2, 1, color.RGBA{100, 100, 100, 255}, false)
		}
	}
}

// drawCodeGame renders bypass code interface.
func (g *Game) drawCodeGame(screen *ebiten.Image, centerX, centerY float32) {
	codeGame, ok := g.activeMinigame.(*minigame.BypassCodeGame)
	if !ok {
		return
	}

	codeLength := len(codeGame.Code)
	digitWidth := float32(30)
	digitHeight := float32(40)
	totalWidth := float32(codeLength) * (digitWidth + 5)
	startX := centerX - totalWidth/2

	// Draw code entry boxes
	for i := 0; i < codeLength; i++ {
		boxX := startX + float32(i)*(digitWidth+5)
		boxY := centerY - digitHeight/2

		boxColor := color.RGBA{30, 30, 30, 255}
		if i < len(codeGame.PlayerInput) {
			boxColor = color.RGBA{0, 150, 0, 255}
		}

		vector.DrawFilledRect(screen, boxX, boxY, digitWidth, digitHeight, boxColor, false)
		vector.StrokeRect(screen, boxX, boxY, digitWidth, digitHeight, 2, color.RGBA{200, 200, 200, 255}, false)
	}
}

// cosf is a helper for float32 cosine.
func cosf(angle float32) float32 {
	return float32(math.Cos(float64(angle)))
}

// sinf is a helper for float32 sine.
func sinf(angle float32) float32 {
	return float32(math.Sin(float64(angle)))
}

// findExitPosition finds the room furthest from player spawn as the exit location.
func (g *Game) findExitPosition(rooms []*bsp.Room, playerX, playerY float64) *quest.Position {
	if len(rooms) == 0 {
		// Fallback to center of map if no rooms available
		return &quest.Position{X: 60, Y: 60}
	}

	// Find the room furthest from player spawn
	maxDist := 0.0
	var exitRoom *bsp.Room
	for _, room := range rooms {
		// Use room center for distance calculation
		roomCenterX := float64(room.X + room.W/2)
		roomCenterY := float64(room.Y + room.H/2)
		dx := roomCenterX - playerX
		dy := roomCenterY - playerY
		dist := dx*dx + dy*dy // squared distance is sufficient for comparison

		if dist > maxDist {
			maxDist = dist
			exitRoom = room
		}
	}

	// Return center of the furthest room
	if exitRoom != nil {
		return &quest.Position{
			X: float64(exitRoom.X + exitRoom.W/2),
			Y: float64(exitRoom.Y + exitRoom.H/2),
		}
	}

	// Fallback
	return &quest.Position{X: 60, Y: 60}
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
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(config.C.VSync)
	ebiten.SetFullscreen(config.C.FullScreen)
	ebiten.SetWindowTitle("VIOLENCE")
	ebiten.SetCursorMode(ebiten.CursorModeVisible) // Start visible for menu

	// Set TPS cap (0 = unlimited, 60 = default)
	if config.C.MaxTPS > 0 {
		ebiten.SetTPS(config.C.MaxTPS)
	}

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
