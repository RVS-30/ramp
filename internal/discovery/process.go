package discovery

import "context"

// ProcInfo is the OS-agnostic snapshot of one running process,
// gathered without judgement about whether it's "dev" or "system".
type ProcInfo struct {
	PID     int
	PPID    int
	Exe     string   // resolved binary path; "" if unresolvable (permission, race)
	Cwd     string   // working directory; "" if unresolvable
	Cmdline []string
	Ports   []int    // ports this process is listening on
}

// Scanner is implemented once per OS (linux, darwin). It has no
// classification logic — it only reports raw facts about processes.
type Scanner interface {
	// ListPIDs returns every PID currently visible to this user.
	ListPIDs(ctx context.Context) ([]int, error)

	// Enrich fills in details for a single PID. Partial data with a
	// nil error is valid and expected (e.g. Cwd == "" on permission
	// denied). Returning an error is reserved for exceptional cases
	// like context cancellation or the PID vanishing mid-read.
	Enrich(ctx context.Context, pid int) (ProcInfo, error)
}

// ErrProcessVanished signals a PID that existed during ListPIDs but
// disappeared before Enrich could read it. Callers should drop the
// entry silently rather than surface this as a failure.
var ErrProcessVanished = errVanished{}

type errVanished struct{}

func (errVanished) Error() string { return "process vanished before it could be read" }