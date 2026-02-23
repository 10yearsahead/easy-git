package git

import (
	"fmt"
	"os"
	"strings"
)

// ── Init ──────────────────────────────────────────────────────────────────────

// GitConfig holds user configuration
type GitConfig struct {
	Name  string
	Email string
}

// GetGlobalConfig reads the global git user config
func GetGlobalConfig() GitConfig {
	name, _ := run("config", "--global", "user.name")
	email, _ := run("config", "--global", "user.email")
	return GitConfig{Name: name, Email: email}
}

// SetGlobalConfig sets global git user name and email
func SetGlobalConfig(name, email string) error {
	if _, err := runCombined("config", "--global", "user.name", name); err != nil {
		return err
	}
	_, err := runCombined("config", "--global", "user.email", email)
	return err
}

// Init initializes a git repository in the current directory
func Init(defaultBranch string) (string, error) {
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	return runCombined("init", "-b", defaultBranch)
}

// AddRemote adds a remote origin
func AddRemote(url string) (string, error) {
	return runCombined("remote", "add", "origin", url)
}

// GetRemotes returns configured remotes
func GetRemotes() []string {
	out, err := run("remote", "-v")
	if err != nil || out == "" {
		return nil
	}
	seen := map[string]bool{}
	var remotes []string
	for _, line := range strings.Split(out, "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 && !seen[parts[0]] {
			seen[parts[0]] = true
			remotes = append(remotes, fmt.Sprintf("%s  →  %s", parts[0], parts[1]))
		}
	}
	return remotes
}

// WriteGitignore creates a .gitignore file with the given content
func WriteGitignore(content string) error {
	return os.WriteFile(".gitignore", []byte(content), 0644)
}

// GitignoreExists checks if .gitignore already exists
func GitignoreExists() bool {
	_, err := os.Stat(".gitignore")
	return err == nil
}

// InitialCommit stages everything and makes the first commit
func InitialCommit() (string, error) {
	if err := AddAll(); err != nil {
		return "", err
	}
	return Commit("🎉 Initial commit")
}

// ── .gitignore templates ──────────────────────────────────────────────────────

type GitignoreTemplate struct {
	Label   string
	Key     string
	Content string
}

var GitignoreTemplates = []GitignoreTemplate{
	{
		Label: "Node.js / JavaScript",
		Key:   "node",
		Content: `# Node.js
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*
.pnp/
.pnp.js
.yarn/

# Build
dist/
build/
.next/
out/

# Env
.env
.env.local
.env.*.local

# Logs
logs/
*.log

# OS
.DS_Store
Thumbs.db
`,
	},
	{
		Label: "Python",
		Key:   "python",
		Content: `# Python
__pycache__/
*.py[cod]
*$py.class
*.so
.Python

# Virtualenv
venv/
env/
ENV/
.venv/

# Build / dist
build/
dist/
*.egg-info/
.eggs/

# Env
.env
.env.local

# Jupyter
.ipynb_checkpoints

# mypy / pytest
.mypy_cache/
.pytest_cache/
.coverage
htmlcov/

# OS
.DS_Store
Thumbs.db
`,
	},
	{
		Label: "Go",
		Key:   "go",
		Content: `# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out

# Build output
/bin/
/dist/

# Vendor
/vendor/

# Env
.env
.env.local

# OS
.DS_Store
Thumbs.db
`,
	},
	{
		Label: "Java / Kotlin",
		Key:   "java",
		Content: `# Java / Kotlin
*.class
*.jar
*.war
*.ear
*.nar
hs_err_pid*

# Build
target/
build/
out/
.gradle/
.mvn/

# IDE
.idea/
*.iml
*.iws
*.ipr
.classpath
.project
.settings/

# OS
.DS_Store
Thumbs.db
`,
	},
	{
		Label: "Rust",
		Key:   "rust",
		Content: `# Rust
/target/
Cargo.lock

# Env
.env
.env.local

# OS
.DS_Store
Thumbs.db
`,
	},
	{
		Label: "Flutter / Dart",
		Key:   "flutter",
		Content: `# Flutter / Dart
.dart_tool/
.flutter-plugins
.flutter-plugins-dependencies
.packages
.pub-cache/
.pub/
build/

# Android
**/android/**/gradle-wrapper.jar
**/android/.gradle
**/android/captures/
**/android/gradlew
**/android/gradlew.bat
**/android/local.properties
**/android/**/GeneratedPluginRegistrant.java

# iOS
**/ios/**/*.mode1v3
**/ios/**/*.mode2v3
**/ios/**/*.moved-aside
**/ios/**/*.pbxuser
**/ios/**/*.perspectivev3
**/ios/**/DerivedData/
**/ios/**/.generated/
**/ios/Flutter/App.framework
**/ios/Flutter/Flutter.framework
**/ios/Flutter/Flutter.podspec
**/ios/Flutter/Generated.xcconfig
**/ios/Flutter/app.flx
**/ios/Flutter/app.zip
**/ios/Flutter/flutter_assets/
**/ios/ServiceDefinitions.json
**/ios/Runner/GeneratedPluginRegistrant.*

# Env
.env
.env.local

# OS
.DS_Store
Thumbs.db
`,
	},
	{
		Label: "Generic / Other",
		Key:   "generic",
		Content: `# Compiled and binaries
*.o
*.a
*.out
*.exe
*.dll
*.so
*.dylib

# Environments
.env
.env.local
.env.*.local

# Logs
*.log
logs/

# Build
dist/
build/
out/

# Dependencies
vendor/
node_modules/

# IDE / Editors
.idea/
.vscode/
*.suo
*.ntvs*
*.njsproj
*.sln
*.sw?

# OS
.DS_Store
Thumbs.db
desktop.ini
`,
	},
}

// GetTemplate returns a template by key
func GetTemplate(key string) *GitignoreTemplate {
	for i, t := range GitignoreTemplates {
		if t.Key == key {
			return &GitignoreTemplates[i]
		}
	}
	return nil
}
