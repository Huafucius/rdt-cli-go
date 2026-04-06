---
name: rdt-cli-go
description: Reddit CLI for browsing feeds, reading posts, searching, viewing users, and exporting data. Use when the user wants to interact with Reddit, scrape posts, or analyze Reddit content. Single binary, no runtime dependencies.
license: Apache-2.0
compatibility: Requires macOS (arm64/amd64), Linux (amd64), or Windows (amd64). No Python or other runtime needed.
metadata:
  author: Huafucius
  version: "0.1.0"
  upstream: https://github.com/public-clis/rdt-cli
---

# rdt-cli-go — Reddit CLI (Go rewrite)

**Binary:** `rdt`
**Install:** automatic via `scripts/run.sh` (downloads pre-built binary on first run)

## Setup

No manual install needed. The skill's `scripts/run.sh` auto-downloads the correct binary for your platform from GitHub Releases on first use.

Manual install (alternative):
```bash
# macOS (Apple Silicon)
curl -sL https://github.com/Huafucius/rdt-cli-go/releases/latest/download/rdt-darwin-arm64 -o rdt && chmod +x rdt
```

## Quick Start

```bash
# All commands go through scripts/run.sh
scripts/run.sh browse popular --limit 10 --json
scripts/run.sh search "golang" --all --output results.csv
scripts/run.sh browse user-posts spez --all --json
```

## Command Reference

### Browsing

| Command | Description | Example |
|---------|-------------|---------|
| `browse popular` | /r/popular | `rdt browse popular --limit 10 --json` |
| `browse all` | /r/all | `rdt browse all --limit 10` |
| `browse sub <name>` | Browse subreddit | `rdt browse sub golang -s top -t week` |
| `browse sub-info <name>` | Subreddit info | `rdt browse sub-info rust --json` |
| `browse user <name>` | User profile | `rdt browse user spez --json` |
| `browse user-posts <name>` | User's posts | `rdt browse user-posts spez --all --json` |
| `browse user-comments <name>` | User's comments | `rdt browse user-comments spez --all --json` |

### Reading

| Command | Description | Example |
|---------|-------------|---------|
| `post read <id>` | Read post + comments | `rdt post read 1abc123 --json` |
| `post show <index>` | Read by index from last listing | `rdt post show 3` |

### Search & Export

| Command | Description | Example |
|---------|-------------|---------|
| `search <query>` | Search Reddit | `rdt search "go cli" --json` |
| `search <query> --sub <name>` | Search in subreddit | `rdt search "error" --sub rust` |
| `search <query> --all` | Fetch all results | `rdt search "ML" --all --output results.csv` |

### Scraping (bulk data)

| Command | Description | Example |
|---------|-------------|---------|
| Any listing + `--all` | Auto-paginate all pages | `rdt browse user-posts kn0thing --all --output posts.csv` |
| Any listing + `--output` | Save to CSV or JSON | `rdt search "harness" --all --output data.json` |

## Listing Flags

All listing commands (browse popular/all/sub, user-posts, user-comments, search) support:

| Flag | Description |
|------|-------------|
| `--json` | JSON output with envelope `{"ok":true,"schema_version":"1","data":...}` |
| `--yaml` | YAML output with envelope |
| `--all` | Auto-paginate until exhausted |
| `--output FILE` | Save to file (.json or .csv, auto-detected by extension) |
| `--limit N` | Items per page (max 100) |
| `--after CURSOR` | Manual pagination cursor |

## Sort Options

- **Listing sort** (`-s`): `hot`, `new`, `top`, `rising`, `controversial`, `best`
- **Search sort** (`-s`): `relevance`, `hot`, `top`, `new`, `comments`
- **Time filter** (`-t`, for top/controversial): `hour`, `day`, `week`, `month`, `year`, `all`
- **Comment sort** (`-s`): `best`, `top`, `new`, `controversial`, `old`, `qa`

## Agent Workflow Examples

### User research
```bash
scripts/run.sh browse user spez --json
scripts/run.sh browse user-posts spez --all --output spez_posts.json
scripts/run.sh browse user-comments spez --all --output spez_comments.csv
```

### Subreddit scrape
```bash
scripts/run.sh browse sub golang --all --sort new --output golang.csv
scripts/run.sh browse sub-info golang --json
```

### Search + export pipeline
```bash
scripts/run.sh search "machine learning" --all --output ml.json
scripts/run.sh search "rust async" --sub rust --limit 50 --json
```

### Browse + read pipeline
```bash
scripts/run.sh browse sub python -s top -t week --limit 5
scripts/run.sh post show 1 --json
```

## Output Schema

Success:
```json
{"ok": true, "schema_version": "1", "data": ...}
```

Error:
```json
{"ok": false, "schema_version": "1", "error": {"code": "not_found", "message": "..."}}
```

## Limitations

- Read-only (no login, upvote, comment, save, subscribe)
- Public Reddit JSON API only
- Built-in rate limiting (~1s between requests with Gaussian jitter)
- Reddit API returns max ~1000 items per listing endpoint
