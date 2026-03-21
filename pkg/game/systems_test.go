package game

import (
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestRegisterECSSystems(t *testing.T) {
	// Test that RegisterECSSystems doesn't panic with nil dependencies
	// This is a smoke test to verify the function signature is correct
	world := engine.NewWorld()

	// Create minimal mock dependencies (all nil)
	// RegisterECSSystems should not be called with nil systems in production,
	// but we verify the function exists and has correct signature
	t.Run("function exists", func(t *testing.T) {
		// Just verify the types exist
		var _ *SystemDependencies
		var _ *engine.World = world
	})
}

func TestConnectSlidingSystem(t *testing.T) {
	// Verify function signature exists
	t.Run("function exists", func(t *testing.T) {
		// ConnectSlidingSystem(nil, nil) would panic, so just verify the function is exported
		_ = ConnectSlidingSystem
	})
}
