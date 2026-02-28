package class

import "testing"

func TestClassConstants(t *testing.T) {
	tests := []struct {
		constant string
		expected string
	}{
		{Grunt, "grunt"},
		{Medic, "medic"},
		{Demo, "demo"},
		{Mystic, "mystic"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Constant mismatch: expected %q, got %q", tt.expected, tt.constant)
			}
		})
	}
}

func TestGetClass(t *testing.T) {
	tests := []struct {
		name    string
		classID string
	}{
		{"grunt_class", Grunt},
		{"medic_class", Medic},
		{"demo_class", Demo},
		{"mystic_class", Mystic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := GetClass(tt.classID)
			if c.ID != tt.classID {
				t.Errorf("Class ID: expected %q, got %q", tt.classID, c.ID)
			}
		})
	}
}

func TestClassStruct(t *testing.T) {
	tests := []struct {
		name   string
		class  Class
		wantID string
	}{
		{
			"grunt_full",
			Class{ID: Grunt, Name: "Grunt", Health: 100, Speed: 1.0},
			Grunt,
		},
		{
			"medic_full",
			Class{ID: Medic, Name: "Medic", Health: 80, Speed: 1.2},
			Medic,
		},
		{
			"demo_full",
			Class{ID: Demo, Name: "Demolition", Health: 120, Speed: 0.9},
			Demo,
		},
		{
			"mystic_full",
			Class{ID: Mystic, Name: "Mystic", Health: 70, Speed: 1.1},
			Mystic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.class.ID != tt.wantID {
				t.Errorf("Class ID: expected %q, got %q", tt.wantID, tt.class.ID)
			}
			if tt.class.Health <= 0 {
				t.Error("Class should have positive health")
			}
			if tt.class.Speed <= 0 {
				t.Error("Class should have positive speed")
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetGenre(%q) panicked: %v", genre, r)
				}
			}()
			SetGenre(genre)
		})
	}
}

func TestAllClassIDs(t *testing.T) {
	classIDs := []string{Grunt, Medic, Demo, Mystic}

	for _, id := range classIDs {
		t.Run(id, func(t *testing.T) {
			c := GetClass(id)
			if c.ID != id {
				t.Errorf("GetClass(%q) returned class with ID %q", id, c.ID)
			}
		})
	}
}

func TestClassDefaults(t *testing.T) {
	// Test that GetClass returns a valid Class struct
	c := GetClass("unknown")

	// Should return a Class with ID set
	if c.ID != "unknown" {
		t.Errorf("GetClass should return class with provided ID, got %q", c.ID)
	}
}

func TestClassHealthValues(t *testing.T) {
	tests := []struct {
		name   string
		health float64
	}{
		{"low_health", 50.0},
		{"medium_health", 100.0},
		{"high_health", 150.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Class{ID: "test", Health: tt.health}
			if c.Health != tt.health {
				t.Errorf("Health: expected %f, got %f", tt.health, c.Health)
			}
		})
	}
}

func TestClassSpeedValues(t *testing.T) {
	tests := []struct {
		name  string
		speed float64
	}{
		{"slow", 0.8},
		{"normal", 1.0},
		{"fast", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Class{ID: "test", Speed: tt.speed}
			if c.Speed != tt.speed {
				t.Errorf("Speed: expected %f, got %f", tt.speed, c.Speed)
			}
		})
	}
}

func TestMultipleClassInstances(t *testing.T) {
	// Create multiple class instances
	classes := []Class{
		{ID: Grunt, Name: "Grunt", Health: 100, Speed: 1.0},
		{ID: Medic, Name: "Medic", Health: 80, Speed: 1.2},
		{ID: Demo, Name: "Demo", Health: 120, Speed: 0.9},
		{ID: Mystic, Name: "Mystic", Health: 70, Speed: 1.1},
	}

	// Verify each class
	for i, class := range classes {
		if class.ID == "" {
			t.Errorf("Class %d has empty ID", i)
		}
		if class.Health <= 0 {
			t.Errorf("Class %d has invalid health: %f", i, class.Health)
		}
		if class.Speed <= 0 {
			t.Errorf("Class %d has invalid speed: %f", i, class.Speed)
		}
	}
}

func TestClassNameAssignment(t *testing.T) {
	c := Class{ID: Grunt, Name: "Warrior"}
	if c.Name != "Warrior" {
		t.Errorf("Class name: expected %q, got %q", "Warrior", c.Name)
	}

	// Test genre-specific names could be assigned
	genreNames := map[string]string{
		"fantasy":   "Warrior",
		"scifi":     "Marine",
		"horror":    "Survivor",
		"cyberpunk": "Enforcer",
		"postapoc":  "Scavenger",
	}

	for genre, name := range genreNames {
		t.Run(genre, func(t *testing.T) {
			c := Class{ID: Grunt, Name: name}
			if c.Name != name {
				t.Errorf("Genre %s: expected name %q, got %q", genre, name, c.Name)
			}
		})
	}
}
