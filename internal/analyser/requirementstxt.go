package analyser

import (
	"strings"
)
func parseRequirementsTxt(data []byte, info *ProjectInfo) {
	lines := strings.Split(string(data), "\n")
	deps := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "-r ") || strings.HasPrefix(line, "--") {
			continue // skip -r other.txt includes and pip flags like --index-url
		}
		deps[pyPackageName(line)] = true
	}

	if matchFramework(deps, pyFrameworkMarkers, pyFrameworkPriority, info) {
		return
	}
	matchFramework(deps, pyToolingMarkers, pyToolingPriority, info)

	// requirements.txt has no project name/version fields at all —
	// info.Name falls back to the directory name in AnalyseProject's
	// caller if still empty after this.
}