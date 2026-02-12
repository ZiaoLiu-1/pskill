# pskill

A universal package manager for LLM skills. One tool to install, link, discover, and monitor skills across **Cursor**, **Claude**, **Codex**, and more.

```
pskill — like pnpm, but for AI skills
```

---

## Why pskill?

Every AI coding assistant stores skills in its own directory (`.cursor/skills/`, `.claude/skills/`, `.codex/skills/`...). If you use multiple assistants, you end up with duplicated skills, no central view, and no way to share a skill set across tools.

**pskill** fixes this:

- **Central store** with symlinks — one copy, available everywhere
- **Cross-CLI** — install once, every supported CLI gets it
- **TUI-first** — interactive dashboard, zero commands to memorize
- **Semantic search** — find skills by meaning, not exact name
- **Trending** — discover popular skills from [skillsmp.com](https://skillsmp.com)
- **Usage monitor** — track which skills are actually used
- **Project defaults** — `pskill init` to bootstrap your skill set

## Quick Start

### Install

**Homebrew** (macOS / Linux):

```bash
brew install ZiaoLiu-1/tap/pskill
```

**curl** (any Unix):

```bash
curl -fsSL https://raw.githubusercontent.com/ZiaoLiu-1/pskill/main/scripts/install.sh | sh
```

**npm / npx**:

```bash
npx pskill@latest
```

**From source**:

```bash
git clone https://github.com/ZiaoLiu-1/pskill.git
cd pskill
make build
./bin/pskill
```

### First Run

Launch `pskill` with no arguments to open the interactive TUI. On first run, an onboarding wizard walks you through:

1. Detecting installed CLIs (Cursor, Claude, Codex, Gemini)
2. Scanning existing skills on your system
3. Importing them into the central store
4. Choosing default skills for new projects

```bash
pskill
```

Or run the wizard explicitly:

```bash
pskill init
```

## TUI

Running `pskill` opens a 6-tab dashboard:

| Tab | Key | Description |
|-----|-----|-------------|
| **Dashboard** | `1` | Overview — skill count, detected CLIs, quick actions |
| **My Skills** | `2` | Browse installed skills with search, grouping, detail pane |
| **Discover** | `3` | Semantic search across local index + remote registry |
| **Trending** | `4` | Popular skills from skillsmp.com with sparkline charts |
| **Monitor** | `5` | Usage analytics — top skills, per-CLI breakdown, stale detection |
| **Settings** | `6` | Configure target CLIs, paths, registry, auto-update |

**Navigation**: `tab` / `shift+tab` to cycle tabs, `1`–`6` to jump, `q` to quit.

Each tab shows its own keyboard shortcuts in the bottom help bar.

## CLI Commands

For scripting or quick operations, every feature is also available as a subcommand:

```bash
pskill add <skill-name>          # Install a skill to store + linked CLIs
pskill add <skill> --cli cursor  # Install to specific CLI only
pskill add <skill> --project     # Also record in pskill.yaml

pskill remove <skill-name>       # Unlink from all CLIs
pskill remove <skill> --prune    # Also delete from central store

pskill ls                        # List installed skills
pskill ls --cli cursor           # List skills linked to Cursor
pskill ls --json                 # JSON output for scripting

pskill search "react hooks"      # Semantic search (local index)
pskill search "react" --online   # Also search skillsmp.com

pskill trending                  # Show trending skills
pskill trending --limit 20       # Top 20

pskill scan                      # Scan system for existing skills
pskill scan --import             # Import found skills into store
pskill scan --json               # JSON output

pskill detect                    # Show detected CLIs and skill dirs
pskill detect --json             # JSON output

pskill monitor                   # Open monitor TUI tab directly

pskill init                      # Interactive onboarding wizard
pskill init --no-tui             # Non-interactive setup
```

## How It Works

### Architecture

```
~/.pskill/
├── config.yaml          # Global configuration
├── store/               # Central skill store (single source of truth)
│   ├── frontend-design/
│   │   └── SKILL.md
│   ├── resume-tailoring/
│   │   └── SKILL.md
│   └── ...
├── cache/               # Registry response cache
├── index/               # Bleve full-text search index
└── stats.db             # SQLite usage tracking database
```

### Symlink Strategy

When you install a skill, pskill:

1. Downloads/copies the skill into `~/.pskill/store/<name>/`
2. Creates symlinks from each target CLI's skill directory:

```
~/.cursor/skills/frontend-design → ~/.pskill/store/frontend-design
~/.claude/skills/frontend-design → ~/.pskill/store/frontend-design
~/.codex/skills/frontend-design  → ~/.pskill/store/frontend-design
```

One copy. Every CLI sees it. On Windows, pskill falls back to directory copies when symlink permissions are unavailable.

### Supported CLIs

| CLI | Skill Directory | Status |
|-----|----------------|--------|
| **Cursor** | `~/.cursor/skills/` | Full support |
| **Claude** | `~/.claude/skills/` | Full support |
| **Codex** | `~/.codex/skills/` | Full support |
| **Gemini** | `~/.gemini/` | Detection only |

### Skill Format

Skills follow the `SKILL.md` convention — a Markdown file with optional YAML frontmatter:

```markdown
---
name: my-skill
description: What this skill does
license: MIT
---

# My Skill

Instructions and content for the LLM...
```

The **directory name** is used as the canonical skill identifier (not the `name` field in frontmatter).

## Project Configuration

Running `pskill add --project` or `pskill init` in a project directory creates a `pskill.yaml`:

```yaml
name: my-project
targetClis:
  - cursor
  - claude
defaultSkills:
  - frontend-design
  - create-rule
installed:
  - frontend-design
  - create-rule
  - resume-tailoring
```

This lets you version-control your team's skill set and bootstrap new clones with `pskill init`.

## Global Configuration

Stored at `~/.pskill/config.yaml`:

```yaml
homeDir: ~/.pskill
storeDir: ~/.pskill/store
cacheDir: ~/.pskill/cache
indexDir: ~/.pskill/index
statsDb: ~/.pskill/stats.db
registryUrl: https://skillsmp.com
targetClis:
  - cursor
  - claude
  - codex
defaultSkills: []
autoUpdateTrending: true
```

## Development

### Prerequisites

- Go 1.22+
- Make

### Build & Run

```bash
make build     # Build to ./bin/pskill
make run       # Build and run
make tidy      # go mod tidy
make test      # Run tests
make lint      # Run golangci-lint
```

### Project Structure

```
cmd/pskill/         # Entry point
internal/
├── adapter/         # CLI-specific adapters (Cursor, Claude, Codex, Gemini)
├── cli/             # Cobra command definitions
├── config/          # Global config management (Viper + YAML)
├── detector/        # Detect installed LLM CLIs
├── monitor/         # SQLite usage tracker
├── project/         # Per-project pskill.yaml management
├── registry/        # Remote registry client + HTTP cache
├── scanner/         # Filesystem skill scanner
├── search/          # Bleve full-text search engine
├── skill/           # Skill model + SKILL.md parser
├── store/           # Central store manager + symlink logic
└── tui/             # Bubble Tea TUI (tabs, layout, styles)
scripts/             # curl installer
npm/                 # npm wrapper package for npx distribution
```

## License

MIT License. See [LICENSE](LICENSE) for details.
