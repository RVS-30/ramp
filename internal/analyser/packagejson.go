package analyser

import (
	"encoding/json"
)

// jsFrameworkMarkers maps an npm package name to its display name.
// These compete for info.Framework via jsFrameworkPriority.
var jsFrameworkMarkers = map[string]string{
	// Meta-frameworks
	"next":      "Next.js",
	"nuxt":      "Nuxt",
	"remix":     "Remix",
	"astro":     "Astro",
	"gatsby":    "Gatsby",
	"sveltekit": "SvelteKit",

	// Frontend frameworks
	"react":          "React",
	"vue":            "Vue",
	"svelte":         "Svelte",
	"@angular/core":  "Angular",
	"solid-js":       "SolidJS",
	"preact":         "Preact",

	// Backend frameworks
	"express":      "Express",
	"fastify":      "Fastify",
	"koa":          "Koa",
	"@nestjs/core": "NestJS",
	"hapi":         "Hapi",
}

// jsFrameworkPriority ranks categories so meta-frameworks outrank the
// underlying library they wrap (e.g. Next.js over React).
var jsFrameworkPriority = []string{
	"Next.js", "Nuxt", "Remix", "Astro", "Gatsby", "SvelteKit",
	"React", "Vue", "Svelte", "Angular", "SolidJS", "Preact",
	"Express", "Fastify", "Koa", "NestJS", "Hapi",
}

// jsToolingMarkers are build tools / test runners — only reported as
// info.Framework when nothing in jsFrameworkMarkers matched at all.
var jsToolingMarkers = map[string]string{
	"vite":    "Vite",
	"webpack": "Webpack",
	"jest":    "Jest",
	"vitest":  "Vitest",
	"mocha":   "Mocha",
}

var jsToolingPriority = []string{
	"Vite", "Webpack", "Jest", "Vitest", "Mocha",
}

// packageJSON mirrors the subset of package.json fields we care about.
type packageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Engines         struct {
		Node string `json:"node"`
	} `json:"engines"`
}

func parsePackageJSON(data []byte, info *ProjectInfo) {
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return // malformed package.json — leave info as-is, caller falls back
	}

	if pkg.Name != "" {
		info.Name = pkg.Name
	}
	if pkg.Engines.Node != "" {
		info.Version = "Node " + pkg.Engines.Node
	}

	allDeps := make(map[string]bool, len(pkg.Dependencies)+len(pkg.DevDependencies))
	for dep := range pkg.Dependencies {
		allDeps[dep] = true
	}
	for dep := range pkg.DevDependencies {
		allDeps[dep] = true
	}

	if matchFramework(allDeps, jsFrameworkMarkers, jsFrameworkPriority, info) {
		return
	}
	matchFramework(allDeps, jsToolingMarkers, jsToolingPriority, info)
}

// matchFramework checks deps against markers, and if any matched, sets
// info.Framework to the highest-priority match. Returns whether a match
// was found, so callers can chain tiers (frameworks, then tooling).
func matchFramework(deps map[string]bool, markers map[string]string, priority []string, info *ProjectInfo) bool {
	found := make(map[string]bool)
	for dep := range deps {
		if name, ok := markers[dep]; ok {
			found[name] = true
		}
	}
	for _, name := range priority {
		if found[name] {
			info.Framework = name
			return true
		}
	}
	return false
}