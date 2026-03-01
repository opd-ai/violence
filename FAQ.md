# Frequently Asked Questions

## General

### What is VIOLENCE?

VIOLENCE is a raycasting first-person shooter built with Go and Ebitengine. All gameplay assets — audio, visuals, textures, and narrative — are 100% procedurally generated at runtime using deterministic algorithms. No pre-rendered or bundled assets exist in the project.

### What are the system requirements?

- **Go**: 1.24 or later
- **C Compiler**: Required for CGo (used by Ebitengine)
- **OS**: Linux (amd64/arm64), macOS (universal), Windows (amd64), or WASM in a browser
- **Graphics**: Any GPU with OpenGL 2.1+ support (Ebitengine requirement)

### How do I build and run the game?

```sh
go build -o violence .
./violence
```

Or run directly with `go run .`.

### Where is the configuration file?

Configuration is loaded from `config.toml` in the working directory, or `$HOME/.violence/config.toml`. See the file for all available options including window size, FOV, mouse sensitivity, and audio volumes.

## Performance

### The game runs slowly. How do I improve performance?

1. **Lower internal resolution**: Reduce `InternalWidth` and `InternalHeight` in `config.toml`. The default 320×200 is intentionally low for the retro aesthetic but can be adjusted.
2. **Enable VSync**: Set `VSync = true` in `config.toml` to prevent excessive frame rendering.
3. **Reduce MaxTPS**: Lower `MaxTPS` from 60 to 30 if CPU-bound.
4. **Close other applications**: Procedural audio synthesis is CPU-intensive.

### Why is audio generation causing lag spikes?

All audio is synthesized at runtime. The first time a music track or sound effect plays, there may be a brief computation period. Subsequent plays use cached samples. Ensure `MasterVolume`, `MusicVolume`, and `SFXVolume` are set appropriately — muted channels skip synthesis entirely.

### What resolution should I use?

The internal framebuffer resolution (`InternalWidth` × `InternalHeight`) controls the raycaster resolution and is upscaled to the window size. The default 320×200 provides a classic retro look. Higher values (e.g., 640×400) improve visual clarity at the cost of CPU/GPU load.

## Multiplayer

### How do I host a multiplayer server?

Run the dedicated server:

```sh
# Build and run the dedicated server
go build -o violence-server ./cmd/server
./violence-server -port 7777
```

Or use Docker:

```sh
docker build -t violence-server .
docker run -p 7777:7777 violence-server
```

See `docs/DOCKER_SERVER.md` for detailed deployment instructions.

### What multiplayer modes are available?

- **Co-op** (2–4 players): Shared level progression with independent inventories
- **Free-for-All Deathmatch** (2–8 players): Configurable frag and time limits
- **Team Deathmatch** (2–16 players): Red vs. Blue with team scoring
- **Territory Control**: Capture and hold control points for score

### What is the maximum supported latency?

| Latency Range | Experience | Behavior |
| ------------- | ---------- | -------- |
| 0–200ms       | Optimal    | Full gameplay with client-side prediction |
| 200–500ms     | Degraded   | Playable with visible interpolation artifacts |
| 500–5000ms    | Poor       | Server rejects stale inputs |
| >5000ms       | Spectator  | Forced into spectator mode with reconnect prompt |

The server stores 500ms of world snapshots (10 snapshots at 20 ticks/second) for lag compensation.

### How does co-op respawning work?

When a player dies, they enter a 10-second bleed-out timer. After the timer expires, they respawn at the nearest living teammate's position with 3 seconds of invulnerability. If all players die (party wipe), the level restarts.

### How do squads work?

Squads support up to 8 members. Any player can create a squad and invite others. Squad members share a dedicated encrypted chat channel and display a configurable 4-character tag in HUD nameplates. Aggregate stats (kills, deaths, wins, play time) are tracked per squad.

### Is chat encrypted?

Yes. All in-game chat messages are encrypted client-side using AES-256-GCM before transmission. The relay server handles only encrypted blobs and never has access to plaintext. Squad chat uses a shared squad encryption key.

### What is federation / cross-server matchmaking?

Federation allows multiple VIOLENCE servers to register with a hub. Players can:
- Browse available servers filtered by region, genre, and player count
- Look up other players across all federated servers
- Queue for matchmaking across co-op, FFA, TDM, and territory modes

## Gameplay

### What genres are available?

Five genre presets alter visuals, audio, level generation, and narrative:

| Genre         | Setting                       |
| ------------- | ----------------------------- |
| Fantasy       | Stone dungeons, torches       |
| Sci-Fi        | Metal hulls, tech terminals   |
| Horror        | Cracked plaster, dim lighting |
| Cyberpunk     | Neon glass, server racks      |
| Post-Apocalyptic | Rusted metal, debris       |

All genre differences are purely cosmetic — generation parameters change, gameplay mechanics remain consistent.

### How does the profanity filter work?

The profanity filter is client-side and enabled by default (`ProfanityFilter = true` in config). It performs case-insensitive substring matching and replaces flagged words with asterisks of equal length. The filter runs after message decryption, so the server never sees plaintext.

### How do saves work?

Save files are stored cross-platform at `$HOME/.violence/saves/`. All game state is serialized to JSON. Since all assets are procedurally generated from seeds, save files only store seeds and game state — not asset data.

## Modding

### Can I create mods?

Yes. See `docs/MODDING.md` for the full modding guide. Mods are Go plugins implementing the `Plugin` interface. They can register hooks for game events (weapon fire, enemy spawn, level generation) and custom procedural content generators.

### Where do I put mods?

Place mod directories in the `mods/` folder. Each mod must contain a `mod.json` manifest with name, version, and configuration.

## Troubleshooting

### Build fails with CGo errors

Ebitengine requires a C compiler. Install one for your platform:
- **Linux**: `sudo apt install gcc` (Debian/Ubuntu) or `sudo dnf install gcc` (Fedora)
- **macOS**: `xcode-select --install`
- **Windows**: Install [MSYS2](https://www.msys2.org/) and add `mingw64/bin` to PATH

### "No gamepad detected"

Gamepads are detected automatically on each frame. Ensure your controller is:
1. Connected before or during gameplay
2. Recognized by your OS (check with system settings)
3. Not claimed by another application

The first connected gamepad is used. Hot-plugging is supported.

### Tests fail with display/window errors

Some tests that involve rendering require a display. Run headless tests with:

```sh
# Linux with virtual framebuffer
xvfb-run go test ./...
```

Most packages use stubs to avoid display dependencies in tests.
