package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/RVS-30/ramp/internal/discovery"
	"github.com/RVS-30/ramp/internal/output"
	"github.com/RVS-30/ramp/internal/tui"
	"github.com/spf13/cobra"
)

var portsKillCmd = &cobra.Command{
	Use:   "kill [port]",
	Short: "Terminate a process listening on a dev port",
	Long: `Terminates the process listening on the given port, after asking
for confirmation. Run without a port to pick interactively from a list.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		scanner := discovery.NewScanner()
		result, err := discovery.Scan(ctx, scanner)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ramp ports kill:", err)
			os.Exit(1)
		}

		var pid int
		var label string

		if len(args) == 1 {
			port, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "ramp ports kill: %q is not a valid port number\n", args[0])
				os.Exit(1)
			}

			var found bool
			pid, label, found = findPIDForPort(result, port)
			if !found {
				output.PrintKillNotFound(port)
				return
			}

			output.PrintKillPrompt(label, pid, port)
			if !confirm() {
				output.PrintKillCancelled()
				return
			}
		} else {
			candidates := buildKillCandidates(result)
			if len(candidates) == 0 {
				fmt.Println("Nothing found to kill.")
				return
			}

			chosen, err := tui.RunKillPicker(candidates)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ramp ports kill:", err)
				os.Exit(1)
			}
			if chosen == nil {
				return // user cancelled in the picker
			}

			pid, label = chosen.PID, chosen.Label
			output.PrintKillPrompt(label, pid, chosen.Port)
			if !confirm() {
				output.PrintKillCancelled()
				return
			}
		}

		res := discovery.Terminate(pid)
		output.PrintKillResult(res.Killed, res.Message)
		if !res.Killed {
			os.Exit(1)
		}
	},
}

// buildKillCandidates converts a DiscoveryResult into the picker's
// input shape, kept deliberately decoupled from discovery's own
// types (see internal/tui's KillCandidate doc comment).
func buildKillCandidates(result *discovery.DiscoveryResult) []tui.KillCandidate {
	candidates := make([]tui.KillCandidate, 0, len(result.DevPorts)+len(result.Databases))

	for _, dp := range result.DevPorts {
		candidates = append(candidates, tui.KillCandidate{
			PID: dp.PID, Port: dp.Port, Label: dp.Project, Stack: dp.Stack,
		})
	}
	for _, db := range result.Databases {
		if db.PID == 0 {
			continue // containerized databases have no host PID to kill — see Step 16 note
		}
		candidates = append(candidates, tui.KillCandidate{
			PID: db.PID, Port: db.Port, Label: db.Name, Stack: db.Source,
		})
	}

	return candidates
}

// findPIDForPort looks up which process/database owns a given port
// within an already-scanned result, so kill never has to re-scan —
// it reuses the exact same data ramp ports just showed the user.
func findPIDForPort(result *discovery.DiscoveryResult, port int) (pid int, label string, found bool) {
	for _, dp := range result.DevPorts {
		if dp.Port == port {
			return dp.PID, dp.Project, true
		}
	}
	for _, db := range result.Databases {
		if db.Port == port {
			return db.PID, db.Name, true
		}
	}
	return 0, "", false
}

func confirm() bool {
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

func init() {
	portsCmd.AddCommand(portsKillCmd)
}
