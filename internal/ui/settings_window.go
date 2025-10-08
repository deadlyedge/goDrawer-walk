package ui

import (
	"fmt"
	"log"

	"github.com/deadlyedge/goDrawer/internal/settings"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

type settingsWindow struct {
	app *App

	window       *walk.MainWindow
	header       *walk.Composite
	body         *walk.Composite
	previewPanel *walk.Composite
	previewLabel *walk.Label

	hueSlider   *walk.Slider
	satSlider   *walk.Slider
	lightSlider *walk.Slider
	alphaSlider *walk.Slider

	hueValue   *walk.Label
	satValue   *walk.Label
	lightValue *walk.Label
	alphaValue *walk.Label

	autostartCheck *walk.CheckBox
	lockCheck      *walk.CheckBox

	previewBrush *walk.SolidColorBrush
}

func (a *App) openSettingsWindow() {
	sw := &settingsWindow{app: a}
	if err := sw.open(); err != nil {
		log.Printf("failed to open settings window: %v", err)
	}
}

func (sw *settingsWindow) open() error {
	theme := sw.app.config.Theme

	dragHandler := func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton && sw.window != nil {
			beginWindowDrag(sw.window)
		}
	}

	mwDef := declarative.MainWindow{
		AssignTo:    &sw.window,
		Title:       "Settings",
		MinSize:     declarative.Size{Width: 300, Height: 400},
		Size:        declarative.Size{Width: 300, Height: 400},
		Layout:      declarative.VBox{MarginsZero: true, Spacing: 0},
		OnMouseDown: dragHandler,
		Children: []declarative.Widget{
			// declarative.Composite{
			// 	AssignTo:    &sw.header,
			// 	OnMouseDown: dragHandler,
			// 	Layout: declarative.HBox{
			// 		Margins: declarative.Margins{Left: 0, Top: 0, Right: 0, Bottom: 0},
			// 	},
			// 	Children: []declarative.Widget{
			// 		declarative.Label{
			// 			Text:        "Background (HSLA)",
			// 			Font:        declarative.Font{Family: brandFontFamily, PointSize: 12},
			// 			OnMouseDown: dragHandler,
			// 		},
			// 	},
			// },
			declarative.Composite{
				AssignTo: &sw.body,
				// OnMouseDown: dragHandler,
				Layout: declarative.VBox{
					Margins: declarative.Margins{Left: 0, Top: 0, Right: 0, Bottom: 0},
					Spacing: 0,
				},
				Children: []declarative.Widget{
					sw.sliderRow("Hue", 0, 360, theme.Hue, &sw.hueSlider, &sw.hueValue, dragHandler, sw.updatePreview),
					sw.sliderRow("Saturation", 0, 100, theme.Saturation, &sw.satSlider, &sw.satValue, dragHandler, sw.updatePreview),
					sw.sliderRow("Lightness", 0, 100, theme.Lightness, &sw.lightSlider, &sw.lightValue, dragHandler, sw.updatePreview),
					sw.sliderRow("Alpha", 0, 100, theme.Alpha, &sw.alphaSlider, &sw.alphaValue, dragHandler, sw.updatePreview),
					declarative.Composite{
						AssignTo:    &sw.previewPanel,
						OnMouseDown: dragHandler,
						Layout: declarative.VBox{
							Margins: declarative.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4},
							Spacing: 4,
						},
						Children: []declarative.Widget{
							declarative.Label{Text: "Preview"},
							declarative.Label{AssignTo: &sw.previewLabel},
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{
							Margins: declarative.Margins{Left: 0, Top: 4, Right: 0, Bottom: 0},
							Spacing: 4,
						},
						Children: []declarative.Widget{
							declarative.Label{Text: "Startup"},
							declarative.CheckBox{
								AssignTo: &sw.autostartCheck,
								Text:     "Launch with Windows",
								Checked:  sw.app.config.Startup.StartWithWindows,
							},
							declarative.CheckBox{
								AssignTo: &sw.lockCheck,
								Text:     "Lock main window position",
								Checked:  sw.app.config.Startup.WindowLocked,
							},
						},
					},
				},
			},
			declarative.Composite{
				Layout: declarative.HBox{
					Margins: declarative.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4},
					Spacing: 4,
				},
				Children: []declarative.Widget{
					declarative.PushButton{
						Text:      "OK",
						OnClicked: func() { sw.applyAndClose() },
					},
					declarative.PushButton{
						Text: "Shutdown APP",
						OnClicked: func() {
							sw.window.Close()
							walk.App().Exit(0)
						},
					},
					declarative.HSpacer{},
					declarative.PushButton{
						Text:      "Cancel",
						OnClicked: func() { sw.window.Close() },
					},
				},
			},
		},
	}

	if err := mwDef.Create(); err != nil {
		return err
	}

	// if err := makeWindowBorderless(sw.window); err != nil {
	// 	return err
	// }
	if err := hideFromTaskbar(sw.window); err != nil {
		return err
	}

	sw.applyTheme()
	sw.updatePreview()

	sw.window.Disposing().Attach(func() {
		if sw.previewBrush != nil {
			sw.previewBrush.Dispose()
		}
	})

	sw.window.Show()
	return nil
}

func (sw *settingsWindow) sliderRow(title string, min, max, value int, slider **walk.Slider, valueLabel **walk.Label, dragHandler func(x, y int, button walk.MouseButton), onChange func()) declarative.Widget {
	return declarative.Composite{
		OnMouseDown: dragHandler,
		Layout:      declarative.VBox{Spacing: 4},
		Children: []declarative.Widget{
			declarative.Label{Text: title},
			declarative.Composite{
				Layout: declarative.HBox{Spacing: 6},
				Children: []declarative.Widget{
					declarative.Slider{
						AssignTo:       slider,
						MinValue:       min,
						MaxValue:       max,
						Value:          value,
						Tracking:       true,
						OnValueChanged: onChange,
					},
					declarative.Label{AssignTo: valueLabel},
				},
			},
		},
	}
}

func (sw *settingsWindow) applyTheme() {
	if sw.window != nil && sw.app.brushes.Background != nil {
		sw.window.SetBackground(sw.app.brushes.Background)
	}
	if sw.header != nil && sw.app.brushes.AccentDark != nil {
		sw.header.SetBackground(sw.app.brushes.AccentDark)
	}
	if sw.body != nil && sw.app.brushes.Surface != nil {
		sw.body.SetBackground(sw.app.brushes.Surface)
	}
	if sw.previewPanel != nil && sw.app.brushes.SurfaceLight != nil && sw.previewBrush == nil {
		sw.previewPanel.SetBackground(sw.app.brushes.SurfaceLight)
	}
}

func (sw *settingsWindow) updatePreview() {
	if sw.hueSlider == nil || sw.satSlider == nil || sw.lightSlider == nil || sw.alphaSlider == nil {
		return
	}

	hue := sw.hueSlider.Value()
	sat := sw.satSlider.Value()
	light := sw.lightSlider.Value()
	alpha := sw.alphaSlider.Value()

	sw.setSliderValue(sw.hueValue, fmt.Sprintf("%d", hue))
	sw.setSliderValue(sw.satValue, fmt.Sprintf("%d%%", sat))
	sw.setSliderValue(sw.lightValue, fmt.Sprintf("%d%%", light))
	sw.setSliderValue(sw.alphaValue, fmt.Sprintf("%d%%", alpha))

	previewTheme := settings.Theme{Hue: hue, Saturation: sat, Lightness: light, Alpha: alpha}
	palette := buildPalette(previewTheme)

	if sw.previewBrush != nil {
		sw.previewBrush.Dispose()
	}
	brush, err := walk.NewSolidColorBrush(palette.Accent)
	if err == nil {
		sw.previewBrush = brush
		if sw.previewPanel != nil {
			sw.previewPanel.SetBackground(brush)
		}
	}

	if sw.previewLabel != nil {
		sw.previewLabel.SetText(fmt.Sprintf("hsla(%d, %d%%, %d%%, %.2f)", hue, sat, light, float64(alpha)/100.0))
	}
}

func (sw *settingsWindow) setSliderValue(label *walk.Label, value string) {
	if label != nil {
		label.SetText(value)
	}
}

func (sw *settingsWindow) applyAndClose() {
	if sw.hueSlider == nil {
		return
	}

	updatedTheme := settings.Theme{
		Hue:        sw.hueSlider.Value(),
		Saturation: sw.satSlider.Value(),
		Lightness:  sw.lightSlider.Value(),
		Alpha:      sw.alphaSlider.Value(),
	}

	sw.app.config.Theme = updatedTheme
	sw.app.config.Startup.StartWithWindows = sw.autostartCheck.Checked()
	sw.app.config.Startup.WindowLocked = sw.lockCheck.Checked()

	if err := sw.app.updateTheme(updatedTheme); err != nil {
		log.Printf("failed to apply theme: %v", err)
	}

	if err := settings.Update(sw.app.settingsPath, sw.app.config); err != nil {
		log.Printf("failed to persist settings: %v", err)
	}

	sw.window.Close()
}
