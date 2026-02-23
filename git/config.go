package git

import (
	"os"
	"os/exec"
	"strings"
)

// ── Config Management ─────────────────────────────────────────────────────────

// ClearGlobalName removes global user.name
func ClearGlobalName() error {
	_, err := runCombined("config", "--global", "--unset", "user.name")
	return err
}

// ClearGlobalEmail removes global user.email
func ClearGlobalEmail() error {
	_, err := runCombined("config", "--global", "--unset", "user.email")
	return err
}

// ClearCredentials removes stored credentials from the credential helper
func ClearCredentials() error {
	// Try to clear via credential helper
	cmd := exec.Command("git", "credential", "reject")
	cmd.Stdin = strings.NewReader("protocol=https\nhost=github.com\n\n")
	cmd.Run()

	cmd2 := exec.Command("git", "credential", "reject")
	cmd2.Stdin = strings.NewReader("protocol=https\nhost=gitlab.com\n\n")
	cmd2.Run()

	// Also try to clear the credential store file directly
	home, err := os.UserHomeDir()
	if err == nil {
		// Remove .git-credentials file (used by credential.helper=store)
		credFile := home + "/.git-credentials"
		if _, err := os.Stat(credFile); err == nil {
			os.WriteFile(credFile, []byte(""), 0600)
		}
	}

	return nil
}

// ClearAllConfig removes name, email and credentials
func ClearAllConfig() error {
	ClearGlobalName()
	ClearGlobalEmail()
	ClearCredentials()
	return nil
}

// SaveCredentialHelper enables git credential store so tokens are saved
func SaveCredentialHelper() error {
	_, err := runCombined("config", "--global", "credential.helper", "store")
	return err
}

// GetCredentialHelper returns the current credential helper
func GetCredentialHelper() string {
	out, _ := run("config", "--global", "credential.helper")
	return out
}

// HasStoredCredentials checks if there are stored credentials
func HasStoredCredentials() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	credFile := home + "/.git-credentials"
	info, err := os.Stat(credFile)
	if err != nil {
		return false
	}
	return info.Size() > 0
}
