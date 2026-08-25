package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPickerModel_EnterSelectsHighlightedItem(t *testing.T) {
	candidates := []KillCandidate{
		{PID: 100, Port: 3000, Label: "events-web", Stack: "Next.js"},
		{PID: 200, Port: 6379, Label: "Redis", Stack: "Local"},
	}

	m := newPickerModel(candidates)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	pm := updated.(pickerModel)

	if pm.choice == nil {
		t.Fatal("expected a selection after enter, got nil")
	}

	if pm.choice.PID != 100 {
		t.Errorf("selected PID = %d, want 100 (first item)", pm.choice.PID)
	}
}

func TestPickerModel_QuitCancelsSelection(t *testing.T) {
	candidates := []KillCandidate{
		{PID: 100, Port: 3000, Label: "events-web", Stack: "Next.js"},
	}

	m := newPickerModel(candidates)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	pm := updated.(pickerModel)

	if !pm.quit {
		t.Error("expected quit=true after esc")
	}

	if pm.choice != nil {
		t.Errorf("expected no selection after quit, got %+v", pm.choice)
	}
}

func TestPickerModel_DownArrowMovesSelectionBeforeEnter(t *testing.T) {
	candidates := []KillCandidate{
		{PID: 100, Port: 3000, Label: "events-web", Stack: "Next.js"},
		{PID: 200, Port: 6379, Label: "Redis", Stack: "Local"},
	}

	m := newPickerModel(candidates)

	// Move from first item to second item.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	pm := updated.(pickerModel)

	if pm.selected != 1 {
		t.Errorf("selected index = %d, want 1", pm.selected)
	}

	// Confirm the second item.
	updated2, _ := pm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	pm2 := updated2.(pickerModel)

	if pm2.choice == nil {
		t.Fatal("expected a selection")
	}

	if pm2.choice.PID != 200 {
		t.Errorf(
			"selected PID = %d, want 200 (second item after moving down)",
			pm2.choice.PID,
		)
	}
}

func TestPickerModel_UpArrowDoesNotMovePastFirstItem(t *testing.T) {
	candidates := []KillCandidate{
		{PID: 100, Port: 3000, Label: "events-web", Stack: "Next.js"},
		{PID: 200, Port: 6379, Label: "Redis", Stack: "Local"},
	}

	m := newPickerModel(candidates)

	// Already at index 0.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	pm := updated.(pickerModel)

	if pm.selected != 0 {
		t.Errorf("selected index = %d, want 0", pm.selected)
	}
}

func TestPickerModel_DownArrowDoesNotMovePastLastItem(t *testing.T) {
	candidates := []KillCandidate{
		{PID: 100, Port: 3000, Label: "events-web", Stack: "Next.js"},
		{PID: 200, Port: 6379, Label: "Redis", Stack: "Local"},
	}

	m := newPickerModel(candidates)

	// Move to second/last item.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	pm := updated.(pickerModel)

	// Try moving past the last item.
	updated2, _ := pm.Update(tea.KeyMsg{Type: tea.KeyDown})
	pm2 := updated2.(pickerModel)

	if pm2.selected != 1 {
		t.Errorf("selected index = %d, want 1", pm2.selected)
	}
}

func TestPickerModel_EmptyCandidates(t *testing.T) {
	m := newPickerModel(nil)

	if m.selected != 0 {
		t.Errorf("selected index = %d, want 0", m.selected)
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	pm := updated.(pickerModel)

	if pm.choice != nil {
		t.Errorf("expected no choice for empty candidates, got %+v", pm.choice)
	}
}

func TestPickerModel_ViewRendersAllCandidates(t *testing.T) {
	candidates := []KillCandidate{
		{PID: 100, Port: 3000, Label: "events-web", Stack: "Next.js"},
		{PID: 200, Port: 6379, Label: "Redis", Stack: "Local"},
		{PID: 300, Port: 8080, Label: "api", Stack: "Express"},
		{PID: 400, Port: 8999, Label: "ramp", Stack: "Cobra"},
	}

	m := newPickerModel(candidates)

	view := m.View()

	expected := []string{
		"3000",
		"events-web",
		"Next.js",
		"6379",
		"Redis",
		"Local",
		"8080",
		"api",
		"Express",
		"8999",
		"ramp",
		"Cobra",
	}

	for _, value := range expected {
		if !contains(view, value) {
			t.Errorf("expected view to contain %q, got:\n%s", value, view)
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}