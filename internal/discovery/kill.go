package discovery

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// KillResult describes the outcome of a terminate attempt in terms a
// CLI user should see — not raw errno values.
type KillResult struct {
	PID     int
	Killed  bool
	Message string
}

// Terminate sends SIGTERM to pid and translates the outcome into a
// user-facing KillResult. It never returns a raw syscall error to the
// caller — every expected failure mode (already dead, no permission)
// is converted into a clear message, per the edge cases scoped back
// when we designed this feature.
func Terminate(pid int) KillResult {
	proc, err := os.FindProcess(pid)
	if err != nil {
		// On Unix, FindProcess essentially never fails on its own —
		// real failures surface from Signal below — but handle it
		// defensively rather than assume.
		return KillResult{PID: pid, Killed: false, Message: fmt.Sprintf("could not locate process: %v", err)}
	}

	err = proc.Signal(syscall.SIGTERM)
	if err == nil {
		return KillResult{PID: pid, Killed: true, Message: "terminated"}
	}

	switch {
	case errors.Is(err, syscall.ESRCH), errors.Is(err, os.ErrProcessDone):
		return KillResult{PID: pid, Killed: false, Message: "process no longer running"}
	case errors.Is(err, syscall.EPERM):
		return KillResult{PID: pid, Killed: false, Message: "permission denied — you don't own this process"}
	default:
		return KillResult{PID: pid, Killed: false, Message: fmt.Sprintf("failed to terminate: %v", err)}
	}
}
