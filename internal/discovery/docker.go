package discovery

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// Container is one running Docker container, as much as ramp cares
// about: identity, image, and its published host ports.
type Container struct {
	ID    string
	Name  string
	Image string
	Ports []int
}

type dockerContainerJSON struct {
	ID     string   `json:"Id"`
	Names  []string `json:"Names"`
	Image  string   `json:"Image"`
	Ports  []struct {
		PublicPort int `json:"PublicPort"`
	} `json:"Ports"`
}

// dockerSocketCandidates lists Docker daemon socket paths to try, in
// order. Docker Desktop for Mac has used different default locations
// across versions, so we check the common ones rather than assuming.
func dockerSocketCandidates() []string {
	if host := os.Getenv("DOCKER_HOST"); strings.HasPrefix(host, "unix://") {
		return []string{strings.TrimPrefix(host, "unix://")}
	}
	home, _ := os.UserHomeDir()
	candidates := []string{"/var/run/docker.sock"}
	if home != "" {
		candidates = append(candidates, home+"/.docker/run/docker.sock")
	}
	return candidates
}

func dockerHTTPClient(socketPath string) *http.Client {
	return &http.Client{
		Timeout: 500 * time.Millisecond, // short, dedicated budget — never let a stuck daemon block ramp ports
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "unix", socketPath)
			},
		},
	}
}

// QueryContainers lists running containers from the local Docker
// daemon. Returns nil, never an error, if Docker isn't installed,
// isn't running, or the socket doesn't respond in time — an absent
// Docker daemon is a normal, expected state, not a failure.
func QueryContainers(ctx context.Context) []Container {
	for _, sock := range dockerSocketCandidates() {
		if _, err := os.Stat(sock); err != nil {
			continue
		}

		client := dockerHTTPClient(sock)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://unix/containers/json", nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		var raw []dockerContainerJSON
		if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
			continue
		}
		return parseContainers(raw)
	}
	return nil
}

// parseContainers converts Docker's raw API shape into Container.
// Pure function, no I/O — independently unit-testable without a real
// daemon, matching this package's I/O-vs-logic split throughout.
func parseContainers(raw []dockerContainerJSON) []Container {
	containers := make([]Container, 0, len(raw))
	for _, c := range raw {
		name := c.ID
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		var ports []int
		seen := make(map[int]bool)
		for _, p := range c.Ports {
			if p.PublicPort > 0 && !seen[p.PublicPort] {
				seen[p.PublicPort] = true
				ports = append(ports, p.PublicPort)
			}
		}

		containers = append(containers, Container{ID: c.ID, Name: name, Image: c.Image, Ports: ports})
	}
	return containers
}

// dbImageSignature maps an image-name substring to a display name,
// mirroring knownDatabases but matched against Docker image tags
// rather than host executable names.
type dbImageSignature struct {
	imageContains []string
	displayName   string
	defaultPort   int
}

var knownDatabaseImages = []dbImageSignature{
	{imageContains: []string{"postgres"}, displayName: "PostgreSQL", defaultPort: 5432},
	{imageContains: []string{"redis"}, displayName: "Redis", defaultPort: 6379},
	{imageContains: []string{"mongo"}, displayName: "MongoDB", defaultPort: 27017},
	{imageContains: []string{"mysql"}, displayName: "MySQL", defaultPort: 3306},
	{imageContains: []string{"mariadb"}, displayName: "MariaDB", defaultPort: 3306},
}

// MatchContainerDatabases scans containers for known database images.
// On macOS this is the ONLY way a containerized database is detected
// at all — Docker Desktop runs containers inside a Linux VM, invisible
// to host-level process scanning (libproc sees none of it).
func MatchContainerDatabases(containers []Container) []DatabaseMatch {
	var matches []DatabaseMatch

	for _, c := range containers {
		image := strings.ToLower(c.Image)
		for _, sig := range knownDatabaseImages {
			matched := false
			for _, substr := range sig.imageContains {
				if strings.Contains(image, substr) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}

			port := sig.defaultPort
			if len(c.Ports) > 0 {
				port = pickDatabasePort(c.Ports, sig.defaultPort)
			}

			matches = append(matches, DatabaseMatch{Name: sig.displayName, Port: port, Source: "Docker"})
			break
		}
	}
	return matches
}

// MergeDatabases combines natively-detected and container-detected
// database matches for display.
func MergeDatabases(native, containerized []DatabaseMatch) []DatabaseMatch {
	return append(native, containerized...)
}