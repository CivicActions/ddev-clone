[![add-on registry](https://img.shields.io/badge/DDEV-Add--on_Registry-blue)](https://addons.ddev.com)
[![tests](https://github.com/civicactions/ddev-clone/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/civicactions/ddev-clone/actions/workflows/tests.yml?query=branch%3Amain)
[![last commit](https://img.shields.io/github/last-commit/civicactions/ddev-clone)](https://github.com/civicactions/ddev-clone/commits)
[![release](https://img.shields.io/github/v/release/civicactions/ddev-clone)](https://github.com/civicactions/ddev-clone/releases/latest)

# ddev-clone

A DDEV addon for creating, managing, and removing independent clones of DDEV projects.

Each clone has its own code copy (via git worktree or directory copy), independent Docker volumes, and a separate DDEV project — enabling parallel development without conflicts.

## Installation

```bash
ddev add-on get civicactions/ddev-clone
```

For global installation (available across all projects):

```bash
ddev add-on get --global civicactions/ddev-clone
```

## Quick Start

```bash
# Create a clone of the current project
ddev clone create feature-x

# List all clones
ddev clone list

# Remove a clone
ddev clone remove feature-x

# Clean up stale references
ddev clone prune
```

## Commands

### `ddev clone create <name>`

Creates a fully independent clone of the current DDEV project:

- Creates a code copy (git worktree or directory copy)
- Copies all Docker volumes (database, config, snapshots)
- Configures a new DDEV project via `config.local.yaml`
- Starts the clone and makes it accessible

**Flags:**

| Flag | Description |
|------|-------------|
| `-b, --branch <name>` | Check out an existing branch (worktree strategy only) |
| `--code-strategy <type>` | Code isolation: `worktree` or `copy` (default: auto-detect) |
| `--no-start` | Configure the clone but do not start it |

**Examples:**

```bash
ddev clone create feature-x
ddev clone create feature-x --branch existing-branch
ddev clone create feature-x --code-strategy=copy
ddev clone create feature-x --no-start
```

### `ddev clone list`

Lists all clones of the current project with name, path, branch, strategy, and status.

**Flags:**

| Flag | Description |
|------|-------------|
| `-j, --json-output` | Output in JSON format |

### `ddev clone remove <name>`

Removes a clone and cleans up all Docker resources, code copy, and project registration.

**Flags:**

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation for dirty worktrees/directories |

### `ddev clone prune`

Cleans up clones whose directories have been manually deleted.

**Flags:**

| Flag | Description |
|------|-------------|
| `--dry-run` | Show what would be pruned without taking action |

## Code Strategies

### Worktree (default for git repos)

Uses `git worktree add` to create a linked worktree. Clones share git history with the source project, enabling efficient disk usage and easy branch management.

**Advantages:**

- Disk-efficient (shares git objects)
- Full git history available
- Easy branch switching

**Requirements:**

- Source directory must be a git repository
- `git` must be on PATH

### Copy (default for non-git projects)

Creates a full recursive copy of the project directory using `otiai10/copy` with preserved timestamps, ownership, and symlinks.

**Advantages:**

- Works with any project (git or not)
- Fully independent copy
- No git dependency

**Auto-detection:** If no `--code-strategy` is specified and the project is not a git repository, the copy strategy is used automatically.

## How It Works

1. **Code isolation**: Creates a code copy using the selected strategy
2. **Config override**: Writes `.ddev/config.local.yaml` with the clone's project name
3. **Volume duplication**: Stops the source DB, copies Docker volumes via tar pipe in Alpine container, restarts source DB
4. **Project start**: Runs `ddev start` on the clone (handles container creation, networking, etc.)

## Volumes Cloned

| Volume | When |
|--------|------|
| Database (MariaDB/MySQL/PostgreSQL) | Always |
| Snapshots | If exists |
| DDEV config | If exists |
| Mutagen sync | If Mutagen is enabled |
| Custom service volumes | If Docker Compose overrides exist |

## Troubleshooting

### Clone creation fails with "image not found"

Pre-pull the Alpine image:

```bash
docker pull alpine:latest
```

### "ddev" command not found

Ensure DDEV is installed and on your PATH:

```bash
which ddev
ddev --version
```

### Volume copy takes too long

Large databases may take several minutes. The addon stops only the DB container during copy to minimize downtime.

### Dirty worktree warning on remove

If a clone has uncommitted changes, you'll be prompted. Use `--force` to skip:

```bash
ddev clone remove feature-x --force
```

## Development

### Building

```bash
make build          # Build for current platform
make build-all      # Build for all 6 platforms
```

### Testing

```bash
make test           # Run all tests
make coverage       # Run tests with coverage
```

### Linting

```bash
make lint           # Run golangci-lint
make fmt            # Format code
make vet            # Run go vet
```

## Requirements

- DDEV >= 1.24.0
- Docker
- Git (for worktree strategy)

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.
