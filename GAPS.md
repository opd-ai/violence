# Implementation Gaps

## v1.0 — Core Engine + Playable Single-Player

### Asset Pipeline
- **Gap**: No procedural audio or texture generation strategy is defined; the roadmap specifies "zero external assets" and all assets must be procedurally generated at runtime, but no synthesis approach for audio is implemented.
- **Impact**: Audio engine (`pkg/audio`) cannot produce sound without a runtime synthesis approach. Texture rendering requires a procedural pixel generator. No pre-rendered, embedded, or bundled audio files (e.g., `.wav`, `.ogg`, `.mp3`) or image files (e.g., `.png`, `.jpg`) are permitted.
- **Resolution needed**: Implement runtime audio synthesis (e.g., generating tones, noise, and waveforms programmatically via deterministic algorithms seeded by `pkg/rng`). All audio and visual assets must be procedurally generated at runtime — embedding or bundling pre-made asset files is not an option.

### ECS Component Query API
- **Gap**: The roadmap specifies "Component registration and query API" but the current `pkg/engine` stub has no component query mechanism (e.g., querying all entities with a specific set of components).
- **Impact**: Systems cannot efficiently iterate over entities matching their required component signature, which is fundamental to ECS operation.
- **Resolution needed**: Define the query API — options include bitmask-based archetype queries, iterator-based component filters, or reflection-based type matching.

### Tile Grid Format
- **Gap**: The raycaster requires a tile grid to ray-march against, and the BSP generator produces rooms/corridors, but no shared tile grid data structure is defined between `pkg/bsp` and `pkg/raycaster`.
- **Impact**: The two core systems cannot integrate without an agreed-upon 2D grid format representing walls, floors, doors, and empty space.
- **Resolution needed**: Define a shared `TileMap` type (e.g., `[][]uint8` with tile-type constants) that BSP outputs and Raycaster consumes. Decide where this type lives — likely `pkg/engine` or a new `pkg/level` package.

### Player Entity Definition
- **Gap**: No player entity schema is defined — there is no specification for which components a player entity carries (position, health, armor, inventory, camera reference).
- **Impact**: Camera, input, HUD, save/load, and combat systems all need to read/write player state, but the component composition is undefined.
- **Resolution needed**: Define the player entity's component set and ensure all v1.0 systems agree on the schema.

### Genre Audio and Visual Assets for Fantasy
- **Gap**: `SetGenre("fantasy")` is wired as a stub in all packages but no concrete fantasy-themed procedural generation parameters exist (e.g., what color palette, what synthesis parameters for SFX, what waveform patterns for music tracks constitute "fantasy").
- **Impact**: Genre swap will compile and run but produce no visible or audible difference until fantasy-specific procedural generation parameters are defined.
- **Resolution needed**: Create a fantasy genre asset generation manifest specifying: fog color, palette RGB values, wall texture generation parameters, SFX synthesis parameters (waveform types, frequency ranges, envelope shapes), and music generation parameters (scale, tempo, instrument synthesis). All assets must be procedurally generated at runtime — no pre-rendered audio or image files are permitted.

### Save Data Schema
- **Gap**: `pkg/save` defines a `Slot` struct with opaque `Data []byte` but no schema for what gets serialized (level seed, player position, health, inventory, discovered map tiles).
- **Impact**: Save/load cannot be implemented without a defined serialization format.
- **Resolution needed**: Define a `SaveState` struct enumerating all fields to persist, and specify JSON field names.

### Gamepad Mapping Specification
- **Gap**: Roadmap requires "analog stick mouselook, trigger fire" but no specific gamepad button/axis mapping is documented.
- **Impact**: Gamepad support cannot be implemented without knowing which Ebitengine `GamepadAxis`/`GamepadButton` constants map to which actions.
- **Resolution needed**: Document default gamepad mappings (e.g., left stick = move, right stick = look, RT = fire, LT = alt-fire, A = interact, B = reload, Start = pause).

### Test Infrastructure
- **Gap**: Zero test files exist in the entire codebase; no test helpers, fixtures, or mocking patterns are established.
- **Impact**: Reaching 82% coverage requires test infrastructure from scratch; testing Ebitengine-dependent packages (render, ui, input) may require mock abstractions.
- **Resolution needed**: Establish testing patterns — decide whether to use standard `testing` package only or add a test framework; create mock interfaces for Ebitengine screen/input dependencies.
