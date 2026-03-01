package test

import (
	"testing"

	"github.com/opd-ai/violence/pkg/ammo"
	"github.com/opd-ai/violence/pkg/automap"
	"github.com/opd-ai/violence/pkg/camera"
	"github.com/opd-ai/violence/pkg/class"
	"github.com/opd-ai/violence/pkg/destruct"
	"github.com/opd-ai/violence/pkg/door"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/inventory"
	"github.com/opd-ai/violence/pkg/loot"
	"github.com/opd-ai/violence/pkg/progression"
	"github.com/opd-ai/violence/pkg/quest"
	"github.com/opd-ai/violence/pkg/shop"
	"github.com/opd-ai/violence/pkg/status"
	"github.com/opd-ai/violence/pkg/tutorial"
)

// TestGenreCascade verifies that genre changes propagate to all package-level SetGenre functions.
// This test addresses AUDIT.md findings:
// - [CRITICAL BUG]: Genre SetGenre Functions Are No-ops
// - [FUNCTIONAL MISMATCH]: Main.go Genre Cascade Calls Non-functional SetGenre
func TestGenreCascade(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			// Simulate main.go's setGenre cascade
			engine.SetGenre(genreID)
			camera.SetGenre(genreID)
			tutorial.SetGenre(genreID)
			automap.SetGenre(genreID)
			door.SetGenre(genreID)
			ammo.SetGenre(genreID)
			status.SetGenre(genreID)
			loot.SetGenre(genreID)
			// progression.SetGenre is now instance method
			p := progression.NewProgression()
			if err := p.SetGenre(genreID); err != nil {
				t.Errorf("progression.SetGenre(%s) failed: %v", genreID, err)
			}
			class.SetGenre(genreID)
			inventory.SetGenre(genreID)
			quest.SetGenre(genreID)
			shop.SetGenre(genreID)
			destruct.SetGenre(genreID)

			// Verify each package received and stored the genre
			tests := []struct {
				name string
				got  string
			}{
				{"engine", engine.GetCurrentGenre()},
				{"camera", camera.GetCurrentGenre()},
				{"tutorial", tutorial.GetCurrentGenre()},
				{"automap", automap.GetCurrentGenre()},
				{"ammo", ammo.GetCurrentGenre()},
				{"status", status.GetCurrentGenre()},
				{"loot", loot.GetCurrentGenre()},
				{"progression", p.GetGenre()},
				{"class", class.GetCurrentGenre()},
				{"inventory", inventory.GetCurrentGenre()},
				{"quest", quest.GetCurrentGenre()},
				{"shop", shop.GetCurrentGenre()},
				{"destruct", destruct.GetCurrentGenre()},
			}

			for _, tt := range tests {
				if tt.got != genreID {
					t.Errorf("%s: GetCurrentGenre() = %v, want %v", tt.name, tt.got, genreID)
				}
			}
		})
	}
}
