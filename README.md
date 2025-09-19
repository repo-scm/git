# git

[![Build Status](https://github.com/repo-scm/git/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/repo-scm/git/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/repo-scm/git)](https://goreportcard.com/report/github.com/repo-scm/git)
[![License](https://img.shields.io/github/license/repo-scm/git.svg)](https://github.com/repo-scm/git/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/repo-scm/git.svg)](https://github.com/repo-scm/git/tags)



## Introduction

git workspace with overlay



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
  mount: "/path/to/overlay"
sshfs:
  mount: "/path/to/sshfs"
  ports: [
    22,
  ]
```



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
git install

# Show install status
git status
```

#### 2. Create git workspace

```bash
# Create workspace for local repo
git create /local/repo [--name string]

# Create workspace for remote repo
git create user@host:/remote/repo [--name string]
```

> **Notes**: Workspace name is set to `<repo_name>-<7_bit_hash>` in default if `--name string` not set.

#### 3. List git workspace

```bash
# List a workspace
git list <workspace_name>

# List a workspace in verbose mode
git list <workspace_name> --verbose

# List all workspaces
git list

# List all workspaces in verbose mode
git list --verbose
```

#### 4. Run git workspace

```bash
# Run a workspace
git run <workspace_name>
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
git delete <workspace_name>

# Delete all workspaces
git delete --all
```

#### 6. Chat with git workspace

```bash
# Chat with workspace in interactive mode
git chat <workspace_name> [prompt] [--model string]

# Chat with workspace in quiet mode
git chat <workspace_name> [prompt] [--model string] --quiet
```

> **Notes**: Model name is set to `litellm/anthropic/claude-opus-4-20250514` in default if `--model string` not set.

#### 7. MCP for git workspace

```bash
git mcp <workspace_name>
```



## FAQ

### Q: Getting "Connection reset by peer" error when creating workspace

**A:** This error typically occurs due to SSH connectivity issues. Follow these steps to resolve:

1. **Check SSH port configuration**
   - Verify your SSH server is running and note the port
   - Test connectivity: `ssh -p <PORT> user@host 'echo "test"'`
   - Update `~/.repo-scm/git.yaml` to include the correct port:
   ```yaml
   sshfs:
     mount: "/path/to/sshfs"
     ports: [
       2220,  # Specific SSH port
       22,    # Default SSH port
     ]
   ```

2. **Set up SSH key authentication**
   ```bash
   # Generate SSH key pair
   ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519 -N '' -C 'user@localhost'

   # Add public key to authorized_keys
   cat ~/.ssh/id_ed25519.pub >> ~/.ssh/authorized_keys
   chmod 600 ~/.ssh/authorized_keys
   chmod 700 ~/.ssh
   ```

3. **Test SSH connectivity**
   ```bash
   ssh -p <PORT> -o StrictHostKeyChecking=no user@host 'echo "SSH works"'
   ```

### Q: "fusermount3: option allow_other only allowed" error

**A:** Enable the `user_allow_other` option in fuse configuration:

1. **Edit fuse configuration**
   ```bash
   sudo sed -i 's/#user_allow_other/user_allow_other/' /etc/fuse.conf
   ```

2. **Verify the change**
   ```bash
   cat /etc/fuse.conf | grep user_allow_other
   ```

3. **No restart required** - change takes effect immediately

### Q: Workspace creation hangs or times out

**A:** This usually indicates network or authentication issues:

1. **Check SSH server status**
   ```bash
   sudo systemctl status ssh
   # Check which port SSH is listening on
   sudo ss -tlnp | grep ssh
   ```

2. **Test with shorter timeout**
   ```bash
   ssh -o ConnectTimeout=5 -p <PORT> user@host 'echo "quick test"'
   ```

### Q: "Directory not empty" errors when cleaning workspace

**A:** Use the built-in clean command instead of `rm -rf`:

```bash
# Inside workspace - use git clean
git clean build/ temp/

# Outside workspace - use with --force
git clean --force /absolute/path/to/directory
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
