# Plan: rdt-cli Go Rewrite

**Date:** 2026-04-06
**Branch:** feat/rewrite

## Goal
Rewrite rdt-cli (Python) as a single Go binary. Zero runtime dependencies. Distributed via GitHub Releases.

## Scope
Read-only commands (no auth required): popular, all, sub, sub-info, read, show, search, user, user-posts, user-comments.
Auth-required commands deferred to a future task.

## Success Criteria
1. `rdt popular` returns post list
2. `rdt sub golang` returns r/golang posts
3. `rdt show 1` after listing reads first post with comments
4. `rdt search "go cli"` returns results
5. `--json` outputs `{"ok":true,"schema_version":"1","data":...}`
6. GitHub Actions builds 4 platform binaries on tag push

## Package Structure
```
cmd/root.go           cobra root + version
cmd/browse.go         popular, all, sub, sub-info, user, user-posts, user-comments
cmd/post.go           read, show
cmd/search.go         search
internal/client/      HTTP client, retry, jitter, fingerprint
internal/models/      Post, Comment, ListingPage, PostDetail, UserProfile, SubredditInfo
internal/parser/      Reddit JSON → typed models
internal/output/      JSON/YAML/text envelope (schema v1)
internal/cache/       index_cache for show N
.github/workflows/release.yml
```

## Docs Impact
- docs/README.md — update after each task group

## Tasks
- [ ] T01: go.mod deps + package skeleton
- [ ] T02: internal/models
- [ ] T03: internal/client (HTTP + retry + jitter)
- [ ] T04: internal/parser
- [ ] T05: internal/output (envelope + text table)
- [ ] T06: internal/cache
- [ ] T07: cmd/root + cmd/browse (popular, all, sub, sub-info)
- [ ] T08: cmd/browse user commands (user, user-posts, user-comments)
- [ ] T09: cmd/post (read, show)
- [ ] T10: cmd/search
- [ ] T11: release.yml GitHub Actions
- [ ] T12: full verification
