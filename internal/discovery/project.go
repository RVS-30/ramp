package discovery

import (
	"sync"

	"github.com/RVS-30/ramp/internal/analyser"
)

// analyseProjectFn is a seam for testing — production code uses
// analyser.AnalyseProject, tests substitute a fake so this file never
// needs a real filesystem to be verified.
var analyseProjectFn = analyser.AnalyseProject

// ProjectResolver resolves a process's cwd into project metadata,
// memoizing per unique cwd within a single discovery run so processes
// sharing a monorepo (e.g. multiple Next.js apps under one repo) don't
// re-walk the filesystem redundantly. Not safe to reuse across runs —
// the filesystem may change between invocations.
type ProjectResolver struct {
	mu    sync.Mutex
	cache map[string]*analyser.ProjectInfo
}

// NewProjectResolver creates a resolver scoped to a single Scan() call.
func NewProjectResolver() *ProjectResolver {
	return &ProjectResolver{
		cache: make(map[string]*analyser.ProjectInfo),
	}
}

// Resolve returns project info for the given cwd, using the cache if
// this exact cwd was already resolved earlier in this run. Returns
// nil if cwd is empty or analysis fails — callers should treat that
// as "no project could be determined" rather than an error, since an
// unresolvable project is expected for many processes (e.g. cwd was
// unreadable, or genuinely isn't a project directory).
func (r *ProjectResolver) Resolve(cwd string) *analyser.ProjectInfo {
	if cwd == "" {
		return nil
	}

	r.mu.Lock()
	if cached, ok := r.cache[cwd]; ok {
		r.mu.Unlock()
		return cached
	}
	r.mu.Unlock()

	info, err := analyseProjectFn(cwd)
	if err != nil {
		info = nil
	}

	r.mu.Lock()
	r.cache[cwd] = info
	r.mu.Unlock()

	return info
}