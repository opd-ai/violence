# Controls Reference

Complete keyboard, mouse, and gamepad control mappings for VIOLENCE.

## Keyboard + Mouse

### Movement

| Action         | Default Key | Description                     |
| -------------- | ----------- | ------------------------------- |
| Move Forward   | `W`         | Walk forward                    |
| Move Backward  | `S`         | Walk backward                   |
| Strafe Left    | `A`         | Sidestep left                   |
| Strafe Right   | `D`         | Sidestep right                  |
| Turn Left      | `←`         | Rotate camera left (keyboard)   |
| Turn Right     | `→`         | Rotate camera right (keyboard)  |

### Combat

| Action         | Default Key | Description                     |
| -------------- | ----------- | ------------------------------- |
| Fire           | `Space`     | Fire current weapon             |
| Weapon 1       | `1`         | Select weapon slot 1            |
| Weapon 2       | `2`         | Select weapon slot 2            |
| Weapon 3       | `3`         | Select weapon slot 3            |
| Weapon 4       | `4`         | Select weapon slot 4            |
| Weapon 5       | `5`         | Select weapon slot 5            |
| Next Weapon    | `Q`         | Cycle to next weapon            |
| Previous Weapon| `Z`         | Cycle to previous weapon        |

### Interaction

| Action         | Default Key | Description                     |
| -------------- | ----------- | ------------------------------- |
| Interact       | `E`         | Use doors, pick up items, talk  |
| Automap        | `Tab`       | Toggle fog-of-war automap       |
| Pause          | `Escape`    | Pause game / open menu          |

### Menus

| Action         | Default Key | Description                     |
| -------------- | ----------- | ------------------------------- |
| Crafting       | `C`         | Open crafting menu              |
| Shop           | `B`         | Open between-level armory shop  |
| Skills         | `K`         | Open skill and talent tree      |
| Multiplayer    | `N`         | Open multiplayer menu           |

### Mouse

| Input          | Function                                |
| -------------- | --------------------------------------- |
| Mouse Movement | Look / aim (horizontal and vertical)    |
| Mouse Sensitivity | Configurable in `config.toml` (`MouseSensitivity`) |

Camera pitch is clamped to ±30 degrees.

## Gamepad

### Buttons

| Action         | Xbox Controller | PlayStation Controller | Description              |
| -------------- | --------------- | ---------------------- | ------------------------ |
| Fire           | A               | Cross (×)              | Fire current weapon      |
| Interact       | B               | Circle (○)             | Use doors, pick up items |
| Automap        | X               | Square (□)             | Toggle automap           |
| Pause          | Start           | Options                | Pause game / open menu   |
| Next Weapon    | LB              | L1                     | Cycle to next weapon     |
| Previous Weapon| RB              | R1                     | Cycle to previous weapon |

### Analog Sticks

| Input          | Function                                |
| -------------- | --------------------------------------- |
| Left Stick X   | Strafe left/right                       |
| Left Stick Y   | Move forward/backward                   |
| Right Stick X  | Turn left/right (look horizontal)       |
| Right Stick Y  | Look up/down (pitch)                    |

Analog stick values range from -1.0 to 1.0. The first connected gamepad is used automatically.

## Chat Controls

When the chat overlay is visible:

| Action         | Key         | Description                     |
| -------------- | ----------- | ------------------------------- |
| Type Message   | Any key     | Adds to input buffer            |
| Send Message   | `Enter`     | Send typed message              |
| Scroll Up      | `Page Up`   | Scroll message history up       |
| Scroll Down    | `Page Down` | Scroll message history down     |

## Customizing Bindings

Key bindings can be overridden in `config.toml` using the `[KeyBindings]` section. Each action maps to an Ebitengine key code (integer):

```toml
[KeyBindings]
move_forward = 87   # W
move_backward = 83  # S
strafe_left = 65    # A
strafe_right = 68   # D
fire = 32           # Space
interact = 69       # E
```

See the [Ebitengine key documentation](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#Key) for key code values.

## Configuration

Audio and display settings are configured in `config.toml`:

| Setting            | Default | Description                      |
| ------------------ | ------- | -------------------------------- |
| `MouseSensitivity` | 1.0     | Mouse look speed multiplier      |
| `FOV`              | 66.0    | Field of view in degrees         |
| `WindowWidth`      | 1280    | Window width in pixels           |
| `WindowHeight`     | 768     | Window height in pixels          |
| `VSync`            | true    | Vertical sync                    |
| `FullScreen`       | false   | Fullscreen mode                  |
| `MaxTPS`           | 60      | Maximum ticks per second         |
