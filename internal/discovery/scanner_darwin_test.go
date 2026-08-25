//go:build darwin

package discovery

import (
	"context"
	"net"
	"os"
	"testing"
)

// TestDarwinScanner_Integration is a smoke test against real
// processes on this machine. It doesn't assert exhaustive
// correctness (that's classify_test.go's job) — just that the
// syscall layer doesn't crash and returns plausible data, including
// finding this very test process by its own PID.
func TestDarwinScanner_Integration(t *testing.T) {
	s := newOSScanner()
	ctx := context.Background()

	pids, err := s.ListPIDs(ctx)
	if err != nil {
		t.Fatalf("ListPIDs failed: %v", err)
	}
	if len(pids) < 10 {
		t.Fatalf("got only %d pids, expected a real machine to have far more", len(pids))
	}

	self := os.Getpid()
	found := false
	for _, p := range pids {
		if p == self {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("own PID %d not found in ListPIDs result", self)
	}

	info, err := s.Enrich(ctx, self)
	if err != nil {
		t.Fatalf("Enrich(self) failed: %v", err)
	}
	if info.Exe == "" {
		t.Errorf("Enrich(self).Exe is empty, expected the test binary's path")
	}
	if info.Cwd == "" {
		t.Errorf("Enrich(self).Cwd is empty, expected our own working directory")
	}
	t.Logf("self: pid=%d ppid=%d exe=%s cwd=%s", info.PID, info.PPID, info.Exe, info.Cwd)
}

func TestDarwinScanner_ListeningPorts(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test listener: %v", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port

	s := newOSScanner()
	info, err := s.Enrich(context.Background(), os.Getpid())
	if err != nil {
		t.Fatalf("Enrich(self) failed: %v", err)
	}

	found := false
	for _, p := range info.Ports {
		if p == port {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected port %d in %v (own process should show as listening)", port, info.Ports)
	}
}

func TestDarwinScanner_Cmdline(t *testing.T) {
	s := newOSScanner()
	info, err := s.Enrich(context.Background(), os.Getpid())
	if err != nil {
		t.Fatalf("Enrich(self) failed: %v", err)
	}
	if len(info.Cmdline) == 0 {
		t.Fatalf("Cmdline is empty, expected at least the exec path/argv[0]")
	}
	t.Logf("self cmdline: %v", info.Cmdline)
}