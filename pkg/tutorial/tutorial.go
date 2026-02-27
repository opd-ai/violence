// Package tutorial provides in-game tutorial prompts.
package tutorial

// Tutorial manages tutorial prompt state.
type Tutorial struct {
	Active  bool
	Current string
}

// NewTutorial creates a new tutorial manager.
func NewTutorial() *Tutorial {
	return &Tutorial{}
}

// ShowPrompt displays a tutorial prompt with the given message.
func (t *Tutorial) ShowPrompt(message string) {
	t.Active = true
	t.Current = message
}

// Dismiss hides the current tutorial prompt.
func (t *Tutorial) Dismiss() {
	t.Active = false
	t.Current = ""
}

// SetGenre configures tutorial content for a genre.
func SetGenre(genreID string) {}
