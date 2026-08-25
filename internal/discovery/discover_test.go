package discovery

import (
	"context"
	"testing"

	"github.com/RVS-30/ramp/internal/analyser"
)

func TestScan_ClassifiesAndAssembles(t *testing.T) {
	home, err := osUserHomeDirForTest()
	if err != nil {
		t.Fatalf("could not resolve home dir for test: %v", err)
	}
	devCwd := home + "/projects/events-web"

	origFn := analyseProjectFn
	defer func() { analyseProjectFn = origFn }()

	analyseProjectFn = func(root string) (*analyser.ProjectInfo, error) {
		if root == devCwd {
			return &analyser.ProjectInfo{Name: "events-web", Framework: "Next.js"}, nil
		}
		return nil, errBoom{}
	}

	scanner := &mockScanner{
		pids: []int{1, 100, 200, 300, 400},
		procs: map[int]ProcInfo{
			1:   {PID: 1, PPID: 0, Exe: "/sbin/launchd"},
			100: {PID: 100, PPID: 1, Exe: "/bin/zsh"},
			200: {
				PID: 200, PPID: 100, Exe: "/usr/local/bin/node",
				Cwd: devCwd, Ports: []int{3000},
			},
			300: {PID: 300, PPID: 1, Exe: "/usr/libexec/rapportd", Ports: []int{5353}},
			400: {PID: 400, PPID: 1, Exe: "/usr/local/bin/postgres", Ports: []int{5432}},
		},
	}

	result, err := Scan(context.Background(), scanner)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.DevPorts) != 1 {
		t.Fatalf("got %d dev ports, want 1: %+v", len(result.DevPorts), result.DevPorts)
	}
	dp := result.DevPorts[0]
	if dp.Port != 3000 || dp.Project != "events-web" || dp.Stack != "Next.js" {
		t.Errorf("dev port = %+v, want {3000 events-web Next.js ...}", dp)
	}

	if len(result.Databases) != 1 {
		t.Fatalf("got %d databases, want 1: %+v", len(result.Databases), result.Databases)
	}
	if result.Databases[0].Name != "PostgreSQL" {
		t.Errorf("database = %+v, want PostgreSQL", result.Databases[0])
	}
}

func TestScan_EmptyWhenNoProcesses(t *testing.T) {
	scanner := &mockScanner{pids: []int{}, procs: map[int]ProcInfo{}}
	result, err := Scan(context.Background(), scanner)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(result.DevPorts) != 0 || len(result.Databases) != 0 {
		t.Errorf("expected empty result, got %+v", result)
	}
}

func TestScan_SkipsVanishedProcess(t *testing.T) {
	scanner := &mockScanner{
		pids: []int{1, 2},
		procs: map[int]ProcInfo{
			1: {PID: 1, PPID: 0, Exe: "/sbin/launchd"},
		},
		errs: map[int]error{
			2: ErrProcessVanished,
		},
	}
	result, err := Scan(context.Background(), scanner)
	if err != nil {
		t.Fatalf("Scan should not fail on a vanished process: %v", err)
	}
	if len(result.DevPorts) != 0 {
		t.Errorf("expected no dev ports, got %+v", result.DevPorts)
	}
}

func TestScan_DatabaseNotDuplicatedInDevPorts(t *testing.T) {
	home, err := osUserHomeDirForTest()
	if err != nil {
		t.Fatalf("could not resolve home dir for test: %v", err)
	}
	// Simulates the real bug: redis-server launched from inside a Go
	// project's own directory. It must appear under Databases only.
	repoCwd := home + "/Projects/GoRamp"

	origFn := analyseProjectFn
	defer func() { analyseProjectFn = origFn }()
	analyseProjectFn = func(root string) (*analyser.ProjectInfo, error) {
		return &analyser.ProjectInfo{Name: "ramp", Framework: "Cobra"}, nil
	}

	scanner := &mockScanner{
		pids: []int{1, 100, 200},
		procs: map[int]ProcInfo{
			1:   {PID: 1, PPID: 0, Exe: "/sbin/launchd"},
			100: {PID: 100, PPID: 1, Exe: "/bin/zsh"},
			200: {
				PID: 200, PPID: 100, Exe: "/usr/local/bin/redis-server",
				Cwd: repoCwd, Ports: []int{6379},
			},
		},
	}

	result, err := Scan(context.Background(), scanner)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	for _, dp := range result.DevPorts {
		if dp.PID == 200 {
			t.Errorf("redis process leaked into DevPorts: %+v", dp)
		}
	}
	if len(result.Databases) != 1 || result.Databases[0].PID != 200 {
		t.Errorf("expected redis under Databases only, got %+v", result.Databases)
	}
}

func TestExcludePortsInDatabases_RemovesHostProxyLeak(t *testing.T) {
	devPorts := []DevPort{
		{Port: 3000, Project: "networking", Stack: "Next.js"},
		{Port: 6380, Project: "Data", Stack: ""}, // simulates the docker-proxy leak
	}
	databases := []DatabaseMatch{
		{Name: "Redis", Port: 6380, Source: "Docker"},
	}

	filtered := ExcludePortsInDatabases(devPorts, databases)

	if len(filtered) != 1 {
		t.Fatalf("got %d dev ports, want 1: %+v", len(filtered), filtered)
	}
	if filtered[0].Port != 3000 {
		t.Errorf("wrong port survived filtering: %+v", filtered[0])
	}
}

type errBoom struct{}

func (errBoom) Error() string { return "boom" }