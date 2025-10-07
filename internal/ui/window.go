package ui

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

func MainWindow() {
	var (
		mw              *walk.MainWindow
		alphaSlider     *walk.Slider
		alphaValueLabel *walk.Label
	)

	if err := ensureOwnerWindow(); err != nil {
		log.Fatalf("failed to create owner window: %v", err)
	}

	dragHandler := makeDragHandler(&mw)

	windowDef := transparentWindowDefinition(
		&mw,
		&alphaSlider,
		&alphaValueLabel,
		dragHandler,
		closeWindowHandler(&mw),
		createNewWindow,
	)

	if err := windowDef.Create(); err != nil {
		log.Fatalf("failed to create main window: %v", err)
	}

	if err := prepareWindow(mw, alphaSlider, alphaValueLabel); err != nil {
		log.Fatalf("failed to prepare main window: %v", err)
	}

	ni, err := walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatalf("failed to create notify icon: %v", err)
	}
	defer ni.Dispose()
	ni.SetToolTip("Semi-transparent Test Window")
	ni.SetVisible(true)

	action := walk.NewAction()
	action.SetText("&Exit")
	action.Triggered().Attach(func() {
		walk.App().Exit(0)
	})

	ni.ContextMenu().Actions().Add(action)

	mw.Run()
}

func createNewWindow() {
	var (
		newMw              *walk.MainWindow
		newAlphaSlider     *walk.Slider
		newAlphaValueLabel *walk.Label
	)

	dragHandler := makeDragHandler(&newMw)

	windowDef := transparentWindowDefinition(
		&newMw,
		&newAlphaSlider,
		&newAlphaValueLabel,
		dragHandler,
		closeWindowHandler(&newMw),
		createNewWindow,
	)

	if err := windowDef.Create(); err != nil {
		log.Printf("failed to create new window: %v", err)
	} else {
		if err := prepareWindow(newMw, newAlphaSlider, newAlphaValueLabel); err != nil {
			log.Printf("failed to prepare new window: %v", err)
		}

		newMw.Show()
	}
}

func transparentWindowDefinition(
	window **walk.MainWindow,
	slider **walk.Slider,
	label **walk.Label,
	dragHandler func(x, y int, button walk.MouseButton),
	onClose func(),
	onOpenNew func(),
) declarative.MainWindow {
	return declarative.MainWindow{
		AssignTo:    window,
		Title:       "Semi-transparent Test Window",
		MinSize:     declarative.Size{Width: 400, Height: 300},
		Size:        declarative.Size{Width: 400, Height: 300},
		Layout:      declarative.VBox{MarginsZero: true, Spacing: 12},
		OnMouseDown: dragHandler,
		Children: []declarative.Widget{
			windowContent(window, slider, label, dragHandler, onClose, onOpenNew),
		},
	}
}

func windowContent(
	window **walk.MainWindow,
	slider **walk.Slider,
	label **walk.Label,
	dragHandler func(x, y int, button walk.MouseButton),
	onClose func(),
	onOpenNew func(),
) declarative.Composite {
	return declarative.Composite{
		Layout:      declarative.VBox{Margins: declarative.Margins{Left: 12, Top: 12, Right: 12, Bottom: 12}, Spacing: 8},
		OnMouseDown: dragHandler,
		Children: []declarative.Widget{
			declarative.Label{
				Text:        "This window uses Walk and Win32 layered styles for alpha blending.",
				OnMouseDown: dragHandler,
			},
			declarative.Label{
				AssignTo:    label,
				Text:        opacityLabelText(int(targetAlpha)),
				OnMouseDown: dragHandler,
			},
			declarative.Slider{
				AssignTo:       slider,
				MinValue:       50,
				MaxValue:       255,
				Tracking:       true,
				OnValueChanged: opacitySliderHandler(window, slider, label),
			},
			declarative.PushButton{
				Text:      "Close",
				OnClicked: onClose,
			},
			declarative.PushButton{
				Text:      "Open New Window",
				OnClicked: onOpenNew,
			},
		},
	}
}

func opacitySliderHandler(window **walk.MainWindow, slider **walk.Slider, label **walk.Label) func() {
	return func() {
		if slider == nil || *slider == nil {
			return
		}

		current := (*slider).Value()

		if label != nil && *label != nil {
			(*label).SetText(opacityLabelText(current))
		}

		if window != nil && *window != nil {
			if err := makeWindowSemiTransparent(*window, byte(current)); err != nil {
				log.Printf("failed to update window transparency: %v", err)
			}
		}
	}
}

func makeDragHandler(window **walk.MainWindow) func(x, y int, button walk.MouseButton) {
	return func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton && window != nil && *window != nil {
			beginWindowDrag(*window)
		}
	}
}

func closeWindowHandler(window **walk.MainWindow) func() {
	return func() {
		if window != nil && *window != nil {
			(*window).Close()
		}
	}
}

func prepareWindow(mw *walk.MainWindow, slider *walk.Slider, label *walk.Label) error {
	if mw == nil {
		return syscall.EINVAL
	}

	if err := makeWindowBorderless(mw); err != nil {
		return fmt.Errorf("make window borderless: %w", err)
	}

	if slider != nil {
		slider.SetValue(int(targetAlpha))
	}

	if label != nil {
		label.SetText(opacityLabelText(int(targetAlpha)))
	}

	if err := makeWindowSemiTransparent(mw, targetAlpha); err != nil {
		return fmt.Errorf("apply transparency: %w", err)
	}

	if err := hideFromTaskbar(mw); err != nil {
		return fmt.Errorf("hide from taskbar: %w", err)
	}

	return nil
}

func ensureOwnerWindow() error {
	if hwndOwner != 0 {
		return nil
	}

	className, _ := syscall.UTF16PtrFromString("STATIC")

	r, _, err := procCreateWindowEx.Call(
		0,                                  // dwExStyle
		uintptr(unsafe.Pointer(className)), // lpClassName
		0,                                  // lpWindowName
		0,                                  // dwStyle
		0, 0, 0, 0,                         // x, y, nWidth, nHeight
		uintptr(win.HWND_MESSAGE),         // hWndParent
		0,                                 // hMenu
		uintptr(win.GetModuleHandle(nil)), // hInstance
		0,                                 // lpParam
	)
	if r == 0 {
		if errno, ok := err.(syscall.Errno); ok && errno != 0 {
			return errno
		}
		return syscall.EINVAL
	}

	hwndOwner = win.HWND(r)

	return nil
}
