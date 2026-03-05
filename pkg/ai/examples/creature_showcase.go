//go:build ignore
// +build ignore

// This program generates example creature sprites to demonstrate body-plan variety.
// Run with: go run pkg/ai/examples/creature_showcase.go
package main

import (
	"fmt"
	"image/png"
	"os"

	"github.com/opd-ai/violence/pkg/ai"
)

func main() {
	creatures := []struct {
		name  string
		ctype ai.CreatureType
	}{
		{"Wolf", ai.CreatureWolf},
		{"Bear", ai.CreatureBear},
		{"Spider", ai.CreatureSpider},
		{"Mantis", ai.CreatureMantis},
		{"Snake", ai.CreatureSnake},
		{"Serpent", ai.CreatureSerpent},
		{"Bat", ai.CreatureBat},
		{"Drake", ai.CreatureDrake},
		{"Slime", ai.CreatureSlime},
		{"Elemental", ai.CreatureElemental},
	}

	fmt.Println("Generating creature sprites...")

	for i, c := range creatures {
		seed := int64(12345 + i)
		img := ai.GenerateCreatureSprite(seed, c.ctype, ai.AnimFrameIdle)

		filename := fmt.Sprintf("/tmp/creature_%s.png", c.name)
		f, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Failed to create %s: %v\n", filename, err)
			continue
		}

		if err := png.Encode(f, img); err != nil {
			fmt.Printf("Failed to encode %s: %v\n", filename, err)
			f.Close()
			continue
		}
		f.Close()

		fmt.Printf("Generated %s: %s (body plan: %v)\n", c.name, filename, ai.GetBodyPlan(c.ctype))
	}

	fmt.Println("\nDone! Sprite files saved to /tmp/")
}
