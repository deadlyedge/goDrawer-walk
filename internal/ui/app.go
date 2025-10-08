package ui

import (
	"log"
	"path/filepath"

	"github.com/deadlyedge/goDrawer/internal/settings"
	"github.com/lxn/walk"
)

type App struct {
	settingsPath string
	config       *settings.Settings
	palette      palette

	mainWindow      *walk.MainWindow
	headerComposite *walk.Composite
	drawerContainer *walk.Composite
	bodyComposite   *walk.Composite
	brandLabel      *walk.Label
	settingsButton  *walk.PushButton
	addDrawerButton *walk.PushButton
	drawerItems     []*drawerItemView
	drawerWindows   []*drawerWindow
	hoveredDrawer   *drawerItemView

	brushes struct {
		AccentLight  *walk.SolidColorBrush
		Accent       *walk.SolidColorBrush
		AccentDark   *walk.SolidColorBrush
		Background   *walk.SolidColorBrush
		Surface      *walk.SolidColorBrush
		SurfaceLight *walk.SolidColorBrush
	}

	buttonStyles map[*walk.PushButton]buttonStyle

	notifyIcon  *walk.NotifyIcon
	trayActions struct {
		ShowHide *walk.Action
	}

	images struct {
		Drawer walk.Image
		Tray   *walk.Icon
	}
}

type buttonStyle struct {
	base  *walk.SolidColorBrush
	hover *walk.SolidColorBrush
}

// MainWindow bootstraps the UI using the provided settings.
func MainWindow(cfg *settings.Settings, settingsPath string) {
	app := &App{
		settingsPath: settingsPath,
		config:       cfg,
	}

	if err := app.run(); err != nil {
		log.Fatalf("failed to launch UI: %v", err)
	}
}

func (a *App) run() error {
	if err := ensureBrandFont(filepath.Join("assets", "fonts", "Lobster-Regular.ttf")); err != nil {
		log.Printf("warn: failed to register brand font: %v", err)
	}

	a.palette = buildPalette(a.config.Theme)
	a.buttonStyles = map[*walk.PushButton]buttonStyle{}

	if err := a.createBrushes(); err != nil {
		return err
	}

	if err := a.loadImages(); err != nil {
		log.Printf("warn: failed to load UI images: %v", err)
	}

	if err := a.createMainWindow(); err != nil {
		return err
	}

	a.mainWindow.Disposing().Attach(func() {
		if a.notifyIcon != nil {
			a.notifyIcon.Dispose()
			a.notifyIcon = nil
		}
		a.disposeImages()
		a.disposeBrushes()
	})

	if err := a.setupNotifyIcon(); err != nil {
		log.Printf("warn: failed to create system tray icon: %v", err)
	}

	a.refreshButtonStyles()
	a.applyPalette()
	if err := a.refreshDrawerList(); err != nil {
		return err
	}

	a.mainWindow.Run()
	return nil
}

func (a *App) createBrushes() error {
	var err error

	if a.brushes.Accent, err = walk.NewSolidColorBrush(a.palette.Accent); err != nil {
		return err
	}
	if a.brushes.AccentLight, err = walk.NewSolidColorBrush(a.palette.AccentLight); err != nil {
		return err
	}
	if a.brushes.AccentDark, err = walk.NewSolidColorBrush(a.palette.AccentDark); err != nil {
		return err
	}
	if a.brushes.Background, err = walk.NewSolidColorBrush(a.palette.Background); err != nil {
		return err
	}
	if a.brushes.Surface, err = walk.NewSolidColorBrush(a.palette.Surface); err != nil {
		return err
	}
	if a.brushes.SurfaceLight, err = walk.NewSolidColorBrush(a.palette.SurfaceLight); err != nil {
		return err
	}

	return nil
}

func (a *App) disposeBrushes() {
	if a.brushes.Accent != nil {
		a.brushes.Accent.Dispose()
	}
	if a.brushes.AccentLight != nil {
		a.brushes.AccentLight.Dispose()
	}
	if a.brushes.AccentDark != nil {
		a.brushes.AccentDark.Dispose()
	}
	if a.brushes.Background != nil {
		a.brushes.Background.Dispose()
	}
	if a.brushes.Surface != nil {
		a.brushes.Surface.Dispose()
	}
	if a.brushes.SurfaceLight != nil {
		a.brushes.SurfaceLight.Dispose()
	}

	a.brushes = struct {
		AccentLight  *walk.SolidColorBrush
		Accent       *walk.SolidColorBrush
		AccentDark   *walk.SolidColorBrush
		Background   *walk.SolidColorBrush
		Surface      *walk.SolidColorBrush
		SurfaceLight *walk.SolidColorBrush
	}{}
}

func (a *App) loadImages() error {
	if a.images.Drawer == nil {
		img, err := walk.NewBitmapFromFileForDPI(filepath.Join("assets", "icons", "folder_icon.png"), 72)
		if err != nil {
			return err
		}

		a.images.Drawer = img
	}

	if a.images.Tray == nil {
		if icon, err := walk.NewIconFromFile(filepath.Join("assets", "drawer.icon.4.ico")); err == nil {
			a.images.Tray = icon
		} else {
			log.Printf("warn: failed to load tray icon: %v", err)
		}
	}

	return nil
}

func (a *App) disposeImages() {
	if bmp, ok := a.images.Drawer.(*walk.Bitmap); ok && bmp != nil {
		bmp.Dispose()
	}
	a.images.Drawer = nil
	if a.images.Tray != nil {
		a.images.Tray.Dispose()
		a.images.Tray = nil
	}
}

func (a *App) applyPalette() {
	if a.mainWindow != nil && a.brushes.Background != nil {
		a.mainWindow.SetBackground(a.brushes.Background)
	}

	if a.headerComposite != nil && a.brushes.AccentDark != nil {
		a.headerComposite.SetBackground(a.brushes.AccentDark)
	}

	if a.bodyComposite != nil && a.brushes.Background != nil {
		a.bodyComposite.SetBackground(a.brushes.Background)
	}

	if a.drawerContainer != nil && a.brushes.Surface != nil {
		a.drawerContainer.SetBackground(a.brushes.Surface)
	}

	if a.brandLabel != nil {
		a.brandLabel.SetTextColor(a.palette.TextPrimary)
	}

	for _, item := range a.drawerItems {
		item.applyPalette(item == a.hoveredDrawer)
	}

	a.applyButtonStyle(a.settingsButton)
	a.applyButtonStyle(a.addDrawerButton)

	for _, dw := range a.drawerWindows {
		dw.applyTheme()
	}
}

func (a *App) decorateActionButton(btn *walk.PushButton, base, hover *walk.SolidColorBrush) {
	if btn == nil {
		return
	}

	btn.SetBackground(base)

	a.buttonStyles[btn] = buttonStyle{base: base, hover: hover}
}

func (a *App) applyButtonStyle(btn *walk.PushButton) {
	if btn == nil {
		return
	}

	if style, ok := a.buttonStyles[btn]; ok {
		if style.base != nil {
			btn.SetBackground(style.base)
		}
	}

}

func (a *App) persistDrawerSettings(updated settings.Drawer) {
	if a.config == nil {
		return
	}

	for i := range a.config.Drawers {
		if a.config.Drawers[i].Path == updated.Path {
			a.config.Drawers[i] = updated
			break
		}
	}

	if err := settings.Update(a.settingsPath, a.config); err != nil {
		log.Printf("failed to persist drawer settings: %v", err)
	}
}

func (a *App) setHoveredDrawer(item *drawerItemView) {
	if a.hoveredDrawer == item {
		return
	}

	if a.hoveredDrawer != nil {
		a.hoveredDrawer.applyPalette(false)
	}

	a.hoveredDrawer = item

	if a.hoveredDrawer != nil {
		a.hoveredDrawer.applyPalette(true)
	}
}

func (a *App) refreshButtonStyles() {
	if a.buttonStyles == nil {
		a.buttonStyles = map[*walk.PushButton]buttonStyle{}
	}

	if a.settingsButton != nil {
		a.buttonStyles[a.settingsButton] = buttonStyle{base: a.brushes.AccentDark, hover: a.brushes.Accent}
	}

	if a.addDrawerButton != nil {
		a.buttonStyles[a.addDrawerButton] = buttonStyle{base: a.brushes.Accent, hover: a.brushes.AccentLight}
	}
}

func (a *App) setupNotifyIcon() error {
	if a.mainWindow == nil {
		return nil
	}

	ni, err := walk.NewNotifyIcon(a.mainWindow)
	if err != nil {
		return err
	}

	if a.images.Tray != nil {
		ni.SetIcon(a.images.Tray)
	}
	ni.SetToolTip("goDrawer")

	titleAction := walk.NewAction()
	titleAction.SetText("goDrawer")
	titleAction.Triggered().Attach(func() {
		a.showMainWindow()
	})

	a.trayActions.ShowHide = walk.NewAction()
	a.trayActions.ShowHide.SetText("Hide app")
	a.trayActions.ShowHide.Triggered().Attach(func() {
		a.toggleMainWindowVisibility()
	})

	exitAction := walk.NewAction()
	exitAction.SetText("Exit app")
	exitAction.Triggered().Attach(func() {
		walk.App().Exit(0)
	})

	menu := ni.ContextMenu()
	menu.Actions().Add(titleAction)
	menu.Actions().Add(a.trayActions.ShowHide)
	menu.Actions().Add(exitAction)

	ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			a.toggleMainWindowVisibility()
		}
	})

	if err := ni.SetVisible(true); err != nil {
		ni.Dispose()
		return err
	}

	a.notifyIcon = ni
	return nil
}

func (a *App) showMainWindow() {
	if a.mainWindow == nil {
		return
	}
	a.mainWindow.Show()
	a.mainWindow.BringToTop()
	a.mainWindow.SetFocus()
	if a.trayActions.ShowHide != nil {
		a.trayActions.ShowHide.SetText("Hide app")
	}
}

func (a *App) hideMainWindow() {
	if a.mainWindow == nil {
		return
	}
	a.mainWindow.Hide()
	if a.trayActions.ShowHide != nil {
		a.trayActions.ShowHide.SetText("Show app")
	}
}

func (a *App) toggleMainWindowVisibility() {
	if a.mainWindow == nil {
		return
	}
	if a.mainWindow.Visible() {
		a.hideMainWindow()
	} else {
		a.showMainWindow()
	}
}

func (a *App) drawerItemAt(x, y int) *drawerItemView {
	for _, item := range a.drawerItems {
		b := item.boundsInContainer()
		if x >= b.X && x < b.X+b.Width && y >= b.Y && y < b.Y+b.Height {
			return item
		}
	}
	return nil
}

func (a *App) updateTheme(theme settings.Theme) error {
	a.palette = buildPalette(theme)

	a.disposeBrushes()
	if err := a.createBrushes(); err != nil {
		return err
	}

	a.refreshButtonStyles()
	a.applyPalette()

	if a.mainWindow != nil {
		if err := makeWindowSemiTransparent(a.mainWindow, a.palette.Alpha); err != nil {
			log.Printf("failed to apply transparency to main window: %v", err)
		}
	}

	for _, dw := range a.drawerWindows {
		if dw != nil && dw.window != nil {
			if err := makeWindowSemiTransparent(dw.window, a.palette.Alpha); err != nil {
				log.Printf("failed to apply transparency to drawer window: %v", err)
			}
			dw.applyTheme()
		}
	}

	return nil
}
