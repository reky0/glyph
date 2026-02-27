package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// gatherContext collects contextual information about the current directory.
// Returns an empty string if nothing interesting is found.
func gatherContext(dir string) string {
	var parts []string

	if ctx := goContext(dir); ctx != "" {
		parts = append(parts, ctx)
	}
	if ctx := gitContext(dir); ctx != "" {
		parts = append(parts, ctx)
	}
	if ctx := nodeContext(dir); ctx != "" {
		parts = append(parts, ctx)
	}

	return strings.Join(parts, "\n")
}

// goContext reads go.mod and extracts module name + Go version.
func goContext(dir string) string {
	modPath := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		return ""
	}

	var module, goVersion string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			module = strings.TrimPrefix(line, "module ")
		}
		if strings.HasPrefix(line, "go ") {
			goVersion = strings.TrimPrefix(line, "go ")
		}
		if module != "" && goVersion != "" {
			break
		}
	}
	if module == "" {
		return ""
	}
	return "Go module: " + module + " (go " + goVersion + ")"
}

// gitContext returns the current branch and last commit subject.
func gitContext(dir string) string {
	branch, err := runGit(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	commit, err := runGit(dir, "log", "-1", "--pretty=%s")
	if err != nil {
		return "Git branch: " + strings.TrimSpace(branch)
	}
	return "Git branch: " + strings.TrimSpace(branch) + "\nLast commit: " + strings.TrimSpace(commit)
}

// nodeContext reads package.json for project name and top-level dependencies.
func nodeContext(dir string) string {
	pkgPath := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return ""
	}

	var pkg struct {
		Name         string            `json:"name"`
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return ""
	}
	if pkg.Name == "" {
		return ""
	}

	result := "Node project: " + pkg.Name
	if len(pkg.Dependencies) > 0 {
		deps := make([]string, 0, len(pkg.Dependencies))
		for k := range pkg.Dependencies {
			deps = append(deps, k)
			if len(deps) >= 5 {
				break
			}
		}
		result += "\nDependencies: " + strings.Join(deps, ", ")
	}
	return result
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}
