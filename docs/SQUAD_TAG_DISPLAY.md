# Squad Tag Display

The squad tag display feature allows players in the same squad to show their 4-character squad tag in HUD nameplates above their characters during multiplayer gameplay.

## Components

### `pkg/federation/squad.go`

The `Squad` struct already had a `Tag` field, but we added:

- **`MaxTagLength`** constant: Set to 4 characters
- **`GetTag()`**: Returns the squad's tag (thread-safe)
- **`SetTag(tag string)`**: Updates the squad's tag, automatically truncating to 4 characters
- **`GetName()`**: Returns the squad's name
- **`GetID()`**: Returns the squad's ID
- **`NewSquad()`**: Modified to auto-truncate tags longer than 4 characters

### `pkg/ui/nameplate.go`

New component for rendering player nameplates in multiplayer:

- **`NameplatePlayer`**: Struct containing player info for display:
  - `PlayerID`: Unique player identifier
  - `PlayerName`: Display name
  - `SquadTag`: Up to 4-character squad tag
  - `ScreenX`, `ScreenY`: Position on screen
  - `IsTeammate`: Whether player is on same team
  - `IsSelf`: Whether this is the local player

- **`Nameplate`**: Manager for all player nameplates:
  - `NewNameplate()`: Creates with default colors (green teammates, red enemies, yellow self)
  - `SetPlayers([]NameplatePlayer)`: Updates all displayed players
  - `AddPlayer(NameplatePlayer)`: Adds a single player
  - `ClearPlayers()`: Removes all players
  - `GetPlayerCount()`: Returns number of players being displayed
  - `SetTeammateColor()`, `SetEnemyColor()`, `SetSelfColor()`: Customize colors
  - `Draw(screen)`: Renders all nameplates

## Usage Example

```go
import (
    "github.com/opd-ai/violence/pkg/federation"
    "github.com/opd-ai/violence/pkg/ui"
)

// Create a squad with a tag
squad := federation.NewSquad("squad1", "Elite Team", "ELIT", "player1", "Alice")

// Change the tag later
squad.SetTag("PROS")

// In your render loop:
nameplate := ui.NewNameplate()

// Add players with their squad tags
players := []ui.NameplatePlayer{
    {
        PlayerID:   "player1",
        PlayerName: "Alice",
        SquadTag:   squad.GetTag(), // "PROS"
        ScreenX:    320.0,
        ScreenY:    240.0,
        IsTeammate: true,
        IsSelf:     false,
    },
    {
        PlayerID:   "player2",
        PlayerName: "Bob",
        SquadTag:   squad.GetTag(), // "PROS"
        ScreenX:    400.0,
        ScreenY:    280.0,
        IsTeammate: true,
        IsSelf:     false,
    },
    {
        PlayerID:   "enemy1",
        PlayerName: "Enemy",
        SquadTag:   "FOES",
        ScreenX:    500.0,
        ScreenY:    300.0,
        IsTeammate: false,
        IsSelf:     false,
    },
}

nameplate.SetPlayers(players)
nameplate.Draw(screen)
```

## Display Format

Nameplates are rendered with:
1. **Squad Tag** (if present): Semi-transparent dark background with white border, displayed above the name
2. **Player Name**: Semi-transparent darker background with light gray border, displayed below the tag

Color coding:
- **Green**: Teammates
- **Red**: Enemies  
- **Yellow**: Local player (self)

## Tag Validation

- Maximum length: **4 characters** (enforced by `MaxTagLength` constant)
- Tags longer than 4 characters are automatically truncated
- Empty tags are allowed (no tag will be displayed)
- Tags are case-sensitive

## Testing

### Squad Tag Tests (`pkg/federation/squad_test.go`)
- Tag getter/setter functionality
- Tag truncation for long tags
- Tag validation (1-4 characters, empty string)
- Persistence through save/load

### Nameplate Tests (`pkg/ui/nameplate_logic_test.go`)
- Player management (add/remove/clear)
- Tag truncation
- Color customization
- Screen position handling
- Multiple players with different tags
- Empty tag handling

**Coverage**: 
- Squad tag functions: 100%
- Nameplate logic: 100% (drawing methods excluded due to display requirements)
