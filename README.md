# git

[![Build Status](https://github.com/repo-scm/git/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/repo-scm/git/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/repo-scm/git)](https://goreportcard.com/report/github.com/repo-scm/git)
[![License](https://img.shields.io/github/license/repo-scm/git.svg)](https://github.com/repo-scm/git/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/repo-scm/git.svg)](https://github.com/repo-scm/git/tags)



## Introduction

git workspace with copy-on-write



## Planning

Below is a summary of the top level plan items.

Legend of annotations:

| Mark | Description                                       |
|:----:|:--------------------------------------------------|
|  🏃  | work in progress                                  |
|  ✋  | blocked task                                      |
|  🔵  | more investigation required to remove uncertainty |
|  ✅  | completed                                         |

### Commands

✅ Add commands of create, run, list and delete for git workspace [repo-scm/git#3](https://github.com/repo-scm/git/issues/3)  
🏃 Add commands of lint, build and exec for git workspace [repo-scm/git#4](https://github.com/repo-scm/git/issues/4)  
🏃 Add commands of chat and agent for git workspace [repo-scm/git#5](https://github.com/repo-scm/git/issues/5)

### Permission

🔵 Add user to repo-scm group for sudo-less git commands [repo-scm/git#7](https://github.com/repo-scm/git/issues/7)  

### Test

🔵 Add playground for git workspace [repo-scm/git#6](https://github.com/repo-scm/git/issues/6)



## Prerequisites

```bash
apt update
apt install -y sshfs
```



## Usage

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
git list workspaces
```

#### 3. Run git workspace

```bash
git run <workspace_name>
```

#### 4. Delete git workspace

```bash
git delete <workspace_name>
```

### Permission

*TBD*

### Test

*TBD*



## Settings

[git](https://github.com/repo-scm/git) parameters can be set in the directory `$HOME/.repo-scm/git.yaml`.

An example of settings can be found in [git.yaml](https://github.com/repo-scm/git/blob/main/config/git.yaml).

```yaml
overlay:
  mount: "/mnt/repo-scm/git/overlay"
sshfs:
  mount: "/mnt/repo-scm/git/sshfs"
  options:
    - "allow_other,default_permissions,follow_symlinks"
    - "cache=yes,kernel_cache,compression=no,big_writes,cache_timeout=115200"
    - "Cipher=aes128-ctr,StrictHostKeyChecking=no,UserKnownHostsFile=/dev/null"
  ports:
    - 22
```



## License

Project License can be found [here](LICENSE).



## Reference

- [cloud-native-build](https://docs.cnb.cool/zh/)
- [git-clone-yyds](https://cloud.tencent.com/developer/article/2456809)
- [git-clone-yyds](https://cnb.cool/cnb/cool/git-clone-yyds)
