package discovery

import "testing"

func TestMatchDatabases(t *testing.T) {
	procs := []ProcInfo{
		{PID: 100, Exe: "/usr/local/bin/postgres", Ports: []int{5432}},
		{PID: 101, Exe: "/usr/local/bin/redis-server", Ports: []int{6379}},
		{PID: 102, Exe: "/usr/bin/node", Ports: []int{3000}}, // not a db
		{PID: 103, Exe: "/opt/mongodb/bin/mongod", Ports: []int{27017}},
		{PID: 104, Exe: "/usr/local/bin/postgres", Ports: []int{5433}}, // non-default port
		{PID: 105, Exe: "/usr/local/bin/redis-server", Ports: []int{}}, // no ports yet, skip
	}

	matches := MatchDatabases(procs)

	if len(matches) != 4 {
		t.Fatalf("got %d matches, want 4: %+v", len(matches), matches)
	}

	byPID := map[int]DatabaseMatch{}
	for _, m := range matches {
		byPID[m.PID] = m
	}

	if got := byPID[100]; got.Name != "PostgreSQL" || got.Port != 5432 {
		t.Errorf("pid 100: got %+v", got)
	}
	if got := byPID[101]; got.Name != "Redis" || got.Port != 6379 {
		t.Errorf("pid 101: got %+v", got)
	}
	if got := byPID[103]; got.Name != "MongoDB" || got.Port != 27017 {
		t.Errorf("pid 103: got %+v", got)
	}
	if got := byPID[104]; got.Name != "PostgreSQL" || got.Port != 5433 {
		t.Errorf("pid 104 (non-default port fallback): got %+v", got)
	}
	if _, ok := byPID[102]; ok {
		t.Errorf("node process should not match as a database")
	}
	if _, ok := byPID[105]; ok {
		t.Errorf("process with no ports should not match")
	}
}