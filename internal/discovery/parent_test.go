package discovery

import "testing"

func TestParentLookup(t *testing.T) {
	procs := []ProcInfo{
		{PID: 1, PPID: 0, Exe: "/sbin/launchd"},
		{PID: 100, PPID: 1, Exe: "/bin/zsh"},
		{PID: 200, PPID: 100, Exe: "/usr/local/bin/node"},
	}

	lookup := NewParentLookup(procs)

	t.Run("finds direct parent", func(t *testing.T) {
		child := procs[2] // node, ppid 100
		parent := lookup.Parent(child)
		if parent == nil {
			t.Fatal("expected parent, got nil")
		}
		if parent.Exe != "/bin/zsh" {
			t.Errorf("parent.Exe = %q, want /bin/zsh", parent.Exe)
		}
	})

	t.Run("ppid zero returns nil", func(t *testing.T) {
		root := procs[0] // launchd, ppid 0
		if got := lookup.Parent(root); got != nil {
			t.Errorf("expected nil for ppid 0, got %+v", got)
		}
	})

	t.Run("parent not in scanned set returns nil", func(t *testing.T) {
		orphan := ProcInfo{PID: 999, PPID: 54321, Exe: "/usr/bin/something"}
		if got := lookup.Parent(orphan); got != nil {
			t.Errorf("expected nil for unknown ppid, got %+v", got)
		}
	})
}