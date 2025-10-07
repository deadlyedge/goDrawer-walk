# project goDrawer

## purpose

combine 
- functions from https://github.com/deadlyedge/iconDrawer (https://github.com/deadlyedge/iconDrawer/blob/master/README.md) and
- architecture from https://github.com/wailsapp/wails

make a mini windows desktop drawer app, for clean view of desktop shortcuts management.

## new architecture

base on golang - wails - toml

- go: 1.25
- wails: instead of pyside6 and qt, use wails for ui: https://github.com/wailsapp/wails
- toml: instead of json, use toml for settings

## difficulties

- did wails support translucent window?
- some study of windows tray icon menu functions
- how can this app use ai for optimizing UX
