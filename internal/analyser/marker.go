// marker maps a root-level file to the language/ecosystem it implies.
package analyser

import (
	"os"
	"path/filepath"
)

type marker struct {
	file     string
	language string
}

var markers = []marker{
	{"go.mod", "Go"},
	{"Cargo.toml", "Rust"},
	{"package.json", "JavaScript"},
	{"pyproject.toml", "Python"},
	{"requirements.txt", "Python"},
	{"pom.xml", "Java"},
	{"build.gradle", "Java"},
	{"Gemfile", "Ruby"},
	{"composer.json", "PHP"},
}

func detectFromMarkers(root string) (*ProjectInfo, bool) {
	for _, m := range markers {
		path := filepath.Join(root, m.file)
		data, err := os.ReadFile(path)
		if err != nil {
			continue // not this ecosystem, try next marker
		}

		info := &ProjectInfo{Language: m.language}

		switch m.file {
		case "go.mod":
			parseGoMod(data, info)
		case "package.json":
			parsePackageJSON(data, info)
		case "Cargo.toml":
			parseCargoToml(data, info)
		case "pyproject.toml":
			parsePyprojectToml(data, info)
		case "requirements.txt":
			parseRequirementsTxt(data, info)
		case "pom.xml":
			parsePomXML(data, info)
		case "build.gradle":
			parseBuildGradle(data, info)
		}

		if info.Name == "" {
			info.Name = filepath.Base(absOrRoot(root))
		}
		return info, true
	}
	return nil, false
}
