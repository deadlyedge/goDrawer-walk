package ui

import (
	"log"
	"path/filepath"

	"github.com/deadlyedge/goDrawer/internal/settings"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

type drawerItemView struct {
	app    *App
	drawer settings.Drawer
	root   *walk.Composite
	icon   *walk.ImageView
	name   *walk.Label
}

func (item *drawerItemView) applyPalette(hover bool) {
	if item == nil || item.app == nil {
		return
	}

	if hover {
		if item.root != nil && item.app.brushes.AccentLight != nil {
			item.root.SetBackground(item.app.brushes.AccentLight)
		}
		if item.name != nil {
			item.name.SetTextColor(item.app.palette.TextPrimary)
		}
	} else {
		if item.root != nil && item.app.brushes.SurfaceLight != nil {
			item.root.SetBackground(item.app.brushes.SurfaceLight)
		}
		if item.name != nil {
			item.name.SetTextColor(item.app.palette.TextSecondary)
		}
	}
}

func (item *drawerItemView) boundsInContainer() walk.Rectangle {
	if item == nil || item.root == nil {
		return walk.Rectangle{}
	}
	return item.root.Bounds()
}

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
						// text should not be selectable, but it is, and drag handler should work here but it did not.
						OnMouseDown: dragHandler,
					},
					declarative.HSpacer{},
					declarative.PushButton{
						AssignTo:  &a.settingsButton,
						Text:      "Settings",
						Font:      declarative.Font{Family: brandFontFamily, PointSize: 10},
						MinSize:   declarative.Size{Width: 80, Height: 16},
						OnClicked: func() { a.openSettingsWindow() },
					},
				},
			},
			declarative.Composite{
				AssignTo:    &a.bodyComposite,
				// OnMouseDown: dragHandler,
				Layout: declarative.VBox{
					Margins: declarative.Margins{Left: 2, Top: 0, Right: 2, Bottom: 0},
					Spacing: 0,
				},
				Children: []declarative.Widget{
					declarative.Composite{
						AssignTo:      &a.drawerContainer,
						StretchFactor: 1,
						Font:          declarative.Font{PointSize: 12},
						// OnMouseDown:   dragHandler,
						Layout: declarative.VBox{
							MarginsZero: true,
							Spacing:     0,
						},
					},
				},
			},
			declarative.Composite{
				Layout: declarative.HBox{
					Margins: declarative.Margins{Left: 2, Top: 0, Right: 2, Bottom: 0},
				},
				Children: []declarative.Widget{
					declarative.HSpacer{},
					declarative.PushButton{
						AssignTo:  &a.addDrawerButton,
						Text:      "Add Drawer",
						Font:      declarative.Font{Family: brandFontFamily, PointSize: 10},
						MinSize:   declarative.Size{Width: 120, Height: 20},
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

	if a.drawerContainer != nil {
		a.drawerContainer.MouseMove().Attach(func(x, y int, button walk.MouseButton) {
			a.setHoveredDrawer(a.drawerItemAt(x, y))
		})
	}

	if a.bodyComposite != nil {
		a.bodyComposite.MouseMove().Attach(func(int, int, walk.MouseButton) {
			a.setHoveredDrawer(nil)
		})
	}

	if a.headerComposite != nil {
		a.headerComposite.MouseMove().Attach(func(int, int, walk.MouseButton) {
			a.setHoveredDrawer(nil)
		})
	}

	a.mainWindow.MouseMove().Attach(func(int, int, walk.MouseButton) {
		a.setHoveredDrawer(nil)
	})

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

	a.drawerItems = nil
	a.setHoveredDrawer(nil)

	for _, drawer := range a.config.Drawers {
		item, err := a.createDrawerItem(drawer)
		if err != nil {
			return err
		}
		a.drawerItems = append(a.drawerItems, item)
	}

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

func (a *App) createDrawerItem(drawer settings.Drawer) (*drawerItemView, error) {
	comp, err := walk.NewComposite(a.drawerContainer)
	if err != nil {
		return nil, err
	}
	comp.SetDoubleBuffering(true)

	layout := walk.NewHBoxLayout()
	// layout.SetMargins(walk.Margins{HNear: 12, VNear: 6, HFar: 12, VFar: 6})
	layout.SetSpacing(0)
	comp.SetLayout(layout)

	item := &drawerItemView{
		app:    a,
		drawer: drawer,
		root:   comp,
	}

	if a.brushes.SurfaceLight != nil {
		comp.SetBackground(a.brushes.SurfaceLight)
	}

	icon, err := walk.NewImageView(comp)
	if err != nil {
		comp.Dispose()
		return nil, err
	}
	icon.SetMode(walk.ImageViewModeZoom)
	icon.SetMinMaxSize(walk.Size{Width: 28, Height: 28}, walk.Size{})
	if a.images.Drawer != nil {
		if err := icon.SetImage(a.images.Drawer); err != nil {
			log.Printf("warn: failed to set drawer icon: %v", err)
		}
	}
	item.icon = icon

	nameLabel, err := walk.NewLabel(comp)
	if err != nil {
		comp.Dispose()
		return nil, err
	}
	nameLabel.SetText(drawer.Name)
	nameLabel.SetTextAlignment(walk.AlignNear)
	nameLabel.SetTextColor(a.palette.TextSecondary)
	layout.SetStretchFactor(nameLabel, 1)
	item.name = nameLabel

	comp.SetCursor(walk.CursorHand())
	icon.SetCursor(walk.CursorHand())
	nameLabel.SetCursor(walk.CursorHand())

	for _, w := range []walk.Widget{comp, icon, nameLabel} {
		wb := w.AsWindowBase()
		wb.MouseMove().Attach(func(int, int, walk.MouseButton) {
			a.setHoveredDrawer(item)
		})
		wb.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
			if button == walk.LeftButton {
				a.openDrawer(drawer)
			}
		})
	}

	item.applyPalette(false)

	return item, nil
}
