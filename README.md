# git

[![Build Status](https://github.com/repo-scm/git/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/repo-scm/git/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/repo-scm/git)](https://goreportcard.com/report/github.com/repo-scm/git)
[![License](https://img.shields.io/github/license/repo-scm/git.svg)](https://github.com/repo-scm/git/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/repo-scm/git.svg)](https://github.com/repo-scm/git/tags)



## Introduction

git workspace with copy-on-write



## Prerequisites

```bash
apt update
apt install -y sshfs
```



## Usage

### Create git workspace

```bash
# Create workspace for local repo
git create /local/repo [--name string]

# Create workspace for remote repo
git create user@host:/remote/repo [--name string]
```

> Notes: workspace name is set to `<repo_name>-<7_bit_hash>` in default if `--name string` not set.

### List git workspaces

```bash
git list workspaces
```

### Run git workspace

```bash
git run <workspace_name>
```

### Delete git workspace

```bash
git delete <workspace_name>
```



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



## Playground

TODO



## License

Project License can be found [here](LICENSE).



## Reference

- [cloud-native-build](https://docs.cnb.cool/zh/)
- [git-clone-yyds](https://cloud.tencent.com/developer/article/2456809)
- [git-clone-yyds](https://cnb.cool/cnb/cool/git-clone-yyds)
