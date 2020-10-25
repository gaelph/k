[![k.supercrabtree.com](https://raw.githubusercontent.com/supercrabtree/k/gh-pages/k-logo.png)](http://k.supercrabtree.com)

This is a golang port of the excellent [supercrabtree/k](https://github.com/supercrabtree/k)

## Directory listings for zsh with git features

**k** makes directory listings more readable, adding a bit of color and some git status information on files and directories.

### Git status on entire repos

![Repository git status](https://raw.githubusercontent.com/supercrabtree/k/gh-pages/repo-dirs.jpg)

### Git status on files within a working tree

![Repository work tree git status](https://raw.githubusercontent.com/supercrabtree/k/gh-pages/inside-work-tree.jpg)

### File weight colours

Files sizes are graded from green for small (< 1k), to red for huge (> 1mb).

**Human readable files sizes**  
Human readable files sizes can be shown by using the `-H` flag, which requires the `numfmt` command to be available. OS X / Darwin does not have a `numfmt` command by default, so GNU coreutils needs to be installed, which provides `gnumfmt` that `k` will also use if available. GNU coreutils can be installed on OS X with [homebrew](http://brew.sh):

```
brew install coreutils
```

![File weight colours](https://raw.githubusercontent.com/supercrabtree/k/gh-pages/file-size-colors.jpg)

### Rotting dates

Dates fade with age.

![Rotting dates](https://raw.githubusercontent.com/supercrabtree/k/gh-pages/dates.jpg)

## Installation

```shell
go install github.com/gaelph/k
```

### Manually

Clone this repository somewhere (~/k for example)

```shell
git clone git@github.com:gaelph/k.git $HOME/k
```

Install it

```shell
cd $HOME/k

go install .
```

## Usage

Hit k in your shell

```shell
k
```

# 😮

## Minimum Requirements

Golang
Git 1.7.2

## Contributors

[gaelph](https://github.com/gaelph)  
Pull requests welcome :smile:

## Thanks

[supercrabtree](https://github.com/supercrabtree) for the original **k**
