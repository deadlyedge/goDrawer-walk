package main

import (
	"log"

	"github.com/deadlyedge/goDrawer/internal/settings"
	"github.com/deadlyedge/goDrawer/internal/ui"
)

func main() {
	const configPath = "goDrawer-settings.toml"

	setting, err := settings.Read(configPath)
	if err != nil {
		log.Fatalf("unable to read settings: %v", err)
	}

	settings.Print(setting)

	ui.MainWindow(setting, configPath)
}
