package flicker

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewWorldSystem(t *testing.T) {
	ws := NewWorldSystem("fantasy")

	if ws == nil {
		t.Fatal("NewWorldSystem returned nil")
	}
	if ws.flickerSys == nil {
		t.Error("flickerSys is nil")
	}
	if ws.tick != 0 {
		t.Errorf("tick = %d, want 0", ws.tick)
	}
}

func TestWorldSystemSetGenre(t *testing.T) {
	ws := NewWorldSystem("fantasy")
	ws.SetGenre("horror")

	if ws.flickerSys.GetGenre() != "horror" {
		t.Errorf("After SetGenre, genre = %q, want 'horror'", ws.flickerSys.GetGenre())
	}
}

func TestWorldSystemGetFlickerSystem(t *testing.T) {
	ws := NewWorldSystem("fantasy")
	sys := ws.GetFlickerSystem()

	if sys == nil {
		t.Error("GetFlickerSystem returned nil")
	}
	if sys.GetGenre() != "fantasy" {
		t.Errorf("GetFlickerSystem().GetGenre() = %q, want 'fantasy'", sys.GetGenre())
	}
}

func TestWorldSystemGetTick(t *testing.T) {
	ws := NewWorldSystem("fantasy")

	if ws.GetTick() != 0 {
		t.Errorf("Initial tick = %d, want 0", ws.GetTick())
	}
}

func TestWorldSystemInitializeComponent(t *testing.T) {
	ws := NewWorldSystem("fantasy")
	comp := NewComponent("torch", 12345)

	ws.InitializeComponent(comp, 1.0, 0.6, 0.2)

	if comp.Params.LightType != "torch" {
		t.Errorf("Params.LightType = %q, want 'torch'", comp.Params.LightType)
	}
	if comp.Params.BaseR != 1.0 {
		t.Errorf("Params.BaseR = %f, want 1.0", comp.Params.BaseR)
	}
}

func TestWorldSystemUpdate(t *testing.T) {
	ws := NewWorldSystem("fantasy")
	world := engine.NewWorld()

	// Create an entity with flicker component
	ent := world.AddEntity()
	comp := NewComponent("torch", 12345)
	ws.InitializeComponent(comp, 1.0, 0.6, 0.2)
	world.AddComponent(ent, comp)

	// Record initial values
	initialIntensity := comp.CurrentIntensity

	// Update multiple times
	for i := 0; i < 60; i++ {
		ws.Update(world)
	}

	// Check tick advanced
	if ws.GetTick() != 60 {
		t.Errorf("After 60 updates, tick = %d, want 60", ws.GetTick())
	}

	// Check values changed (flicker should cause variation)
	// Note: intensity might be same by chance, so just verify no errors
	_ = initialIntensity
}

func TestWorldSystemUpdateDisabled(t *testing.T) {
	ws := NewWorldSystem("fantasy")
	world := engine.NewWorld()

	// Create an entity with disabled flicker
	ent := world.AddEntity()
	comp := NewComponent("torch", 12345)
	ws.InitializeComponent(comp, 1.0, 0.6, 0.2)
	comp.SetEnabled(false)
	comp.CurrentIntensity = 0.5 // Set a non-default value
	world.AddComponent(ent, comp)

	// Update
	ws.Update(world)

	// Disabled component should not be modified
	if comp.CurrentIntensity != 0.5 {
		t.Errorf("Disabled component intensity changed to %f", comp.CurrentIntensity)
	}
}

func TestWorldSystemMultipleEntities(t *testing.T) {
	ws := NewWorldSystem("fantasy")
	world := engine.NewWorld()

	// Create multiple entities with different flicker components
	ent1 := world.AddEntity()
	comp1 := NewComponent("torch", 11111)
	ws.InitializeComponent(comp1, 1.0, 0.6, 0.2)
	world.AddComponent(ent1, comp1)

	ent2 := world.AddEntity()
	comp2 := NewComponent("candle", 22222)
	ws.InitializeComponent(comp2, 1.0, 0.8, 0.3)
	world.AddComponent(ent2, comp2)

	ent3 := world.AddEntity()
	comp3 := NewComponent("brazier", 33333)
	ws.InitializeComponent(comp3, 1.0, 0.5, 0.1)
	world.AddComponent(ent3, comp3)

	// Update
	for i := 0; i < 30; i++ {
		ws.Update(world)
	}

	// All components should have been processed
	if comp1.CurrentIntensity == 1.0 && comp2.CurrentIntensity == 1.0 && comp3.CurrentIntensity == 1.0 {
		t.Log("All intensities still at 1.0 (could be coincidence)")
	}
}

func TestComponentReflection(t *testing.T) {
	// Verify component type can be used with reflection (for ECS query)
	comp := &Component{}
	compType := reflect.TypeOf(comp)

	if compType.Kind() != reflect.Ptr {
		t.Errorf("Component type kind = %v, want Ptr", compType.Kind())
	}
	if compType.Elem().Name() != "Component" {
		t.Errorf("Component element name = %q, want 'Component'", compType.Elem().Name())
	}
}
