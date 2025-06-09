# git

[![Build Status](https://github.com/repo-scm/git/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/repo-scm/git/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/repo-scm/git)](https://goreportcard.com/report/github.com/repo-scm/git)
[![License](https://img.shields.io/github/license/repo-scm/git.svg)](https://github.com/repo-scm/git/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/repo-scm/git.svg)](https://github.com/repo-scm/git/tags)



## Introduction

git with copy-on-write



## Prerequisites

```bash
sudo apt update
sudo apt install -y sshfs
```



## Usage

```
Usage:
  git [flags]

Flags:
  -h, --help                help for git
  -m, --mount string        mount path
  -r, --repository string   repository path (user@host:/remote/repo:/local/repo)
  -s, --sshkey string       sshkey file (/path/to/id_rsa)
  -u, --unmount string      unmount path
  -v, --version             version for git
```



## Example

### 1. Overlay

#### Mount

```bash
sudo ./git --mount /mnt/overlay/repo --repository /path/to/repo

sudo chown -R $USER:$USER /mnt/overlay/repo
sudo chown -R $USER:$USER /path/to/cow-repo
```

#### Test

```bash
cd /mnt/overlay/repo

echo "new file" | tee newfile.txt
echo "modified" | tee README.md

git commit -m "repo changes"
git push origin main
```

#### Unmount

```bash
sudo ./git --unmount /mnt/overlay/repo --repository /path/to/repo
```

### 2. SSHFS and Overlay

#### Config

```bash
cat $HOME/.ssh/config
```

```
Host *
    HostName <host>
    User <user>
    Port 22
    IdentityFile ~/.ssh/id_rsa
```

#### Mount

```bash
sudo ./git --mount /mnt/overlay/repo --repository user@host:/remote/repo:/local/repo --sshkey /path/to/id_rsa

sudo chown -R $USER:$USER /mnt/overlay/repo
sudo chown -R $USER:$USER /path/to/cow-repo
```

#### Test

```bash
cd /mnt/overlay/repo

echo "new file" | tee newfile.txt
echo "modified" | tee README.md

git commit -m "repo changes"
git push origin main
```

#### Unmount

```bash
sudo ./git --unmount /mnt/overlay/repo --repository /local/repo
```



## License

Project License can be found [here](LICENSE).



## Reference

- [cloud-native-build](https://docs.cnb.cool/zh/)
- [git-clone-yyds](https://cloud.tencent.com/developer/article/2456809)
- [git-clone-yyds](https://cnb.cool/cnb/cool/git-clone-yyds)
