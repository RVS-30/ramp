package discovery

import (
	"testing"

	"github.com/RVS-30/ramp/internal/analyser"
)

func TestProjectResolver_ResolveAndCache(t *testing.T) {
	calls := 0
	origFn := analyseProjectFn
	defer func() { analyseProjectFn = origFn }()

	analyseProjectFn = func(root string) (*analyser.ProjectInfo, error) {
		calls++
		return &analyser.ProjectInfo{Name: "events-web", Framework: "Next.js"}, nil
	}

	r := NewProjectResolver()

	info1 := r.Resolve("/home/dev/events-web")
	info2 := r.Resolve("/home/dev/events-web") // same cwd, should hit cache
	info3 := r.Resolve("/home/dev/other-app")  // different cwd, real call

	if calls != 2 {
		t.Errorf("analyseProjectFn called %d times, want 2 (cache should dedupe repeat cwd)", calls)
	}
	if info1 == nil || info1.Name != "events-web" {
		t.Errorf("info1 = %+v, want events-web", info1)
	}
	if info1 != info2 {
		t.Errorf("expected cached call to return same pointer, got different results")
	}
	if info3 == nil || info3.Name != "events-web" {
		// fake always returns same struct regardless of root; just
		// confirming it was actually invoked for the new cwd
		t.Errorf("info3 = %+v, want a resolved result", info3)
	}
}

func TestProjectResolver_EmptyCwd(t *testing.T) {
	r := NewProjectResolver()
	if got := r.Resolve(""); got != nil {
		t.Errorf("Resolve(\"\") = %+v, want nil", got)
	}
}

func TestProjectResolver_AnalyseErrorReturnsNil(t *testing.T) {
	origFn := analyseProjectFn
	defer func() { analyseProjectFn = origFn }()

	analyseProjectFn = func(root string) (*analyser.ProjectInfo, error) {
		return nil, assertErr{}
	}

	r := NewProjectResolver()
	if got := r.Resolve("/some/path"); got != nil {
		t.Errorf("Resolve on error = %+v, want nil", got)
	}
}

type assertErr struct{}

func (assertErr) Error() string { return "boom" }