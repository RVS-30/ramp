package discovery

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/RVS-30/ramp/internal/analyser"
	"golang.org/x/sync/errgroup"
)

// DevPort is one row of the "dev ports" table: a listening port
// attributed to a project and stack.

type DevPort struct {
	Port       int
	Project    string
	Stack      string
	PID        int
	Supervisor string // "" if not supervised, otherwise e.g. "nodemon"
}

// DiscoveryResult is the full output of a Scan() run.
type DiscoveryResult struct {
	DevPorts  []DevPort
	Databases []DatabaseMatch
}

// knownSupervisors are process names that automatically restart
// their children on exit. If a process we're about to show/kill has
// one of these as its direct parent, the user should know killing
// the child alone likely won't stop it for long.
var knownSupervisors = map[string]bool{
	"nodemon":     true,
	"pm2":         true,
	"ts-node-dev": true,
	"forever":     true,
}

// NewScanner returns the Scanner implementation for the current OS.
func NewScanner() Scanner {
	return newOSScanner()
}

// Scan performs a full discovery pass: list processes, enrich them
// concurrently, classify, resolve project/stack, and match databases.
// Docker containers are intentionally not included here — that's a
// separate, independently-timed data source added at the caller level.
func Scan(ctx context.Context, scanner Scanner) (*DiscoveryResult, error) {
	pids, err := scanner.ListPIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing processes: %w", err)
	}

	procs, err := enrichAll(ctx, scanner, pids)
	if err != nil {
		return nil, fmt.Errorf("enriching processes: %w", err)
	}

	// Match databases first — a process identified as a known database
	// binary is excluded from the dev-ports section below, since its
	// cwd/project attribution (e.g. "started from inside a Go repo")
	// is incidental and would produce a misleading project/stack label.
	databases := MatchDatabases(procs)
	isDatabase := make(map[int]bool, len(databases))
	for _, db := range databases {
		isDatabase[db.PID] = true
	}

	parents := NewParentLookup(procs)
	resolver := NewProjectResolver()

	var devPorts []DevPort
	for _, p := range procs {
		if len(p.Ports) == 0 || isDatabase[p.PID] {
			continue
		}

		class := Classify(p, parents.Parent(p))
		if !class.IsDev || class.Confidence < ConfidenceMedium {
			continue
		}

		name, stack := projectNameAndStack(resolver, p)
		supervisor := detectSupervisor(parents.Parent(p))
		for _, port := range p.Ports {
			devPorts = append(devPorts, DevPort{
				Port:       port,
				Project:    name,
				Stack:      stack,
				PID:        p.PID,
				Supervisor: supervisor,
			})
		}
	}

	return &DiscoveryResult{
		DevPorts:  devPorts,
		Databases: MatchDatabases(procs),
	}, nil
}

// enrichAll fans Enrich() out across a bounded worker pool sized to
// available CPUs. A process that vanished mid-scan (ErrProcessVanished)
// or otherwise failed to enrich is dropped silently — one bad process
// must never abort the whole scan. If ctx is cancelled/times out,
// in-flight and unscheduled work stops and that error propagates up,
// since a caller-imposed deadline is a real, intentional signal to stop.
func enrichAll(ctx context.Context, scanner Scanner, pids []int) ([]ProcInfo, error) {
	var mu sync.Mutex
	procs := make([]ProcInfo, 0, len(pids))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.NumCPU())

	for _, pid := range pids {
		pid := pid
		g.Go(func() error {
			if err := gctx.Err(); err != nil {
				return err
			}

			info, err := scanner.Enrich(gctx, pid)
			if err != nil {
				if errors.Is(err, ErrProcessVanished) {
					return nil
				}
				return nil // unknown enrich failure: skip, don't abort the scan
			}

			mu.Lock()
			procs = append(procs, info)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return procs, nil
}

// projectNameAndStack resolves display name and stack label for a
// dev-classified process, falling back to the cwd's base directory
// name when project analysis found nothing more specific.
func projectNameAndStack(resolver *ProjectResolver, p ProcInfo) (name, stack string) {
	name = filepath.Base(p.Cwd)
	info := resolver.Resolve(p.Cwd)
	if info == nil {
		return name, ""
	}
	if info.Name != "" {
		name = info.Name
	}
	return name, stackLabel(info)
}

// excludePortsIn removes any DevPort whose Port matches one already
// claimed by a database — whether local or containerized. This
// catches host-side artifacts of database processes (e.g. Docker
// Desktop's per-container port-forwarding proxy) that the PID-based
// exclusion in Scan() can't see, since that proxy process is
// genuinely distinct from the database process itself and was never
// matched by MatchDatabases directly.
func ExcludePortsInDatabases(devPorts []DevPort, databases []DatabaseMatch) []DevPort {
	claimed := make(map[int]bool, len(databases))
	for _, db := range databases {
		claimed[db.Port] = true
	}

	filtered := make([]DevPort, 0, len(devPorts))
	for _, dp := range devPorts {
		if !claimed[dp.Port] {
			filtered = append(filtered, dp)
		}
	}
	return filtered
}

// stackLabel prefers the detected framework (e.g. "Next.js") and
// falls back to the language (e.g. "Go") when no framework matched.
func stackLabel(info *analyser.ProjectInfo) string {
	if info.Framework != "" {
		return info.Framework
	}
	return info.Language
}

// detectSupervisor returns the supervisor's display name if parent
// looks like a process supervisor, or "" otherwise. Pure function,
// no I/O — same testing philosophy as classify.go.
func detectSupervisor(parent *ProcInfo) string {
	if parent == nil {
		return ""
	}
	base := filepath.Base(parent.Exe)
	// nodemon/pm2/etc are typically invoked as
	// ".../node_modules/.bin/nodemon" or similar — match on the
	// cmdline's script name too, since Exe is often just "node".
	if knownSupervisors[base] {
		return base
	}
	for _, arg := range parent.Cmdline {
		name := filepath.Base(arg)
		if knownSupervisors[name] {
			return name
		}
	}
	return ""
}