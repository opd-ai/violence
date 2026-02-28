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
	tracker.Generate(555, 50) // Generate many to hit all types
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
		}
	}
	if len(types) < 5 {
		t.Errorf("expected all 5 objective types represented, got %d", len(types))
	}
}
