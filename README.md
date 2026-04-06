# rdt-cli-go

A Go rewrite of [rdt-cli](https://github.com/public-clis/rdt-cli) (by [@jackwener](https://github.com/jackwener)) — browse Reddit from the terminal. Single binary, zero runtime dependencies.

## Why Go?

The original rdt-cli is written in Python and requires a Python runtime, virtual environments, and pip/uv to install. This Go rewrite compiles to a single static binary that runs anywhere — no Python, no venv, no dependencies.

This project is also an [Agent Skill](https://skill.md). Clone it and use it directly with Claude Code or any compatible AI agent.

## Install

**As a skill** (recommended for AI agents):
```bash
# Clone and use immediately — scripts/run.sh handles everything
git clone https://github.com/Huafucius/rdt-cli-go.git
```

**Manual download:**
```bash
# macOS (Apple Silicon)
curl -sL https://github.com/Huafucius/rdt-cli-go/releases/latest/download/rdt-darwin-arm64 -o rdt
chmod +x rdt

# macOS (Intel)
curl -sL https://github.com/Huafucius/rdt-cli-go/releases/latest/download/rdt-darwin-amd64 -o rdt
chmod +x rdt

# Linux
curl -sL https://github.com/Huafucius/rdt-cli-go/releases/latest/download/rdt-linux-amd64 -o rdt
chmod +x rdt
```

**Build from source:**
```bash
go install github.com/Huafucius/rdt-cli-go@latest
```

## Usage

```bash
# Browse
rdt browse popular --limit 10
rdt browse sub golang -s top -t week
rdt browse sub-info rust --json

# Read posts
rdt post show 1
rdt post read 1abc123 --json

# Search
rdt search "go cli" --limit 50

# User research
rdt browse user spez --json
rdt browse user-posts spez --all --output spez.csv
rdt browse user-comments spez --all --output comments.json

# Bulk export
rdt search "machine learning" --all --output ml.csv
rdt browse sub golang --all --sort new --output golang.json
```

## Structured Output

All commands support `--json` and `--yaml` flags with a consistent envelope:

```json
{
  "ok": true,
  "schema_version": "1",
  "data": { ... }
}
```

## Features

- Browse /r/popular, /r/all, any subreddit
- Read posts with full comment trees
- Search with subreddit filtering
- View user profiles, posts, and comments
- `--all` flag: auto-paginate to collect all results
- `--output file.csv` or `file.json`: export to file
- Index cache: `browse sub X` then `post show 3` to read the 3rd result
- Chrome 133 fingerprinting with Gaussian jitter for anti-detection
- Exponential backoff on rate limits and server errors

## Limitations

- Read-only (no login, no voting, no commenting)
- Public Reddit JSON API only (~1000 items max per endpoint)
- Built-in rate limiting (~1s between requests)

## Acknowledgments

This project is a Go rewrite of [rdt-cli](https://github.com/public-clis/rdt-cli) by [@jackwener](https://github.com/jackwener). The original Python implementation provided the API design, anti-detection strategy, and command structure that this project is based on.

## License

Apache-2.0
