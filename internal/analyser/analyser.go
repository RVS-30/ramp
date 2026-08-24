package analyser

import "path/filepath"

type ProjectInfo struct {
	Name      string
	Language  string
	Framework string
	Version   string
	FileCount int
}


func AnalyseProject(root string) (*ProjectInfo, error) {
	if info, ok := detectFromMarkers(root); ok {
		info.FileCount = countFiles(root)
		return info, nil
	}
	return detectFromWalk(root)
}


func absOrRoot(root string) string {
	abs, err := filepath.Abs(root)
	if err != nil {
		return root
	}
	return abs
}