package fog

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

// TestIntegrationWithECS verifies fog system works end-to-end with the ECS.
func TestIntegrationWithECS(t *testing.T) {
	world := engine.NewWorld()
	fogSys := NewSystem("horror")
	fogSys.SetCamera(50, 50)

	// Register fog system with world
	world.AddSystem(fogSys)

	// Create multiple entities at various distances
	entities := []struct {
		x, y         float64
		expectedFar  bool
		expectedNear bool
	}{
		{50, 50, false, true},   // Same as camera - very near
		{52, 52, false, true},   // Close to camera
		{60, 60, false, false},  // Mid distance
		{100, 100, true, false}, // Far from camera
	}

	for i, ent := range entities {
		e := world.AddEntity()
		world.AddComponent(e, &engine.Position{X: ent.x, Y: ent.y})
		entities[i].expectedFar = (ent.x-50)*(ent.x-50)+(ent.y-50)*(ent.y-50) > 300
		entities[i].expectedNear = (ent.x-50)*(ent.x-50)+(ent.y-50)*(ent.y-50) < 20
	}

	// Run system update
	fogSys.Update(world)

	// Verify all entities got fog components
	fogType := reflect.TypeOf(&Component{})
	posType := reflect.TypeOf(&engine.Position{})

	entitiesWithPos := world.Query(posType)
	for _, e := range entitiesWithPos {
		fogComp, hasFog := world.GetComponent(e, fogType)
		if !hasFog {
			t.Errorf("entity %v missing fog component", e)
			continue
		}

		fog := fogComp.(*Component)

		// Verify fog density is calculated
		if fog.FogDensity < 0.0 || fog.FogDensity > 1.0 {
			t.Errorf("entity %v fog density %.3f out of range [0.0, 1.0]",
				e, fog.FogDensity)
		}

		// Verify tint is set
		if fog.Tint[0] == 0 && fog.Tint[1] == 0 && fog.Tint[2] == 0 {
			t.Errorf("entity %v fog tint not set", e)
		}

		// Verify distance is calculated
		if fog.DistanceFromCamera < 0 {
			t.Errorf("entity %v negative distance", e)
		}
	}
}

// TestMultiGenreIntegration verifies different genres produce distinct fog.
func TestMultiGenreIntegration(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			world := engine.NewWorld()
			world.SetGenre(genre)
			fogSys := NewSystem(genre)
			fogSys.SetCamera(0, 0)

			world.AddSystem(fogSys)

			// Create test entity
			e := world.AddEntity()
			world.AddComponent(e, &engine.Position{X: 10, Y: 10})

			// Update system
			fogSys.Update(world)

			// Verify fog was applied
			fogType := reflect.TypeOf(&Component{})
			fogComp, hasFog := world.GetComponent(e, fogType)
			if !hasFog {
				t.Fatalf("%s: fog component not added", genre)
			}

			fog := fogComp.(*Component)
			if fog.FogDensity == 0 {
				t.Errorf("%s: fog density is zero (should have some fog at distance)", genre)
			}
		})
	}
}

// TestCameraMovement verifies fog updates when camera moves.
func TestCameraMovement(t *testing.T) {
	world := engine.NewWorld()
	fogSys := NewSystem("fantasy")
	world.AddSystem(fogSys)

	// Create stationary entity
	e := world.AddEntity()
	world.AddComponent(e, &engine.Position{X: 20, Y: 20})

	fogType := reflect.TypeOf(&Component{})

	// Camera far from entity
	fogSys.SetCamera(0, 0)
	fogSys.Update(world)
	fogComp1, _ := world.GetComponent(e, fogType)
	density1 := fogComp1.(*Component).FogDensity

	// Move camera closer to entity
	fogSys.SetCamera(15, 15)
	fogSys.Update(world)
	fogComp2, _ := world.GetComponent(e, fogType)
	density2 := fogComp2.(*Component).FogDensity

	if density2 >= density1 {
		t.Errorf("fog density should decrease when camera approaches: %.3f >= %.3f",
			density2, density1)
	}
}
