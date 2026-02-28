package quest

import (
	"testing"
)

func TestNewTracker(t *testing.T) {
	tracker := NewTracker()
	if tracker == nil {
		t.Fatal("NewTracker returned nil")
	}
	if len(tracker.Objectives) != 0 {
		t.Errorf("expected empty objectives, got %d", len(tracker.Objectives))
	}
}

func TestTracker_Generate(t *testing.T) {
	tracker := NewTracker()
	tracker.Generate(12345, 5)
	if len(tracker.Objectives) != 5 {
		t.Errorf("expected 5 objectives, got %d", len(tracker.Objectives))
	}
	for i, obj := range tracker.Objectives {
		if obj.ID == "" {
			t.Errorf("objective %d has empty ID", i)
		}
		if obj.Desc == "" {
			t.Errorf("objective %d has empty description", i)
		}
		if obj.Complete {
			t.Errorf("objective %d should not be complete initially", i)
		}
	}
}

func TestTracker_GenerateDeterministic(t *testing.T) {
	t1 := NewTracker()
	t1.Generate(999, 3)
	t2 := NewTracker()
	t2.Generate(999, 3)
	if len(t1.Objectives) != len(t2.Objectives) {
		t.Errorf("expected same count, got %d vs %d", len(t1.Objectives), len(t2.Objectives))
	}
	for i := range t1.Objectives {
		if t1.Objectives[i].Type != t2.Objectives[i].Type {
			t.Errorf("objective %d type mismatch", i)
		}
		if t1.Objectives[i].Desc != t2.Objectives[i].Desc {
			t.Errorf("objective %d desc mismatch", i)
		}
	}
}

func TestTracker_Add(t *testing.T) {
	tracker := NewTracker()
	obj := Objective{ID: "test1", Desc: "Test objective", Count: 1}
	tracker.Add(obj)
	if len(tracker.Objectives) != 1 {
		t.Errorf("expected 1 objective, got %d", len(tracker.Objectives))
	}
	if tracker.Objectives[0].ID != "test1" {
		t.Errorf("expected ID test1, got %s", tracker.Objectives[0].ID)
	}
}

func TestTracker_Complete(t *testing.T) {
	tracker := NewTracker()
	tracker.Add(Objective{ID: "obj1", Desc: "Test"})
	tracker.Complete("obj1")
	if !tracker.Objectives[0].Complete {
		t.Error("objective should be complete")
	}
}

func TestTracker_UpdateProgress(t *testing.T) {
	tracker := NewTracker()
	tracker.Add(Objective{ID: "kill10", Count: 10})
	tracker.UpdateProgress("kill10", 5)
	if tracker.Objectives[0].Progress != 5 {
		t.Errorf("expected progress 5, got %d", tracker.Objectives[0].Progress)
	}
	if tracker.Objectives[0].Complete {
		t.Error("objective should not be complete yet")
	}
	tracker.UpdateProgress("kill10", 5)
	if tracker.Objectives[0].Progress != 10 {
		t.Errorf("expected progress 10, got %d", tracker.Objectives[0].Progress)
	}
	if !tracker.Objectives[0].Complete {
		t.Error("objective should be complete")
	}
}

func TestTracker_UpdateProgressOverflow(t *testing.T) {
	tracker := NewTracker()
	tracker.Add(Objective{ID: "kill5", Count: 5})
	tracker.UpdateProgress("kill5", 10)
	if tracker.Objectives[0].Progress != 10 {
		t.Errorf("expected progress 10, got %d", tracker.Objectives[0].Progress)
	}
	if !tracker.Objectives[0].Complete {
		t.Error("objective should be complete with overflow")
	}
}

func TestTracker_GetActive(t *testing.T) {
	tracker := NewTracker()
	tracker.Add(Objective{ID: "obj1", Desc: "Active"})
	tracker.Add(Objective{ID: "obj2", Desc: "Done", Complete: true})
	tracker.Add(Objective{ID: "obj3", Desc: "Active"})
	active := tracker.GetActive()
	if len(active) != 2 {
		t.Errorf("expected 2 active objectives, got %d", len(active))
	}
	for _, obj := range active {
		if obj.Complete {
			t.Error("GetActive returned completed objective")
		}
	}
}

func TestTracker_AllComplete(t *testing.T) {
	tracker := NewTracker()
	if tracker.AllComplete() {
		t.Error("empty tracker should not be all complete")
	}
	tracker.Add(Objective{ID: "obj1"})
	tracker.Add(Objective{ID: "obj2"})
	if tracker.AllComplete() {
		t.Error("incomplete objectives should not be all complete")
	}
	tracker.Complete("obj1")
	if tracker.AllComplete() {
		t.Error("partial completion should not be all complete")
	}
	tracker.Complete("obj2")
	if !tracker.AllComplete() {
		t.Error("all objectives complete, should return true")
	}
}

func TestTracker_SetGenre(t *testing.T) {
	tests := []struct {
		genre    string
		expected map[ObjectiveType]string
	}{
		{"fantasy", map[ObjectiveType]string{ObjFindExit: "Find the exit"}},
		{"scifi", map[ObjectiveType]string{ObjFindExit: "Locate extraction point"}},
		{"horror", map[ObjectiveType]string{ObjFindExit: "Escape the facility"}},
		{"cyberpunk", map[ObjectiveType]string{ObjFindExit: "Reach the downlink"}},
		{"postapoc", map[ObjectiveType]string{ObjFindExit: "Find shelter"}},
	}
	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			tracker := NewTracker()
			tracker.SetGenre(tt.genre)
			tracker.Generate(42, 10)
			// Find an ObjFindExit objective
			found := false
			for _, obj := range tracker.Objectives {
				if obj.Type == ObjFindExit {
					found = true
					expected := tt.expected[ObjFindExit]
					if obj.Desc != expected {
						t.Errorf("genre %s: expected %q, got %q", tt.genre, expected, obj.Desc)
					}
				}
			}
			if !found {
				// Generate more until we find one
				tracker.Generate(123, 20)
				for _, obj := range tracker.Objectives {
					if obj.Type == ObjFindExit {
						found = true
						break
					}
				}
			}
		})
	}
}

func TestTracker_ObjectiveTypes(t *testing.T) {
	tracker := NewTracker()
	tracker.Generate(555, 70) // Generate many to hit all types
	types := make(map[ObjectiveType]bool)
	for _, obj := range tracker.Objectives {
		types[obj.Type] = true
		// Validate each type
		switch obj.Type {
		case ObjFindExit:
			if obj.Target != "exit" || obj.Count != 1 {
				t.Errorf("ObjFindExit has wrong target/count")
			}
		case ObjKillAll:
			if obj.Target != "enemy" || obj.Count < 5 || obj.Count >= 15 {
				t.Errorf("ObjKillAll has wrong target/count")
			}
		case ObjFindItem:
			if obj.Target != "item" || obj.Count != 1 {
				t.Errorf("ObjFindItem has wrong target/count")
			}
		case ObjDestroyTarget:
			if obj.Target != "destroy" || obj.Count < 2 || obj.Count >= 5 {
				t.Errorf("ObjDestroyTarget has wrong target/count")
			}
		case ObjSurvive:
			if obj.Target != "time" || obj.Count < 60 || obj.Count >= 180 {
				t.Errorf("ObjSurvive has wrong target/count")
			}
		case ObjRetrieveItem:
			if obj.Target != "retrieve" || obj.Count != 1 {
				t.Errorf("ObjRetrieveItem has wrong target/count")
			}
		case ObjRescueHostage:
			if obj.Target != "hostage" || obj.Count < 1 || obj.Count >= 4 {
				t.Errorf("ObjRescueHostage has wrong target/count: %d", obj.Count)
			}
		}
	}
	if len(types) < 7 {
		t.Errorf("expected all 7 objective types represented, got %d", len(types))
	}
}

func TestTracker_GenerateWithLayout(t *testing.T) {
	layout := LevelLayout{
		Width:       100,
		Height:      100,
		ExitPos:     &Position{X: 90, Y: 90},
		SecretCount: 5,
		Rooms:       []Room{{X: 10, Y: 10, Width: 20, Height: 20}},
	}
	tracker := NewTracker()
	tracker.GenerateWithLayout(12345, layout)

	// Should have at least 4 objectives (1 main + 3 bonus)
	if len(tracker.Objectives) < 4 {
		t.Errorf("expected at least 4 objectives, got %d", len(tracker.Objectives))
	}

	// First objective should be main exit objective
	if tracker.Objectives[0].Type != ObjFindExit {
		t.Errorf("first objective should be FindExit, got %v", tracker.Objectives[0].Type)
	}
	if tracker.Objectives[0].Category != CategoryMain {
		t.Error("first objective should be main category")
	}
	if tracker.Objectives[0].PosX != 90 || tracker.Objectives[0].PosY != 90 {
		t.Errorf("exit position not set correctly: got (%f, %f)", tracker.Objectives[0].PosX, tracker.Objectives[0].PosY)
	}

	// Should have bonus objectives
	bonusCount := 0
	for _, obj := range tracker.Objectives {
		if obj.Category == CategoryBonus {
			bonusCount++
		}
	}
	if bonusCount != 3 {
		t.Errorf("expected 3 bonus objectives, got %d", bonusCount)
	}
}

func TestTracker_GetMainObjectives(t *testing.T) {
	tracker := NewTracker()
	tracker.Add(Objective{ID: "main1", Category: CategoryMain})
	tracker.Add(Objective{ID: "bonus1", Category: CategoryBonus})
	tracker.Add(Objective{ID: "main2", Category: CategoryMain})

	main := tracker.GetMainObjectives()
	if len(main) != 2 {
		t.Errorf("expected 2 main objectives, got %d", len(main))
	}
	for _, obj := range main {
		if obj.Category != CategoryMain {
			t.Error("GetMainObjectives returned non-main objective")
		}
	}
}

func TestTracker_GetBonusObjectives(t *testing.T) {
	tracker := NewTracker()
	tracker.Add(Objective{ID: "main1", Category: CategoryMain})
	tracker.Add(Objective{ID: "bonus1", Category: CategoryBonus})
	tracker.Add(Objective{ID: "bonus2", Category: CategoryBonus, Complete: true})

	bonus := tracker.GetBonusObjectives()
	if len(bonus) != 1 {
		t.Errorf("expected 1 active bonus objective, got %d", len(bonus))
	}
	for _, obj := range bonus {
		if obj.Category != CategoryBonus {
			t.Error("GetBonusObjectives returned non-bonus objective")
		}
		if obj.Complete {
			t.Error("GetBonusObjectives returned completed objective")
		}
	}
}

func TestTracker_BonusObjectiveTypes(t *testing.T) {
	layout := LevelLayout{
		Width:       50,
		Height:      50,
		ExitPos:     &Position{X: 40, Y: 40},
		SecretCount: 10,
	}
	tracker := NewTracker()
	tracker.GenerateWithLayout(999, layout)

	bonus := tracker.GetBonusObjectives()

	// Should have 3 bonus types: secrets, kills, speedrun
	bonusTypes := make(map[string]bool)
	for _, obj := range bonus {
		bonusTypes[obj.ID] = true
	}

	if !bonusTypes["bonus_secrets"] {
		t.Error("missing bonus_secrets objective")
	}
	if !bonusTypes["bonus_kills"] {
		t.Error("missing bonus_kills objective")
	}
	if !bonusTypes["bonus_speed"] {
		t.Error("missing bonus_speed objective")
	}
}

func TestTracker_GenerateWithLayoutDeterministic(t *testing.T) {
	layout := LevelLayout{
		Width:       100,
		Height:      100,
		ExitPos:     &Position{X: 50, Y: 50},
		SecretCount: 3,
	}

	t1 := NewTracker()
	t1.GenerateWithLayout(777, layout)

	t2 := NewTracker()
	t2.GenerateWithLayout(777, layout)

	if len(t1.Objectives) != len(t2.Objectives) {
		t.Errorf("expected same count, got %d vs %d", len(t1.Objectives), len(t2.Objectives))
	}

	for i := range t1.Objectives {
		if t1.Objectives[i].ID != t2.Objectives[i].ID {
			t.Errorf("objective %d ID mismatch", i)
		}
		if t1.Objectives[i].Desc != t2.Objectives[i].Desc {
			t.Errorf("objective %d desc mismatch", i)
		}
		if t1.Objectives[i].Count != t2.Objectives[i].Count {
			t.Errorf("objective %d count mismatch", i)
		}
	}
}

func TestObjectiveCategory(t *testing.T) {
	tests := []struct {
		category ObjectiveCategory
		name     string
	}{
		{CategoryMain, "main"},
		{CategoryBonus, "bonus"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := Objective{Category: tt.category}
			if obj.Category != tt.category {
				t.Errorf("expected category %v, got %v", tt.category, obj.Category)
			}
		})
	}
}

func TestTracker_UpdateProgressBonusObjective(t *testing.T) {
	tracker := NewTracker()
	tracker.Add(Objective{ID: "bonus_kills", Category: CategoryBonus, Count: 50})

	tracker.UpdateProgress("bonus_kills", 25)
	if tracker.Objectives[0].Progress != 25 {
		t.Errorf("expected progress 25, got %d", tracker.Objectives[0].Progress)
	}

	tracker.UpdateProgress("bonus_kills", 25)
	if !tracker.Objectives[0].Complete {
		t.Error("bonus objective should be complete")
	}
}

func TestLevelLayout(t *testing.T) {
	layout := LevelLayout{
		Width:       200,
		Height:      150,
		ExitPos:     &Position{X: 100, Y: 75},
		SecretCount: 8,
		Rooms: []Room{
			{X: 10, Y: 10, Width: 30, Height: 20},
			{X: 50, Y: 50, Width: 40, Height: 30},
		},
	}

	if layout.Width != 200 {
		t.Errorf("expected width 200, got %d", layout.Width)
	}
	if layout.ExitPos.X != 100 {
		t.Errorf("expected exit X 100, got %f", layout.ExitPos.X)
	}
	if len(layout.Rooms) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(layout.Rooms))
	}
}

func TestUpdateProgressInt64(t *testing.T) {
	tracker := NewTracker()
	tracker.Add(Objective{
		ID:       "test_overflow",
		Type:     ObjKillAll,
		Category: CategoryMain,
		Desc:     "Test large numbers",
		Target:   "enemy",
		Count:    1000000000,
		Progress: 0,
	})

	// Simulate many updates (would overflow int32 at 2.1 billion)
	for i := 0; i < 1000; i++ {
		tracker.UpdateProgress("test_overflow", 1000000)
	}

	// Progress should be tracked correctly with int64
	obj := tracker.Objectives[0]
	expected := int64(1000000000)
	if obj.Progress != expected {
		t.Errorf("Progress = %d, want %d", obj.Progress, expected)
	}
	if !obj.Complete {
		t.Error("Objective should be marked complete")
	}
}

func TestUpdateProgressLargeValue(t *testing.T) {
	tracker := NewTracker()
	tracker.Add(Objective{
		ID:       "test_large",
		Type:     ObjKillAll,
		Category: CategoryBonus,
		Desc:     "Very large target",
		Target:   "enemy",
		Count:    100000000,
		Progress: 0,
	})

	// Add a very large amount at once
	tracker.UpdateProgress("test_large", 100000000)

	obj := tracker.Objectives[0]
	if obj.Progress != 100000000 {
		t.Errorf("Progress = %d, want 100000000", obj.Progress)
	}
	if !obj.Complete {
		t.Error("Objective should be complete")
	}
}
