# envbox

Developer-first CLI tool for .env & .env.example management.

[![Go Version](https://img.shields.io/github/go-mod/go-version/sarner/envbox)](go.mod)
[![License](https://img.shields.io/github/license/sarner/envbox)](LICENSE)

## Features

- **Sync** `.env` and `.env.example` automatically
- **Scan** project files for hardcoded secrets
- **CRUD** operations for environment variables
- **Interactive** developer onboarding wizard

## Installation

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/sarner/envbox/main/scripts/install.sh | bash
```

### Windows

```powershell
irm https://raw.githubusercontent.com/sarner/envbox/main/scripts/install.ps1 | iex
```

### From Release

Download the latest release for your platform from [GitHub Releases](https://github.com/sarner/envbox/releases).

## Usage

### Sync new keys

```bash
envbox sync
```

### Check consistency

```bash
envbox check
```

### Manage variables

```bash
envbox get <KEY>          # Get value
envbox set <KEY> <VALUE>  # Set value
envbox unset <KEY>       # Remove variable
envbox list              # List all variables
envbox list -r           # List with unmasked secrets
```

### Scan for secrets

```bash
envbox scan
```

### Interactive setup

```bash
envbox init
```

## Configuration

Create `envbox.toml` in your project root:

```toml
[env]
env_file = ".env"
example_file = ".env.example"

[required]
DB_PASSWORD = { type = "string" }
API_TOKEN = { type = "string" }

[scan]
ignore_paths = [".git", "vendor", "node_modules"]
```

## License

MIT
