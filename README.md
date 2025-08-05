# git

[![Build Status](https://github.com/repo-scm/git/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/repo-scm/git/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/repo-scm/git)](https://goreportcard.com/report/github.com/repo-scm/git)
[![License](https://img.shields.io/github/license/repo-scm/git.svg)](https://github.com/repo-scm/git/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/repo-scm/git.svg)](https://github.com/repo-scm/git/tags)



## Introduction

git workspace with overlay



## Usage

### Prerequisites

```bash
# Install sshfs
sudo apt update
sudo apt install -y sshfs util-linux
```

### Commands

#### 1. Install toolchains

```bash
# Install toolchains
sudo git install

# Show install status
sudo git status
```

#### 2. Create git workspace

```bash
# Create workspace for local repo
sudo git create /local/repo [--name string]

# Create workspace for remote repo
sudo git create user@host:/remote/repo [--name string]
```

> **Notes**: Workspace name is set to `<repo_name>-<7_bit_hash>` in default if `--name string` not set.

#### 3. List git workspaces

```bash
# List a workspace
sudo git list <workspace_name>

# List a workspace in verbose mode
sudo git list <workspace_name> --verbose

# List all workspaces
sudo git list

# List all workspaces in verbose mode
sudo git list --verbose
```

#### 4. Run git workspace

```bash
# Run a workspace
sudo git run <workspace_name>
```

##### 4.1. Clean directories in workspace

When working inside an overlayfs workspace, use the clean command to remove directories safely:

```bash
# Clean specific directories (use inside workspace)
git clean build/
git clean build/ temp/ cache/

# Clean with absolute paths (requires --force flag)
git clean --force /absolute/path/to/directory
```

> **Notes**: The `git clean` command uses overlayfs-aware removal methods to avoid "Directory not empty" errors that can
> occur with standard `rm -rf` commands in overlayfs mounted workspaces.

#### 5. Delete git workspace

```bash
# Delete a workspace
sudo git delete <workspace_name>

# Delete all workspaces
sudo git delete --all
```

#### 6. Chat with git workspace

```bash
# Chat with workspace in interactive mode
sudo git chat <workspace_name> [prompt] [--model string]

# Chat with workspace in quiet mode
sudo git chat <workspace_name> [prompt] [--model string] --quiet
```

> **Notes**: Model name is set to `litellm/anthropic/claude-opus-4-20250514` in default if `--model string` not set.

#### 7. MCP for git workspace

```bash
sudo git mcp <workspace_name>
```



## Settings

[git](https://github.com/repo-scm/git) parameters can be set in the directory `$HOME/.repo-scm/git.yaml`.

An example of settings can be found in [git.yaml](https://github.com/repo-scm/git/blob/main/config/git.yaml).

```yaml
models:
  - provider_name: "litellm"
    api_base: "http://localhost:4000"
    api_key: "noop"
    model_id: "anthropic/claude-opus-4-20250514"
overlay:
  mount: "/mnt/repo-scm/git/overlay"
sshfs:
  mount: "/mnt/repo-scm/git/sshfs"
  ports: [
    22,
  ]
```



## Sandbox

*TBD*



## License

Project License can be found [here](LICENSE).



## Reference

- [claude-code](https://github.com/anthropics/claude-code)
- [cloud-native-build](https://docs.cnb.cool/zh/)
- [gemini-cli](https://github.com/google-gemini/gemini-cli)
- [git-clone-yyds](https://cloud.tencent.com/developer/article/2456809)
- [git-clone-yyds](https://cnb.cool/cnb/cool/git-clone-yyds)
- [openai-codex](https://github.com/openai/codex)
