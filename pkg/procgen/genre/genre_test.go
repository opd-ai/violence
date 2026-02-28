package genre

import "testing"

// TestNewRegistry verifies registry initialization.
func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()
	if reg == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	if reg.genres == nil {
		t.Fatal("Registry.genres is nil")
	}
}

// TestRegisterAndGet verifies genre registration and retrieval.
func TestRegisterAndGet(t *testing.T) {
	tests := []struct {
		name  string
		genre Genre
	}{
		{
			name:  "fantasy genre",
			genre: Genre{ID: Fantasy, Name: "Fantasy"},
		},
		{
			name:  "scifi genre",
			genre: Genre{ID: SciFi, Name: "Science Fiction"},
		},
		{
			name:  "horror genre",
			genre: Genre{ID: Horror, Name: "Horror"},
		},
		{
			name:  "cyberpunk genre",
			genre: Genre{ID: Cyberpunk, Name: "Cyberpunk"},
		},
		{
			name:  "postapoc genre",
			genre: Genre{ID: PostApoc, Name: "Post-Apocalyptic"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewRegistry()

			// Register genre
			reg.Register(tt.genre)

			// Retrieve genre
			g, ok := reg.Get(tt.genre.ID)
			if !ok {
				t.Errorf("Get(%s) returned not found", tt.genre.ID)
			}
			if g.ID != tt.genre.ID {
				t.Errorf("Expected ID %s, got %s", tt.genre.ID, g.ID)
			}
			if g.Name != tt.genre.Name {
				t.Errorf("Expected Name %s, got %s", tt.genre.Name, g.Name)
			}
		})
	}
}

// TestGetNonexistent verifies retrieval of nonexistent genre.
func TestGetNonexistent(t *testing.T) {
	reg := NewRegistry()

	_, ok := reg.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) returned found, expected not found")
	}
}

// TestRegisterOverwrite verifies genre overwriting.
func TestRegisterOverwrite(t *testing.T) {
	reg := NewRegistry()

	// Register first version
	g1 := Genre{ID: Fantasy, Name: "First Fantasy"}
	reg.Register(g1)

	// Register second version with same ID
	g2 := Genre{ID: Fantasy, Name: "Second Fantasy"}
	reg.Register(g2)

	// Retrieve should get second version
	g, ok := reg.Get(Fantasy)
	if !ok {
		t.Fatal("Get(Fantasy) returned not found")
	}
	if g.Name != "Second Fantasy" {
		t.Errorf("Expected Name %s, got %s", g2.Name, g.Name)
	}
}

// TestMultipleGenres verifies multiple genre registration.
func TestMultipleGenres(t *testing.T) {
	reg := NewRegistry()

	genres := []Genre{
		{ID: Fantasy, Name: "Fantasy"},
		{ID: SciFi, Name: "Science Fiction"},
		{ID: Horror, Name: "Horror"},
	}

	// Register all
	for _, g := range genres {
		reg.Register(g)
	}

	// Retrieve all
	for _, expected := range genres {
		g, ok := reg.Get(expected.ID)
		if !ok {
			t.Errorf("Get(%s) returned not found", expected.ID)
			continue
		}
		if g.ID != expected.ID {
			t.Errorf("Expected ID %s, got %s", expected.ID, g.ID)
		}
		if g.Name != expected.Name {
			t.Errorf("Expected Name %s, got %s", expected.Name, g.Name)
		}
	}
}

// TestGenreConstants verifies genre constant values.
func TestGenreConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"Fantasy", Fantasy, "fantasy"},
		{"SciFi", SciFi, "scifi"},
		{"Horror", Horror, "horror"},
		{"Cyberpunk", Cyberpunk, "cyberpunk"},
		{"PostApoc", PostApoc, "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %s = %s, got %s", tt.name, tt.expected, tt.constant)
			}
		})
	}
}

// BenchmarkRegister benchmarks genre registration.
func BenchmarkRegister(b *testing.B) {
	reg := NewRegistry()
	g := Genre{ID: Fantasy, Name: "Fantasy"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reg.Register(g)
	}
}

// BenchmarkGet benchmarks genre retrieval.
func BenchmarkGet(b *testing.B) {
	reg := NewRegistry()
	reg.Register(Genre{ID: Fantasy, Name: "Fantasy"})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reg.Get(Fantasy)
	}
}
