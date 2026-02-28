package tutorial

import "testing"

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name     string
		genreID  string
		expected string
	}{
		{"fantasy", "fantasy", "fantasy"},
		{"scifi", "scifi", "scifi"},
		{"horror", "horror", "horror"},
		{"cyberpunk", "cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetGenre(tt.genreID)
			if got := GetCurrentGenre(); got != tt.expected {
				t.Errorf("GetCurrentGenre() = %v, want %v", got, tt.expected)
			}
		})
	}
}
