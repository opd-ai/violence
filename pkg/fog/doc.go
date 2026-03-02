/*
Package fog provides atmospheric depth rendering through distance-based fog effects.

The fog system adds visual depth to the game world by applying genre-specific
atmospheric effects that fade distant entities. This creates immersion and
improves readability by focusing attention on nearby action.

# Features

  - Genre-aware fog presets (fantasy dungeons, horror atmosphere, cyberpunk smog)
  - Configurable falloff curves (linear, exponential, exponential squared)
  - Per-entity fog density based on distance from camera
  - Automatic visibility culling for heavily obscured entities
  - Zero-allocation updates (component pooling via ECS)

# Usage

	// Create fog system with genre preset
	fogSystem := fog.NewSystem("horror")

	// Update camera position each frame
	fogSystem.SetCamera(playerX, playerY)

	// Process all entities (called by engine)
	fogSystem.Update(world, deltaTime)

	// Renderer reads fog component for blending
	fogComp := world.GetComponent(entity, "fog")
	fog := fogComp.(*fog.Component)
	if fog.Visible {
		// Apply fog tint and density to sprite
		blendedColor := lerpColor(spriteColor, fog.Tint, fog.FogDensity)
	}

# Integration

The system automatically adds fog.Component to entities with positions.
Rendering code should check for fog components and blend sprite colors
accordingly. Entities with fog.Visible == false can be culled.

# Performance

The system processes all positioned entities each frame. Spatial indexing
is recommended for scenes with >500 entities. Fog components are transient
and not serialized.
*/
package fog
