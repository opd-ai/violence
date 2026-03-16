// Package network provides match utility functions shared across match types.
package network

import (
	"sync"
	"time"
)

// PlayerState defines the interface for player state access needed by shared respawn logic.
// This consolidates duplicate respawn handling from FFAMatch and TeamMatch.
type PlayerState interface {
	GetMutex() *sync.RWMutex
	IsDead() bool
	GetRespawnTime() time.Time
	SetDead(dead bool)
	SetPosition(x, y float64)
	SetHealth(health float64)
	GetMaxHealth() float64
	ClearRespawnTime()
}

// playerStateAdapter provides PlayerState interface for FFAPlayerState.
type ffaPlayerAdapter struct {
	*FFAPlayerState
}

// GetMutex returns the player's mutex for thread-safe access.
func (p *ffaPlayerAdapter) GetMutex() *sync.RWMutex { return &p.mu }

// IsDead returns whether the player is dead.
func (p *ffaPlayerAdapter) IsDead() bool { return p.Dead }

// GetRespawnTime returns the player's respawn time.
func (p *ffaPlayerAdapter) GetRespawnTime() time.Time { return p.RespawnTime }

// SetDead sets the player's death state.
func (p *ffaPlayerAdapter) SetDead(dead bool) { p.Dead = dead }

// SetPosition sets the player's position.
func (p *ffaPlayerAdapter) SetPosition(x, y float64) { p.PosX, p.PosY = x, y }

// SetHealth sets the player's health.
func (p *ffaPlayerAdapter) SetHealth(health float64) { p.Health = health }

// GetMaxHealth returns the player's maximum health.
func (p *ffaPlayerAdapter) GetMaxHealth() float64 { return p.MaxHealth }

// ClearRespawnTime clears the player's respawn time.
func (p *ffaPlayerAdapter) ClearRespawnTime() { p.RespawnTime = time.Time{} }

// teamPlayerAdapter provides PlayerState interface for TeamPlayerState.
type teamPlayerAdapter struct {
	*TeamPlayerState
}

// GetMutex returns the player's mutex for thread-safe access.
func (p *teamPlayerAdapter) GetMutex() *sync.RWMutex { return &p.mu }

// IsDead returns whether the player is dead.
func (p *teamPlayerAdapter) IsDead() bool { return p.Dead }

// GetRespawnTime returns the player's respawn time.
func (p *teamPlayerAdapter) GetRespawnTime() time.Time { return p.RespawnTime }

// SetDead sets the player's death state.
func (p *teamPlayerAdapter) SetDead(dead bool) { p.Dead = dead }

// SetPosition sets the player's position.
func (p *teamPlayerAdapter) SetPosition(x, y float64) { p.PosX, p.PosY = x, y }

// SetHealth sets the player's health.
func (p *teamPlayerAdapter) SetHealth(health float64) { p.Health = health }

// GetMaxHealth returns the player's maximum health.
func (p *teamPlayerAdapter) GetMaxHealth() float64 { return p.MaxHealth }

// ClearRespawnTime clears the player's respawn time.
func (p *teamPlayerAdapter) ClearRespawnTime() { p.RespawnTime = time.Time{} }

// canRespawn checks if a player is ready to respawn based on their state.
// This shared helper consolidates duplicate logic from FFAMatch.ProcessRespawns
// and TeamMatch.ProcessRespawns.
func canRespawn(player PlayerState) bool {
	mu := player.GetMutex()
	mu.RLock()
	isDead := player.IsDead()
	respawnTime := player.GetRespawnTime()
	mu.RUnlock()

	return isDead && !respawnTime.IsZero() && time.Now().After(respawnTime)
}

// applyRespawn applies respawn state to a player at the given spawn point.
// This shared helper consolidates duplicate logic from FFAMatch.RespawnPlayer
// and TeamMatch.RespawnPlayer.
func applyRespawn(player PlayerState, spawn SpawnPoint) {
	mu := player.GetMutex()
	mu.Lock()
	player.SetDead(false)
	player.SetPosition(spawn.X, spawn.Y)
	player.SetHealth(player.GetMaxHealth())
	player.ClearRespawnTime()
	mu.Unlock()
}
