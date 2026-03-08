# Devcontainer Go Setup Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Modify the devcontainer to support Go development with lazygit and Powerlevel10k matching the host shell.

**Architecture:** All changes are in `.devcontainer/`. No application code is touched. The Go toolchain is installed via a devcontainer feature; lazygit and powerlevel10k are installed directly in the Dockerfile; p10k config is mounted from the host.

**Tech Stack:** Docker, devcontainer spec, Oh My Zsh, Powerlevel10k, Go 1.24, lazygit

---

### Task 1: Add Go devcontainer feature

**Files:**
- Modify: `.devcontainer/devcontainer.json`

**Step 1: Add Go feature to `features` block**

In `.devcontainer/devcontainer.json`, the current `features` block is:
```json
"features": {
  "ghcr.io/devcontainers/features/github-cli:1": {}
},
```

Change it to:
```json
"features": {
  "ghcr.io/devcontainers/features/github-cli:1": {},
  "ghcr.io/devcontainers/features/go:1": {
    "version": "1.24",
    "golangciLintVersion": "latest"
  }
},
```

**Step 2: Commit**

```bash
git add .devcontainer/devcontainer.json
git commit -m "feat: add Go 1.24 devcontainer feature"
```

---

### Task 2: Fix workspace path

**Files:**
- Modify: `.devcontainer/devcontainer.json`

**Step 1: Update `workspaceMount` and `workspaceFolder`**

Find these two lines in `.devcontainer/devcontainer.json`:
```json
"workspaceMount": "source=${localWorkspaceFolder},target=/workspace,type=bind,consistency=delegated",
"workspaceFolder": "/workspace",
```

Replace with:
```json
"workspaceMount": "source=${localWorkspaceFolder},target=/workspaces/${localWorkspaceFolderBasename},type=bind,consistency=delegated",
"workspaceFolder": "/workspaces/${localWorkspaceFolderBasename}",
```

`${localWorkspaceFolderBasename}` resolves to the folder name on the host (e.g. `go-sqlglot`), so the shell prompt shows the real project name instead of `workspace`.

**Step 2: Commit**

```bash
git add .devcontainer/devcontainer.json
git commit -m "fix: use project folder name as workspace path"
```

---

### Task 3: Mount `.p10k.zsh` from host

**Files:**
- Modify: `.devcontainer/devcontainer.json`

**Step 1: Add `.p10k.zsh` bind mount**

In `.devcontainer/devcontainer.json`, find the `mounts` array. It currently ends with:
```json
"source=${localWorkspaceFolder}/.devcontainer,target=/workspace/.devcontainer,type=bind,readonly"
```

Add a new entry after it (note: also update the `.devcontainer` mount target to match the new workspace path):
```json
"source=${localEnv:HOME}/.p10k.zsh,target=/home/vscode/.p10k.zsh,type=bind,readonly"
```

Also update the `.devcontainer` mount target from `/workspace/.devcontainer` to `/workspaces/${localWorkspaceFolderBasename}/.devcontainer`.

The full updated `mounts` array should look like:
```json
"mounts": [
  "source=claude-code-bashhistory-${devcontainerId},target=/commandhistory,type=volume",
  "source=claude-code-config-${devcontainerId},target=/home/vscode/.claude,type=volume",
  "source=claude-code-gh-${devcontainerId},target=/home/vscode/.config/gh,type=volume",
  "source=${localEnv:HOME}/.gitconfig,target=/home/vscode/.gitconfig,type=bind,readonly",
  "source=${localWorkspaceFolder}/.devcontainer,target=/workspaces/${localWorkspaceFolderBasename}/.devcontainer,type=bind,readonly",
  "source=${localEnv:HOME}/.p10k.zsh,target=/home/vscode/.p10k.zsh,type=bind,readonly"
],
```

**Step 2: Commit**

```bash
git add .devcontainer/devcontainer.json
git commit -m "feat: mount host .p10k.zsh into container"
```

---

### Task 4: Install lazygit in Dockerfile

**Files:**
- Modify: `.devcontainer/Dockerfile`

**Step 1: Add lazygit install block**

In `.devcontainer/Dockerfile`, find the fzf install block (lines ~47-54). Add the lazygit install **after** it, before the `mkdir` step:

```dockerfile
# Install lazygit
ARG LAZYGIT_VERSION=0.44.1
RUN ARCH=$(dpkg --print-architecture) && \
  case "${ARCH}" in \
    amd64) LG_ARCH="Linux_x86_64" ;; \
    arm64) LG_ARCH="Linux_arm64" ;; \
    *) echo "Unsupported architecture: ${ARCH}" && exit 1 ;; \
  esac && \
  curl -fsSL "https://github.com/jesseduffield/lazygit/releases/download/v${LAZYGIT_VERSION}/lazygit_${LAZYGIT_VERSION}_${LG_ARCH}.tar.gz" | tar -xz -C /usr/local/bin lazygit
```

Note: This runs as root (before the `USER vscode` line), so no sudo needed.

**Step 2: Commit**

```bash
git add .devcontainer/Dockerfile
git commit -m "feat: install lazygit in devcontainer"
```

---

### Task 5: Install Powerlevel10k and set ZSH_THEME

**Files:**
- Modify: `.devcontainer/Dockerfile`

**Step 1: Clone powerlevel10k after Oh My Zsh install**

In `.devcontainer/Dockerfile`, find the Oh My Zsh install block:
```dockerfile
# Install Oh My Zsh
ARG ZSH_IN_DOCKER_VERSION=1.2.1
RUN sh -c "$(curl -fsSL https://github.com/deluan/zsh-in-docker/releases/download/v${ZSH_IN_DOCKER_VERSION}/zsh-in-docker.sh)" -- \
  -p git \
  -x
```

Add two `RUN` steps immediately after it:

```dockerfile
# Install Powerlevel10k theme
RUN git clone --depth=1 https://github.com/romkatv/powerlevel10k.git \
  ~/.oh-my-zsh/custom/themes/powerlevel10k

# Set p10k as the ZSH theme
RUN sed -i 's/^ZSH_THEME=.*/ZSH_THEME="powerlevel10k\/powerlevel10k"/' ~/.zshrc
```

**Step 2: Commit**

```bash
git add .devcontainer/Dockerfile
git commit -m "feat: install powerlevel10k theme in devcontainer"
```

---

### Task 6: Configure p10k in `.zshrc`

**Files:**
- Modify: `.devcontainer/.zshrc`

**Step 1: Append p10k config to `.zshrc`**

Add these two lines at the end of `.devcontainer/.zshrc`:

```bash
# Powerlevel10k config (instant prompt disabled — can't be at top of .zshrc in this setup)
typeset -g POWERLEVEL9K_INSTANT_PROMPT=off
[[ ! -f ~/.p10k.zsh ]] || source ~/.p10k.zsh
```

**Step 2: Commit**

```bash
git add .devcontainer/.zshrc
git commit -m "feat: source p10k config in container zshrc"
```

---

## Verification

After all tasks are committed, rebuild the devcontainer and verify:

1. `go version` → shows `go1.24.x`
2. `gopls version` → installed
3. `golangci-lint version` → installed
4. `lazygit --version` → installed
5. `pwd` → shows `/workspaces/go-sqlglot`
6. Shell prompt uses Powerlevel10k styling
7. `which lazygit` → `/usr/local/bin/lazygit`
