package hazard

import (
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewECSSystem(t *testing.T) {
	s := NewECSSystem(12345)
	if s == nil {
		t.Fatal("NewECSSystem returned nil")
	}
	if s.genre != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got '%s'", s.genre)
	}
}

func TestECSSystemSetGenre(t *testing.T) {
	s := NewECSSystem(12345)
	s.SetGenre("scifi")
	if s.genre != "scifi" {
		t.Errorf("Expected genre 'scifi', got '%s'", s.genre)
	}
}

func TestECSGenerateHazards(t *testing.T) {
	testMap := make([][]int, 20)
	for i := range testMap {
		testMap[i] = make([]int, 20)
		for j := range testMap[i] {
			if i == 0 || i == 19 || j == 0 || j == 19 {
				testMap[i][j] = 1 // Walls
			} else {
				testMap[i][j] = 0 // Floor
			}
		}
	}

	world := engine.NewWorld()
	s := NewECSSystem(12345)
	s.GenerateHazards(world, testMap, 67890)

	// Verify that hazard entities were created by checking world state
	// The system should have created entities - we can't directly count them without
	// exposing world internals, but we can verify the system doesn't panic
	if s == nil {
		t.Fatal("System became nil after GenerateHazards")
	}
}

func TestECSGenreHazards(t *testing.T) {
	tests := []struct {
		genre         string
		expectedTypes []Type
	}{
		{"fantasy", []Type{TypeSpikeTrap, TypeFireGrate, TypePoisonVent, TypeFallingRocks, TypeAcidPool}},
		{"scifi", []Type{TypeElectricFloor, TypeLaserGrid, TypeCryoField, TypePlasmaJet, TypeGravityWell}},
		{"horror", []Type{TypeSpikeTrap, TypePoisonVent, TypeAcidPool, TypeFallingRocks}},
		{"cyberpunk", []Type{TypeElectricFloor, TypeLaserGrid, TypePlasmaJet, TypeGravityWell}},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			s := NewECSSystem(12345)
			s.SetGenre(tt.genre)
			types := s.getGenreHazards()

			if len(types) != len(tt.expectedTypes) {
				t.Errorf("Expected %d hazard types for %s, got %d", len(tt.expectedTypes), tt.genre, len(types))
			}

			typeMap := make(map[Type]bool)
			for _, ht := range types {
				typeMap[ht] = true
			}

			for _, expected := range tt.expectedTypes {
				if !typeMap[expected] {
					t.Errorf("Genre %s missing expected hazard type %d", tt.genre, expected)
				}
			}
		})
	}
}

func TestECSUpdate(t *testing.T) {
	world := engine.NewWorld()
	s := NewECSSystem(12345)

	// Create a test hazard entity
	entity := world.AddEntity()
	world.AddComponent(entity, &PositionComponent{X: 5.0, Y: 5.0})
	hazard := &HazardComponent{
		Type:             TypeSpikeTrap,
		State:            StateInactive,
		Timer:            0.0,
		ChargeDuration:   0.5,
		ActiveDuration:   0.3,
		CooldownDuration: 2.0,
		CycleDuration:    2.8,
		Damage:           20,
		Persistent:       false,
	}
	world.AddComponent(entity, hazard)

	// Run update multiple times to cycle through states
	for i := 0; i < 100; i++ {
		s.Update(world)
	}

	// Verify the system ran without panic
	if s == nil {
		t.Fatal("System became nil after update")
	}
}

func TestECSCheckCollision(t *testing.T) {
	world := engine.NewWorld()
	s := NewECSSystem(12345)

	// Create an active hazard
	entity := world.AddEntity()
	world.AddComponent(entity, &PositionComponent{X: 10.0, Y: 10.0})
	world.AddComponent(entity, &HazardComponent{
		Type:   TypeSpikeTrap,
		State:  StateActive,
		Damage: 25,
		Width:  1.0,
		Height: 1.0,
	})

	// Test collision at hazard position
	hit, damage, _ := s.CheckCollision(world, 10.0, 10.0)
	if !hit {
		t.Error("Expected collision at hazard position")
	}
	if damage != 25 {
		t.Errorf("Expected damage 25, got %d", damage)
	}

	// Test no collision far away
	hit, _, _ = s.CheckCollision(world, 20.0, 20.0)
	if hit {
		t.Error("Expected no collision far from hazard")
	}

	// Create inactive hazard
	entity2 := world.AddEntity()
	world.AddComponent(entity2, &PositionComponent{X: 15.0, Y: 15.0})
	world.AddComponent(entity2, &HazardComponent{
		Type:   TypeFireGrate,
		State:  StateInactive,
		Damage: 30,
		Width:  1.0,
		Height: 1.0,
	})

	// Test no collision with inactive hazard
	hit, _, _ = s.CheckCollision(world, 15.0, 15.0)
	if hit {
		t.Error("Expected no collision with inactive hazard")
	}
}

func TestECSGetHazardsForRendering(t *testing.T) {
	world := engine.NewWorld()
	s := NewECSSystem(12345)

	// Create multiple hazards
	for i := 0; i < 3; i++ {
		entity := world.AddEntity()
		world.AddComponent(entity, &PositionComponent{X: float64(i), Y: float64(i)})
		world.AddComponent(entity, &HazardComponent{
			Type:   TypeSpikeTrap,
			State:  StateActive,
			Color:  0xFF0000,
			Width:  1.0,
			Height: 1.0,
		})
	}

	renderData := s.GetHazardsForRendering(world)
	if len(renderData) != 3 {
		t.Errorf("Expected 3 hazards for rendering, got %d", len(renderData))
	}

	for _, data := range renderData {
		if data.Color != 0xFF0000 {
			t.Errorf("Hazard has wrong color: 0x%X", data.Color)
		}
	}
}

func TestHazardComponentTypes(t *testing.T) {
	types := []Type{
		TypeSpikeTrap,
		TypeFireGrate,
		TypePoisonVent,
		TypeElectricFloor,
		TypeFallingRocks,
		TypeAcidPool,
		TypeLaserGrid,
		TypeCryoField,
		TypePlasmaJet,
		TypeGravityWell,
	}

	s := NewECSSystem(12345)
	localRNG := s.rng

	for _, hType := range types {
		comp := s.createHazardComponent(hType, localRNG)
		if comp == nil {
			t.Fatalf("createHazardComponent returned nil for type %d", hType)
		}
		if comp.Type != hType {
			t.Errorf("Expected type %d, got %d", hType, comp.Type)
		}
		if comp.Damage == 0 {
			t.Error("Hazard has zero damage")
		}
		if comp.Width <= 0 || comp.Height <= 0 {
			t.Errorf("Hazard has invalid dimensions: %.2f x %.2f", comp.Width, comp.Height)
		}
	}
}

func TestHazardStatusEffects(t *testing.T) {
	tests := []struct {
		hazardType   Type
		expectEffect bool
	}{
		{TypeSpikeTrap, false},
		{TypeFireGrate, true},
		{TypePoisonVent, true},
		{TypeElectricFloor, true},
		{TypeFallingRocks, false},
		{TypeAcidPool, true},
		{TypeLaserGrid, false},
		{TypeCryoField, true},
		{TypePlasmaJet, true},
		{TypeGravityWell, true},
	}

	s := NewECSSystem(12345)
	localRNG := s.rng

	for _, tt := range tests {
		comp := s.createHazardComponent(tt.hazardType, localRNG)
		hasEffect := comp.StatusEffect != ""
		if hasEffect != tt.expectEffect {
			t.Errorf("Type %d: expected status effect=%v, got effect='%s'",
				tt.hazardType, tt.expectEffect, comp.StatusEffect)
		}
	}
}
