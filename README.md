# gitclone

[![Build Status](https://github.com/craftslab/gitclone/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/craftslab/gitclone/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/craftslab/gitclone)](https://goreportcard.com/report/github.com/craftslab/gitclone)
[![License](https://img.shields.io/github/license/craftslab/gitclone.svg)](https://github.com/craftslab/gitclone/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/craftslab/gitclone.svg)](https://github.com/craftslab/gitclone/tags)



## Introduction

git clone with copy-on-write



## Prerequisites

- Go >= 1.22.0



## Build

```bash
make build
```



## Usage

```
git clone with copy-on-write

Usage:
  clone [flags]

Flags:
  -c, --config-file string   config file
  -d, --dest-dir string      dest dir
  -h, --help                 help for clone
  -b, --repo-branch string   repo branch
  -r, --repo-url string      repo url
  -u, --unmount-dir string   unmount dir
```



## Settings

*gitclone* parameters can be set in [config.yml](https://github.com/craftslab/gitclone/blob/main/config.yml).

An example of configuration in [config.yml](https://github.com/craftslab/gitclone/blob/main/config.yml):

```yaml
clone:
  depth: 1
  single_branch: true
```



## Overlay

```
gitclone
├── lower
├── merged
├── source
├── upper
└── work
```

- `lower`: Read-only base layer(s)
- `merged`: The combined view
- `upper`: Read-write layer for changes
- `work`: Temporary working space



## Example

### Clone

```bash
sudo ./clone -c config.yml -r https://github.com/craftslab/gitclone.git -b main -d gitclone
```

```
gitclone
├── lower
├── merged
│   ├── clone.go
│   ├── clone_test.go
│   ├── config.yml
│   ├── go.mod
│   ├── go.sum
│   ├── LICENSE
│   ├── Makefile
│   ├── README.md
│   └── script
├── source
│   ├── clone.go
│   ├── clone_test.go
│   ├── config.yml
│   ├── go.mod
│   ├── go.sum
│   ├── LICENSE
│   ├── Makefile
│   ├── README.md
│   └── script
├── upper
└── work
```

### Test

```bash
echo "new file" | sudo tee gitclone/merged/newfile.txt
echo "modified" | sudo tee gitclone/merged/README.md
```

### Umount

```bash
sudo ./clone -c config.yml -u gitclone
```



## License

Project License can be found [here](LICENSE).



## Reference

- [cloud-native-build](https://docs.cnb.cool/zh/)
- [git-clone-yyds](https://cloud.tencent.com/developer/article/2456809)
- [git-clone-yyds](https://cnb.cool/cnb/cool/git-clone-yyds)
