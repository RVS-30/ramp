package discovery

import (
	"path/filepath"
	"strings"
)

// DatabaseMatch is a recognized dev database process, independent of
// whether it's running natively or inside a container (that
// association is added later by cross-referencing Docker data).
type DatabaseMatch struct {
	PID     int
	Name    string // display name, e.g. "PostgreSQL"
	Port    int
	Source  string // "Local" until Docker cross-reference says otherwise
}

// dbSignature maps a recognizable executable basename to a display
// name and the default port used to pick which listening port is
// "the" database port when a process holds several.
type dbSignature struct {
	exeNames    []string
	displayName string
	defaultPort int
}

var knownDatabases = []dbSignature{
	{exeNames: []string{"postgres", "postmaster"}, displayName: "PostgreSQL", defaultPort: 5432},
	{exeNames: []string{"redis-server"}, displayName: "Redis", defaultPort: 6379},
	{exeNames: []string{"mongod"}, displayName: "MongoDB", defaultPort: 27017},
	{exeNames: []string{"mysqld"}, displayName: "MySQL", defaultPort: 3306},
	{exeNames: []string{"mariadbd"}, displayName: "MariaDB", defaultPort: 3306},
}

// MatchDatabases scans a set of processes for known database
// executables. This runs independently of Classify — a native
// Postgres install is always shown, regardless of cwd/parent signals.
func MatchDatabases(procs []ProcInfo) []DatabaseMatch {
	var matches []DatabaseMatch

	for _, proc := range procs {
		if proc.Exe == "" || len(proc.Ports) == 0 {
			continue
		}
		base := strings.ToLower(filepath.Base(proc.Exe))

		for _, sig := range knownDatabases {
			if !containsName(sig.exeNames, base) {
				continue
			}

			port := pickDatabasePort(proc.Ports, sig.defaultPort)
			matches = append(matches, DatabaseMatch{
				PID:    proc.PID,
				Name:   sig.displayName,
				Port:   port,
				Source: "Local",
			})
			break // one signature match per process is enough
		}
	}

	return matches
}

func containsName(names []string, target string) bool {
	for _, n := range names {
		if n == target {
			return true
		}
	}
	return false
}

// pickDatabasePort prefers the conventional default port if the
// process happens to be listening on it, otherwise falls back to the
// first port we saw (covers non-standard port configs).
func pickDatabasePort(ports []int, preferred int) int {
	for _, p := range ports {
		if p == preferred {
			return preferred
		}
	}
	return ports[0]
}