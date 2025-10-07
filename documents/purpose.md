# project goDrawer

## purpose

combine 
- functions from https://github.com/deadlyedge/iconDrawer (https://github.com/deadlyedge/iconDrawer/blob/master/README.md) and
- architecture from https://github.com/lxn/walk

make a mini windows desktop drawer app, for clean view of desktop shortcuts management.

## new architecture

base on golang - walk - toml

- go: 1.25
- walk: instead of pyside6 and qt, use walk for ui: https://github.com/lxn/walk
- toml: instead of json, use toml for settings
