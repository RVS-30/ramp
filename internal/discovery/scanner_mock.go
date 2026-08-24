package discovery

import "context"

// mockScanner is a test double for Scanner. Not OS-specific, used by
// every test in this package so classify/database/discovery logic
// can be verified with zero real syscalls.
type mockScanner struct {
	pids   []int
	procs  map[int]ProcInfo
	errs   map[int]error
}

func (m *mockScanner) ListPIDs(ctx context.Context) ([]int, error) {
	return m.pids, nil
}

func (m *mockScanner) Enrich(ctx context.Context, pid int) (ProcInfo, error) {
	if err, ok := m.errs[pid]; ok {
		return ProcInfo{}, err
	}
	return m.procs[pid], nil
}