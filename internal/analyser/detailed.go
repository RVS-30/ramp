package analyser

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
)

// detailedExtToLang extends extToLang with non-code file types worth
// showing in a breakdown, but not worth guessing as the "primary" language.
var detailedExtToLang = map[string]string{
	".go":    "Go",
	".py":    "Python",
	".js":    "JavaScript",
	".jsx":   "JavaScript",
	".mjs":   "JavaScript",
	".ts":    "TypeScript",
	".tsx":   "TypeScript",
	".java":  "Java",
	".kt":    "Kotlin",
	".cpp":   "C++",
	".cxx":   "C++",
	".cc":    "C++",
	".hpp":   "C++",
	".c":     "C",
	".h":     "C",
	".rb":    "Ruby",
	".php":   "PHP",
	".rs":    "Rust",
	".swift": "Swift",
	".cs":    "C#",
	".md":    "Markdown",
	".yaml":  "YAML",
	".yml":   "YAML",
	".json":  "JSON",
	".html":  "HTML",
	".css":   "CSS",
	".scss":  "SCSS",
	".sh":    "Shell",
	".sql":   "SQL",
	".toml":  "TOML",
	".xml":   "XML",
}

// LangStat holds file/line counts for one detected language.
type LangStat struct {
	Name  string
	Files int
	Lines int
}

// DetailedInfo holds a full language breakdown for a project.
type DetailedInfo struct {
	Languages  []LangStat // sorted by Lines, descending
	TotalFiles int
	TotalLines int
}

// DetailedStats walks root once, counting files and lines per language.
// Used only when --detailed is requested — the base AnalyseProject path
// stays cheap and doesn't pay this cost unless asked.
func DetailedStats(root string) (*DetailedInfo, error) {
	counts := make(map[string]*LangStat)
	totalFiles := 0
	totalLines := 0

	err := filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries rather than aborting the whole walk
		}
		if fi.IsDir() {
			if skipDirs[fi.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		lang, ok := detailedExtToLang[filepath.Ext(path)]
		if !ok {
			return nil
		}

		lines := countLines(path)
		totalFiles++
		totalLines += lines

		stat, exists := counts[lang]
		if !exists {
			stat = &LangStat{Name: lang}
			counts[lang] = stat
		}
		stat.Files++
		stat.Lines += lines
		return nil
	})
	if err != nil {
		return nil, err
	}

	langs := make([]LangStat, 0, len(counts))
	for _, s := range counts {
		langs = append(langs, *s)
	}
	sort.Slice(langs, func(i, j int) bool {
		return langs[i].Lines > langs[j].Lines
	})

	return &DetailedInfo{
		Languages:  langs,
		TotalFiles: totalFiles,
		TotalLines: totalLines,
	}, nil
}

// countLines does a fast line count via bufio.Scanner rather than reading
// the whole file into memory and splitting — matters once this runs
// across every source file in a real project, not just a handful.
func countLines(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024) // handle long lines (minified files etc.)

	count := 0
	for scanner.Scan() {
		count++
	}
	return count
}