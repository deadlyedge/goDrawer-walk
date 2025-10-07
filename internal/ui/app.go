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
	drawerButtons   []*walk.PushButton
	drawerWindows   []*drawerWindow

	brushes struct {
		AccentLight  *walk.SolidColorBrush
		Accent       *walk.SolidColorBrush
		AccentDark   *walk.SolidColorBrush
		Background   *walk.SolidColorBrush
		Surface      *walk.SolidColorBrush
		SurfaceLight *walk.SolidColorBrush
	}

	buttonStyles map[*walk.PushButton]buttonStyle
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

	if err := a.createMainWindow(); err != nil {
		return err
	}

	a.mainWindow.Disposing().Attach(func() { a.disposeBrushes() })

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

	a.applyButtonStyle(a.settingsButton)
	a.applyButtonStyle(a.addDrawerButton)
	for _, btn := range a.drawerButtons {
		a.applyButtonStyle(btn)
	}

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

	for _, btn := range a.drawerButtons {
		a.buttonStyles[btn] = buttonStyle{base: a.brushes.SurfaceLight, hover: a.brushes.AccentLight}
	}
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
