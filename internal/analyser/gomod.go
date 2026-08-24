package analyser

import (
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// frameworkMarkers maps a Go module import path (or prefix of one) to
// its display name.
var frameworkMarkers = map[string]string{
	// Web frameworks
	"gin-gonic/gin": "Gin",
	"labstack/echo": "Echo",
	"gofiber/fiber": "Fiber",
	"go-chi/chi":    "Chi",
	"gorilla/mux":   "Gorilla Mux",
	"beego":         "Beego",
	"kataras/iris":  "Iris",
	"revel/revel":   "Revel",

	// CLI frameworks
	"spf13/cobra":        "Cobra",
	"urfave/cli":         "urfave/cli",
	"alecthomas/kingpin": "Kingpin",

	// RPC / API
	"google.golang.org/grpc": "gRPC",
	"99designs/gqlgen":       "gqlgen",
	"graphql-go/graphql":     "graphql-go",

	// ORM / database
	"gorm.io/gorm": "GORM",
	"jmoiron/sqlx": "sqlx",
	"ent.io":       "Ent",
	"uptrace/bun":  "Bun ORM",

	// TUI
	"charmbracelet/bubbletea": "Bubble Tea",
	"charmbracelet/lipgloss":  "Lip Gloss",

	// Testing / mocking (only matters if nothing else found)
	"stretchr/testify": "Testify",
}

// frameworkPriority ranks framework categories so that when several
// direct dependencies match, the most "defining" one wins.
var frameworkPriority = []string{
	"Gin", "Echo", "Fiber", "Chi", "Gorilla Mux", "Beego", "Iris", "Revel",
	"gRPC", "gqlgen", "graphql-go",
	"Cobra", "urfave/cli", "Kingpin",
	"Bubble Tea",
	"GORM", "sqlx", "Ent", "Bun ORM",
	"Lip Gloss",
	"Testify",
}

func parseGoMod(data []byte, info *ProjectInfo) {
	f, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return // malformed go.mod — leave info as-is, caller falls back
	}

	if f.Module != nil {
		info.Name = filepath.Base(f.Module.Mod.Path)
	}
	if f.Go != nil {
		info.Version = "Go " + f.Go.Version
	}

	found := make(map[string]bool)
	for _, req := range f.Require {
		if req.Indirect {
			continue // skip transitive deps entirely
		}
		for importPath, name := range frameworkMarkers {
			if strings.Contains(req.Mod.Path, importPath) {
				found[name] = true
			}
		}
	}

	for _, name := range frameworkPriority {
		if found[name] {
			info.Framework = name
			return
		}
	}
}
