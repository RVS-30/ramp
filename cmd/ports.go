package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/RVS-30/ramp/internal/discovery"
	"github.com/RVS-30/ramp/internal/output"
	"github.com/spf13/cobra"
)

var portsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Show ports used by your active development environment",
	Long: `Scans your machine for active dev servers, local databases, and
Docker containers — filtering out normal OS/system processes.

Examples:
  ramp ports
  ramp ports --all`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var (
			wg         sync.WaitGroup
			result     *discovery.DiscoveryResult
			scanErr    error
			containers []discovery.Container
		)

		wg.Add(2)

		go func() {
			defer wg.Done()
			scanner := discovery.NewScanner()
			result, scanErr = discovery.Scan(ctx, scanner)
		}()

		go func() {
			defer wg.Done()
			// QueryContainers never errors — an absent/unreachable
			// Docker daemon is a normal state, not a failure, and it
			// has its own internal 500ms budget so it can never make
			// this command hang even if ctx's outer deadline is longer.
			containers = discovery.QueryContainers(ctx)
		}()

		wg.Wait()

		if scanErr != nil {
			fmt.Fprintln(os.Stderr, "ramp ports:", scanErr)
			os.Exit(1)
		}

		containerDBs := discovery.MatchContainerDatabases(containers)
		result.Databases = discovery.MergeDatabases(result.Databases, containerDBs)
		result.DevPorts = discovery.ExcludePortsInDatabases(result.DevPorts, result.Databases)

		output.PrintPorts(result, containers)
	},
}

func init() {
	rootCmd.AddCommand(portsCmd)
}
