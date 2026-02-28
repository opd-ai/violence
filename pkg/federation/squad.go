// Package federation provides squad group management.
package federation

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const MaxSquadMembers = 8

var (
	ErrSquadFull      = errors.New("squad is full")
	ErrNotInSquad     = errors.New("player not in squad")
	ErrAlreadyInSquad = errors.New("player already in squad")
	ErrInvalidSquadID = errors.New("invalid squad ID")
	ErrNoInvite       = errors.New("no pending invite")
	ErrSelfInvite     = errors.New("cannot invite self")
)

// SquadMember represents a member in a squad.
type SquadMember struct {
	PlayerID   string    `json:"player_id"`
	PlayerName string    `json:"player_name"`
	JoinedAt   time.Time `json:"joined_at"`
	IsLeader   bool      `json:"is_leader"`
}

// SquadInvite represents a pending squad invitation.
type SquadInvite struct {
	SquadID  string    `json:"squad_id"`
	PlayerID string    `json:"player_id"`
	SentAt   time.Time `json:"sent_at"`
}

// Squad represents a group of up to 8 players.
type Squad struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Tag     string                 `json:"tag"`
	Members map[string]SquadMember `json:"members"`
	Invites map[string]SquadInvite `json:"invites"`
	mu      sync.RWMutex
}

// NewSquad creates a new squad with a founding member as leader.
func NewSquad(id, name, tag, founderID, founderName string) *Squad {
	s := &Squad{
		ID:      id,
		Name:    name,
		Tag:     tag,
		Members: make(map[string]SquadMember),
		Invites: make(map[string]SquadInvite),
	}
	s.Members[founderID] = SquadMember{
		PlayerID:   founderID,
		PlayerName: founderName,
		JoinedAt:   time.Now(),
		IsLeader:   true,
	}
	return s
}

// Invite sends an invitation to a player to join the squad.
func (s *Squad) Invite(playerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if inviting self
	if _, isMember := s.Members[playerID]; isMember {
		return ErrAlreadyInSquad
	}

	// Check if squad is full
	if len(s.Members) >= MaxSquadMembers {
		return ErrSquadFull
	}

	// Create invite
	s.Invites[playerID] = SquadInvite{
		SquadID:  s.ID,
		PlayerID: playerID,
		SentAt:   time.Now(),
	}

	return nil
}

// Accept allows a player to accept a pending invitation.
func (s *Squad) Accept(playerID, playerName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for pending invite
	invite, exists := s.Invites[playerID]
	if !exists {
		return ErrNoInvite
	}

	// Check if already in squad
	if _, isMember := s.Members[playerID]; isMember {
		delete(s.Invites, playerID)
		return ErrAlreadyInSquad
	}

	// Check if squad is full
	if len(s.Members) >= MaxSquadMembers {
		delete(s.Invites, playerID)
		return ErrSquadFull
	}

	// Add member
	s.Members[playerID] = SquadMember{
		PlayerID:   playerID,
		PlayerName: playerName,
		JoinedAt:   invite.SentAt,
		IsLeader:   false,
	}

	// Remove invite
	delete(s.Invites, playerID)

	return nil
}

// Leave removes a player from the squad.
func (s *Squad) Leave(playerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	member, exists := s.Members[playerID]
	if !exists {
		return ErrNotInSquad
	}

	// If leader leaves, promote another member
	if member.IsLeader && len(s.Members) > 1 {
		for id, m := range s.Members {
			if id != playerID {
				m.IsLeader = true
				s.Members[id] = m
				break
			}
		}
	}

	delete(s.Members, playerID)
	return nil
}

// GetMembers returns a copy of the members map.
func (s *Squad) GetMembers() map[string]SquadMember {
	s.mu.RLock()
	defer s.mu.RUnlock()

	members := make(map[string]SquadMember, len(s.Members))
	for k, v := range s.Members {
		members[k] = v
	}
	return members
}

// GetMemberCount returns the number of members.
func (s *Squad) GetMemberCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Members)
}

// IsMember checks if a player is in the squad.
func (s *Squad) IsMember(playerID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.Members[playerID]
	return exists
}

// IsLeader checks if a player is the squad leader.
func (s *Squad) IsLeader(playerID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	member, exists := s.Members[playerID]
	return exists && member.IsLeader
}

// GetInvites returns a copy of the invites map.
func (s *Squad) GetInvites() map[string]SquadInvite {
	s.mu.RLock()
	defer s.mu.RUnlock()

	invites := make(map[string]SquadInvite, len(s.Invites))
	for k, v := range s.Invites {
		invites[k] = v
	}
	return invites
}

// SquadManager manages multiple squads and provides persistence.
type SquadManager struct {
	squads   map[string]*Squad
	mu       sync.RWMutex
	savePath string
}

// NewSquadManager creates a new squad manager.
func NewSquadManager() *SquadManager {
	return &SquadManager{
		squads: make(map[string]*Squad),
	}
}

// CreateSquad creates and registers a new squad.
func (sm *SquadManager) CreateSquad(id, name, tag, founderID, founderName string) (*Squad, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.squads[id]; exists {
		return nil, fmt.Errorf("squad %s already exists", id)
	}

	squad := NewSquad(id, name, tag, founderID, founderName)
	sm.squads[id] = squad

	return squad, nil
}

// GetSquad retrieves a squad by ID.
func (sm *SquadManager) GetSquad(id string) (*Squad, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	squad, exists := sm.squads[id]
	if !exists {
		return nil, ErrInvalidSquadID
	}

	return squad, nil
}

// DeleteSquad removes a squad.
func (sm *SquadManager) DeleteSquad(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.squads[id]; !exists {
		return ErrInvalidSquadID
	}

	delete(sm.squads, id)
	return nil
}

// ListSquads returns all registered squads.
func (sm *SquadManager) ListSquads() []*Squad {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	squads := make([]*Squad, 0, len(sm.squads))
	for _, squad := range sm.squads {
		squads = append(squads, squad)
	}
	return squads
}

// Save persists all squads to disk.
func (sm *SquadManager) Save() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	savePath, err := sm.getSquadSavePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(sm.squads, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal squads: %w", err)
	}

	if err := os.WriteFile(savePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write squad file: %w", err)
	}

	return nil
}

// Load reads all squads from disk.
func (sm *SquadManager) Load() error {
	savePath, err := sm.getSquadSavePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return nil // No saved squads yet
	}

	data, err := os.ReadFile(savePath)
	if err != nil {
		return fmt.Errorf("failed to read squad file: %w", err)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	squads := make(map[string]*Squad)
	if err := json.Unmarshal(data, &squads); err != nil {
		return fmt.Errorf("failed to unmarshal squads: %w", err)
	}

	sm.squads = squads
	return nil
}

// getSquadSavePath returns the platform-specific squad storage path.
func (sm *SquadManager) getSquadSavePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	savePath := filepath.Join(home, ".violence", "squads")
	if err := os.MkdirAll(savePath, 0o755); err != nil {
		return "", fmt.Errorf("failed to create squad directory: %w", err)
	}

	return filepath.Join(savePath, "squads.json"), nil
}
