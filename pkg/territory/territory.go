// Package territory implements dynamic faction territory control and warfare.
//
// The territory system divides the game world into controllable zones that factions
// can contest and capture. Territory ownership affects resource spawning, enemy patrols,
// reinforcement availability, and creates dynamic battle fronts between opposing factions.
//
// Example usage:
//
//	// Initialize territory system
//	territorySys := territory.NewControlSystem(mapWidth, mapHeight, factionSystem)
//	world.AddSystem(territorySys)
//
//	// Assign initial territory control based on room placement
//	rooms := bsp.GetRooms(bspTree)
//	for i, room := range rooms {
//	    factionID := faction.FactionMercenaries
//	    territorySys.ClaimRoom(room, factionID)
//	}
//
//	// System automatically handles:
//	// - Territory contestation during combat
//	// - Patrol spawning in controlled zones
//	// - Reinforcements when territories are threatened
//	// - Battle front detection between faction borders
package territory

import (
	"math"
	"math/rand"
	"reflect"
	"sync"

	"github.com/opd-ai/violence/pkg/bsp"
	"github.com/opd-ai/violence/pkg/combat"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/faction"
	"github.com/sirupsen/logrus"
)

// TerritoryID uniquely identifies a territory zone.
type TerritoryID string

// Territory represents a controlled zone in the game world.
type Territory struct {
	ID             TerritoryID
	CenterX        float64
	CenterY        float64
	Radius         float64
	ControlFaction faction.FactionID
	Contested      bool
	ControlPoints  int
	MaxControl     int
	PatrolSpawns   []SpawnPoint
	LastBattleTime float64
}

// SpawnPoint defines a location for patrol spawning.
type SpawnPoint struct {
	X, Y   float64
	Active bool
}

// TerritoryComponent marks an entity as within a specific territory.
type TerritoryComponent struct {
	CurrentTerritory TerritoryID
	LastCheckTime    float64
}

// Type returns component type identifier.
func (t *TerritoryComponent) Type() string {
	return "TerritoryComponent"
}

// PatrolComponent marks an entity as a territorial patrol.
type PatrolComponent struct {
	TerritoryID TerritoryID
	PatrolRoute []Position
	RouteIndex  int
	HomeX       float64
	HomeY       float64
}

// Type returns component type identifier.
func (p *PatrolComponent) Type() string {
	return "PatrolComponent"
}

// Position represents a 2D coordinate.
type Position struct {
	X, Y float64
}

// PositionComponent provides entity position.
type PositionComponent struct {
	X, Y float64
}

// Type returns component type identifier.
func (p *PositionComponent) Type() string {
	return "PositionComponent"
}

// ReinforcementComponent marks an entity as a reinforcement spawn.
type ReinforcementComponent struct {
	TerritoryID TerritoryID
	SpawnedAt   float64
	Duration    float64
}

// Type returns component type identifier.
func (r *ReinforcementComponent) Type() string {
	return "ReinforcementComponent"
}

// ControlSystem manages territorial control and faction warfare.
type ControlSystem struct {
	territories           map[TerritoryID]*Territory
	factionSystem         *faction.ReputationSystem
	mapWidth              int
	mapHeight             int
	gridSize              int
	territoryGrid         [][]TerritoryID
	contestThreshold      int
	mu                    sync.RWMutex
	gameTime              float64
	reinforcementCooldown float64
}

// NewControlSystem creates a territory control system.
func NewControlSystem(mapWidth, mapHeight int, factionSys *faction.ReputationSystem) *ControlSystem {
	gridSize := 8
	gridW := (mapWidth + gridSize - 1) / gridSize
	gridH := (mapHeight + gridSize - 1) / gridSize

	grid := make([][]TerritoryID, gridH)
	for i := range grid {
		grid[i] = make([]TerritoryID, gridW)
	}

	return &ControlSystem{
		territories:           make(map[TerritoryID]*Territory),
		factionSystem:         factionSys,
		mapWidth:              mapWidth,
		mapHeight:             mapHeight,
		gridSize:              gridSize,
		territoryGrid:         grid,
		contestThreshold:      50,
		reinforcementCooldown: 30.0,
	}
}

// Update processes territory control, patrol spawning, and contestation.
func (s *ControlSystem) Update(w *engine.World) {
	s.mu.Lock()
	s.gameTime += 1.0 / 60.0
	s.mu.Unlock()

	s.updateTerritoryAssignments(w)
	s.updateContestStatus(w)
	s.spawnPatrols(w)
	s.spawnReinforcements(w)
	s.updatePatrolMovement(w)
}

func (s *ControlSystem) updateTerritoryAssignments(w *engine.World) {
	posType := reflect.TypeOf((*PositionComponent)(nil))
	terType := reflect.TypeOf((*TerritoryComponent)(nil))

	entities := w.Query(posType)
	for _, ent := range entities {
		posComp, ok := w.GetComponent(ent, posType)
		if !ok {
			continue
		}

		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		territoryID := s.getTerritoryAt(pos.X, pos.Y)

		terComp, hasTer := w.GetComponent(ent, terType)
		if !hasTer {
			tc := &TerritoryComponent{
				CurrentTerritory: territoryID,
				LastCheckTime:    s.gameTime,
			}
			w.AddComponent(ent, tc)
		} else if tc, ok := terComp.(*TerritoryComponent); ok {
			if tc.CurrentTerritory != territoryID {
				tc.CurrentTerritory = territoryID
				tc.LastCheckTime = s.gameTime
			}
		}
	}
}

func (s *ControlSystem) updateContestStatus(w *engine.World) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, territory := range s.territories {
		if territory.Contested {
			s.processContestation(w, territory)
		}
	}
}

func (s *ControlSystem) processContestation(w *engine.World, territory *Territory) {
	factionCounts := s.countLivingFactionMembers(w, territory.ID)
	controllingCount, contestingCount, contestingFaction := s.analyzeContestationState(territory, factionCounts)

	s.updateTerritoryControl(territory, controllingCount, contestingCount, contestingFaction)
	s.updateTerritoryContestedState(territory, contestingCount)
}

// countLivingFactionMembers counts alive faction members in a territory.
func (s *ControlSystem) countLivingFactionMembers(w *engine.World, territoryID TerritoryID) map[faction.FactionID]int {
	terType := reflect.TypeOf((*TerritoryComponent)(nil))
	facType := reflect.TypeOf((*faction.FactionMemberComponent)(nil))
	healthType := reflect.TypeOf((*combat.HealthComponent)(nil))

	factionCounts := make(map[faction.FactionID]int)
	entities := w.Query(terType, facType, healthType)

	for _, ent := range entities {
		if s.isLivingTerritoryMember(w, ent, territoryID, terType, facType, healthType) {
			facComp, _ := w.GetComponent(ent, facType)
			fc, _ := facComp.(*faction.FactionMemberComponent)
			factionCounts[fc.FactionID]++
		}
	}

	return factionCounts
}

// isLivingTerritoryMember checks if an entity is alive and in the specified territory.
func (s *ControlSystem) isLivingTerritoryMember(w *engine.World, ent engine.Entity, territoryID TerritoryID, terType, facType, healthType reflect.Type) bool {
	terComp, _ := w.GetComponent(ent, terType)
	facComp, _ := w.GetComponent(ent, facType)
	healthComp, _ := w.GetComponent(ent, healthType)

	tc, okT := terComp.(*TerritoryComponent)
	_, okF := facComp.(*faction.FactionMemberComponent)
	hc, okH := healthComp.(*combat.HealthComponent)

	return okT && okF && okH && tc.CurrentTerritory == territoryID && hc.Current > 0
}

// analyzeContestationState determines the strongest contesting faction.
func (s *ControlSystem) analyzeContestationState(territory *Territory, factionCounts map[faction.FactionID]int) (int, int, faction.FactionID) {
	controllingCount := factionCounts[territory.ControlFaction]
	contestingCount := 0
	var contestingFaction faction.FactionID

	for factionID, count := range factionCounts {
		if factionID != territory.ControlFaction && count > contestingCount {
			contestingCount = count
			contestingFaction = factionID
		}
	}

	return controllingCount, contestingCount, contestingFaction
}

// updateTerritoryControl adjusts control points and handles territory transfer.
func (s *ControlSystem) updateTerritoryControl(territory *Territory, controllingCount, contestingCount int, contestingFaction faction.FactionID) {
	if contestingCount > controllingCount {
		territory.ControlPoints--
		if territory.ControlPoints <= 0 {
			s.transferControl(territory, contestingFaction)
		}
	} else if controllingCount > 0 {
		territory.ControlPoints++
		if territory.ControlPoints > territory.MaxControl {
			territory.ControlPoints = territory.MaxControl
		}
	}
}

// updateTerritoryContestedState updates the contested flag and battle time.
func (s *ControlSystem) updateTerritoryContestedState(territory *Territory, contestingCount int) {
	territory.Contested = (contestingCount > 0)
	if territory.Contested {
		territory.LastBattleTime = s.gameTime
	}
}

func (s *ControlSystem) transferControl(territory *Territory, newFaction faction.FactionID) {
	oldFaction := territory.ControlFaction
	territory.ControlFaction = newFaction
	territory.ControlPoints = territory.MaxControl / 2
	territory.Contested = false

	logrus.WithFields(logrus.Fields{
		"system":       "territory",
		"territory_id": territory.ID,
		"old_faction":  oldFaction,
		"new_faction":  newFaction,
	}).Info("Territory control transferred")

	s.updateTerritoryGrid(territory)
}

func (s *ControlSystem) updateTerritoryGrid(territory *Territory) {
	gridX := int(territory.CenterX) / s.gridSize
	gridY := int(territory.CenterY) / s.gridSize

	radius := int(territory.Radius) / s.gridSize
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			gx := gridX + dx
			gy := gridY + dy
			if gx >= 0 && gx < len(s.territoryGrid[0]) && gy >= 0 && gy < len(s.territoryGrid) {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist <= float64(radius) {
					s.territoryGrid[gy][gx] = territory.ID
				}
			}
		}
	}
}

func (s *ControlSystem) getTerritoryAt(x, y float64) TerritoryID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	gridX := int(x) / s.gridSize
	gridY := int(y) / s.gridSize

	if gridX >= 0 && gridX < len(s.territoryGrid[0]) && gridY >= 0 && gridY < len(s.territoryGrid) {
		return s.territoryGrid[gridY][gridX]
	}
	return ""
}

func (s *ControlSystem) spawnPatrols(w *engine.World) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, territory := range s.territories {
		if territory.Contested || len(territory.PatrolSpawns) == 0 {
			continue
		}

		if s.gameTime-territory.LastBattleTime < 15.0 {
			continue
		}

		patType := reflect.TypeOf((*PatrolComponent)(nil))
		existingPatrols := w.Query(patType)
		patrolCount := 0
		for _, ent := range existingPatrols {
			patComp, _ := w.GetComponent(ent, patType)
			if pc, ok := patComp.(*PatrolComponent); ok && pc.TerritoryID == territory.ID {
				patrolCount++
			}
		}

		maxPatrols := 2
		if patrolCount >= maxPatrols {
			continue
		}

		if rand.Float64() > 0.01 {
			continue
		}

		s.spawnPatrolEntity(w, territory)
	}
}

func (s *ControlSystem) spawnPatrolEntity(w *engine.World, territory *Territory) {
	if len(territory.PatrolSpawns) == 0 {
		return
	}

	spawn := territory.PatrolSpawns[rand.Intn(len(territory.PatrolSpawns))]
	if !spawn.Active {
		return
	}

	patrolEnt := w.AddEntity()

	w.AddComponent(patrolEnt, &PositionComponent{
		X: spawn.X,
		Y: spawn.Y,
	})

	w.AddComponent(patrolEnt, &faction.FactionMemberComponent{
		FactionID: territory.ControlFaction,
		Rank:      1,
	})

	w.AddComponent(patrolEnt, &combat.HealthComponent{
		Current: 100.0,
		Max:     100.0,
	})

	route := s.generatePatrolRoute(territory, spawn)
	w.AddComponent(patrolEnt, &PatrolComponent{
		TerritoryID: territory.ID,
		PatrolRoute: route,
		RouteIndex:  0,
		HomeX:       spawn.X,
		HomeY:       spawn.Y,
	})

	w.AddComponent(patrolEnt, &TerritoryComponent{
		CurrentTerritory: territory.ID,
		LastCheckTime:    s.gameTime,
	})

	logrus.WithFields(logrus.Fields{
		"system":    "territory",
		"territory": territory.ID,
		"faction":   territory.ControlFaction,
		"spawn_x":   spawn.X,
		"spawn_y":   spawn.Y,
	}).Debug("Spawned patrol")
}

func (s *ControlSystem) generatePatrolRoute(territory *Territory, start SpawnPoint) []Position {
	route := make([]Position, 0, 4)
	route = append(route, Position{X: start.X, Y: start.Y})

	numWaypoints := 3
	for i := 0; i < numWaypoints; i++ {
		angle := rand.Float64() * 2 * math.Pi
		dist := territory.Radius * (0.3 + rand.Float64()*0.5)

		x := territory.CenterX + math.Cos(angle)*dist
		y := territory.CenterY + math.Sin(angle)*dist

		route = append(route, Position{X: x, Y: y})
	}

	route = append(route, Position{X: start.X, Y: start.Y})
	return route
}

func (s *ControlSystem) spawnReinforcements(w *engine.World) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, territory := range s.territories {
		if !territory.Contested {
			continue
		}

		timeSinceBattle := s.gameTime - territory.LastBattleTime
		if timeSinceBattle < s.reinforcementCooldown {
			continue
		}

		if rand.Float64() > 0.02 {
			continue
		}

		s.spawnReinforcementEntity(w, territory)
		territory.LastBattleTime = s.gameTime
	}
}

func (s *ControlSystem) spawnReinforcementEntity(w *engine.World, territory *Territory) {
	if len(territory.PatrolSpawns) == 0 {
		return
	}

	spawn := territory.PatrolSpawns[rand.Intn(len(territory.PatrolSpawns))]

	reinforcementEnt := w.AddEntity()

	w.AddComponent(reinforcementEnt, &PositionComponent{
		X: spawn.X,
		Y: spawn.Y,
	})

	w.AddComponent(reinforcementEnt, &faction.FactionMemberComponent{
		FactionID: territory.ControlFaction,
		Rank:      2,
	})

	w.AddComponent(reinforcementEnt, &combat.HealthComponent{
		Current: 150.0,
		Max:     150.0,
	})

	w.AddComponent(reinforcementEnt, &ReinforcementComponent{
		TerritoryID: territory.ID,
		SpawnedAt:   s.gameTime,
		Duration:    60.0,
	})

	w.AddComponent(reinforcementEnt, &TerritoryComponent{
		CurrentTerritory: territory.ID,
		LastCheckTime:    s.gameTime,
	})

	logrus.WithFields(logrus.Fields{
		"system":    "territory",
		"territory": territory.ID,
		"faction":   territory.ControlFaction,
	}).Info("Spawned reinforcements")
}

func (s *ControlSystem) updatePatrolMovement(w *engine.World) {
	patType := reflect.TypeOf((*PatrolComponent)(nil))
	posType := reflect.TypeOf((*PositionComponent)(nil))

	entities := w.Query(patType, posType)
	for _, ent := range entities {
		patComp, _ := w.GetComponent(ent, patType)
		posComp, _ := w.GetComponent(ent, posType)

		pc, okP := patComp.(*PatrolComponent)
		pos, okPos := posComp.(*PositionComponent)

		if !okP || !okPos || len(pc.PatrolRoute) == 0 {
			continue
		}

		target := pc.PatrolRoute[pc.RouteIndex]
		dx := target.X - pos.X
		dy := target.Y - pos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < 0.5 {
			pc.RouteIndex = (pc.RouteIndex + 1) % len(pc.PatrolRoute)
		} else {
			speed := 0.05
			pos.X += (dx / dist) * speed
			pos.Y += (dy / dist) * speed
		}
	}
}

// ClaimRoom assigns a faction to control a room-based territory.
func (s *ControlSystem) ClaimRoom(room *bsp.Room, factionID faction.FactionID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	territoryID := TerritoryID("territory_" + string(rune(room.X+'0')) + "_" + string(rune(room.Y+'0')))

	centerX := float64(room.X + room.W/2)
	centerY := float64(room.Y + room.H/2)
	radius := math.Max(float64(room.W), float64(room.H)) / 2.0

	spawns := make([]SpawnPoint, 0, 4)
	spawns = append(spawns, SpawnPoint{
		X:      centerX - float64(room.W)/4,
		Y:      centerY - float64(room.H)/4,
		Active: true,
	})
	spawns = append(spawns, SpawnPoint{
		X:      centerX + float64(room.W)/4,
		Y:      centerY + float64(room.H)/4,
		Active: true,
	})

	territory := &Territory{
		ID:             territoryID,
		CenterX:        centerX,
		CenterY:        centerY,
		Radius:         radius,
		ControlFaction: factionID,
		Contested:      false,
		ControlPoints:  100,
		MaxControl:     100,
		PatrolSpawns:   spawns,
		LastBattleTime: 0,
	}

	s.territories[territoryID] = territory
	s.updateTerritoryGrid(territory)

	logrus.WithFields(logrus.Fields{
		"system":    "territory",
		"territory": territoryID,
		"faction":   factionID,
		"center_x":  centerX,
		"center_y":  centerY,
	}).Debug("Territory claimed")
}

// GetTerritory retrieves territory by ID.
func (s *ControlSystem) GetTerritory(id TerritoryID) *Territory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.territories[id]
}

// GetTerritoryByPosition finds the territory at a given position.
func (s *ControlSystem) GetTerritoryByPosition(x, y float64) *Territory {
	id := s.getTerritoryAt(x, y)
	if id == "" {
		return nil
	}
	return s.GetTerritory(id)
}

// IsContested returns whether a territory is currently under attack.
func (s *ControlSystem) IsContested(id TerritoryID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	territory, ok := s.territories[id]
	if !ok {
		return false
	}
	return territory.Contested
}

// GetControllingFaction returns the faction controlling a territory.
func (s *ControlSystem) GetControllingFaction(id TerritoryID) faction.FactionID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	territory, ok := s.territories[id]
	if !ok {
		return ""
	}
	return territory.ControlFaction
}

// GetBattleFronts returns territories that are currently contested.
func (s *ControlSystem) GetBattleFronts() []*Territory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fronts := make([]*Territory, 0)
	for _, territory := range s.territories {
		if territory.Contested {
			fronts = append(fronts, territory)
		}
	}
	return fronts
}
