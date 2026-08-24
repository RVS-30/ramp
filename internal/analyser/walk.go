package analyser

import (
	"os"
	"path/filepath"
)

var skipDirs = map[string]bool{
	".git":         true,
	".hg":          true,
	".svn":         true,
	"node_modules": true,
	"vendor":       true,
	".venv":        true,
	"venv":         true,
	"__pycache__":  true,
	"dist":         true,
	"build":        true,
	"target":       true,
	".idea":        true,
	".vscode":      true,
	"bin":          true,
	"obj":          true,
}

var extToLang = map[string]string{
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
}

func countFiles(root string) int {
	count := 0
	filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries, don't abort the count
		}
		if fi.IsDir() {
			if skipDirs[fi.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		count++
		return nil
	})
	return count
}

func detectFromWalk(root string) (*ProjectInfo, error) {
	languageCount := make(map[string]int)
	fileCount := 0

	err := filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			if skipDirs[fi.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		fileCount++
		if lang, ok := extToLang[filepath.Ext(path)]; ok {
			languageCount[lang]++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var dominant string
	max := 0
	for lang, c := range languageCount {
		if c > max {
			max, dominant = c, lang
		}
	}

	return &ProjectInfo{
		Name:      filepath.Base(absOrRoot(root)),
		Language:  dominant,
		FileCount: fileCount,
	}, nil
}