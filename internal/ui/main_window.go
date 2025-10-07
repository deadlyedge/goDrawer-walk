package ui

import (
	"log"
	"path/filepath"

	"github.com/deadlyedge/goDrawer/internal/settings"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

func (a *App) createMainWindow() error {
	dragHandler := func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton && a.mainWindow != nil {
			beginWindowDrag(a.mainWindow)
		}
	}

	mwDef := declarative.MainWindow{
		AssignTo:    &a.mainWindow,
		Title:       "goDrawer",
		Icon:        filepath.Join("assets", "drawer.icon.4.ico"),
		MinSize:     declarative.Size{Width: 256, Height: 256},
		Size:        declarative.Size{Width: 256, Height: 256},
		Layout:      declarative.VBox{MarginsZero: true, Spacing: 0},
		OnMouseDown: dragHandler,
		Children: []declarative.Widget{
			declarative.Composite{
				AssignTo:    &a.headerComposite,
				OnMouseDown: dragHandler,
				Layout: declarative.HBox{
					Margins: declarative.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4},
				},
				Children: []declarative.Widget{
					declarative.Label{
						AssignTo:    &a.brandLabel,
						Text:        "goDrawer",
						Font:        declarative.Font{Family: brandFontFamily, PointSize: 16},
						OnMouseDown: dragHandler,
					},
					declarative.HSpacer{},
					declarative.PushButton{
						AssignTo:  &a.settingsButton,
						Text:      "Settings",
						MinSize:   declarative.Size{Width: 96, Height: 34},
						OnClicked: func() { a.openSettingsWindow() },
					},
				},
			},
			declarative.Composite{
				AssignTo:    &a.bodyComposite,
				OnMouseDown: dragHandler,
				Layout: declarative.VBox{
					Margins: declarative.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4},
					Spacing: 8,
				},
				Children: []declarative.Widget{
					declarative.Composite{
						AssignTo:      &a.drawerContainer,
						StretchFactor: 1,
						OnMouseDown:   dragHandler,
						Layout: declarative.VBox{
							MarginsZero: true,
							Spacing:     6,
						},
					},
				},
			},
			declarative.Composite{
				Layout: declarative.HBox{
					Margins: declarative.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4},
				},
				Children: []declarative.Widget{
					declarative.HSpacer{},
					declarative.PushButton{
						AssignTo:  &a.addDrawerButton,
						Text:      "Add Drawer",
						MinSize:   declarative.Size{Width: 120, Height: 38},
						OnClicked: func() { a.onAddDrawer() },
					},
				},
			},
		},
	}

	if err := mwDef.Create(); err != nil {
		return err
	}

	if err := makeWindowBorderless(a.mainWindow); err != nil {
		return err
	}
	if err := makeWindowSemiTransparent(a.mainWindow, a.palette.Alpha); err != nil {
		return err
	}
	if err := hideFromTaskbar(a.mainWindow); err != nil {
		return err
	}

	a.decorateActionButton(a.settingsButton, a.brushes.AccentDark, a.brushes.Accent)
	a.decorateActionButton(a.addDrawerButton, a.brushes.Accent, a.brushes.AccentLight)

	return nil
}

func (a *App) refreshDrawerList() error {
	if a.drawerContainer == nil {
		return nil
	}

	for a.drawerContainer.Children().Len() > 0 {
		child := a.drawerContainer.Children().At(0)
		child.SetParent(nil)
		child.Dispose()
	}

	a.drawerButtons = a.drawerButtons[:0]

	for _, drawer := range a.config.Drawers {
		comp, err := walk.NewComposite(a.drawerContainer)
		if err != nil {
			return err
		}

		layout := walk.NewHBoxLayout()
		layout.SetMargins(walk.Margins{HNear: 0, VNear: 0, HFar: 0, VFar: 0})
		layout.SetSpacing(0)
		comp.SetLayout(layout)
		comp.SetBackground(a.brushes.SurfaceLight)

		btn, err := walk.NewPushButton(comp)
		if err != nil {
			return err
		}
		btn.SetText(drawer.Name)
		btn.SetMinMaxSize(walk.Size{Width: 0, Height: 36}, walk.Size{})
		a.decorateActionButton(btn, a.brushes.SurfaceLight, a.brushes.AccentLight)

		current := drawer
		btn.Clicked().Attach(func() { a.openDrawer(current) })

		a.drawerButtons = append(a.drawerButtons, btn)
	}

	a.refreshButtonStyles()
	return nil
}

func (a *App) onAddDrawer() {
	if a.mainWindow == nil {
		return
	}

	dlg := walk.FileDialog{
		Title: "Select drawer folder",
	}
	ok, err := dlg.ShowBrowseFolder(a.mainWindow)
	if err != nil {
		log.Printf("failed to open folder dialog: %v", err)
		return
	}
	if !ok || dlg.FilePath == "" {
		return
	}

	folder := dlg.FilePath
	name := filepath.Base(folder)
	if name == "" {
		name = folder
	}

	newDrawer := settings.Drawer{
		Name: name,
		Path: folder,
		Size: settings.Size{Width: 420, Height: 360},
	}

	a.config.Drawers = append(a.config.Drawers, newDrawer)

	if err := settings.Update(a.settingsPath, a.config); err != nil {
		log.Printf("failed to persist settings: %v", err)
	}

	if err := a.refreshDrawerList(); err != nil {
		log.Printf("failed to refresh drawer list: %v", err)
	}
}
