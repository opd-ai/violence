// Package quest manages level objectives and quest tracking.
package quest

// Objective represents a single quest objective.
type Objective struct {
	ID       string
	Desc     string
	Complete bool
}

// Tracker tracks active objectives.
type Tracker struct {
	Objectives []Objective
}

// NewTracker creates a new quest tracker.
func NewTracker() *Tracker {
	return &Tracker{}
}

// Complete marks an objective as completed by ID.
func (t *Tracker) Complete(id string) {
	for i := range t.Objectives {
		if t.Objectives[i].ID == id {
			t.Objectives[i].Complete = true
		}
	}
}

// SetGenre configures quest types for a genre.
func SetGenre(genreID string) {}
