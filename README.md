# easy-git 🌿
> git made simple for humans

Interactive terminal interface for using git without memorizing commands.

---

## ✨ Features

- **🚀 Initialize repository** — complete 5-step wizard: configure name/email, choose main branch, create `.gitignore`, connect remote and make the first commit
- **📋 Visual status** — see the repository state in organized panels (staged, modified, new)
- **✅ Guided commit** — step by step: choose files → write message → confirm
- **🔄 Push / Pull** — send or download changes in one click
- **🌿 Branches** — create, switch or delete branches easily

---

## 🚀 Installation

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- Git installed on the system

### Steps

```bash
# 1. Clone the repository
git clone https://github.com/10yearsahead/easy-git.git
cd easy-git

# 2. Download dependencies
go mod tidy

# 3. Build
go build -o easy-git .

# 4. (Optional) Install globally
sudo mv easy-git /usr/local/bin/
```

### Run directly without installing

```bash
go run .
```

---

## 📦 Dependencies

| Package | Usage |
|---------|-------|
| [bubbletea](https://github.com/charmbracelet/bubbletea) | TUI framework (Model/View/Update) |
| [lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling and colors |
| [bubbles](https://github.com/charmbracelet/bubbles) | Components (spinner, input, etc.) |

---

## 🛠️ Development

```bash
# Run in dev mode
go run .

# Production build
go build -ldflags="-s -w" -o easy-git .

# Cross-platform build
GOOS=linux   GOARCH=amd64 go build -o easy-git-linux .
GOOS=darwin  GOARCH=amd64 go build -o easy-git-mac .
GOOS=windows GOARCH=amd64 go build -o easy-git.exe .
```