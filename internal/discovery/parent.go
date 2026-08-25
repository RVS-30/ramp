package discovery
// ParentLookup resolves a process's parent ProcInfo from a
// pre-fetched set of processes, avoiding a second scan/enrich call
// per lookup. Built once per Scan() run from already-gathered data.
type ParentLookup struct {
	byPID map[int]ProcInfo
}

// NewParentLookup indexes a slice of ProcInfo by PID for O(1) parent
// lookups. procs should be the full set already enriched in this run.
func NewParentLookup(procs []ProcInfo) *ParentLookup {
	byPID := make(map[int]ProcInfo, len(procs))
	for _, p := range procs {
		byPID[p.PID] = p
	}
	return &ParentLookup{byPID: byPID}
}

// Parent returns the parent ProcInfo for proc, or nil if the parent
// wasn't part of the scanned set — this is expected, not an error:
// the parent may be a system process we successfully scanned but
// that's still fine, OR it may be PID 1 / a process that exited
// between scans / one we lack permission to enrich at all.
func (l *ParentLookup) Parent(proc ProcInfo) *ProcInfo {
	if proc.PPID <= 0 {
		return nil
	}
	if parent, ok := l.byPID[proc.PPID]; ok {
		return &parent
	}
	return nil
}