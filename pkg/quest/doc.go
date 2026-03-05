// Package quest manages level objectives and quest tracking for procedurally
// generated gameplay objectives. It supports genre-specific text variants and
// integrates with level layout systems.
//
// # Basic Usage
//
// Create a tracker and generate objectives:
//
//	tracker := quest.NewTracker()
//	tracker.SetGenre("scifi")
//	tracker.Generate(12345, 3) // seed=12345, count=3 objectives
//
// # Layout-Based Generation
//
// Generate objectives from level layout:
//
//	layout := quest.LevelLayout{
//	    Rooms: []quest.Room{
//	        {Center: quest.Position{X: 10, Y: 20}, Width: 5, Height: 5},
//	    },
//	}
//	tracker.GenerateWithLayout(67890, 2, layout)
//
// # Progress Tracking
//
// Update and complete objectives:
//
//	tracker.UpdateProgress("obj1", 5) // add 5 progress
//	tracker.Complete("obj1")          // mark complete
//	if tracker.AllComplete() {
//	    fmt.Println("All objectives done!")
//	}
package quest
