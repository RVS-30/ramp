package discovery

import (
	"os"
	"path/filepath"
	"strings"
)

// Confidence expresses how sure we are that a process is a
// developer-relevant one, not a system/background process.
type Confidence int

const (
	ConfidenceNone Confidence = iota
	ConfidenceLow
	ConfidenceMedium
	ConfidenceHigh
)

// Classification is the classifier's verdict on one ProcInfo, along
// with a human-readable reason — useful for --all/debug output and
// for making test failures legible.
type Classification struct {
	IsDev      bool
	Confidence Confidence
	Reason     string
}

// systemPathPrefixes are directories system/OS-managed binaries run
// from. A process whose exe lives here is essentially never a dev
// server, regardless of what port it happens to hold.
var systemPathPrefixes = []string{
	"/System/",
	"/usr/libexec/",
	"/usr/sbin/",
	"/sbin/",
	"/Library/Apple/",
}

// shellOrIDENames are processes commonly found in a dev server's
// parent chain. Presence doesn't guarantee "dev" but strongly
// supports it when combined with other signals.
var shellOrIDENames = map[string]bool{
	"bash": true, "zsh": true, "fish": true, "sh": true,
	"code": true, "code-helper": true, "cursor": true,
	"iterm2": true, "terminal": true, "tmux": true, "screen": true,
}

// Classify inspects a single process and decides whether it looks
// developer-relevant. It never performs I/O — proc is a snapshot the
// caller already gathered.
func Classify(proc ProcInfo, parent *ProcInfo) Classification {
	// Strongest system signal: exe path is under a known OS-managed
	// directory. This alone is enough to classify as non-dev.
	if proc.Exe != "" {
		for _, prefix := range systemPathPrefixes {
			if strings.HasPrefix(proc.Exe, prefix) {
				return Classification{
					IsDev:      false,
					Confidence: ConfidenceHigh,
					Reason:     "executable under system path " + prefix,
				}
			}
		}
	}

	// No cwd at all (permission denied, or a kernel/system process
	// with no meaningful working directory) — can't confirm dev
	// relevance, so treat as unknown/low confidence.
	if proc.Cwd == "" {
		return Classification{
			IsDev:      false,
			Confidence: ConfidenceNone,
			Reason:     "working directory unavailable",
		}
	}

	// cwd under the user's home directory is a meaningful positive
	// signal — system daemons don't run from ~/projects/whatever.
	home, _ := os.UserHomeDir()
	underHome := home != "" && strings.HasPrefix(proc.Cwd, home)

	// Parent is a shell/terminal/IDE — classic dev-server ancestry.
	parentIsShell := false
	if parent != nil {
		base := strings.ToLower(filepath.Base(parent.Exe))
		parentIsShell = shellOrIDENames[base]
	}

	switch {
	case underHome && parentIsShell:
		return Classification{
			IsDev:      true,
			Confidence: ConfidenceHigh,
			Reason:     "cwd under home, parent is shell/IDE",
		}
	case underHome:
		return Classification{
			IsDev:      true,
			Confidence: ConfidenceMedium,
			Reason:     "cwd under home directory",
		}
	case parentIsShell:
		return Classification{
			IsDev:      true,
			Confidence: ConfidenceLow,
			Reason:     "parent is shell/IDE, but cwd outside home",
		}
	default:
		return Classification{
			IsDev:      false,
			Confidence: ConfidenceLow,
			Reason:     "no dev signals matched",
		}
	}
}