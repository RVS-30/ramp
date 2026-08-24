package analyser

import (
	"regexp"
)

// gradleDepPattern matches common Gradle dependency declaration styles:
//   implementation 'org.springframework.boot:spring-boot-starter'
//   implementation("org.springframework.boot:spring-boot-starter:3.2.0")
//   api group: 'junit', name: 'junit', version: '4.13'
var gradleDepPattern = regexp.MustCompile(`['"]([a-zA-Z0-9_.\-]+):([a-zA-Z0-9_.\-]+)`)

func parseBuildGradle(data []byte, info *ProjectInfo) {
	deps := make(map[string]bool)

	matches := gradleDepPattern.FindAllStringSubmatch(string(data), -1)
	for _, m := range matches {
		// m[2] is the artifactId portion, e.g. "spring-boot-starter"
		deps[m[2]] = true
	}

	if matchFramework(deps, javaFrameworkMarkers, javaFrameworkPriority, info) {
		return
	}
	matchFramework(deps, javaToolingMarkers, javaToolingPriority, info)

	// build.gradle has no reliable single field for project name/version
	// the way pom.xml does (it's set via a separate settings.gradle
	// rootProject.name, or the directory name by default) — leaving
	// info.Name to fall back to the directory name is the honest choice.
}