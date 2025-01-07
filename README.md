# gitclone

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
  -r, --repo-url string      repo url
```



## Settings

*gitclone* parameters can be set in [clone.yml](https://github.com/craftslab/gitclone/blob/main/clone.yml).

An example of configuration in [clone.yml](https://github.com/craftslab/gitclone/blob/main/clone.yml):

```yaml
clone:
  depth: 1
  single_branch: true
mount:
  overlay:
    base_dir: base
    upper_dir: upper
    work_dir: work
    index: off
    merged_dir: merged
```



## License

Project License can be found [here](LICENSE).



## Reference

- [cloud-native-build](https://docs.cnb.cool/zh/)
- [git-clone-yyds](https://cloud.tencent.com/developer/article/2456809)
- [git-clone-yyds](https://cnb.cool/cnb/cool/git-clone-yyds)
