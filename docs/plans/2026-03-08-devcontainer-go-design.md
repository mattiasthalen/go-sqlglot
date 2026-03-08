# Devcontainer Go Setup Design

**Date:** 2026-03-08
**Status:** Approved

## Goal

Configure the Claude Code devcontainer for Go development on the `go-sqlglot` project. Adds Go toolchain, lazygit, and Powerlevel10k theme matching the host shell, while fixing the workspace path to show the real project name.

## Changes

### `devcontainer.json`

- Add `ghcr.io/devcontainers/features/go:1` feature with version `1.24` and `golangciLintVersion: latest`. Installs Go, sets `GOPATH`, adds `$GOPATH/bin` to `PATH`, and installs gopls and delve automatically.
- Change `workspaceMount` target and `workspaceFolder` from `/workspace` to `/workspaces/${localWorkspaceFolderBasename}`. This makes the terminal show the real project name (`go-sqlglot`) instead of `workspace`.
- Add read-only bind mount for `~/.p10k.zsh` from host (same pattern as `.gitconfig`).

### `Dockerfile`

- Install **lazygit** from GitHub releases — pinned version, arch-aware binary download, same pattern as fzf and git-delta.
- Clone **powerlevel10k** into `~/.oh-my-zsh/custom/themes/powerlevel10k` after the Oh My Zsh install step.
- Use `sed` to set `ZSH_THEME="powerlevel10k/powerlevel10k"` in `~/.zshrc`.

### `.zshrc` (custom)

Append to end:

```bash
typeset -g POWERLEVEL9K_INSTANT_PROMPT=off
[[ ! -f ~/.p10k.zsh ]] || source ~/.p10k.zsh
```

`POWERLEVEL9K_INSTANT_PROMPT=off` disables the instant prompt optimization (which requires being at the very top of `.zshrc`). The prompt appearance is identical.

## Decisions

- **Go installation:** devcontainer feature with explicit version pin (`1.24`) — avoids Dockerfile complexity while keeping reproducibility.
- **tmux:** already installed, no change needed.
- **VS Code extensions:** omitted — terminal-only workflow.
- **p10k config:** mounted from host rather than copied, so host changes propagate automatically.
