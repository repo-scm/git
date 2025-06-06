# git

[![Build Status](https://github.com/repo-scm/git/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/repo-scm/git/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/repo-scm/git)](https://goreportcard.com/report/github.com/repo-scm/git)
[![License](https://img.shields.io/github/license/repo-scm/git.svg)](https://github.com/repo-scm/git/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/repo-scm/git.svg)](https://github.com/repo-scm/git/tags)



## Introduction

git with copy-on-write



## Prerequisites

- Go >= 1.24.0



## Usage

```
Usage:
  git [flags]

Flags:
  -h, --help                help for git
  -m, --mount string        mount path
  -r, --repository string   repository path
  -u, --unmount string      unmount path
  -v, --version             version for git
```



## Example

### Mount

```bash
sudo ./git --mount /mnt/overlay/project --repository /path/to/project

sudo chown -R $USER:$USER /mnt/overlay/project
sudo chown -R $USER:$USER /path/to/cow-project
```

### Test

```bash
cd /mnt/overlay/project

echo "new file" | tee newfile.txt
echo "modified" | tee README.md

git commit -m "project changes"
git push origin main
```

### Unmount

```bash
sudo ./git --unmount /mnt/overlay/project --repository /path/to/project
```



## Overlay

```
/path/to/project
/path/to/cow-project
```

```
/mnt/overlay/
├── repo
│   ├── LICENSE
│   └── README.md
└── work-repo
    └── work
```

- `/path/to/project`: Read-only base layer (lower)
- `/path/to/cow-project`: Read-write layer for changes (upper)
- `/mnt/overlay/work-repo/work`: Temporary working space (work)
- `/mnt/overlay/repo`: The combined view (merged)



## License

Project License can be found [here](LICENSE).



## Reference

- [cloud-native-build](https://docs.cnb.cool/zh/)
- [git-clone-yyds](https://cloud.tencent.com/developer/article/2456809)
- [git-clone-yyds](https://cnb.cool/cnb/cool/git-clone-yyds)
