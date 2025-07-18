# git

[![Build Status](https://github.com/repo-scm/git/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/repo-scm/git/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/repo-scm/git)](https://goreportcard.com/report/github.com/repo-scm/git)
[![License](https://img.shields.io/github/license/repo-scm/git.svg)](https://github.com/repo-scm/git/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/repo-scm/git.svg)](https://github.com/repo-scm/git/tags)



## Introduction

git workspace with copy-on-write



## Usage

### Prerequisites

```bash
# Install overlayfs
curl -L https://github.com/containers/fuse-overlayfs/releases/download/v1.15/fuse-overlayfs-x86_64 -o fuse-overlayfs
chmod +x fuse-overlayfs
sudo mv fuse-overlayfs /usr/local/bin/

# Install sshfs
sudo apt update
sudo apt install -y sshfs util-linux
```

### Commands

#### 1. Create git workspace

```bash
# Create workspace for local repo
git create /local/repo [--name string]

# Create workspace for remote repo
git create user@host:/remote/repo [--name string]
```

> **Notes**: Workspace name is set to `<repo_name>-<7_bit_hash>` in default if `--name string` not set.

#### 2. List git workspaces

```bash
# List a workspace
git list <workspace_name>

# List a workspace in verbose mode
git list <workspace_name> --verbose

# List all workspaces
git list --all

# List all workspaces in verbose mode
git list --all --verbose
```

#### 3. Run git workspace

```bash
git run <workspace_name>
```

#### 4. Delete git workspace

```bash
# Delete a workspace
git delete <workspace_name>

# Delete all workspaces
git delete --all
```

#### 5. Chat with git workspace

```bash
# Chat with workspace in interactive mode
git chat <workspace_name> [prompt] [--model string]

# Chat with workspace in quiet mode
git chat <workspace_name> [prompt] [--model string] --quiet
```

> **Notes**: Model name is set to `litellm/anthropic/claude-opus-4-20250514` in default if `--model string` not set.

#### 6. MCP for git workspace

```bash
git mcp <workspace_name>
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
  options: [
    "allow_other,default_permissions,follow_symlinks",
    "cache=yes,kernel_cache,compression=no,big_writes,cache_timeout=115200",
    "Cipher=aes128-ctr,StrictHostKeyChecking=no,UserKnownHostsFile=/dev/null",
  ]
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
