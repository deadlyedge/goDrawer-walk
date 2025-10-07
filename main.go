package main

import (
	"github.com/deadlyedge/goDrawer/internal/settings"
	"github.com/deadlyedge/goDrawer/internal/ui"
)

func main() {
	// Read and print settings
	if setting, err := settings.Read("goDrawer-settings.toml"); err == nil {
		settings.Print(setting)
	} else {
		panic(err)
	}

	ui.MainWindow()
}
