package discovery

import (
	"os/exec"
	"testing"
)

func TestTerminate_RealProcess(t *testing.T) {
	// Spawn a real short-lived process we own, so this test exercises
	// the actual syscall path rather than mocking it — Signal
	// permission/ESRCH behavior is OS-level, worth verifying for real.
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start test process: %v", err)
	}
	pid := cmd.Process.Pid

	result := Terminate(pid)
	if !result.Killed {
		t.Errorf("expected Killed=true, got %+v", result)
	}

	_ = cmd.Wait() // reap it, avoid a zombie in the test run
}

func TestTerminate_AlreadyDeadProcess(t *testing.T) {
	cmd := exec.Command("true") // exits immediately
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start test process: %v", err)
	}
	pid := cmd.Process.Pid
	_ = cmd.Wait() // ensure it's actually reaped/dead before we try to signal it

	result := Terminate(pid)
	if result.Killed {
		t.Errorf("expected Killed=false for a dead process, got %+v", result)
	}
	if result.Message != "process no longer running" {
		t.Errorf("Message = %q, want %q", result.Message, "process no longer running")
	}
}

func TestTerminate_PermissionDenied(t *testing.T) {
	// PID 1 (launchd on macOS) is always running and never owned by
	// a normal user — a reliable, real EPERM case without needing
	// root or a second user account.
	result := Terminate(1)
	if result.Killed {
		t.Errorf("expected Killed=false against pid 1, got %+v", result)
	}
	if result.Message != "permission denied — you don't own this process" {
		t.Errorf("Message = %q, want permission denied message", result.Message)
	}
}