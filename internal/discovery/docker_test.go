package discovery

import "testing"

func TestParseContainers(t *testing.T) {
	raw := []dockerContainerJSON{
		{
			ID:    "abc123",
			Names: []string{"/seedicon-postgres"},
			Image: "postgres:16",
			Ports: []struct {
				PublicPort int `json:"PublicPort"`
			}{{PublicPort: 5432}, {PublicPort: 5432}}, // duplicate on purpose — should dedupe
		},
		{
			ID:    "def456",
			Names: []string{},
			Image: "redis:7-alpine",
			Ports: []struct {
				PublicPort int `json:"PublicPort"`
			}{{PublicPort: 6379}},
		},
	}

	containers := parseContainers(raw)
	if len(containers) != 2 {
		t.Fatalf("got %d containers, want 2", len(containers))
	}

	if containers[0].Name != "seedicon-postgres" {
		t.Errorf("Name = %q, want seedicon-postgres (leading slash stripped)", containers[0].Name)
	}
	if len(containers[0].Ports) != 1 || containers[0].Ports[0] != 5432 {
		t.Errorf("Ports = %v, want [5432] deduped", containers[0].Ports)
	}

	if containers[1].Name != "def456" {
		t.Errorf("Name = %q, want fallback to ID when Names is empty", containers[1].Name)
	}
}

func TestMatchContainerDatabases(t *testing.T) {
	containers := []Container{
		{Name: "seedicon-postgres", Image: "postgres:16", Ports: []int{5432}},
		{Name: "seedicon-redis", Image: "redis:7-alpine", Ports: []int{6379}},
		{Name: "seedicon-events", Image: "seedicon/events:latest", Ports: []int{8080}}, // not a db
	}

	matches := MatchContainerDatabases(containers)
	if len(matches) != 2 {
		t.Fatalf("got %d matches, want 2: %+v", len(matches), matches)
	}

	byName := map[string]DatabaseMatch{}
	for _, m := range matches {
		byName[m.Name] = m
	}

	if pg := byName["PostgreSQL"]; pg.Port != 5432 || pg.Source != "Docker" {
		t.Errorf("PostgreSQL match = %+v", pg)
	}
	if r := byName["Redis"]; r.Port != 6379 || r.Source != "Docker" {
		t.Errorf("Redis match = %+v", r)
	}
}

func TestQueryContainers_NoDaemon_ReturnsNilNotError(t *testing.T) {
	// Doesn't assert on real Docker state (may or may not be running
	// on the test machine) — just proves the function never panics
	// and always returns a usable (possibly empty) slice.
	ctx := contextWithTimeout(t)
	_ = QueryContainers(ctx) // no daemon required to reach this line
}

func TestDetectSupervisor(t *testing.T) {
	tests := []struct {
		name   string
		parent *ProcInfo
		want   string
	}{
		{"nil parent", nil, ""},
		{"nodemon by exe name", &ProcInfo{Exe: "/usr/local/bin/nodemon"}, "nodemon"},
		{
			"nodemon invoked via node with script arg",
			&ProcInfo{Exe: "/usr/local/bin/node", Cmdline: []string{"node", "/path/node_modules/.bin/nodemon", "src/index.js"}},
			"nodemon",
		},
		{"plain shell, no supervisor", &ProcInfo{Exe: "/bin/zsh"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectSupervisor(tt.parent)
			if got != tt.want {
				t.Errorf("detectSupervisor() = %q, want %q", got, tt.want)
			}
		})
	}
}