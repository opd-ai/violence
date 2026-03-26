package edgeao_test

import (
	"fmt"

	"github.com/opd-ai/violence/pkg/edgeao"
)

// Example_basicUsage demonstrates creating an edge AO system and building an AO map.
func Example_basicUsage() {
	// Create edge AO system for horror genre (strong shadows in corners)
	sys := edgeao.NewSystem("horror", 12345)

	// Create a simple room: walls around edges, floor in center
	// 1 = wall, 0 = floor
	tiles := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}

	// Build the AO map from level geometry
	sys.BuildAOMap(tiles)

	// Query AO at different positions
	cornerAO := sys.GetAO(1.5, 1.5) // Inside corner - high AO
	centerAO := sys.GetAO(2.5, 2.5) // Room center - lower AO

	// Verify corner is darker than center
	if cornerAO > centerAO {
		fmt.Println("Corner correctly darker than center")
	}

	// Output:
	// Corner correctly darker than center
}

// Example_genreComparison shows how different genres have different AO intensities.
func Example_genreComparison() {
	tiles := [][]int{
		{1, 1, 1},
		{1, 0, 1},
		{1, 1, 1},
	}

	// Horror has strong shadows
	horrorSys := edgeao.NewSystem("horror", 42)
	horrorSys.BuildAOMap(tiles)

	// Sci-Fi has subtle edges
	scifiSys := edgeao.NewSystem("scifi", 42)
	scifiSys.BuildAOMap(tiles)

	horrorAO := horrorSys.GetAO(1.5, 1.5)
	scifiAO := scifiSys.GetAO(1.5, 1.5)

	if horrorAO > scifiAO {
		fmt.Println("Horror genre has stronger shadows")
	}

	// Output:
	// Horror genre has stronger shadows
}

// Example_applyAO demonstrates applying AO to a color.
func Example_applyAO() {
	// Start with a bright color
	r, g, b := uint8(200), uint8(180), uint8(160)

	// Apply 30% ambient occlusion (deterministic, no noise)
	darkR, darkG, darkB := edgeao.ApplyAO(r, g, b, 0.30)

	fmt.Printf("Original: RGB(%d, %d, %d)\n", r, g, b)
	fmt.Printf("After 30%% AO: RGB(%d, %d, %d)\n", darkR, darkG, darkB)

	// Output:
	// Original: RGB(200, 180, 160)
	// After 30% AO: RGB(140, 125, 112)
}
