package proximityui

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewSystem(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.genre != genre {
				t.Errorf("genre = %q, want %q", sys.genre, genre)
			}
			// Check config was set
			if sys.config.AdjacentDistance <= 0 {
				t.Error("AdjacentDistance should be positive")
			}
			if sys.config.FarDistance <= sys.config.AdjacentDistance {
				t.Error("FarDistance should be greater than AdjacentDistance")
			}
		})
	}
}

func TestDistanceToDetailLevel(t *testing.T) {
	sys := NewSystem("fantasy")

	tests := []struct {
		name     string
		distance float64
		want     DetailLevel
	}{
		{"at camera", 0.0, DetailFull},
		{"very close", 1.0, DetailFull},
		{"adjacent boundary", 3.0, DetailFull},
		{"near range", 5.0, DetailModerate},
		{"near boundary", 8.0, DetailModerate},
		{"mid range", 10.0, DetailMinimal},
		{"mid boundary", 15.0, DetailMinimal},
		{"far range", 18.0, DetailNone},
		{"very far", 100.0, DetailNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.distanceToDetailLevel(tt.distance)
			if got != tt.want {
				t.Errorf("distanceToDetailLevel(%f) = %v, want %v", tt.distance, got, tt.want)
			}
		})
	}
}

func TestPriorityOverrides(t *testing.T) {
	sys := NewSystem("fantasy")

	tests := []struct {
		name     string
		comp     *Component
		distance float64
		want     DetailLevel
	}{
		{
			name:     "targeted always full",
			comp:     &Component{IsTargeted: true},
			distance: 100.0,
			want:     DetailFull,
		},
		{
			name:     "boss at distance stays moderate",
			comp:     &Component{IsBoss: true},
			distance: 50.0,
			want:     DetailModerate,
		},
		{
			name:     "boss close gets full",
			comp:     &Component{IsBoss: true},
			distance: 2.0,
			want:     DetailFull,
		},
		{
			name:     "player at distance stays minimal",
			comp:     &Component{IsPlayer: true},
			distance: 50.0,
			want:     DetailMinimal,
		},
		{
			name:     "quest NPC stays moderate",
			comp:     &Component{IsQuestNPC: true},
			distance: 50.0,
			want:     DetailModerate,
		},
		{
			name:     "manual override",
			comp:     &Component{PriorityOverride: DetailFull},
			distance: 50.0,
			want:     DetailFull,
		},
		{
			name:     "no override uses distance",
			comp:     &Component{PriorityOverride: -1},
			distance: 50.0,
			want:     DetailNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.calculateDetailLevel(tt.comp, tt.distance)
			if got != tt.want {
				t.Errorf("calculateDetailLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenreConfigDifferences(t *testing.T) {
	horror := NewSystem("horror")
	scifi := NewSystem("scifi")

	// Horror should have shorter ranges than sci-fi
	if horror.config.FarDistance >= scifi.config.FarDistance {
		t.Errorf("horror FarDistance (%f) should be less than scifi (%f)",
			horror.config.FarDistance, scifi.config.FarDistance)
	}
	if horror.config.AdjacentDistance >= scifi.config.AdjacentDistance {
		t.Errorf("horror AdjacentDistance (%f) should be less than scifi (%f)",
			horror.config.AdjacentDistance, scifi.config.AdjacentDistance)
	}
}

func TestFadeAlphaCalculation(t *testing.T) {
	sys := NewSystem("fantasy")

	tests := []struct {
		name        string
		targetLevel DetailLevel
		distance    float64
		wantMin     float64
		wantMax     float64
	}{
		{
			name:        "full detail close",
			targetLevel: DetailFull,
			distance:    1.0,
			wantMin:     0.9,
			wantMax:     1.0,
		},
		{
			name:        "none detail far",
			targetLevel: DetailNone,
			distance:    25.0,
			wantMin:     0.0,
			wantMax:     0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &Component{TargetDetailLevel: tt.targetLevel}
			got := sys.calculateFadeAlpha(comp, tt.distance)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateFadeAlpha() = %f, want between %f and %f",
					got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestDetailLevelString(t *testing.T) {
	tests := []struct {
		level DetailLevel
		want  string
	}{
		{DetailNone, "none"},
		{DetailMinimal, "minimal"},
		{DetailModerate, "moderate"},
		{DetailFull, "full"},
		{DetailLevel(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComponentHelpers(t *testing.T) {
	tests := []struct {
		name        string
		level       DetailLevel
		showHealth  bool
		showName    bool
		showStatus  bool
		showFaction bool
	}{
		{"none", DetailNone, false, false, false, false},
		{"minimal", DetailMinimal, true, false, false, false},
		{"moderate", DetailModerate, true, true, false, false},
		{"full", DetailFull, true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &Component{CurrentDetailLevel: tt.level}

			if got := comp.ShouldShowHealthBar(); got != tt.showHealth {
				t.Errorf("ShouldShowHealthBar() = %v, want %v", got, tt.showHealth)
			}
			if got := comp.ShouldShowName(); got != tt.showName {
				t.Errorf("ShouldShowName() = %v, want %v", got, tt.showName)
			}
			if got := comp.ShouldShowStatusIcons(); got != tt.showStatus {
				t.Errorf("ShouldShowStatusIcons() = %v, want %v", got, tt.showStatus)
			}
			if got := comp.ShouldShowFactionBadge(); got != tt.showFaction {
				t.Errorf("ShouldShowFactionBadge() = %v, want %v", got, tt.showFaction)
			}
		})
	}
}

func TestSystemUpdate(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()

	// Create an entity with position and proximity component
	entity := world.AddEntity()
	world.AddComponent(entity, &engine.Position{X: 5.0, Y: 0.0})
	world.AddComponent(entity, NewComponent())

	// Set camera at origin
	sys.SetCameraPosition(0, 0)

	// Run update
	sys.Update(world)

	// Check component was updated
	compType := reflect.TypeOf(&Component{})
	comp, ok := world.GetComponent(entity, compType)
	if !ok {
		t.Fatal("Component not found after update")
	}

	proxComp := comp.(*Component)
	if proxComp.LastDistance != 5.0 {
		t.Errorf("LastDistance = %f, want 5.0", proxComp.LastDistance)
	}
	if proxComp.TargetDetailLevel != DetailModerate {
		t.Errorf("TargetDetailLevel = %v, want DetailModerate", proxComp.TargetDetailLevel)
	}
}

func TestTargetedEntity(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()

	// Create an entity far away
	entity := world.AddEntity()
	world.AddComponent(entity, &engine.Position{X: 50.0, Y: 0.0})
	world.AddComponent(entity, NewComponent())

	sys.SetCameraPosition(0, 0)
	sys.SetTargetedEntity(entity)

	sys.Update(world)

	// Targeted entity should be full detail despite distance
	level := sys.GetDetailLevel(world, entity)
	if level != DetailFull {
		t.Errorf("GetDetailLevel() = %v, want DetailFull for targeted entity", level)
	}

	// Clear target
	sys.ClearTargetedEntity()
	sys.Update(world)

	// Now should be based on distance
	level = sys.GetDetailLevel(world, entity)
	if level != DetailNone {
		t.Errorf("GetDetailLevel() = %v, want DetailNone for untargeted far entity", level)
	}
}

func TestTransitionSmoothing(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()

	entity := world.AddEntity()
	world.AddComponent(entity, &engine.Position{X: 2.0, Y: 0.0})
	comp := NewComponent()
	comp.CurrentDetailLevel = DetailNone
	comp.TargetDetailLevel = DetailNone
	comp.TransitionProgress = 1.0
	world.AddComponent(entity, comp)

	sys.SetCameraPosition(0, 0)

	// First update should start transition to new level
	sys.Update(world)

	compType := reflect.TypeOf(&Component{})
	compRaw, _ := world.GetComponent(entity, compType)
	proxComp := compRaw.(*Component)

	// Target should be DetailFull (close to camera)
	if proxComp.TargetDetailLevel != DetailFull {
		t.Errorf("TargetDetailLevel = %v, want DetailFull", proxComp.TargetDetailLevel)
	}

	// Should be transitioning (progress not 1.0 if levels differ)
	// After one update, progress should have increased
	if proxComp.CurrentDetailLevel == proxComp.TargetDetailLevel && proxComp.TransitionProgress != 1.0 {
		t.Error("Transition should complete when levels match")
	}
}

func TestShouldRenderHelpers(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()

	// Entity with no component
	noCompEntity := world.AddEntity()
	world.AddComponent(noCompEntity, &engine.Position{X: 2.0, Y: 0.0})

	sys.SetCameraPosition(0, 0)

	// Should use distance-based calculation
	if !sys.ShouldRenderHealthBar(world, noCompEntity) {
		t.Error("ShouldRenderHealthBar should return true for close entity")
	}
	if !sys.ShouldRenderName(world, noCompEntity) {
		t.Error("ShouldRenderName should return true for close entity")
	}
	if !sys.ShouldRenderStatusIcons(world, noCompEntity) {
		t.Error("ShouldRenderStatusIcons should return true for close entity")
	}

	// Far entity
	farEntity := world.AddEntity()
	world.AddComponent(farEntity, &engine.Position{X: 50.0, Y: 0.0})

	if sys.ShouldRenderHealthBar(world, farEntity) {
		t.Error("ShouldRenderHealthBar should return false for far entity")
	}
}

func TestNewComponentVariants(t *testing.T) {
	boss := NewBossComponent()
	if !boss.IsBoss {
		t.Error("NewBossComponent should set IsBoss=true")
	}

	player := NewPlayerComponent()
	if !player.IsPlayer {
		t.Error("NewPlayerComponent should set IsPlayer=true")
	}

	quest := NewQuestNPCComponent()
	if !quest.IsQuestNPC {
		t.Error("NewQuestNPCComponent should set IsQuestNPC=true")
	}
}

func TestSetTargeted(t *testing.T) {
	comp := NewComponent()

	comp.SetTargeted(true)
	if !comp.IsTargeted {
		t.Error("SetTargeted(true) should set IsTargeted=true")
	}
	if comp.TargetDetailLevel != DetailFull {
		t.Error("SetTargeted(true) should set TargetDetailLevel=DetailFull")
	}

	comp.SetTargeted(false)
	if comp.IsTargeted {
		t.Error("SetTargeted(false) should set IsTargeted=false")
	}
}

func TestGetEffectiveAlpha(t *testing.T) {
	tests := []struct {
		name       string
		fadeAlpha  float64
		transition float64
		want       float64
	}{
		{"full transition", 1.0, 1.0, 1.0},
		{"half alpha", 0.5, 1.0, 0.5},
		{"mid transition", 1.0, 0.5, 0.5},
		{"both half", 0.8, 0.5, 0.4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &Component{
				FadeAlpha:          tt.fadeAlpha,
				TransitionProgress: tt.transition,
			}
			got := comp.GetEffectiveAlpha()
			if got != tt.want {
				t.Errorf("GetEffectiveAlpha() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestConfigSetGet(t *testing.T) {
	sys := NewSystem("fantasy")
	originalConfig := sys.GetConfig()

	newConfig := Config{
		AdjacentDistance:  5.0,
		NearDistance:      10.0,
		MidDistance:       15.0,
		FarDistance:       20.0,
		TransitionSpeed:   5.0,
		DistanceFadeStart: 0.5,
		DistanceFadeEnd:   0.8,
	}

	sys.SetConfig(newConfig)
	gotConfig := sys.GetConfig()

	if gotConfig.AdjacentDistance != newConfig.AdjacentDistance {
		t.Errorf("AdjacentDistance = %f, want %f", gotConfig.AdjacentDistance, newConfig.AdjacentDistance)
	}

	// Restore
	sys.SetConfig(originalConfig)
}

func TestComponentType(t *testing.T) {
	comp := NewComponent()
	if comp.Type() != "proximityui" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "proximityui")
	}
}

// Benchmark tests
func BenchmarkDistanceToDetailLevel(b *testing.B) {
	sys := NewSystem("fantasy")
	distances := []float64{1.0, 5.0, 10.0, 18.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.distanceToDetailLevel(distances[i%len(distances)])
	}
}

func BenchmarkCalculateFadeAlpha(b *testing.B) {
	sys := NewSystem("fantasy")
	comp := &Component{TargetDetailLevel: DetailModerate}
	distances := []float64{1.0, 5.0, 10.0, 18.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.calculateFadeAlpha(comp, distances[i%len(distances)])
	}
}

func BenchmarkSystemUpdate(b *testing.B) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()

	// Create 100 entities
	for i := 0; i < 100; i++ {
		entity := world.AddEntity()
		world.AddComponent(entity, &engine.Position{
			X: float64(i % 20),
			Y: float64(i / 20),
		})
		world.AddComponent(entity, NewComponent())
	}

	sys.SetCameraPosition(10, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(world)
	}
}
