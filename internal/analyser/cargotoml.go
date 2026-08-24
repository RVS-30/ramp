package analyser

import (
	"github.com/BurntSushi/toml"
)

// rustFrameworkMarkers maps a crates.io crate name to its display name.
// These compete for info.Framework via rustFrameworkPriority.
var rustFrameworkMarkers = map[string]string{
	// Web frameworks
	"actix-web": "Actix Web",
	"axum":      "Axum",
	"rocket":    "Rocket",
	"warp":      "Warp",
	"tide":      "Tide",

	// Async runtimes (only meaningful as "framework" if no web framework found)
	"tokio":    "Tokio",
	"async-std": "async-std",

	// CLI frameworks
	"clap":    "Clap",
	"structopt": "StructOpt",

	// ORM / database
	"diesel":  "Diesel",
	"sqlx":    "SQLx",
	"sea-orm": "SeaORM",
}

var rustFrameworkPriority = []string{
	"Actix Web", "Axum", "Rocket", "Warp", "Tide",
	"Diesel", "SQLx", "SeaORM",
	"Clap", "StructOpt",
	"Tokio", "async-std",
}

// rustToolingMarkers are build/test helpers — only reported as
// info.Framework when nothing in rustFrameworkMarkers matched at all.
var rustToolingMarkers = map[string]string{
	"criterion": "Criterion",
	"proptest":  "Proptest",
}

var rustToolingPriority = []string{
	"Criterion", "Proptest",
}

// cargoToml mirrors the subset of Cargo.toml fields we care about.
// Dependency values in Cargo.toml can be either a plain version string
// ("1.0") or a table ({ version = "1.0", features = [...] }), so we use
// toml.Primitive to accept either shape without needing two passes.
type cargoToml struct {
	Package struct {
		Name    string `toml:"name"`
		Version string `toml:"version"`
		Edition string `toml:"edition"`
	} `toml:"package"`
	Dependencies    map[string]toml.Primitive `toml:"dependencies"`
	DevDependencies map[string]toml.Primitive `toml:"dev-dependencies"`
}

func parseCargoToml(data []byte, info *ProjectInfo) {
	var cargo cargoToml
	if _, err := toml.Decode(string(data), &cargo); err != nil {
		return // malformed Cargo.toml — leave info as-is, caller falls back
	}

	if cargo.Package.Name != "" {
		info.Name = cargo.Package.Name
	}
	if cargo.Package.Edition != "" {
		info.Version = "Rust " + cargo.Package.Edition + " edition"
	}

	allDeps := make(map[string]bool, len(cargo.Dependencies)+len(cargo.DevDependencies))
	for dep := range cargo.Dependencies {
		allDeps[dep] = true
	}
	for dep := range cargo.DevDependencies {
		allDeps[dep] = true
	}

	if matchFramework(allDeps, rustFrameworkMarkers, rustFrameworkPriority, info) {
		return
	}
	matchFramework(allDeps, rustToolingMarkers, rustToolingPriority, info)
}