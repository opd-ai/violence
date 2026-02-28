package squad

import (
	"testing"
)

func TestNewSquad(t *testing.T) {
	tests := []struct {
		name       string
		maxMembers int
		want       int
	}{
		{"Default max", 3, 3},
		{"Custom max", 5, 5},
		{"Zero defaults to 3", 0, 3},
		{"Negative defaults to 3", -1, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSquad(tt.maxMembers)
			if s == nil {
				t.Fatal("NewSquad returned nil")
			}
			if s.MaxMembers != tt.want {
				t.Errorf("MaxMembers = %d, want %d", s.MaxMembers, tt.want)
			}
			if s.Behavior != BehaviorFollow {
				t.Errorf("Behavior = %v, want %v", s.Behavior, BehaviorFollow)
			}
			if s.Formation != FormationWedge {
				t.Errorf("Formation = %v, want %v", s.Formation, FormationWedge)
			}
			if len(s.Members) != 0 {
				t.Errorf("Members = %d, want 0", len(s.Members))
			}
		})
	}
}

func TestAddMember(t *testing.T) {
	tests := []struct {
		name      string
		classID   string
		wantHP    float64
		wantSpeed float64
	}{
		{"Grunt class", "grunt", 100, 0.035},
		{"Medic class", "medic", 80, 0.04},
		{"Demo class", "demo", 90, 0.03},
		{"Mystic class", "mystic", 70, 0.038},
		{"Unknown class", "unknown", 100, 0.035},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSquad(3)
			err := s.AddMember("member1", tt.classID, "pistol", 5.0, 5.0, 12345)
			if err != nil {
				t.Errorf("AddMember returned error: %v", err)
			}
			if len(s.Members) != 1 {
				t.Fatalf("Members = %d, want 1", len(s.Members))
			}

			m := s.Members[0]
			if m.ID != "member1" {
				t.Errorf("ID = %s, want member1", m.ID)
			}
			if m.ClassID != tt.classID {
				t.Errorf("ClassID = %s, want %s", m.ClassID, tt.classID)
			}
			if m.MaxHealth != tt.wantHP {
				t.Errorf("MaxHealth = %f, want %f", m.MaxHealth, tt.wantHP)
			}
			if m.Health != tt.wantHP {
				t.Errorf("Health = %f, want %f", m.Health, tt.wantHP)
			}
			if m.Speed != tt.wantSpeed {
				t.Errorf("Speed = %f, want %f", m.Speed, tt.wantSpeed)
			}
			if m.WeaponID != "pistol" {
				t.Errorf("WeaponID = %s, want pistol", m.WeaponID)
			}
			if m.Agent == nil {
				t.Error("Agent is nil")
			}
			if m.BehaviorTree == nil {
				t.Error("BehaviorTree is nil")
			}
		})
	}
}

func TestAddMember_MaxCapacity(t *testing.T) {
	s := NewSquad(2)

	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)
	s.AddMember("m2", "medic", "pistol", 6.0, 6.0, 2)

	// Third member should be silently ignored
	err := s.AddMember("m3", "demo", "shotgun", 7.0, 7.0, 3)
	if err != nil {
		t.Errorf("AddMember returned error: %v", err)
	}

	if len(s.Members) != 2 {
		t.Errorf("Members = %d, want 2 (max capacity)", len(s.Members))
	}
}

func TestRemoveMember(t *testing.T) {
	s := NewSquad(3)
	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)
	s.AddMember("m2", "medic", "pistol", 6.0, 6.0, 2)
	s.AddMember("m3", "demo", "shotgun", 7.0, 7.0, 3)

	t.Run("Remove middle member", func(t *testing.T) {
		s.RemoveMember("m2")
		if len(s.Members) != 2 {
			t.Errorf("Members = %d, want 2", len(s.Members))
		}
		if s.Members[0].ID != "m1" || s.Members[1].ID != "m3" {
			t.Error("Wrong members remained")
		}
	})

	t.Run("Remove non-existent member", func(t *testing.T) {
		initialCount := len(s.Members)
		s.RemoveMember("nonexistent")
		if len(s.Members) != initialCount {
			t.Error("Members count changed when removing non-existent member")
		}
	})
}

func TestCommand(t *testing.T) {
	tests := []struct {
		name          string
		cmd           string
		wantBehavior  BehaviorState
		wantFormation Formation
	}{
		{"Follow command", "follow", BehaviorFollow, FormationWedge},
		{"Hold command", "hold", BehaviorHold, FormationWedge},
		{"Attack command", "attack", BehaviorAttack, FormationWedge},
		{"Formation line", "formation_line", BehaviorFollow, FormationLine},
		{"Formation wedge", "formation_wedge", BehaviorFollow, FormationWedge},
		{"Formation column", "formation_column", BehaviorFollow, FormationColumn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSquad(3)
			s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)

			s.Command(tt.cmd)

			if s.Behavior != tt.wantBehavior {
				t.Errorf("Behavior = %v, want %v", s.Behavior, tt.wantBehavior)
			}
			if tt.cmd == "formation_line" || tt.cmd == "formation_wedge" || tt.cmd == "formation_column" {
				if s.Formation != tt.wantFormation {
					t.Errorf("Formation = %v, want %v", s.Formation, tt.wantFormation)
				}
			}
		})
	}
}

func TestCommand_Hold(t *testing.T) {
	s := NewSquad(3)
	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)
	s.AddMember("m2", "medic", "pistol", 6.0, 6.0, 2)

	s.Command("hold")

	// Check that hold positions are set
	for _, m := range s.Members {
		if m.HoldX != m.X || m.HoldY != m.Y {
			t.Errorf("Hold position not set correctly: (%f, %f) vs (%f, %f)",
				m.HoldX, m.HoldY, m.X, m.Y)
		}
	}
}

func TestSetTarget(t *testing.T) {
	s := NewSquad(3)
	s.SetTarget(10.5, 20.5)

	if s.TargetX != 10.5 {
		t.Errorf("TargetX = %f, want 10.5", s.TargetX)
	}
	if s.TargetY != 20.5 {
		t.Errorf("TargetY = %f, want 20.5", s.TargetY)
	}
}

func TestUpdateFollow(t *testing.T) {
	s := NewSquad(3)
	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)

	tileMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	}

	// Leader at (8, 5), member starts at (5, 5)
	s.Update(8.0, 5.0, tileMap, 0, 0, 12345)

	member := s.Members[0]

	// Member should have moved closer to leader + formation offset
	if member.X == 5.0 && member.Y == 5.0 {
		// Check if it's already close enough to not move
		dx := (8.0 + member.FormationOffsetX) - member.X
		dy := (5.0 + member.FormationOffsetY) - member.Y
		dist := dx*dx + dy*dy
		if dist > 0.5*0.5 {
			t.Error("Member did not move when it should have")
		}
	}
}

func TestUpdateHold(t *testing.T) {
	s := NewSquad(3)
	s.AddMember("m1", "grunt", "rifle", 3.0, 3.0, 1)

	tileMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	}

	s.Command("hold")

	// Move member away from hold position
	s.Members[0].X = 5.0
	s.Members[0].Y = 3.0

	initialX := s.Members[0].X

	// Update should move member back toward hold position
	s.Update(10.0, 10.0, tileMap, 0, 0, 12345)

	member := s.Members[0]
	// Member should have moved closer to hold position (3.0, 3.0)
	if member.X >= initialX {
		t.Errorf("Member X did not move back toward hold position: %f >= %f", member.X, initialX)
	}
}

func TestUpdateAttack(t *testing.T) {
	s := NewSquad(3)
	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)

	tileMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	}

	s.Command("attack")
	s.SetTarget(8.0, 5.0)

	// Update should use behavior tree for combat
	s.Update(10.0, 10.0, tileMap, 8.0, 5.0, 12345)

	// Member should have behavior tree executed (hard to test without mocking)
	// At minimum, verify it doesn't panic
	if s.Members[0].Agent == nil {
		t.Error("Agent is nil after attack update")
	}
}

func TestFormationLine(t *testing.T) {
	s := NewSquad(3)
	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)
	s.AddMember("m2", "medic", "pistol", 6.0, 6.0, 2)
	s.AddMember("m3", "demo", "shotgun", 7.0, 7.0, 3)

	s.Command("formation_line")

	// Check that formation offsets are calculated
	for i, m := range s.Members {
		if m.FormationOffsetY != -2.0 {
			t.Errorf("Member %d: FormationOffsetY = %f, want -2.0", i, m.FormationOffsetY)
		}
		// X offsets should be spread horizontally
		if i > 0 {
			prev := s.Members[i-1]
			if m.FormationOffsetX == prev.FormationOffsetX {
				t.Error("Members have same X offset in line formation")
			}
		}
	}
}

func TestFormationWedge(t *testing.T) {
	s := NewSquad(4)
	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)
	s.AddMember("m2", "medic", "pistol", 6.0, 6.0, 2)
	s.AddMember("m3", "demo", "shotgun", 7.0, 7.0, 3)
	s.AddMember("m4", "mystic", "staff", 8.0, 8.0, 4)

	s.Command("formation_wedge")

	// Check that Y offsets increase with row
	for i, m := range s.Members {
		row := i / 2
		expectedY := -float64(row+1) * 1.5
		if m.FormationOffsetY != expectedY {
			t.Errorf("Member %d: FormationOffsetY = %f, want %f", i, m.FormationOffsetY, expectedY)
		}
	}
}

func TestFormationColumn(t *testing.T) {
	s := NewSquad(3)
	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)
	s.AddMember("m2", "medic", "pistol", 6.0, 6.0, 2)
	s.AddMember("m3", "demo", "shotgun", 7.0, 7.0, 3)

	s.Command("formation_column")

	// Check that all members are in a column (X offset = 0)
	for i, m := range s.Members {
		if m.FormationOffsetX != 0 {
			t.Errorf("Member %d: FormationOffsetX = %f, want 0", i, m.FormationOffsetX)
		}
		expectedY := -float64(i+1) * 1.5
		if m.FormationOffsetY != expectedY {
			t.Errorf("Member %d: FormationOffsetY = %f, want %f", i, m.FormationOffsetY, expectedY)
		}
	}
}

func TestGetMembers(t *testing.T) {
	s := NewSquad(3)
	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)
	s.AddMember("m2", "medic", "pistol", 6.0, 6.0, 2)

	members := s.GetMembers()
	if len(members) != 2 {
		t.Errorf("GetMembers returned %d members, want 2", len(members))
	}
	if members[0].ID != "m1" || members[1].ID != "m2" {
		t.Error("GetMembers returned wrong members")
	}
}

func TestGetBehavior(t *testing.T) {
	s := NewSquad(3)

	if s.GetBehavior() != BehaviorFollow {
		t.Errorf("GetBehavior = %v, want %v", s.GetBehavior(), BehaviorFollow)
	}

	s.Command("hold")
	if s.GetBehavior() != BehaviorHold {
		t.Errorf("GetBehavior = %v, want %v", s.GetBehavior(), BehaviorHold)
	}
}

func TestGetFormation(t *testing.T) {
	s := NewSquad(3)

	if s.GetFormation() != FormationWedge {
		t.Errorf("GetFormation = %v, want %v", s.GetFormation(), FormationWedge)
	}

	s.Command("formation_line")
	if s.GetFormation() != FormationLine {
		t.Errorf("GetFormation = %v, want %v", s.GetFormation(), FormationLine)
	}
}

func TestSetGenre(t *testing.T) {
	// SetGenre should not panic
	SetGenre("fantasy")
	SetGenre("scifi")
	SetGenre("horror")
	SetGenre("cyberpunk")
	SetGenre("postapoc")
}

func TestIsWalkable(t *testing.T) {
	tileMap := [][]int{
		{1, 1, 1},
		{1, 0, 1},
		{1, 2, 1},
		{1, 1, 1},
	}

	tests := []struct {
		name string
		x, y float64
		want bool
	}{
		{"Walkable floor", 1.5, 1.5, true},
		{"Walkable tile 2", 1.5, 2.5, true},
		{"Wall tile", 0.5, 0.5, false},
		{"Out of bounds negative", -1.0, 1.0, false},
		{"Out of bounds large", 10.0, 10.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isWalkable(tt.x, tt.y, tileMap)
			if got != tt.want {
				t.Errorf("isWalkable(%f, %f) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestIsWalkable_NilMap(t *testing.T) {
	// Should return true for nil map (allow movement)
	if !isWalkable(5.0, 5.0, nil) {
		t.Error("isWalkable should return true for nil map")
	}

	// Empty map
	if !isWalkable(5.0, 5.0, [][]int{}) {
		t.Error("isWalkable should return true for empty map")
	}
}

func TestHealthSync(t *testing.T) {
	s := NewSquad(3)
	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)

	tileMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	}

	// Damage the agent
	s.Members[0].Agent.Health = 50.0

	// Update should sync health
	s.Update(5.0, 5.0, tileMap, 0, 0, 12345)

	if s.Members[0].Health != 50.0 {
		t.Errorf("Health = %f, want 50.0 (synced from agent)", s.Members[0].Health)
	}
}

func TestAddHumanPlayer(t *testing.T) {
	s := NewSquad(3)

	s.AddHumanPlayer(101, "Player1", 10.0, 15.0)
	s.AddHumanPlayer(102, "Player2", 12.0, 18.0)

	if len(s.HumanPlayers) != 2 {
		t.Errorf("HumanPlayers = %d, want 2", len(s.HumanPlayers))
	}

	p1 := s.HumanPlayers[0]
	if p1.PlayerID != 101 || p1.Name != "Player1" {
		t.Errorf("Player1 incorrect: ID=%d Name=%s", p1.PlayerID, p1.Name)
	}
	if p1.X != 10.0 || p1.Y != 15.0 {
		t.Errorf("Player1 position incorrect: (%f, %f)", p1.X, p1.Y)
	}
	if !p1.Active {
		t.Error("Player1 should be active")
	}
}

func TestAddHumanPlayer_Reactivate(t *testing.T) {
	s := NewSquad(3)

	s.AddHumanPlayer(101, "Player1", 10.0, 15.0)
	s.RemoveHumanPlayer(101)

	if s.HumanPlayers[0].Active {
		t.Error("Player should be inactive after removal")
	}

	// Re-add should reactivate
	s.AddHumanPlayer(101, "Player1", 20.0, 25.0)

	if !s.HumanPlayers[0].Active {
		t.Error("Player should be active after re-adding")
	}
	if s.HumanPlayers[0].X != 20.0 || s.HumanPlayers[0].Y != 25.0 {
		t.Errorf("Player position not updated: (%f, %f)", s.HumanPlayers[0].X, s.HumanPlayers[0].Y)
	}
}

func TestRemoveHumanPlayer(t *testing.T) {
	s := NewSquad(3)

	s.AddHumanPlayer(101, "Player1", 10.0, 15.0)
	s.AddHumanPlayer(102, "Player2", 12.0, 18.0)

	s.RemoveHumanPlayer(101)

	if s.HumanPlayers[0].Active {
		t.Error("Player1 should be inactive")
	}
	if !s.HumanPlayers[1].Active {
		t.Error("Player2 should still be active")
	}

	// Remove non-existent player should not panic
	s.RemoveHumanPlayer(999)
}

func TestUpdateHumanPlayer(t *testing.T) {
	s := NewSquad(3)

	s.AddHumanPlayer(101, "Player1", 10.0, 15.0)

	s.UpdateHumanPlayer(101, 25.0, 30.0, 75.0)

	p := s.HumanPlayers[0]
	if p.X != 25.0 || p.Y != 30.0 {
		t.Errorf("Position not updated: (%f, %f)", p.X, p.Y)
	}
	if p.Health != 75.0 {
		t.Errorf("Health not updated: %f", p.Health)
	}

	// Update non-existent player should not panic
	s.UpdateHumanPlayer(999, 0, 0, 0)
}

func TestGetHumanPlayers(t *testing.T) {
	s := NewSquad(3)

	s.AddHumanPlayer(101, "Player1", 10.0, 15.0)
	s.AddHumanPlayer(102, "Player2", 12.0, 18.0)
	s.AddHumanPlayer(103, "Player3", 14.0, 20.0)

	s.RemoveHumanPlayer(102)

	active := s.GetHumanPlayers()

	if len(active) != 2 {
		t.Errorf("GetHumanPlayers returned %d, want 2 active players", len(active))
	}

	// Check that inactive player is not in the list
	for _, p := range active {
		if p.PlayerID == 102 {
			t.Error("Inactive player 102 should not be in active list")
		}
	}
}

func TestCommandTargetPlayer_Follow(t *testing.T) {
	s := NewSquad(3)
	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)
	s.AddMember("m2", "medic", "pistol", 6.0, 6.0, 2)

	s.AddHumanPlayer(101, "Player1", 20.0, 25.0)

	s.CommandTargetPlayer("follow_player", 101)

	if s.Behavior != BehaviorFollow {
		t.Errorf("Behavior = %v, want %v", s.Behavior, BehaviorFollow)
	}

	for i, m := range s.Members {
		if m.TargetPlayerID != 101 {
			t.Errorf("Member %d TargetPlayerID = %d, want 101", i, m.TargetPlayerID)
		}
	}
}

func TestCommandTargetPlayer_Attack(t *testing.T) {
	s := NewSquad(3)
	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)

	s.AddHumanPlayer(101, "Player1", 20.0, 25.0)

	s.CommandTargetPlayer("attack_player_target", 101)

	if s.Behavior != BehaviorAttack {
		t.Errorf("Behavior = %v, want %v", s.Behavior, BehaviorAttack)
	}

	if s.Members[0].TargetPlayerID != 101 {
		t.Errorf("TargetPlayerID = %d, want 101", s.Members[0].TargetPlayerID)
	}
}

func TestUpdateFollow_WithTargetPlayer(t *testing.T) {
	s := NewSquad(3)
	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)

	s.AddHumanPlayer(101, "Player1", 20.0, 25.0)

	tileMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	}

	s.CommandTargetPlayer("follow_player", 101)

	initialX := s.Members[0].X

	// Update should move member toward player instead of leader
	s.Update(8.0, 5.0, tileMap, 0, 0, 12345)

	// Member should have moved toward player (20, 25) not leader (8, 5)
	member := s.Members[0]

	// Check member moved in the direction of the target player
	dx := 20.0 - initialX
	if dx > 0 && member.X <= initialX {
		t.Errorf("Member should have moved toward target player at (20, 25), but stayed at (%f, %f)", member.X, member.Y)
	}
}

func TestUpdateFollow_WithInactiveTargetPlayer(t *testing.T) {
	s := NewSquad(3)
	s.AddMember("m1", "grunt", "rifle", 5.0, 5.0, 1)

	s.AddHumanPlayer(101, "Player1", 20.0, 25.0)
	s.RemoveHumanPlayer(101) // Mark as inactive

	tileMap := [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	}

	s.CommandTargetPlayer("follow_player", 101)

	// Should fall back to following leader when target player is inactive
	s.Update(8.0, 5.0, tileMap, 0, 0, 12345)

	// Member should follow leader, not the inactive player
	// Just verify it doesn't crash
	if s.Members[0].Agent == nil {
		t.Error("Agent should not be nil")
	}
}
