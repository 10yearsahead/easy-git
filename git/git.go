package git

import (
	"os"
	"os/exec"
	"strings"
)

// FileStatus represents a file in the working tree
type FileStatus struct {
	Name   string
	Status string // "staged", "unstaged", "untracked"
	Short  string // git short status code (M, A, D, ?)
}

// RepoStatus holds the full status of the repo
type RepoStatus struct {
	Branch    string
	Staged    []FileStatus
	Unstaged  []FileStatus
	Untracked []FileStatus
	IsRepo    bool
}

// run executes a git command and returns stdout
func run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// runCombined executes a git command and returns combined stdout+stderr
func runCombined(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// IsGitRepo checks if current directory is a git repo
func IsGitRepo() bool {
	_, err := run("rev-parse", "--git-dir")
	return err == nil
}

// CurrentBranch returns the current branch name
func CurrentBranch() string {
	branch, err := run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "unknown"
	}
	return branch
}

// Status returns the full repository status
func Status() RepoStatus {
	if !IsGitRepo() {
		return RepoStatus{IsRepo: false}
	}

	status := RepoStatus{
		IsRepo: true,
		Branch: CurrentBranch(),
	}

	out, err := run("status", "--porcelain")
	if err != nil || out == "" {
		return status
	}

	for _, line := range strings.Split(out, "\n") {
		if len(line) < 3 {
			continue
		}
		x := string(line[0]) // index (staged)
		y := string(line[1]) // worktree (unstaged)
		name := strings.TrimSpace(line[3:])

		// Handle renames (old -> new)
		if strings.Contains(name, " -> ") {
			parts := strings.Split(name, " -> ")
			name = parts[1]
		}

		switch {
		case x == "?" && y == "?":
			status.Untracked = append(status.Untracked, FileStatus{
				Name: name, Status: "untracked", Short: "?",
			})
		default:
			if x != " " && x != "?" {
				status.Staged = append(status.Staged, FileStatus{
					Name: name, Status: "staged", Short: x,
				})
			}
			if y != " " && y != "?" {
				status.Unstaged = append(status.Unstaged, FileStatus{
					Name: name, Status: "unstaged", Short: y,
				})
			}
		}
	}

	return status
}

// Add stages specific files
func Add(files []string) error {
	args := append([]string{"add"}, files...)
	_, err := runCombined(args...)
	return err
}

// AddAll stages all files
func AddAll() error {
	_, err := runCombined("add", "-A")
	return err
}

// Commit creates a commit with the given message
func Commit(message string) (string, error) {
	return runCombined("commit", "-m", message)
}

// Push pushes to origin on the current branch — runs interactively so git can prompt for credentials
func Push() (string, error) {
	branch := CurrentBranch()
	return runInteractive("push", "origin", branch)
}

// Pull pulls from origin — runs interactively so git can prompt for credentials
func Pull() (string, error) {
	return runInteractive("pull")
}

// runInteractive runs a git command connected to the real terminal (stdin/stdout/stderr).
// This allows git to show credential prompts to the user.
func runInteractive(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err.Error(), err
	}
	return "ok", nil
}

// Branches returns all local branches
func Branches() ([]string, error) {
	out, err := run("branch")
	if err != nil {
		return nil, err
	}

	var branches []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "* ")
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// SwitchBranch checks out a branch
func SwitchBranch(name string) (string, error) {
	return runCombined("checkout", name)
}

// CreateBranch creates and switches to a new branch
func CreateBranch(name string) (string, error) {
	return runCombined("checkout", "-b", name)
}

// DeleteBranch deletes a local branch
func DeleteBranch(name string) (string, error) {
	return runCombined("branch", "-d", name)
}

// ShortStatusLabel returns a human-readable label for a git status code
func ShortStatusLabel(code string) string {
	labels := map[string]string{
		"M": "modified",
		"A": "added",
		"D": "deleted",
		"R": "renamed",
		"C": "copied",
		"U": "conflict",
		"?": "new file",
	}
	if l, ok := labels[code]; ok {
		return l
	}
	return code
}
