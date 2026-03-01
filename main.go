package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"net"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/ai"
	"github.com/opd-ai/violence/pkg/ammo"
	"github.com/opd-ai/violence/pkg/audio"
	"github.com/opd-ai/violence/pkg/automap"
	"github.com/opd-ai/violence/pkg/bsp"
	"github.com/opd-ai/violence/pkg/camera"
	"github.com/opd-ai/violence/pkg/chat"
	"github.com/opd-ai/violence/pkg/class"
	"github.com/opd-ai/violence/pkg/combat"
	"github.com/opd-ai/violence/pkg/config"
	"github.com/opd-ai/violence/pkg/crafting"
	"github.com/opd-ai/violence/pkg/destruct"
	"github.com/opd-ai/violence/pkg/door"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/event"
	"github.com/opd-ai/violence/pkg/federation"
	"github.com/opd-ai/violence/pkg/hazard"
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
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font/basicfont"
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
	shadowSystem    *lighting.ShadowSystem
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
	networkConn     net.Conn    // Active network connection for key exchange
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

	// E2E encrypted chat system
	chatManager     *chat.Chat
	chatInput       string   // Current chat message being typed
	chatMessages    []string // Recent chat messages to display
	chatInputActive bool     // Whether chat input is active

	// Environmental hazard system
	hazardSystem *hazard.System

	// Enemy role and squad tactics system
	roleBasedAISystem *ai.RoleBasedAISystem
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
		shadowSystem:    lighting.NewShadowSystem(config.C.InternalWidth, config.C.InternalHeight, "fantasy"),
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
		hazardSystem:       hazard.NewSystem(int64(seed)),
		roleBasedAISystem:  ai.NewRoleBasedAISystem(),
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
	g.state = StateLoading
	g.loadingScreen.Show(g.seed, "Generating level...")

	g.generateLevel()
	g.populateLevel()
	g.initializePlayer()
	g.initializeGameSystems()
	g.finalizeGameStart()
}

// generateLevel generates the BSP level and initializes core map systems.
func (g *Game) generateLevel() {
	g.bspGenerator.SetGenre(g.genreID)
	bspTree, tiles := g.bspGenerator.Generate()
	g.currentMap = tiles
	g.currentBSPTree = bspTree
	g.raycaster.SetMap(tiles)

	if len(tiles) > 0 && len(tiles[0]) > 0 {
		g.automap = automap.NewMap(len(tiles[0]), len(tiles))
		g.lightMap = lighting.NewSectorLightMap(len(tiles[0]), len(tiles), 0.3)
		g.weatherEmitter = particle.NewWeatherEmitter(g.particleSystem, g.genreID, 0, 0, float64(len(tiles[0])), float64(len(tiles)))
	} else {
		g.lightMap = lighting.NewSectorLightMap(0, 0, 0.3)
		g.weatherEmitter = nil
	}
	g.setGenre(g.genreID)
}

// populateLevel populates the generated level with content and entities.
func (g *Game) populateLevel() {
	rooms := bsp.GetRooms(g.currentBSPTree)
	g.placeDecorativeProps(rooms)
	g.placeLoreItems(rooms)
	g.scanSecretWalls()
	g.spawnEnemies()
	g.spawnDestructibles()
	g.initializeSquad()
	g.setupQuests(rooms)
	g.setupEventTriggers()
	g.generateHazards()
}

// placeDecorativeProps places decorative props in BSP rooms.
func (g *Game) placeDecorativeProps(rooms []*bsp.Room) {
	g.propsManager.Clear()
	g.propsManager.SetGenre(g.genreID)
	for _, room := range rooms {
		propRoom := &props.Room{X: room.X, Y: room.Y, W: room.W, H: room.H}
		g.propsManager.PlaceProps(propRoom, 0.2, g.seed+uint64(room.X*1000+room.Y))
	}
}

// placeLoreItems generates and places lore items in level rooms.
func (g *Game) placeLoreItems(rooms []*bsp.Room) {
	g.loreItems = make([]*lore.LoreItem, 0)
	g.loreGenerator.SetGenre(g.genreID)
	loreItemsPerLevel := 5 + len(rooms)/3
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
}

// scanSecretWalls scans the map for secret walls and registers them.
func (g *Game) scanSecretWalls() {
	tiles := g.currentMap
	g.secretManager = secret.NewManager(len(tiles[0]))
	for y := 0; y < len(tiles); y++ {
		for x := 0; x < len(tiles[y]); x++ {
			if tiles[y][x] == bsp.TileSecret {
				dir := g.determineSecretDirection(x, y, tiles)
				g.secretManager.Add(x, y, dir)
			}
		}
	}
}

// determineSecretDirection determines the slide direction for a secret wall.
func (g *Game) determineSecretDirection(x, y int, tiles [][]int) secret.Direction {
	dir := secret.DirNorth
	if y+1 < len(tiles) && isWalkableTile(tiles[y+1][x]) {
		dir = secret.DirSouth
	} else if y > 0 && isWalkableTile(tiles[y-1][x]) {
		dir = secret.DirNorth
	} else if x+1 < len(tiles[y]) && isWalkableTile(tiles[y][x+1]) {
		dir = secret.DirEast
	} else if x > 0 && isWalkableTile(tiles[y][x-1]) {
		dir = secret.DirWest
	}
	return dir
}

// spawnEnemies spawns AI enemies in the level.
func (g *Game) spawnEnemies() {
	g.aiAgents = make([]*ai.Agent, 0)
	ai.SetGenre(g.genreID)
	rooms := bsp.GetRooms(g.currentBSPTree)
	for i := 0; i < 3; i++ {
		var spawnX, spawnY float64
		if i+1 < len(rooms) {
			// Spawn in different rooms, skip room 0 (player spawn)
			r := rooms[i+1]
			spawnX = float64(r.X+r.W/2) + 0.5
			spawnY = float64(r.Y+r.H/2) + 0.5
		} else if len(rooms) > 1 {
			r := rooms[len(rooms)-1]
			spawnX = float64(r.X+r.W/2) + 0.5
			spawnY = float64(r.Y+r.H/2) + 0.5
		} else {
			spawnX = float64(10 + i*5)
			spawnY = float64(10 + i*3)
		}
		agent := ai.NewAgent("enemy_"+string(rune(i+'0')), spawnX, spawnY)
		g.aiAgents = append(g.aiAgents, agent)
	}
}

// spawnDestructibles spawns destructible objects like barrels and crates.
func (g *Game) spawnDestructibles() {
	g.destructibleSystem = destruct.NewSystem()
	destruct.SetGenre(g.genreID)
	rooms := bsp.GetRooms(g.currentBSPTree)
	for i := 0; i < 5; i++ {
		// Place destructibles inside BSP rooms so they're on walkable tiles
		var bx, by, cx, cy float64
		if i < len(rooms) {
			r := rooms[i]
			bx = float64(r.X+1) + 0.5
			by = float64(r.Y+1) + 0.5
			cx = float64(r.X+r.W-2) + 0.5
			cy = float64(r.Y+r.H-2) + 0.5
		} else {
			bx = float64(15 + i*4)
			by = float64(8 + i*2)
			cx = float64(12 + i*3)
			cy = float64(12 + i*3)
		}
		barrel := destruct.NewDestructibleObject(
			"barrel_"+string(rune(i+'0')),
			"barrel",
			50.0,
			bx,
			by,
			true,
		)
		barrel.AddDropItem("ammo_shells")
		g.destructibleSystem.Add(&barrel.Destructible)

		crate := destruct.NewDestructibleObject(
			"crate_"+string(rune(i+'0')),
			"crate",
			30.0,
			cx,
			cy,
			false,
		)
		crate.AddDropItem("health_small")
		g.destructibleSystem.Add(&crate.Destructible)
	}
}

// initializeSquad initializes squad companions near the player.
func (g *Game) initializeSquad() {
	g.squadCompanions = squad.NewSquad(3)
	squad.SetGenre(g.genreID)
	g.squadCompanions.AddMember("companion_1", "grunt", "assault_rifle", g.camera.X-2, g.camera.Y+1, g.seed)
	g.squadCompanions.AddMember("companion_2", "medic", "pistol", g.camera.X-2, g.camera.Y-1, g.seed)
}

// setupQuests initializes the quest tracker with level objectives.
func (g *Game) setupQuests(rooms []*bsp.Room) {
	g.questTracker = quest.NewTracker()
	g.questTracker.SetGenre(g.genreID)

	questRooms := make([]quest.Room, len(rooms))
	for i, r := range rooms {
		questRooms[i] = quest.Room{X: r.X, Y: r.Y, Width: r.W, Height: r.H}
	}

	exitPos := g.findExitPosition(rooms, g.camera.X, g.camera.Y)
	layout := quest.LevelLayout{
		Width:       len(g.currentMap[0]),
		Height:      len(g.currentMap),
		ExitPos:     exitPos,
		SecretCount: len(g.secretManager.GetAll()),
		Rooms:       questRooms,
	}
	g.questTracker.GenerateWithLayout(g.seed, layout)
}

// setupEventTriggers initializes event triggers for alarms, lockdowns, and boss arenas.
func (g *Game) setupEventTriggers() {
	g.alarmTrigger = event.NewAlarmTrigger("alarm_1", 30.0)
	g.lockdownTrigger = event.NewTimedLockdown("lockdown_1", 180.0)

	centerX := len(g.currentMap[0]) / 2
	centerY := len(g.currentMap) / 2
	g.bossArena = event.NewBossArenaEvent("boss_1", "center_room", 3, 5.0)

	if int(g.camera.X) == centerX && int(g.camera.Y) == centerY {
		g.bossArena.Trigger()
	}

	event.SetGenre(g.genreID)
}

// generateHazards creates environmental hazards for the current level.
func (g *Game) generateHazards() {
	if g.hazardSystem != nil && g.currentMap != nil {
		g.hazardSystem.SetGenre(g.genreID)
		g.hazardSystem.GenerateHazards(g.currentMap, int64(g.seed))
	}
}

// initializePlayer sets up the player's starting position, stats, and equipment.
func (g *Game) initializePlayer() {
	rooms := bsp.GetRooms(g.currentBSPTree)
	spawnX, spawnY := g.findSpawnPosition(rooms)

	g.camera.X = spawnX
	g.camera.Y = spawnY
	g.camera.DirX = 1.0
	g.camera.DirY = 0.0
	g.camera.Pitch = 0.0

	g.hud.Health = 100
	g.hud.Armor = 0
	g.hud.MaxHealth = 100
	g.hud.MaxArmor = 100

	g.ammoPool.Add("bullets", 50)
	g.ammoPool.Add("shells", 8)
	g.ammoPool.Add("cells", 20)
	g.ammoPool.Add("rockets", 0)

	currentWeapon := g.arsenal.GetCurrentWeapon()
	g.hud.Ammo = g.ammoPool.Get(currentWeapon.AmmoType)

	g.keycards = make(map[string]bool)
	g.automapVisible = false
}

// findSpawnPosition finds a safe starting position for the player.
func (g *Game) findSpawnPosition(rooms []*bsp.Room) (float64, float64) {
	spawnX, spawnY := 5.0, 5.0
	if len(rooms) > 0 {
		spawnX = float64(rooms[0].X+rooms[0].W/2) + 0.5
		spawnY = float64(rooms[0].Y+rooms[0].H/2) + 0.5
	} else {
		tiles := g.currentMap
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
	return spawnX, spawnY
}

// initializeGameSystems initializes progression, shop, crafting, and mod systems.
func (g *Game) initializeGameSystems() {
	g.progression = progression.NewProgression()
	if err := g.progression.SetGenre(g.genreID); err != nil {
		logrus.WithError(err).Warn("Failed to set progression genre")
	}
	g.statusReg = status.NewRegistry()
	status.SetGenre(g.genreID)

	g.shopCredits = shop.NewCredit(100)
	g.shopArmory = shop.NewArmory(g.genreID)
	g.shopInventory = &g.shopArmory.Inventory
	g.scrapStorage = crafting.NewScrapStorage()
	scrapName := crafting.GetScrapNameForGenre(g.genreID)
	g.scrapStorage.Add(scrapName, 10)
	g.craftingMenu = crafting.NewCraftingMenu(g.scrapStorage, g.genreID)
	g.craftingResult = ""

	g.skillManager = skills.NewManager()
	g.skillManager.AddPoints(3)
	g.skillsTreeIdx = 0
	g.skillsNodeIdx = 0

	if g.modLoader == nil {
		g.modLoader = mod.NewLoader()
	}
	g.scanMods()

	g.playerInventory = inventory.NewInventory()
	inventory.SetGenre(g.genreID)
}

// finalizeGameStart completes the game initialization and transitions to playing state.
func (g *Game) finalizeGameStart() {
	g.levelStartTime = time.Now()
	g.audioEngine.PlayMusic("theme", 0.5)
	g.loadingScreen.Hide()
	g.state = StatePlaying
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
	if err := g.progression.SetGenre(genreID); err != nil {
		logrus.WithError(err).Warn("Failed to set progression genre")
	}
	class.SetGenre(genreID)
	ai.SetGenre(genreID)

	// v3.0 systems
	g.textureAtlas.SetGenre(genreID)
	g.lightMap.SetGenre(genreID)
	g.shadowSystem.SetGenre(genreID)
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
	if g.hazardSystem != nil {
		g.hazardSystem.SetGenre(genreID)
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
		// Create new progression and add XP to restore state
		g.progression = progression.NewProgression()
		// Set level by adding appropriate XP
		// Calculate total XP needed to reach saved level
		totalXP := 0
		for lvl := 1; lvl < state.Progression.Level; lvl++ {
			totalXP += lvl * 100 // Base XP per level
		}
		totalXP += state.Progression.XP // Add remaining XP
		if err := g.progression.AddXP(totalXP); err != nil {
			logrus.WithError(err).Warn("Failed to restore progression XP")
		}
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

	// Restore inventory
	if g.playerInventory != nil && len(state.Inventory.Items) > 0 {
		g.playerInventory = inventory.NewInventory()
		for _, saveItem := range state.Inventory.Items {
			g.playerInventory.Add(inventory.Item{
				ID:   saveItem.ID,
				Name: saveItem.Name,
				Qty:  saveItem.Qty,
			})
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
	if handled := g.handleMenuActions(); handled {
		return nil
	}

	g.handlePlayerActions()
	g.handleWeaponFiring()

	g.arsenal.Update()
	g.updateAIAgents()
	g.statusReg.Tick()
	g.updateSquadAndEventTriggers()
	g.updateQuestObjectives()
	g.updateV3Systems()
	g.updateLightingAndAudio()

	g.animationTicker++

	deltaX, deltaY, deltaPitch := g.processPlayerMovement()
	g.handleCollisionAndMovement(deltaX, deltaY, deltaPitch)
	g.checkTutorialCompletion(deltaX, deltaY)

	return nil
}

// handleMenuActions processes menu-related input actions and returns true if handled.
func (g *Game) handleMenuActions() bool {
	if g.input.IsJustPressed(input.ActionPause) {
		g.state = StatePaused
		g.menuManager.Show(ui.MenuTypePause)
		return true
	}

	if g.input.IsJustPressed(input.ActionAutomap) {
		g.automapVisible = !g.automapVisible
	}

	if g.input.IsJustPressed(input.ActionShop) {
		g.openShop()
		return true
	}

	if g.input.IsJustPressed(input.ActionCraft) {
		g.openCrafting()
		return true
	}

	if g.input.IsJustPressed(input.ActionSkills) {
		g.openSkills()
		return true
	}

	if g.input.IsJustPressed(input.ActionMultiplayer) {
		g.openMultiplayer()
		return true
	}

	if g.input.IsJustPressed(input.ActionCodex) {
		g.state = StateCodex
		g.codexScrollIdx = 0
		return true
	}

	return false
}

// handlePlayerActions processes player interaction actions.
func (g *Game) handlePlayerActions() {
	if g.input.IsJustPressed(input.ActionUseItem) {
		g.useQuickSlotItem()
	}

	if g.input.IsJustPressed(input.ActionInteract) {
		g.tryCollectLore()
		g.tryInteractDoor()
	}
}

// handleWeaponFiring processes weapon firing and hit detection.
func (g *Game) handleWeaponFiring() {
	if !g.input.IsJustPressed(input.ActionFire) {
		return
	}

	currentWeapon := g.arsenal.GetCurrentWeapon()
	if currentWeapon.Name == "" {
		return
	}

	ammoType := currentWeapon.AmmoType
	availableAmmo := g.ammoPool.Get(ammoType)

	if currentWeapon.Type != weapon.TypeMelee && availableAmmo <= 0 {
		return
	}

	raycastFn := g.createEnemyRaycastFunction()
	hitResults := g.arsenal.Fire(g.camera.X, g.camera.Y, g.camera.DirX, g.camera.DirY, raycastFn)

	if currentWeapon.Type != weapon.TypeMelee {
		g.ammoPool.Consume(ammoType, 1)
		g.hud.Ammo = g.ammoPool.Get(ammoType)
	}

	g.processWeaponHits(hitResults, currentWeapon)
	g.checkDestructibleHits(hitResults, currentWeapon)
	g.audioEngine.PlaySFX("weapon_fire", g.camera.X, g.camera.Y)
}

// createEnemyRaycastFunction creates a raycast function for enemy hit detection.
func (g *Game) createEnemyRaycastFunction() func(float64, float64, float64, float64, float64) (bool, float64, float64, float64, uint64) {
	return func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64) {
		for i, agent := range g.aiAgents {
			if agent.Health <= 0 {
				continue
			}
			agentDist := (agent.X-x)*(agent.X-x) + (agent.Y-y)*(agent.Y-y)
			if agentDist < maxDist*maxDist {
				toAgentX := agent.X - x
				toAgentY := agent.Y - y
				dot := toAgentX*dx + toAgentY*dy
				if dot > 0 {
					return true, agentDist, agent.X, agent.Y, uint64(i + 1)
				}
			}
		}
		return false, 0, 0, 0, 0
	}
}

// processWeaponHits applies damage to enemies hit by weapon fire.
func (g *Game) processWeaponHits(hitResults []weapon.HitResult, currentWeapon weapon.Weapon) {
	for _, hitResult := range hitResults {
		if !hitResult.Hit || hitResult.EntityID == 0 {
			continue
		}

		agentIdx := int(hitResult.EntityID - 1)
		if agentIdx < 0 || agentIdx >= len(g.aiAgents) {
			continue
		}

		agent := g.aiAgents[agentIdx]
		if agent.Health <= 0 {
			continue
		}

		upgradedDamage := g.getUpgradedWeaponDamage(currentWeapon)
		agent.Health -= upgradedDamage

		if g.masteryManager != nil {
			g.masteryManager.AddMasteryXP(g.arsenal.CurrentSlot, 10)
		}

		if agent.Health <= 0 {
			g.handleEnemyDeath()
		}
	}
}

// handleEnemyDeath processes enemy death rewards and progression.
func (g *Game) handleEnemyDeath() {
	oldLevel := g.progression.GetLevel()
	if err := g.progression.AddXP(50); err != nil {
		logrus.WithError(err).Warn("Failed to add XP")
	}

	newLevel := g.progression.GetLevel()
	if newLevel > oldLevel {
		if g.skillManager != nil {
			g.skillManager.AddPoints(1)
		}
		g.hud.ShowMessage("Level Up! Skill point earned!")
	}

	if g.shopCredits != nil {
		g.shopCredits.Add(25)
	}

	if g.upgradeManager != nil {
		g.upgradeManager.GetTokens().Add(1)
	}

	if g.scrapStorage != nil {
		scrapName := crafting.GetScrapNameForGenre(g.genreID)
		g.scrapStorage.Add(scrapName, 3)
	}

	if g.questTracker != nil {
		g.questTracker.UpdateProgress("bonus_kills", 1)
	}
}

// checkDestructibleHits checks for and processes hits on destructible objects.
func (g *Game) checkDestructibleHits(hitResults []weapon.HitResult, currentWeapon weapon.Weapon) {
	if hitResults != nil && len(hitResults) > 0 {
		return
	}

	allDestructibles := g.destructibleSystem.GetAll()
	upgradedDamage := g.getUpgradedWeaponDamage(currentWeapon)

	for _, obj := range allDestructibles {
		if obj.IsDestroyed() {
			continue
		}

		objDist := (obj.X-g.camera.X)*(obj.X-g.camera.X) + (obj.Y-g.camera.Y)*(obj.Y-g.camera.Y)
		if objDist >= 100 {
			continue
		}

		toObjX := obj.X - g.camera.X
		toObjY := obj.Y - g.camera.Y
		dot := toObjX*g.camera.DirX + toObjY*g.camera.DirY
		if dot <= 0 {
			continue
		}

		destroyed := obj.Damage(upgradedDamage)
		if destroyed {
			g.handleDestructibleDestroyed(obj)
		}
		break
	}
}

// handleDestructibleDestroyed processes the destruction of a destructible object.
func (g *Game) handleDestructibleDestroyed(obj *destruct.Destructible) {
	if g.particleSystem != nil {
		debrisColor := color.RGBA{R: 100, G: 80, B: 60, A: 255}
		g.particleSystem.SpawnBurst(obj.X, obj.Y, 0, 15, 8.0, 1.0, 1.5, 1.0, debrisColor)
	}

	if g.scrapStorage != nil {
		scrapName := crafting.GetScrapNameForGenre(g.genreID)
		g.scrapStorage.Add(scrapName, 2)
	}

	if g.shopCredits != nil {
		g.shopCredits.Add(10)
	}

	g.audioEngine.PlaySFX("barrel_explode", obj.X, obj.Y)
}

// updateAIAgents updates all AI agents' behavior and combat actions.
func (g *Game) updateAIAgents() {
	for _, agent := range g.aiAgents {
		if agent.Health <= 0 {
			continue
		}

		dx := g.camera.X - agent.X
		dy := g.camera.Y - agent.Y
		distSq := dx*dx + dy*dy

		if distSq < 100 && agent.Cooldown <= 0 {
			g.handleAgentAttack(agent)
		}

		if agent.Cooldown > 0 {
			agent.Cooldown--
		}
	}
}

// handleAgentAttack processes an AI agent's attack on the player.
func (g *Game) handleAgentAttack(agent *ai.Agent) {
	damage := agent.Damage
	healthDamage := damage

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
	agent.Cooldown = 60
	g.audioEngine.PlaySFX("enemy_attack", agent.X, agent.Y)
	g.hud.ShowMessage("Taking damage!")

	if g.hud.Health <= 0 {
		g.hud.Health = 0
	}
}

// updateSquadAndEventTriggers updates squad companions and event trigger systems.
func (g *Game) updateSquadAndEventTriggers() {
	if g.squadCompanions != nil {
		g.squadCompanions.Update(g.camera.X, g.camera.Y, g.currentMap, g.camera.X, g.camera.Y, g.seed)
	}

	deltaTime := 1.0 / 60.0
	g.updateAlarmTrigger(deltaTime)
	g.updateLockdownTrigger(deltaTime)
	g.checkBossArenaTrigger()
}

// updateAlarmTrigger updates the alarm event trigger if active.
func (g *Game) updateAlarmTrigger(deltaTime float64) {
	if g.alarmTrigger != nil && g.alarmTrigger.IsActive() {
		g.alarmTrigger.Update(deltaTime)
	}
}

// updateLockdownTrigger updates the lockdown event trigger and checks time limits.
func (g *Game) updateLockdownTrigger(deltaTime float64) {
	if g.lockdownTrigger == nil || !g.lockdownTrigger.IsActive() {
		return
	}

	g.lockdownTrigger.Update(deltaTime)

	if g.questTracker == nil {
		return
	}

	remainingTime := g.lockdownTrigger.GetRemaining()
	if g.lockdownTrigger.IsExpired() {
		g.hud.ShowMessage("Lockdown complete - you are trapped!")
	} else if remainingTime < 10 {
		g.hud.ShowMessage("WARNING: 10 seconds remaining!")
	}
}

// checkBossArenaTrigger checks if the player has entered the boss arena.
func (g *Game) checkBossArenaTrigger() {
	if g.bossArena == nil || g.bossArena.IsTriggered() {
		return
	}

	centerX := float64(len(g.currentMap[0]) / 2)
	centerY := float64(len(g.currentMap) / 2)
	distToCenterSq := (g.camera.X-centerX)*(g.camera.X-centerX) + (g.camera.Y-centerY)*(g.camera.Y-centerY)

	if distToCenterSq >= 25 {
		return
	}

	g.bossArena.Trigger()
	_ = event.GenerateEventAudioSting(g.seed, event.EventBossArena)
	g.audioEngine.PlaySFX("boss_encounter", g.camera.X, g.camera.Y)
	eventText := event.GenerateEventText(g.seed, event.EventBossArena)
	g.hud.ShowMessage(eventText)
}

// updateQuestObjectives updates quest progress and speedrun timers.
func (g *Game) updateQuestObjectives() {
	if g.questTracker == nil {
		return
	}

	elapsedTime := time.Since(g.levelStartTime).Seconds()
	for i := range g.questTracker.Objectives {
		obj := &g.questTracker.Objectives[i]
		if obj.ID == "bonus_speed" && !obj.Complete {
			if elapsedTime > float64(obj.Count) {
				obj.Complete = false
			}
		}
	}
}

// updateV3Systems updates particle, weather, and secret wall systems.
func (g *Game) updateV3Systems() {
	deltaTime := 1.0 / 60.0

	if g.particleSystem != nil {
		g.particleSystem.Update(deltaTime)
	}

	if g.weatherEmitter != nil {
		g.weatherEmitter.Update(deltaTime)
	}

	if g.secretManager != nil {
		g.secretManager.Update(deltaTime)
	}

	if g.hazardSystem != nil {
		g.hazardSystem.Update(deltaTime)
		g.checkHazardCollisions()
	}

	// Update enemy role-based AI and squad tactics
	if g.roleBasedAISystem != nil {
		g.roleBasedAISystem.Update(g.world)
	}
}

// checkHazardCollisions tests player collision with environmental hazards.
func (g *Game) checkHazardCollisions() {
	if g.hazardSystem == nil {
		return
	}

	hit, damage, statusEffect := g.hazardSystem.CheckCollision(g.camera.X, g.camera.Y)
	if !hit {
		return
	}

	// Apply damage
	healthDamage := damage
	if g.hud.Armor > 0 {
		armorDamage := damage / 2
		g.hud.Armor -= armorDamage
		if g.hud.Armor < 0 {
			healthDamage = -g.hud.Armor
			g.hud.Armor = 0
		} else {
			healthDamage = damage / 2
		}
	}

	g.hud.Health -= healthDamage
	if g.hud.Health < 0 {
		g.hud.Health = 0
	}

	// Apply status effect if present
	if statusEffect != "" && g.statusReg != nil {
		g.statusReg.Apply(statusEffect)
		g.hud.ShowMessage("Hazard! " + statusEffect)
	}

	// Screen shake on hazard hit
	g.audioEngine.PlaySFX("hit", g.camera.X, g.camera.Y)
}

// updateLightingAndAudio updates lighting calculations and audio positioning.
func (g *Game) updateLightingAndAudio() {
	if g.lightMap != nil {
		preset := lighting.GetFlashlightPreset(g.genreID)
		flashlight := lighting.NewConeLight(g.camera.X, g.camera.Y, g.camera.DirX, g.camera.DirY, preset)
		g.lightMap.Clear()
		g.lightMap.AddLight(flashlight.GetContributionAsPointLight())
		g.lightMap.Calculate()
	}

	if g.currentBSPTree != nil {
		g.audioEngine.UpdateReverb(int(g.camera.X), int(g.camera.Y), g.currentBSPTree)
	}
}

// processPlayerMovement calculates player movement delta based on input.
func (g *Game) processPlayerMovement() (float64, float64, float64) {
	moveSpeed := 0.05
	rotSpeed := 0.03
	deltaX := 0.0
	deltaY := 0.0
	deltaPitch := 0.0

	if g.input.IsPressed(input.ActionMoveForward) {
		deltaX += g.camera.DirX * moveSpeed
		deltaY += g.camera.DirY * moveSpeed
	}
	if g.input.IsPressed(input.ActionMoveBackward) {
		deltaX -= g.camera.DirX * moveSpeed
		deltaY -= g.camera.DirY * moveSpeed
	}

	if g.input.IsPressed(input.ActionStrafeLeft) {
		deltaX += g.camera.DirY * moveSpeed
		deltaY -= g.camera.DirX * moveSpeed
	}
	if g.input.IsPressed(input.ActionStrafeRight) {
		deltaX -= g.camera.DirY * moveSpeed
		deltaY += g.camera.DirX * moveSpeed
	}

	g.processGamepadMovement(&deltaX, &deltaY, moveSpeed)
	g.processCameraRotation(rotSpeed, &deltaPitch)

	return deltaX, deltaY, deltaPitch
}

// processGamepadMovement processes gamepad stick input for player movement.
func (g *Game) processGamepadMovement(deltaX, deltaY *float64, moveSpeed float64) {
	leftX, leftY := g.input.GamepadLeftStick()
	deadzone := 0.15
	if leftX*leftX+leftY*leftY > deadzone*deadzone {
		*deltaX += (g.camera.DirX*leftY - g.camera.DirY*leftX) * moveSpeed
		*deltaY += (g.camera.DirY*leftY + g.camera.DirX*leftX) * moveSpeed
	}
}

// processCameraRotation processes keyboard, mouse, and gamepad camera rotation.
func (g *Game) processCameraRotation(rotSpeed float64, deltaPitch *float64) {
	if g.input.IsPressed(input.ActionTurnLeft) {
		g.camera.Rotate(-rotSpeed)
	}
	if g.input.IsPressed(input.ActionTurnRight) {
		g.camera.Rotate(rotSpeed)
	}

	mouseDX, mouseDY := g.input.MouseDelta()
	if mouseDX != 0 || mouseDY != 0 {
		sensitivity := config.C.MouseSensitivity * 0.002
		g.camera.Rotate(mouseDX * sensitivity)
		*deltaPitch = -mouseDY * sensitivity * 3.0
	}

	rightX, rightY := g.input.GamepadRightStick()
	deadzone := 0.15
	if rightX*rightX+rightY*rightY > deadzone*deadzone {
		g.camera.Rotate(rightX * rotSpeed * 1.5)
		*deltaPitch = -rightY * rotSpeed * 15.0
	}
}

// handleCollisionAndMovement applies collision detection and updates camera position.
func (g *Game) handleCollisionAndMovement(deltaX, deltaY, deltaPitch float64) {
	newX := g.camera.X + deltaX
	newY := g.camera.Y + deltaY

	if g.isWalkable(newX, newY) {
		g.camera.Update(deltaX, deltaY, 0, 0, deltaPitch)
	} else if g.isWalkable(newX, g.camera.Y) {
		g.camera.Update(deltaX, 0, 0, 0, deltaPitch)
	} else if g.isWalkable(g.camera.X, newY) {
		g.camera.Update(0, deltaY, 0, 0, deltaPitch)
	} else {
		g.camera.Update(0, 0, 0, 0, deltaPitch)
	}

	if g.automap != nil {
		g.automap.Reveal(int(g.camera.X), int(g.camera.Y))
	}

	g.world.Update()
	g.audioEngine.SetListenerPosition(g.camera.X, g.camera.Y)
}

// checkTutorialCompletion checks and completes active tutorial prompts based on player actions.
func (g *Game) checkTutorialCompletion(deltaX, deltaY float64) {
	if !g.tutorialSystem.Active {
		return
	}

	if g.tutorialSystem.Type == tutorial.PromptMovement && (deltaX != 0 || deltaY != 0) {
		g.tutorialSystem.Complete()
	}

	if g.input.IsJustPressed(input.ActionFire) || g.input.IsJustPressed(input.ActionInteract) {
		g.tutorialSystem.Complete()
	}
}

// isWalkableTile returns true if the tile type permits player movement.
func isWalkableTile(tile int) bool {
	switch {
	case tile == bsp.TileFloor:
		return true
	case tile == bsp.TileDoor:
		return false // Doors block movement until opened via interaction
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
	mapX, mapY, valid := g.getInteractionTileCoords()
	if !valid {
		return
	}

	tile := g.currentMap[mapY][mapX]

	if tile == bsp.TileSecret {
		g.handleSecretWall(mapX, mapY)
		return
	}

	if tile == bsp.TileDoor {
		g.handleDoorInteraction(mapX, mapY)
	}
}

// getInteractionTileCoords calculates the tile coordinates the player is facing.
func (g *Game) getInteractionTileCoords() (int, int, bool) {
	checkDist := 1.0
	checkX := g.camera.X + g.camera.DirX*checkDist
	checkY := g.camera.Y + g.camera.DirY*checkDist
	mapX := int(checkX)
	mapY := int(checkY)

	if g.currentMap == nil || len(g.currentMap) == 0 {
		return 0, 0, false
	}
	if mapY < 0 || mapY >= len(g.currentMap) || mapX < 0 || mapX >= len(g.currentMap[0]) {
		return 0, 0, false
	}

	return mapX, mapY, true
}

// handleSecretWall triggers a secret wall and updates quest progress.
func (g *Game) handleSecretWall(mapX, mapY int) {
	if g.secretManager != nil && g.secretManager.TriggerAt(mapX, mapY, "player") {
		// Convert the secret wall tile to floor so the player can walk through
		g.currentMap[mapY][mapX] = bsp.TileFloor
		g.raycaster.SetMap(g.currentMap)
		g.audioEngine.PlaySFX("secret_open", float64(mapX), float64(mapY))
		g.hud.ShowMessage("Secret discovered!")
		if g.questTracker != nil {
			g.questTracker.UpdateProgress("bonus_secrets", 1)
		}
	}
}

// handleDoorInteraction opens a door or starts lockpicking minigame.
func (g *Game) handleDoorInteraction(mapX, mapY int) {
	requiredColor := g.getDoorColor(mapX, mapY)
	if requiredColor == "" || g.keycards[requiredColor] {
		g.currentMap[mapY][mapX] = bsp.TileFloor
		g.raycaster.SetMap(g.currentMap)
		g.audioEngine.PlaySFX("door_open", float64(mapX), float64(mapY))
	} else {
		g.startMinigame(mapX, mapY)
	}
}

// startMinigame initiates a minigame for the current genre.
func (g *Game) startMinigame(doorX, doorY int) {
	// Determine difficulty based on progression level
	difficulty := g.progression.GetLevel() / 3
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
	case "ammo_bullets", "ammo_shells", "ammo_cells", "ammo_rockets", "ammo_arrows", "ammo_bolts":
		g.applyAmmoItem(itemID)
	case "medkit", "grenade", "plasma_grenade", "emp_grenade", "bomb", "proximity_mine":
		g.applyConsumableItem(itemID)
	case "armor_vest":
		g.applyArmorItem()
	case "upgrade_damage", "upgrade_firerate", "upgrade_clipsize", "upgrade_accuracy", "upgrade_range":
		g.applyWeaponUpgrade(itemID)
	}
	g.updateHUDAmmo()
}

// applyAmmoItem adds ammunition to the ammo pool.
func (g *Game) applyAmmoItem(itemID string) {
	ammoAmounts := map[string]int{
		"ammo_bullets": 20,
		"ammo_shells":  10,
		"ammo_cells":   15,
		"ammo_rockets": 5,
		"ammo_arrows":  20,
		"ammo_bolts":   10,
	}

	if amount, ok := ammoAmounts[itemID]; ok {
		ammoType := itemID[5:]
		g.ammoPool.Add(ammoType, amount)
	}
}

// applyConsumableItem adds consumable items to inventory.
func (g *Game) applyConsumableItem(itemID string) {
	switch itemID {
	case "medkit":
		g.playerInventory.Add(inventory.Item{ID: "medkit", Name: "Medkit", Qty: 1})
	case "grenade", "plasma_grenade", "emp_grenade", "bomb":
		g.playerInventory.Add(inventory.Item{ID: "grenade", Name: "Grenade", Qty: 1})
	case "proximity_mine":
		g.playerInventory.Add(inventory.Item{ID: "proximity_mine", Name: "Proximity Mine", Qty: 1})
	}
}

// applyArmorItem increases player armor.
func (g *Game) applyArmorItem() {
	g.hud.Armor += 50
	if g.hud.Armor > g.hud.MaxArmor {
		g.hud.Armor = g.hud.MaxArmor
	}
}

// applyWeaponUpgrade applies an upgrade to the current weapon.
func (g *Game) applyWeaponUpgrade(itemID string) {
	currentWeapon := g.arsenal.GetCurrentWeapon()
	weaponID := currentWeapon.Name

	upgradeMap := map[string]upgrade.UpgradeType{
		"upgrade_damage":   upgrade.UpgradeDamage,
		"upgrade_firerate": upgrade.UpgradeFireRate,
		"upgrade_clipsize": upgrade.UpgradeClipSize,
		"upgrade_accuracy": upgrade.UpgradeAccuracy,
		"upgrade_range":    upgrade.UpgradeRange,
	}

	upgradeMessages := map[string]string{
		"upgrade_damage":   "Damage upgrade applied!",
		"upgrade_firerate": "Fire rate upgrade applied!",
		"upgrade_clipsize": "Clip size upgrade applied!",
		"upgrade_accuracy": "Accuracy upgrade applied!",
		"upgrade_range":    "Range upgrade applied!",
	}

	if upgradeType, ok := upgradeMap[itemID]; ok {
		if g.upgradeManager.ApplyUpgrade(weaponID, upgradeType, 2) {
			if msg, exists := upgradeMessages[itemID]; exists {
				g.hud.ShowMessage(msg)
			}
		}
	}
}

// updateHUDAmmo refreshes the HUD ammo display.
func (g *Game) updateHUDAmmo() {
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
	if g.input.IsJustPressed(input.ActionPause) || g.input.IsJustPressed(input.ActionSkills) {
		g.state = StatePlaying
		g.menuManager.Hide()
		return nil
	}

	handleSkillTreeNavigation(g)
	handleSkillNodeNavigation(g)

	if g.input.IsJustPressed(input.ActionFire) || g.input.IsJustPressed(input.ActionInteract) {
		g.handleSkillAllocate()
	}

	return nil
}

// handleSkillTreeNavigation switches between skill tree tabs using strafe keys.
func handleSkillTreeNavigation(g *Game) {
	if g.input.IsJustPressed(input.ActionStrafeLeft) {
		if g.skillsTreeIdx > 0 {
			g.skillsTreeIdx--
			g.skillsNodeIdx = 0
		}
	}
	if g.input.IsJustPressed(input.ActionStrafeRight) {
		if g.skillsTreeIdx < 2 {
			g.skillsTreeIdx++
			g.skillsNodeIdx = 0
		}
	}
}

// handleSkillNodeNavigation moves cursor between skill nodes within the active tree.
func handleSkillNodeNavigation(g *Game) {
	if g.input.IsJustPressed(input.ActionMoveForward) {
		if g.skillsNodeIdx > 0 {
			g.skillsNodeIdx--
		}
	}
	if g.input.IsJustPressed(input.ActionMoveBackward) {
		g.skillsNodeIdx++
	}
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

	trees := buildSkillTrees(g)
	totalPoints := getTotalAvailablePoints(g.skillManager)

	return &ui.SkillsState{
		Trees:       trees,
		ActiveTree:  g.skillsTreeIdx,
		Selected:    g.skillsNodeIdx,
		TotalPoints: totalPoints,
	}
}

// buildSkillTrees constructs UI representations of all skill trees.
func buildSkillTrees(g *Game) []ui.SkillTreeState {
	treeIDs := []string{"combat", "survival", "tech"}
	treeNames := []string{"Combat", "Survival", "Tech"}
	trees := make([]ui.SkillTreeState, 0, len(treeIDs))

	for i, treeID := range treeIDs {
		tree, err := g.skillManager.GetTree(treeID)
		if err != nil {
			continue
		}

		treeState := buildSingleSkillTree(g, tree, treeNames[i], treeID)
		trees = append(trees, treeState)
	}

	return trees
}

// buildSingleSkillTree creates a UI skill tree state from a skill tree.
func buildSingleSkillTree(g *Game, tree *skills.Tree, treeName, treeID string) ui.SkillTreeState {
	nodeList := g.getTreeNodeList(tree)
	uiNodes := buildSkillNodes(tree, nodeList)

	return ui.SkillTreeState{
		TreeName: treeName,
		TreeID:   treeID,
		Nodes:    uiNodes,
		Points:   tree.GetPoints(),
		Selected: g.skillsNodeIdx,
	}
}

// buildSkillNodes converts skill nodes to UI representations with availability status.
func buildSkillNodes(tree *skills.Tree, nodeList []skills.Node) []ui.SkillNode {
	uiNodes := make([]ui.SkillNode, len(nodeList))
	for j, node := range nodeList {
		available := checkNodeAvailability(tree, &node)
		uiNodes[j] = ui.SkillNode{
			ID:          node.ID,
			Name:        node.Name,
			Description: node.Description,
			Cost:        node.Cost,
			Allocated:   tree.IsAllocated(node.ID),
			Available:   available,
		}
	}
	return uiNodes
}

// checkNodeAvailability determines if a skill node can be allocated.
func checkNodeAvailability(tree *skills.Tree, node *skills.Node) bool {
	if tree.IsAllocated(node.ID) {
		return false
	}

	for _, reqID := range node.Requires {
		if !tree.IsAllocated(reqID) {
			return false
		}
	}

	return tree.GetPoints() >= node.Cost
}

// getTotalAvailablePoints retrieves the total skill points available across all trees.
func getTotalAvailablePoints(skillManager *skills.Manager) int {
	if t, err := skillManager.GetTree("combat"); err == nil {
		return t.GetPoints()
	}
	return 0
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

	// Initialize E2E encrypted chat
	g.initializeEncryptedChat()
}

// updateMultiplayer handles multiplayer lobby input.
func (g *Game) updateMultiplayer() error {
	g.handleChatInput()

	if g.chatInputActive {
		return nil
	}

	if handled := g.handleMultiplayerNavigation(); handled {
		return nil
	}

	g.handleMultiplayerModeToggle()
	g.handleMultiplayerServerNavigation()
	g.handleMultiplayerRefresh()
	g.handleMultiplayerAction()

	return nil
}

// handleMultiplayerNavigation handles back to playing action and returns true if handled.
func (g *Game) handleMultiplayerNavigation() bool {
	if g.input.IsJustPressed(input.ActionPause) || g.input.IsJustPressed(input.ActionMultiplayer) {
		g.state = StatePlaying
		g.menuManager.Hide()
		return true
	}
	return false
}

// handleMultiplayerModeToggle toggles between local and federation multiplayer modes.
func (g *Game) handleMultiplayerModeToggle() {
	if g.input.IsJustPressed(input.ActionCodex) {
		g.useFederation = !g.useFederation
		if g.useFederation {
			g.refreshServerBrowser()
		}
	}
}

// handleMultiplayerServerNavigation handles navigation through servers or modes.
func (g *Game) handleMultiplayerServerNavigation() {
	if g.input.IsJustPressed(input.ActionMoveForward) {
		g.handleNavigationUp()
	}

	if g.input.IsJustPressed(input.ActionMoveBackward) {
		g.handleNavigationDown()
	}
}

// handleNavigationUp moves selection up in the multiplayer menu.
func (g *Game) handleNavigationUp() {
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

// handleNavigationDown moves selection down in the multiplayer menu.
func (g *Game) handleNavigationDown() {
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

// handleMultiplayerRefresh refreshes the federation server browser.
func (g *Game) handleMultiplayerRefresh() {
	if g.useFederation && g.input.IsJustPressed(input.ActionCraft) {
		g.refreshServerBrowser()
	}
}

// handleMultiplayerAction handles mode selection or server join action.
func (g *Game) handleMultiplayerAction() {
	if g.input.IsJustPressed(input.ActionFire) || g.input.IsJustPressed(input.ActionInteract) {
		if g.useFederation {
			g.handleFederationJoin()
		} else {
			g.handleMultiplayerSelect()
		}
	}
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
	// If a federation hub URL is configured, query it remotely
	if config.C.FederationHubURL != "" {
		genre := g.genreID
		query := &federation.ServerQuery{
			Genre: &genre,
		}

		servers, err := federation.DiscoverServers(config.C.FederationHubURL, query, 5*time.Second)
		if err != nil {
			logrus.WithError(err).Warn("failed to discover servers from federation hub")
			g.mpStatusMsg = "Failed to connect to federation hub. Press R to retry."
			g.serverBrowser = nil
			return
		}

		g.serverBrowser = make([]*federation.ServerAnnouncement, len(servers))
		for i := range servers {
			g.serverBrowser[i] = &servers[i]
		}
		g.browserIdx = 0

		if len(servers) == 0 {
			g.mpStatusMsg = "No servers found. Press R to refresh."
		} else {
			g.mpStatusMsg = "Found " + string(rune(len(servers)+'0')) + " servers. Press L for local mode."
		}
		return
	}

	// Otherwise use local federation hub (for testing/local servers)
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

	// Draw encrypted chat interface
	g.drawEncryptedChat(screen)
}

// drawEncryptedChat renders the encrypted chat UI.
func (g *Game) drawEncryptedChat(screen *ebiten.Image) {
	if g.chatManager == nil {
		return
	}

	// Draw chat messages in bottom-left corner
	startY := float32(config.C.InternalHeight - 150)
	startX := float32(10)

	// Show last 5 messages
	numMsgs := len(g.chatMessages)
	startIdx := 0
	if numMsgs > 5 {
		startIdx = numMsgs - 5
	}

	for i := startIdx; i < numMsgs; i++ {
		msg := g.chatMessages[i]
		y := startY + float32((i-startIdx)*15)
		// Draw semi-transparent background
		vector.DrawFilledRect(screen, startX-2, y-2, 300, 14, color.RGBA{0, 0, 0, 128}, false)
		// Would draw text here with proper text rendering
		// For now, the infrastructure is in place
		_ = msg
		_ = y
	}

	// Draw chat input box if active
	if g.chatInputActive {
		inputY := float32(config.C.InternalHeight - 30)
		vector.DrawFilledRect(screen, startX-2, inputY-2, 400, 24, color.RGBA{0, 0, 0, 180}, false)
		vector.StrokeRect(screen, startX-2, inputY-2, 400, 24, 2, color.RGBA{0, 255, 0, 255}, false)
		// Would draw chat input text here
		// Prompt: "> " + g.chatInput + "_"
	}
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

// convertInventoryToSaveItems converts inventory.Item slice to save.Item slice
func convertInventoryToSaveItems(inv *inventory.Inventory) []save.Item {
	if inv == nil {
		return []save.Item{}
	}

	// Thread-safe access to inventory items
	inv.Items = append([]inventory.Item{}, inv.Items...)

	saveItems := make([]save.Item, len(inv.Items))
	for i, item := range inv.Items {
		saveItems[i] = save.Item{
			ID:   item.ID,
			Name: item.Name,
			Qty:  item.Qty,
		}
	}
	return saveItems
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
		Inventory: save.Inventory{Items: convertInventoryToSaveItems(g.playerInventory)},
		Progression: save.ProgressionState{
			Level: g.progression.GetLevel(),
			XP:    g.progression.GetXP(),
		},
		Keycards: g.keycards,
		AmmoPool: ammoPoolState,
	}
	save.Save(slot, state)
}

// updateLoading handles loading screen updates.
func (g *Game) updateLoading() error {
	// Update loading screen animation
	g.loadingScreen.Update()
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

	// Render dynamic shadows for entities/props
	if g.shadowSystem != nil && g.lightMap != nil {
		g.renderShadows(screen)
	}

	// Render particles on top of 3D world (simple sprite overlay for now)
	// TODO: Add particle rendering to renderer or as separate overlay
	if g.particleSystem != nil {
		g.renderParticles(screen)
	}

	// Render environmental hazards
	if g.hazardSystem != nil {
		g.renderHazards(screen)
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
	// Use optimized visible particles query with frustum culling
	const maxDistSq = 400.0
	particles := g.particleSystem.GetVisibleParticles(
		g.camera.X, g.camera.Y,
		g.camera.DirX, g.camera.DirY,
		maxDistSq,
	)

	for _, p := range particles {
		dx := p.X - g.camera.X
		dy := p.Y - g.camera.Y

		// Project to screen space
		screenX := config.C.InternalWidth/2 + int(dx*10)
		screenY := config.C.InternalHeight/2 + int(dy*10)

		// Screen bounds check
		if screenX >= 0 && screenX < config.C.InternalWidth && screenY >= 0 && screenY < config.C.InternalHeight {
			particleColor := color.RGBA{R: p.R, G: p.G, B: p.B, A: p.A}
			vector.DrawFilledRect(screen, float32(screenX), float32(screenY), 2, 2, particleColor, false)
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

// renderShadows draws dynamic shadows for props and entities based on active lights.
func (g *Game) renderShadows(screen *ebiten.Image) {
	// Collect shadow casters from props
	var casters []lighting.ShadowCaster

	if g.propsManager != nil {
		allProps := g.propsManager.GetProps()
		for _, prop := range allProps {
			// Calculate distance to camera for culling
			dx := prop.X - g.camera.X
			dy := prop.Y - g.camera.Y
			distSq := dx*dx + dy*dy

			// Only cast shadows for nearby props (within visible range)
			if distSq > 400 {
				continue
			}

			// Determine shadow parameters based on prop type
			radius := 0.5
			height := 1.0
			opacity := 0.7

			switch prop.SpriteType {
			case props.PropPillar:
				radius = 0.6
				height = 2.0
				opacity = 0.8
			case props.PropBarrel, props.PropCrate, props.PropContainer:
				radius = 0.5
				height = 1.2
				opacity = 0.7
			case props.PropTable:
				radius = 0.7
				height = 0.8
				opacity = 0.6
			case props.PropTorch:
				// Torches are light sources - smaller shadow
				radius = 0.2
				height = 0.5
				opacity = 0.4
			case props.PropTerminal:
				radius = 0.4
				height = 1.5
				opacity = 0.7
			case props.PropBones, props.PropDebris:
				// Small debris - subtle shadows
				radius = 0.3
				height = 0.3
				opacity = 0.5
			case props.PropPlant:
				radius = 0.4
				height = 0.8
				opacity = 0.6
			}

			casters = append(casters, lighting.ShadowCaster{
				X:          prop.X,
				Y:          prop.Y,
				Radius:     radius,
				Height:     height,
				Opacity:    opacity,
				CastShadow: true,
			})
		}
	}

	// Add player shadow
	casters = append(casters, lighting.ShadowCaster{
		X:          g.camera.X,
		Y:          g.camera.Y,
		Radius:     0.4,
		Height:     1.8,
		Opacity:    0.75,
		CastShadow: true,
	})

	// Add lore item shadows
	for _, item := range g.loreItems {
		dx := item.PosX - g.camera.X
		dy := item.PosY - g.camera.Y
		distSq := dx*dx + dy*dy
		if distSq <= 400 {
			casters = append(casters, lighting.ShadowCaster{
				X:          item.PosX,
				Y:          item.PosY,
				Radius:     0.3,
				Height:     0.5,
				Opacity:    0.6,
				CastShadow: true,
			})
		}
	}

	// Collect lights from lightMap
	// For now, we'll use a simple approach: add point lights based on torches and ambient sources
	var lights []lighting.Light
	var coneLights []lighting.ConeLight

	// Add torch props as light sources
	if g.propsManager != nil {
		allProps := g.propsManager.GetProps()
		for _, prop := range allProps {
			if prop.SpriteType == props.PropTorch {
				lights = append(lights, lighting.Light{
					X:         prop.X,
					Y:         prop.Y,
					Radius:    8.0,
					Intensity: 0.9,
					R:         1.0,
					G:         0.8,
					B:         0.4,
				})
			}
		}
	}

	// Add player flashlight as cone light (always present for gameplay visibility)
	coneLights = append(coneLights, lighting.ConeLight{
		X:         g.camera.X,
		Y:         g.camera.Y,
		DirX:      g.camera.DirX,
		DirY:      g.camera.DirY,
		Radius:    12.0,
		Angle:     0.5,
		Intensity: 0.8,
		R:         1.0,
		G:         1.0,
		B:         1.0,
		IsActive:  true,
	})

	// Render shadows
	g.shadowSystem.RenderShadows(screen, casters, lights, coneLights, g.camera.X, g.camera.Y)
}

// renderHazards draws environmental hazards as floor sprites in world space.
func (g *Game) renderHazards(screen *ebiten.Image) {
	hazards := g.hazardSystem.GetHazards()

	// Calculate camera plane for sprite projection
	planeX := 0.0
	planeY := 0.66
	if g.camera.DirX != 0 || g.camera.DirY != 0 {
		planeX = -g.camera.DirY * 0.66
		planeY = g.camera.DirX * 0.66
	}

	for _, h := range hazards {
		dx := h.X - g.camera.X
		dy := h.Y - g.camera.Y

		// Simple distance culling
		distSq := dx*dx + dy*dy
		if distSq > 400 {
			continue
		}

		// Transform to camera space
		invDet := 1.0 / (planeX*g.camera.DirY - g.camera.DirX*planeY)
		transformX := invDet * (g.camera.DirY*dx - g.camera.DirX*dy)
		transformY := invDet * (-planeY*dx + planeX*dy)

		// Skip hazards behind camera
		if transformY <= 0.1 {
			continue
		}

		// Calculate screen position
		spriteScreenX := int((float64(config.C.InternalWidth) / 2.0) * (1.0 + transformX/transformY))

		// Calculate size based on distance and hazard dimensions
		spriteHeight := int(float64(config.C.InternalHeight) / transformY * h.Height)
		spriteWidth := int(float64(config.C.InternalWidth) / transformY * h.Width)

		// Draw bounds
		drawStartX := spriteScreenX - spriteWidth/2
		drawEndX := spriteScreenX + spriteWidth/2
		drawStartY := config.C.InternalHeight/2 + spriteHeight/4 // Offset down for floor effect
		drawEndY := config.C.InternalHeight/2 + spriteHeight/4 + spriteHeight/2

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
		if drawStartY < 0 {
			drawStartY = 0
		}
		if drawEndY >= config.C.InternalHeight {
			drawEndY = config.C.InternalHeight - 1
		}

		// Determine visual representation based on state
		var hazardColor color.RGBA
		alpha := uint8(180)

		// Extract RGB from color
		r := uint8((h.Color >> 16) & 0xFF)
		g := uint8((h.Color >> 8) & 0xFF)
		b := uint8(h.Color & 0xFF)

		switch h.State {
		case hazard.StateInactive:
			alpha = 60
			hazardColor = color.RGBA{r / 2, g / 2, b / 2, alpha}
		case hazard.StateCharging:
			// Pulsate warning
			alpha = uint8(100 + int(h.Timer*400)%100)
			hazardColor = color.RGBA{r, g, b, alpha}
		case hazard.StateActive:
			// Full brightness when active
			alpha = 220
			hazardColor = color.RGBA{r, g, b, alpha}
		case hazard.StateCooldown:
			alpha = 100
			hazardColor = color.RGBA{r / 2, g / 2, b / 2, alpha}
		}

		// Draw hazard sprite
		vector.DrawFilledRect(screen,
			float32(drawStartX), float32(drawStartY),
			float32(drawEndX-drawStartX), float32(drawEndY-drawStartY),
			hazardColor, false)

		// Draw warning indicator when charging
		if h.State == hazard.StateCharging {
			warningColor := color.RGBA{255, 255, 0, uint8(150 + int(h.Timer*600)%100)}
			vector.StrokeRect(screen,
				float32(drawStartX), float32(drawStartY),
				float32(drawEndX-drawStartX), float32(drawEndY-drawStartY),
				2, warningColor, false)
		}
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
			switch {
			case tile == bsp.TileWall || (tile >= 10 && tile <= 14): // Generic or genre-specific wall
				tileColor = color.RGBA{150, 150, 150, 255}
			case tile == bsp.TileFloor || tile == bsp.TileEmpty || (tile >= 20 && tile <= 29): // Floor variants
				tileColor = color.RGBA{50, 50, 50, 255}
			case tile == bsp.TileDoor:
				tileColor = color.RGBA{100, 100, 200, 255}
			case tile == bsp.TileSecret:
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
	planeX, planeY := calculateCameraPlane(g.camera)

	for _, loreItem := range g.loreItems {
		if loreItem.Activated {
			continue
		}

		if !shouldRenderLoreItem(loreItem, g.camera) {
			continue
		}

		drawLoreItemSprite(screen, loreItem, g.camera, planeX, planeY, g.animationTicker)
	}
}

// calculateCameraPlane computes the camera's view plane vectors.
func calculateCameraPlane(camera *camera.Camera) (float64, float64) {
	fov := camera.FOV
	planeX := -camera.DirY * fov / 66.0
	planeY := camera.DirX * fov / 66.0
	return planeX, planeY
}

// shouldRenderLoreItem checks if a lore item is close enough to render.
func shouldRenderLoreItem(loreItem *lore.LoreItem, camera *camera.Camera) bool {
	dx := loreItem.PosX - camera.X
	dy := loreItem.PosY - camera.Y
	dist := dx*dx + dy*dy
	return dist <= 400
}

// drawLoreItemSprite renders a single lore item sprite on screen.
func drawLoreItemSprite(screen *ebiten.Image, loreItem *lore.LoreItem, camera *camera.Camera, planeX, planeY float64, animationTicker int) {
	transformX, transformY := calculateSpriteTransform(loreItem, camera, planeX, planeY)
	if transformY <= 0.1 {
		return
	}

	spriteScreenX, spriteHeight, spriteWidth := calculateSpriteDimensions(transformX, transformY)
	drawStartX, drawEndX, drawStartY, drawEndY := calculateSpriteDrawBounds(spriteScreenX, spriteHeight, spriteWidth)

	if !isWithinScreenBounds(drawStartX, drawEndX) {
		return
	}

	loreColor := calculateLoreItemColor(loreItem, animationTicker)
	vector.DrawFilledRect(screen, float32(drawStartX), float32(drawStartY),
		float32(drawEndX-drawStartX), float32(drawEndY-drawStartY), loreColor, false)
}

// calculateSpriteTransform computes the sprite's transform coordinates.
func calculateSpriteTransform(loreItem *lore.LoreItem, camera *camera.Camera, planeX, planeY float64) (float64, float64) {
	dx := loreItem.PosX - camera.X
	dy := loreItem.PosY - camera.Y
	invDet := 1.0 / (planeX*camera.DirY - camera.DirX*planeY)
	transformX := invDet * (camera.DirY*dx - camera.DirX*dy)
	transformY := invDet * (-planeY*dx + planeX*dy)
	return transformX, transformY
}

// calculateSpriteDimensions computes sprite screen position and size.
func calculateSpriteDimensions(transformX, transformY float64) (int, int, int) {
	spriteScreenX := int((float64(config.C.InternalWidth) / 2.0) * (1.0 + transformX/transformY))
	spriteHeight := int(float64(config.C.InternalHeight) / transformY / 2)
	spriteWidth := spriteHeight
	return spriteScreenX, spriteHeight, spriteWidth
}

// calculateSpriteDrawBounds computes the screen-space draw boundaries for the sprite.
func calculateSpriteDrawBounds(spriteScreenX, spriteHeight, spriteWidth int) (int, int, int, int) {
	drawStartX := spriteScreenX - spriteWidth/2
	drawEndX := spriteScreenX + spriteWidth/2
	drawStartY := config.C.InternalHeight/2 - spriteHeight/2
	drawEndY := config.C.InternalHeight/2 + spriteHeight/2

	if drawStartX < 0 {
		drawStartX = 0
	}
	if drawEndX >= config.C.InternalWidth {
		drawEndX = config.C.InternalWidth - 1
	}

	return drawStartX, drawEndX, drawStartY, drawEndY
}

// isWithinScreenBounds checks if the sprite is visible on screen.
func isWithinScreenBounds(drawStartX, drawEndX int) bool {
	return drawEndX >= 0 && drawStartX < config.C.InternalWidth
}

// calculateLoreItemColor determines the color for a lore item with pulsing effect.
func calculateLoreItemColor(loreItem *lore.LoreItem, animationTicker int) color.RGBA {
	pulse := float32(0.7 + 0.3*float64(animationTicker%60)/60.0)
	switch loreItem.Type {
	case lore.LoreItemNote:
		return color.RGBA{uint8(255 * pulse), uint8(255 * pulse), 200, 255}
	case lore.LoreItemAudioLog:
		return color.RGBA{100, uint8(200 * pulse), uint8(255 * pulse), 255}
	case lore.LoreItemGraffiti:
		return color.RGBA{uint8(255 * pulse), 100, 100, 255}
	case lore.LoreItemBodyArrangement:
		return color.RGBA{150, 150, uint8(150 * pulse), 255}
	default:
		return color.RGBA{150, 150, 150, 255}
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

	// Draw title
	titleText := "LOCKPICKING"
	titleBounds := text.BoundString(basicfont.Face7x13, titleText)
	titleX := int(centerX) - titleBounds.Dx()/2
	titleY := int(centerY) - 90
	text.Draw(screen, titleText, basicfont.Face7x13, titleX, titleY, color.RGBA{255, 255, 255, 255})

	// Draw instructions
	instrText := "Press SPACE when pick is in GREEN zone"
	instrBounds := text.BoundString(basicfont.Face7x13, instrText)
	instrX := int(centerX) - instrBounds.Dx()/2
	instrY := int(centerY) - 75
	text.Draw(screen, instrText, basicfont.Face7x13, instrX, instrY, color.RGBA{200, 200, 200, 255})

	// Draw lock cylinder
	cylinderWidth := float32(200)
	cylinderHeight := float32(30)
	cylinderX := centerX - cylinderWidth/2
	cylinderY := centerY - cylinderHeight/2

	vector.DrawFilledRect(screen, cylinderX, cylinderY, cylinderWidth, cylinderHeight, color.RGBA{100, 100, 100, 255}, false)
	vector.StrokeRect(screen, cylinderX, cylinderY, cylinderWidth, cylinderHeight, 2, color.RGBA{200, 200, 200, 255}, false)

	// Draw target zone with label
	targetX := cylinderX + cylinderWidth*float32(lpGame.Target)
	targetWidth := cylinderWidth * float32(lpGame.Tolerance*2)
	vector.DrawFilledRect(screen, targetX-targetWidth/2, cylinderY, targetWidth, cylinderHeight, color.RGBA{0, 200, 0, 100}, false)

	// Draw lockpick position indicator
	pickX := cylinderX + cylinderWidth*float32(lpGame.Position)
	vector.DrawFilledRect(screen, pickX-2, cylinderY-10, 4, cylinderHeight+20, color.RGBA{255, 255, 0, 255}, false)

	// Draw pins status
	pinsText := fmt.Sprintf("Pins: %d/%d", lpGame.UnlockedPins, lpGame.Pins)
	pinsBounds := text.BoundString(basicfont.Face7x13, pinsText)
	pinsX := int(centerX) - pinsBounds.Dx()/2
	pinsY := int(centerY) - 50
	text.Draw(screen, pinsText, basicfont.Face7x13, pinsX, pinsY, color.RGBA{255, 255, 255, 255})

	// Draw unlocked pins visualization
	for i := 0; i < lpGame.Pins; i++ {
		pinX := centerX - float32(lpGame.Pins*12)/2 + float32(i*25)
		pinY := centerY - 30
		pinColor := color.RGBA{80, 80, 80, 255}
		if i < lpGame.UnlockedPins {
			pinColor = color.RGBA{0, 255, 0, 255}
		}
		vector.DrawFilledCircle(screen, pinX, pinY, 8, pinColor, false)
		vector.StrokeCircle(screen, pinX, pinY, 8, 2, color.RGBA{200, 200, 200, 255}, false)

		// Draw pin number
		pinNumText := fmt.Sprintf("%d", i+1)
		numBounds := text.BoundString(basicfont.Face7x13, pinNumText)
		numX := int(pinX) - numBounds.Dx()/2
		numY := int(pinY) + 4
		text.Draw(screen, pinNumText, basicfont.Face7x13, numX, numY, color.RGBA{255, 255, 255, 255})
	}
}

// drawHackGame renders hacking interface.
func (g *Game) drawHackGame(screen *ebiten.Image, centerX, centerY float32) {
	hackGame, ok := g.activeMinigame.(*minigame.HackGame)
	if !ok {
		return
	}

	drawHackGameHeader(screen, centerX, centerY, hackGame)
	drawHackGameNodeGrid(screen, centerX, centerY, hackGame)
	drawHackGameSequenceIndicators(screen, centerX, centerY, hackGame)
}

// drawHackGameHeader renders the title, instructions, and sequence status for the hack game.
func drawHackGameHeader(screen *ebiten.Image, centerX, centerY float32, hackGame *minigame.HackGame) {
	titleText := "NETWORK BREACH"
	titleBounds := text.BoundString(basicfont.Face7x13, titleText)
	titleX := int(centerX) - titleBounds.Dx()/2
	titleY := int(centerY) - 90
	text.Draw(screen, titleText, basicfont.Face7x13, titleX, titleY, color.RGBA{0, 255, 255, 255})

	instrText := "Use number keys (1-6) to match sequence"
	instrBounds := text.BoundString(basicfont.Face7x13, instrText)
	instrX := int(centerX) - instrBounds.Dx()/2
	instrY := int(centerY) - 75
	text.Draw(screen, instrText, basicfont.Face7x13, instrX, instrY, color.RGBA{200, 200, 200, 255})

	seqText := fmt.Sprintf("Sequence: %d/%d", len(hackGame.PlayerInput), len(hackGame.Sequence))
	seqBounds := text.BoundString(basicfont.Face7x13, seqText)
	seqX := int(centerX) - seqBounds.Dx()/2
	seqY := int(centerY) - 110
	text.Draw(screen, seqText, basicfont.Face7x13, seqX, seqY, color.RGBA{255, 255, 255, 255})
}

// drawHackGameNodeGrid renders the circular node grid for the hack game.
func drawHackGameNodeGrid(screen *ebiten.Image, centerX, centerY float32, hackGame *minigame.HackGame) {
	nodeRadius := float32(80)
	for i := 0; i < 6; i++ {
		angle := float32(i) * 3.14159 * 2.0 / 6.0
		nodeX := centerX + nodeRadius*float32(cosf(angle))
		nodeY := centerY + nodeRadius*float32(sinf(angle))

		nodeColor, nextNode := getHackNodeColor(hackGame, i)
		vector.DrawFilledCircle(screen, nodeX, nodeY, 15, nodeColor, false)
		vector.StrokeCircle(screen, nodeX, nodeY, 15, 2, color.RGBA{255, 255, 255, 255}, false)

		drawHackNodeNumber(screen, nodeX, nodeY, i, nextNode)
	}
}

// getHackNodeColor determines the color and next node status for a hack node.
func getHackNodeColor(hackGame *minigame.HackGame, nodeIndex int) (color.RGBA, bool) {
	nodeColor := color.RGBA{100, 100, 200, 255}

	for j, node := range hackGame.Sequence {
		if j < len(hackGame.PlayerInput) && node == nodeIndex {
			nodeColor = color.RGBA{0, 255, 0, 255}
		}
	}

	nextNode := false
	if len(hackGame.PlayerInput) < len(hackGame.Sequence) {
		if hackGame.Sequence[len(hackGame.PlayerInput)] == nodeIndex {
			nodeColor = color.RGBA{255, 255, 0, 255}
			nextNode = true
		}
	}

	return nodeColor, nextNode
}

// drawHackNodeNumber renders the number label on a hack node.
func drawHackNodeNumber(screen *ebiten.Image, nodeX, nodeY float32, nodeIndex int, nextNode bool) {
	nodeNumText := fmt.Sprintf("%d", nodeIndex+1)
	numBounds := text.BoundString(basicfont.Face7x13, nodeNumText)
	numX := int(nodeX) - numBounds.Dx()/2
	numY := int(nodeY) + 4
	numColor := color.RGBA{255, 255, 255, 255}
	if nextNode {
		numColor = color.RGBA{0, 0, 0, 255}
	}
	text.Draw(screen, nodeNumText, basicfont.Face7x13, numX, numY, numColor)
}

// drawHackGameSequenceIndicators renders the sequence progress indicators at the top.
func drawHackGameSequenceIndicators(screen *ebiten.Image, centerX, centerY float32, hackGame *minigame.HackGame) {
	for i := range hackGame.Sequence {
		boxX := centerX - float32(len(hackGame.Sequence)*10) + float32(i*20)
		boxY := centerY - 50
		boxColor := color.RGBA{50, 50, 50, 255}
		if i < len(hackGame.PlayerInput) {
			boxColor = color.RGBA{0, 200, 0, 255}
		}
		vector.DrawFilledRect(screen, boxX, boxY, 15, 15, boxColor, false)
		vector.StrokeRect(screen, boxX, boxY, 15, 15, 1, color.RGBA{200, 200, 200, 255}, false)

		targetNodeText := fmt.Sprintf("%d", hackGame.Sequence[i]+1)
		targetBounds := text.BoundString(basicfont.Face7x13, targetNodeText)
		targetX := int(boxX) + 8 - targetBounds.Dx()/2
		targetY := int(boxY) + 11
		targetColor := color.RGBA{150, 150, 150, 255}
		if i < len(hackGame.PlayerInput) {
			targetColor = color.RGBA{255, 255, 255, 255}
		}
		text.Draw(screen, targetNodeText, basicfont.Face7x13, targetX, targetY, targetColor)
	}
}

// drawCircuitGame renders circuit trace interface.
func (g *Game) drawCircuitGame(screen *ebiten.Image, centerX, centerY float32) {
	circuitGame, ok := g.activeMinigame.(*minigame.CircuitTraceGame)
	if !ok {
		return
	}

	g.drawCircuitGameHeader(screen, centerX, centerY, circuitGame)

	gridSize := len(circuitGame.Grid)
	cellSize := float32(30)
	startX := centerX - float32(gridSize)*cellSize/2
	startY := centerY - float32(gridSize)*cellSize/2

	g.drawCircuitGameGrid(screen, circuitGame, startX, startY, cellSize, gridSize)
	g.drawCircuitGameLegend(screen, startX, startY, cellSize, gridSize)
}

// drawCircuitGameHeader renders title, instructions, and move counter.
func (g *Game) drawCircuitGameHeader(screen *ebiten.Image, centerX, centerY float32, circuitGame *minigame.CircuitTraceGame) {
	titleText := "CIRCUIT TRACE"
	titleBounds := text.BoundString(basicfont.Face7x13, titleText)
	titleX := int(centerX) - titleBounds.Dx()/2
	titleY := int(centerY) - 110
	text.Draw(screen, titleText, basicfont.Face7x13, titleX, titleY, color.RGBA{0, 255, 200, 255})

	instrText := "Arrow keys to navigate. Reach BLUE target!"
	instrBounds := text.BoundString(basicfont.Face7x13, instrText)
	instrX := int(centerX) - instrBounds.Dx()/2
	instrY := int(centerY) - 95
	text.Draw(screen, instrText, basicfont.Face7x13, instrX, instrY, color.RGBA{200, 200, 200, 255})

	movesText := fmt.Sprintf("Moves: %d/%d", circuitGame.Moves, circuitGame.MaxMoves)
	movesBounds := text.BoundString(basicfont.Face7x13, movesText)
	movesX := int(centerX) - movesBounds.Dx()/2
	movesY := int(centerY) - 80
	text.Draw(screen, movesText, basicfont.Face7x13, movesX, movesY, color.RGBA{255, 255, 255, 255})
}

// drawCircuitGameGrid renders the circuit trace grid with cells.
func (g *Game) drawCircuitGameGrid(screen *ebiten.Image, circuitGame *minigame.CircuitTraceGame, startX, startY, cellSize float32, gridSize int) {
	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			cellX := startX + float32(x)*cellSize
			cellY := startY + float32(y)*cellSize

			cellColor, cellLabel, labelColor := g.getCellAppearance(circuitGame, x, y)

			vector.DrawFilledRect(screen, cellX, cellY, cellSize-2, cellSize-2, cellColor, false)
			vector.StrokeRect(screen, cellX, cellY, cellSize-2, cellSize-2, 1, color.RGBA{100, 100, 100, 255}, false)

			if cellLabel != "" {
				labelBounds := text.BoundString(basicfont.Face7x13, cellLabel)
				labelX := int(cellX) + int(cellSize-2)/2 - labelBounds.Dx()/2
				labelY := int(cellY) + int(cellSize-2)/2 + 4
				text.Draw(screen, cellLabel, basicfont.Face7x13, labelX, labelY, labelColor)
			}
		}
	}
}

// getCellAppearance determines color and label for a circuit game cell.
func (g *Game) getCellAppearance(circuitGame *minigame.CircuitTraceGame, x, y int) (color.RGBA, string, color.RGBA) {
	cellColor := color.RGBA{50, 50, 50, 255}
	cellLabel := ""
	labelColor := color.RGBA{150, 150, 150, 255}

	if circuitGame.Grid[y][x] == 2 {
		cellColor = color.RGBA{200, 0, 0, 255}
		cellLabel = "X"
		labelColor = color.RGBA{255, 255, 255, 255}
	}
	if x == circuitGame.CurrentX && y == circuitGame.CurrentY {
		cellColor = color.RGBA{0, 255, 0, 255}
		cellLabel = "P"
		labelColor = color.RGBA{0, 0, 0, 255}
	}
	if x == circuitGame.TargetX && y == circuitGame.TargetY {
		cellColor = color.RGBA{0, 200, 255, 255}
		cellLabel = "T"
		labelColor = color.RGBA{0, 0, 0, 255}
	}

	return cellColor, cellLabel, labelColor
}

// drawCircuitGameLegend renders the legend showing cell type meanings.
func (g *Game) drawCircuitGameLegend(screen *ebiten.Image, startX, startY, cellSize float32, gridSize int) {
	legendY := int(startY + float32(gridSize)*cellSize + 15)
	legendX := int(startX)

	text.Draw(screen, "P=You", basicfont.Face7x13, legendX, legendY, color.RGBA{0, 255, 0, 255})
	text.Draw(screen, "T=Target", basicfont.Face7x13, legendX+60, legendY, color.RGBA{0, 200, 255, 255})
	text.Draw(screen, "X=Blocked", basicfont.Face7x13, legendX+140, legendY, color.RGBA{200, 0, 0, 255})
}

// drawCodeGame renders bypass code interface.
func (g *Game) drawCodeGame(screen *ebiten.Image, centerX, centerY float32) {
	codeGame, ok := g.activeMinigame.(*minigame.BypassCodeGame)
	if !ok {
		return
	}

	// Draw title
	titleText := "ACCESS CODE BYPASS"
	titleBounds := text.BoundString(basicfont.Face7x13, titleText)
	titleX := int(centerX) - titleBounds.Dx()/2
	titleY := int(centerY) - 80
	text.Draw(screen, titleText, basicfont.Face7x13, titleX, titleY, color.RGBA{255, 255, 0, 255})

	// Draw instructions
	instrText := "Enter code using number keys (0-9)"
	instrBounds := text.BoundString(basicfont.Face7x13, instrText)
	instrX := int(centerX) - instrBounds.Dx()/2
	instrY := int(centerY) - 65
	text.Draw(screen, instrText, basicfont.Face7x13, instrX, instrY, color.RGBA{200, 200, 200, 255})

	// Draw code length hint
	hintText := fmt.Sprintf("Code Length: %d digits", len(codeGame.Code))
	hintBounds := text.BoundString(basicfont.Face7x13, hintText)
	hintX := int(centerX) - hintBounds.Dx()/2
	hintY := int(centerY) - 50
	text.Draw(screen, hintText, basicfont.Face7x13, hintX, hintY, color.RGBA{255, 255, 255, 255})

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
		borderColor := color.RGBA{200, 200, 200, 255}
		if i < len(codeGame.PlayerInput) {
			boxColor = color.RGBA{0, 150, 0, 255}
			borderColor = color.RGBA{0, 255, 0, 255}
		}

		vector.DrawFilledRect(screen, boxX, boxY, digitWidth, digitHeight, boxColor, false)
		vector.StrokeRect(screen, boxX, boxY, digitWidth, digitHeight, 2, borderColor, false)

		// Draw entered digit
		if i < len(codeGame.PlayerInput) {
			digitText := fmt.Sprintf("%d", codeGame.PlayerInput[i])
			digitBounds := text.BoundString(basicfont.Face7x13, digitText)
			digitX := int(boxX) + int(digitWidth)/2 - digitBounds.Dx()/2
			digitY := int(boxY) + int(digitHeight)/2 + 4
			text.Draw(screen, digitText, basicfont.Face7x13, digitX, digitY, color.RGBA{255, 255, 255, 255})
		} else if i == len(codeGame.PlayerInput) {
			// Show cursor in next empty box
			cursorText := "_"
			cursorBounds := text.BoundString(basicfont.Face7x13, cursorText)
			cursorX := int(boxX) + int(digitWidth)/2 - cursorBounds.Dx()/2
			cursorY := int(boxY) + int(digitHeight)/2 + 4
			text.Draw(screen, cursorText, basicfont.Face7x13, cursorX, cursorY, color.RGBA{255, 255, 0, 255})
		}
	}

	// Draw backspace hint
	backspaceText := "Press BACKSPACE to clear"
	backspaceBounds := text.BoundString(basicfont.Face7x13, backspaceText)
	backspaceX := int(centerX) - backspaceBounds.Dx()/2
	backspaceY := int(centerY) + 40
	text.Draw(screen, backspaceText, basicfont.Face7x13, backspaceX, backspaceY, color.RGBA{150, 150, 150, 255})
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

// initializeEncryptedChat sets up E2E encrypted chat for multiplayer.
// For networked sessions, performs ECDH key exchange.
// For local/single-player, uses deterministic seed-based key.
func (g *Game) initializeEncryptedChat() {
	var encryptionKey []byte

	// Network mode: perform proper key exchange
	if g.networkMode && g.networkConn != nil {
		key, err := chat.PerformKeyExchange(g.networkConn)
		if err != nil {
			// Fallback to seed-based key on key exchange failure
			logrus.WithError(err).Warn("Key exchange failed, using seed-based key")
			encryptionKey = g.deriveSeedKey()
		} else {
			encryptionKey = key
			logrus.Info("Chat encryption key exchanged successfully")
		}
	} else {
		// Single-player or local mode: use deterministic seed-based key
		encryptionKey = g.deriveSeedKey()
	}

	g.chatManager = chat.NewChatWithKey(encryptionKey)
	g.chatMessages = make([]string, 0, 50)
	g.chatInput = ""
	g.chatInputActive = false

	g.hud.ShowMessage("Encrypted chat initialized - Press T to chat")
}

// deriveSeedKey derives a 32-byte encryption key from the game seed.
// Used for deterministic local multiplayer or as fallback.
func (g *Game) deriveSeedKey() []byte {
	seedBytes := make([]byte, 32)
	for i := 0; i < 32; i++ {
		seedBytes[i] = byte((g.seed >> (i * 8)) & 0xFF)
	}
	return seedBytes
}

// handleChatInput processes chat input and encryption.
func (g *Game) handleChatInput() {
	if g.chatManager == nil {
		return
	}

	if g.handleChatToggle() {
		return
	}

	if g.handleChatExit() {
		return
	}

	if g.handleChatSend() {
		return
	}

	g.handleChatTextInput()
}

// handleChatToggle activates chat input mode and returns true if activated.
func (g *Game) handleChatToggle() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyT) && !g.chatInputActive {
		g.chatInputActive = true
		g.chatInput = ""
		return true
	}
	return false
}

// handleChatExit exits chat input mode and returns true if exited.
func (g *Game) handleChatExit() bool {
	if g.chatInputActive && g.input.IsJustPressed(input.ActionPause) {
		g.chatInputActive = false
		g.chatInput = ""
		return true
	}
	return false
}

// handleChatSend sends the current chat message and returns true if sent.
func (g *Game) handleChatSend() bool {
	if !g.chatInputActive || !inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		return false
	}

	if g.chatInput != "" {
		encrypted, err := g.chatManager.Encrypt(g.chatInput)
		if err == nil {
			decrypted, err := g.chatManager.Decrypt(encrypted)
			if err == nil {
				g.addChatMessage("[You]: " + decrypted)
			}
		}
		g.chatInput = ""
	}
	g.chatInputActive = false
	return true
}

// handleChatTextInput processes text input and backspace for chat.
func (g *Game) handleChatTextInput() {
	if !g.chatInputActive {
		return
	}

	g.chatInput = g.chatInput + string(ebiten.AppendInputChars(nil))

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.chatInput) > 0 {
		g.chatInput = g.chatInput[:len(g.chatInput)-1]
	}

	if len(g.chatInput) > 100 {
		g.chatInput = g.chatInput[:100]
	}
}

// addChatMessage adds a message to the chat history.
func (g *Game) addChatMessage(message string) {
	g.chatMessages = append(g.chatMessages, message)

	// Keep only last 50 messages
	if len(g.chatMessages) > 50 {
		g.chatMessages = g.chatMessages[len(g.chatMessages)-50:]
	}
}

func main() {
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}

	initializeEbitenWindow()
	stopWatch := setupConfigHotReload()
	defer func() {
		if stopWatch != nil {
			stopWatch()
		}
	}()

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// initializeEbitenWindow configures the initial Ebiten window settings.
func initializeEbitenWindow() {
	ebiten.SetWindowSize(config.C.WindowWidth, config.C.WindowHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(config.C.VSync)
	ebiten.SetFullscreen(config.C.FullScreen)
	ebiten.SetWindowTitle("VIOLENCE")
	ebiten.SetCursorMode(ebiten.CursorModeVisible)

	if config.C.MaxTPS > 0 {
		ebiten.SetTPS(config.C.MaxTPS)
	}
}

// setupConfigHotReload enables configuration hot-reloading and returns a stop function.
func setupConfigHotReload() func() {
	stopWatch, err := config.Watch(applyConfigChanges)
	if err != nil {
		log.Printf("Warning: config hot-reload failed to start: %v", err)
		return nil
	}
	return stopWatch
}

// applyConfigChanges handles runtime-changeable configuration settings.
func applyConfigChanges(old, new config.Config) {
	if new.VSync != old.VSync {
		ebiten.SetVsyncEnabled(new.VSync)
	}
	if new.FullScreen != old.FullScreen {
		ebiten.SetFullscreen(new.FullScreen)
	}
	if new.WindowWidth != old.WindowWidth || new.WindowHeight != old.WindowHeight {
		ebiten.SetWindowSize(new.WindowWidth, new.WindowHeight)
	}
	if new.MaxTPS != old.MaxTPS {
		if new.MaxTPS > 0 {
			ebiten.SetTPS(new.MaxTPS)
		} else {
			ebiten.SetTPS(60)
		}
	}
}
