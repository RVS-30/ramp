package analyser

import (
	"strings"

	"github.com/BurntSushi/toml"
)

var pyFrameworkMarkers = map[string]string{
	// Web frameworks
	"django":  "Django",
	"flask":   "Flask",
	"fastapi": "FastAPI",
	"tornado": "Tornado",
	"pyramid": "Pyramid",
	"bottle":  "Bottle",
	"sanic":   "Sanic",

	// Data / ML
	"pandas":     "Pandas",
	"numpy":      "NumPy",
	"torch":      "PyTorch",
	"tensorflow": "TensorFlow",
	"scikit-learn": "scikit-learn",

	// CLI
	"click":   "Click",
	"typer":   "Typer",
}

var pyFrameworkPriority = []string{
	"Django", "Flask", "FastAPI", "Tornado", "Pyramid", "Bottle", "Sanic",
	"PyTorch", "TensorFlow", "scikit-learn", "Pandas", "NumPy",
	"Click", "Typer",
}

var pyToolingMarkers = map[string]string{
	"pytest": "pytest",
	"tox":    "tox",
}

var pyToolingPriority = []string{
	"pytest", "tox",
}

// pyprojectToml covers both PEP 621 (project.dependencies) and Poetry
// (tool.poetry.dependencies) layouts, since either may be present.
type pyprojectToml struct {
	Project struct {
		Name           string   `toml:"name"`
		Version        string   `toml:"version"`
		RequiresPython string   `toml:"requires-python"`
		Dependencies   []string `toml:"dependencies"`
	} `toml:"project"`
	Tool struct {
		Poetry struct {
			Name         string                     `toml:"name"`
			Dependencies map[string]toml.Primitive `toml:"dependencies"`
		} `toml:"poetry"`
	} `toml:"tool"`
}

func parsePyprojectToml(data []byte, info *ProjectInfo) {
	var proj pyprojectToml
	if _, err := toml.Decode(string(data), &proj); err != nil {
		return // malformed pyproject.toml — leave info as-is, caller falls back
	}

	switch {
	case proj.Project.Name != "":
		info.Name = proj.Project.Name
	case proj.Tool.Poetry.Name != "":
		info.Name = proj.Tool.Poetry.Name
	}
	if proj.Project.RequiresPython != "" {
		info.Version = "Python " + proj.Project.RequiresPython
	}

	deps := make(map[string]bool)

	// PEP 621: dependencies is a flat list like "django>=4.2" — extract
	// the package name before any version specifier.
	for _, dep := range proj.Project.Dependencies {
		deps[pyPackageName(dep)] = true
	}
	// Poetry: dependencies is a table keyed by package name.
	for dep := range proj.Tool.Poetry.Dependencies {
		if dep == "python" {
			continue // this is the Python version constraint, not a dependency
		}
		deps[dep] = true
	}

	if matchFramework(deps, pyFrameworkMarkers, pyFrameworkPriority, info) {
		return
	}
	matchFramework(deps, pyToolingMarkers, pyToolingPriority, info)
}

// pyPackageName strips version specifiers and extras from a PEP 508
// dependency string, e.g. "django[bcrypt]>=4.2" -> "django".
func pyPackageName(dep string) string {
	dep = strings.TrimSpace(dep)
	for _, sep := range []string{"[", "==", ">=", "<=", "~=", "!=", ">", "<", " "} {
		if idx := strings.Index(dep, sep); idx != -1 {
			dep = dep[:idx]
		}
	}
	return strings.ToLower(strings.TrimSpace(dep))
}