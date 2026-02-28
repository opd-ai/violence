// Package quest manages level objectives and quest tracking.
package quest

import (
	"fmt"

	"github.com/opd-ai/violence/pkg/rng"
)

// ObjectiveType represents the category of objective.
type ObjectiveType int

const (
	ObjFindExit ObjectiveType = iota
	ObjKillAll
	ObjFindItem
	ObjDestroyTarget
	ObjSurvive
	ObjRetrieveItem
	ObjRescueHostage
)

// ObjectiveCategory indicates if objective is main or bonus.
type ObjectiveCategory int

const (
	CategoryMain ObjectiveCategory = iota
	CategoryBonus
)

// Objective represents a single quest objective.
type Objective struct {
	ID       string
	Type     ObjectiveType
	Category ObjectiveCategory
	Desc     string
	Target   string
	Count    int
	Progress int
	Complete bool
	PosX     float64 // Objective position in level
	PosY     float64
}

// Tracker tracks active objectives.
type Tracker struct {
	Objectives []Objective
	genreID    string
}

// NewTracker creates a new quest tracker.
func NewTracker() *Tracker {
	return &Tracker{genreID: "fantasy"}
}

// Generate creates procedural objectives from seed.
func (t *Tracker) Generate(seed uint64, count int) {
	r := rng.NewRNG(seed)
	t.Objectives = make([]Objective, 0, count)
	for i := 0; i < count; i++ {
		objType := ObjectiveType(r.Intn(7))
		obj := t.generateObjective(r, objType, i)
		t.Objectives = append(t.Objectives, obj)
	}
}

// GenerateWithLayout creates objectives with level layout integration for positioning.
func (t *Tracker) GenerateWithLayout(seed uint64, layout LevelLayout) {
	r := rng.NewRNG(seed)
	t.Objectives = make([]Objective, 0, 3)

	// Main objective (always FindExit)
	mainObj := t.generateObjective(r, ObjFindExit, 0)
	if layout.ExitPos != nil {
		mainObj.PosX = layout.ExitPos.X
		mainObj.PosY = layout.ExitPos.Y
	}
	t.Objectives = append(t.Objectives, mainObj)

	// Generate bonus objectives
	t.generateBonusObjectives(r, layout)
}

func (t *Tracker) generateBonusObjectives(r *rng.RNG, layout LevelLayout) {
	// Secret count bonus
	if layout.SecretCount > 0 {
		threshold := (layout.SecretCount + 1) / 2 // 50% of secrets
		obj := Objective{
			ID:       "bonus_secrets",
			Type:     ObjFindItem,
			Category: CategoryBonus,
			Desc:     t.genreText(fmt.Sprintf("Find %d secrets", threshold), fmt.Sprintf("Discover %d hidden areas", threshold), fmt.Sprintf("Reveal %d secret chambers", threshold), fmt.Sprintf("Access %d hidden zones", threshold), fmt.Sprintf("Uncover %d caches", threshold)),
			Target:   "secret",
			Count:    threshold,
		}
		t.Objectives = append(t.Objectives, obj)
	}

	// Kill count bonus
	killTarget := 20 + r.Intn(30)
	killObj := Objective{
		ID:       "bonus_kills",
		Type:     ObjKillAll,
		Category: CategoryBonus,
		Desc:     t.genreText(fmt.Sprintf("Slay %d foes", killTarget), fmt.Sprintf("Eliminate %d enemies", killTarget), fmt.Sprintf("Kill %d monsters", killTarget), fmt.Sprintf("Neutralize %d targets", killTarget), fmt.Sprintf("Destroy %d hostiles", killTarget)),
		Target:   "enemy",
		Count:    killTarget,
	}
	t.Objectives = append(t.Objectives, killObj)

	// Speed run bonus (in seconds)
	timeTarget := 180 + r.Intn(120) // 3-5 minutes
	speedObj := Objective{
		ID:       "bonus_speed",
		Type:     ObjSurvive,
		Category: CategoryBonus,
		Desc:     t.genreText(fmt.Sprintf("Complete in %d seconds", timeTarget), fmt.Sprintf("Finish within %d seconds", timeTarget), fmt.Sprintf("Escape in %d seconds", timeTarget), fmt.Sprintf("Beat the clock: %d seconds", timeTarget), fmt.Sprintf("Time limit: %d seconds", timeTarget)),
		Target:   "speedrun",
		Count:    timeTarget,
	}
	t.Objectives = append(t.Objectives, speedObj)
}

// LevelLayout represents level structure for objective placement.
type LevelLayout struct {
	Width       int
	Height      int
	ExitPos     *Position
	SecretCount int
	Rooms       []Room
}

// Position represents a 2D coordinate.
type Position struct {
	X float64
	Y float64
}

// Room represents a level room for objective placement.
type Room struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (t *Tracker) generateObjective(r *rng.RNG, objType ObjectiveType, idx int) Objective {
	obj := Objective{
		ID:       fmt.Sprintf("obj_%d", idx),
		Type:     objType,
		Category: CategoryMain,
	}
	switch objType {
	case ObjFindExit:
		obj.Desc = t.genreText("Find the exit", "Locate extraction point", "Escape the facility", "Reach the downlink", "Find shelter")
		obj.Target = "exit"
		obj.Count = 1
	case ObjKillAll:
		count := 5 + r.Intn(10)
		obj.Desc = t.genreText(
			fmt.Sprintf("Slay %d enemies", count),
			fmt.Sprintf("Eliminate %d hostiles", count),
			fmt.Sprintf("Kill %d creatures", count),
			fmt.Sprintf("Neutralize %d targets", count),
			fmt.Sprintf("Destroy %d mutants", count),
		)
		obj.Target = "enemy"
		obj.Count = count
	case ObjFindItem:
		obj.Desc = t.genreText("Retrieve the artifact", "Recover data core", "Find ritual tome", "Hack terminal", "Salvage supplies")
		obj.Target = "item"
		obj.Count = 1
	case ObjDestroyTarget:
		count := 2 + r.Intn(3)
		obj.Desc = t.genreText(
			fmt.Sprintf("Destroy %d altars", count),
			fmt.Sprintf("Sabotage %d generators", count),
			fmt.Sprintf("Exorcise %d shrines", count),
			fmt.Sprintf("Disable %d nodes", count),
			fmt.Sprintf("Demolish %d caches", count),
		)
		obj.Target = "destroy"
		obj.Count = count
	case ObjSurvive:
		time := 60 + r.Intn(120)
		obj.Desc = t.genreText(
			fmt.Sprintf("Survive %d seconds", time),
			fmt.Sprintf("Hold position for %d seconds", time),
			fmt.Sprintf("Endure %d seconds", time),
			fmt.Sprintf("Defend for %d seconds", time),
			fmt.Sprintf("Last %d seconds", time),
		)
		obj.Target = "time"
		obj.Count = time
	case ObjRetrieveItem:
		obj.Desc = t.genreText("Retrieve the ancient relic", "Recover the datapad", "Find the cursed book", "Hack the mainframe", "Salvage the tech")
		obj.Target = "retrieve"
		obj.Count = 1
	case ObjRescueHostage:
		count := 1 + r.Intn(3)
		obj.Desc = t.genreText(
			fmt.Sprintf("Rescue %d captives", count),
			fmt.Sprintf("Extract %d survivors", count),
			fmt.Sprintf("Free %d prisoners", count),
			fmt.Sprintf("Liberate %d hostages", count),
			fmt.Sprintf("Save %d refugees", count),
		)
		obj.Target = "hostage"
		obj.Count = count
	}
	return obj
}

func (t *Tracker) genreText(fantasy, scifi, horror, cyberpunk, postapoc string) string {
	switch t.genreID {
	case "scifi":
		return scifi
	case "horror":
		return horror
	case "cyberpunk":
		return cyberpunk
	case "postapoc":
		return postapoc
	default:
		return fantasy
	}
}

// Add adds a new objective to tracker.
func (t *Tracker) Add(obj Objective) {
	t.Objectives = append(t.Objectives, obj)
}

// UpdateProgress increments objective progress.
func (t *Tracker) UpdateProgress(id string, amount int) {
	for i := range t.Objectives {
		if t.Objectives[i].ID == id {
			t.Objectives[i].Progress += amount
			if t.Objectives[i].Progress >= t.Objectives[i].Count {
				t.Objectives[i].Complete = true
			}
		}
	}
}

// Complete marks an objective as completed by ID.
func (t *Tracker) Complete(id string) {
	for i := range t.Objectives {
		if t.Objectives[i].ID == id {
			t.Objectives[i].Complete = true
		}
	}
}

// GetActive returns all incomplete objectives.
func (t *Tracker) GetActive() []Objective {
	active := []Objective{}
	for _, obj := range t.Objectives {
		if !obj.Complete {
			active = append(active, obj)
		}
	}
	return active
}

// GetMainObjectives returns incomplete main objectives.
func (t *Tracker) GetMainObjectives() []Objective {
	main := []Objective{}
	for _, obj := range t.Objectives {
		if !obj.Complete && obj.Category == CategoryMain {
			main = append(main, obj)
		}
	}
	return main
}

// GetBonusObjectives returns incomplete bonus objectives.
func (t *Tracker) GetBonusObjectives() []Objective {
	bonus := []Objective{}
	for _, obj := range t.Objectives {
		if !obj.Complete && obj.Category == CategoryBonus {
			bonus = append(bonus, obj)
		}
	}
	return bonus
}

// AllComplete returns true if all objectives are done.
func (t *Tracker) AllComplete() bool {
	for _, obj := range t.Objectives {
		if !obj.Complete {
			return false
		}
	}
	return len(t.Objectives) > 0
}

// SetGenre configures quest types for a genre.
func SetGenre(genreID string) {}

// SetGenre on instance configures genre-specific text.
func (t *Tracker) SetGenre(genreID string) {
	t.genreID = genreID
}
