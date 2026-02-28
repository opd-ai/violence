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
)

// Objective represents a single quest objective.
type Objective struct {
	ID       string
	Type     ObjectiveType
	Desc     string
	Target   string
	Count    int
	Progress int
	Complete bool
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
		objType := ObjectiveType(r.Intn(5))
		obj := t.generateObjective(r, objType, i)
		t.Objectives = append(t.Objectives, obj)
	}
}

func (t *Tracker) generateObjective(r *rng.RNG, objType ObjectiveType, idx int) Objective {
	obj := Objective{
		ID:   fmt.Sprintf("obj_%d", idx),
		Type: objType,
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
