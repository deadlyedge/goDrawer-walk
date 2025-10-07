package ui

import (
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

const (
	lwaAlpha       uint32 = 0x2
	wsExAppWindow  uint32 = 0x00040000
	wsExToolWindow uint32 = 0x00000080
)

var (
	user32                         = syscall.NewLazyDLL("user32.dll")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procCreateWindowEx             = user32.NewProc("CreateWindowExW")
	hwndOwner                      win.HWND
)

func makeWindowBorderless(mw *walk.MainWindow) error {
	hwnd := mw.Handle()
	if hwnd == 0 {
		return syscall.EINVAL
	}

	style := uint32(win.GetWindowLong(hwnd, win.GWL_STYLE))
	style &^= uint32(win.WS_CAPTION | win.WS_THICKFRAME | win.WS_MINIMIZE | win.WS_MAXIMIZE | win.WS_SYSMENU)
	style |= uint32(win.WS_POPUP)

	win.SetLastError(0)
	if prev := win.SetWindowLong(hwnd, win.GWL_STYLE, int32(style)); prev == 0 {
		if err := syscall.GetLastError(); err != nil && err != syscall.Errno(0) {
			return err
		}
	}

	if !win.SetWindowPos(hwnd, 0, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOZORDER|win.SWP_FRAMECHANGED|win.SWP_NOACTIVATE) {
		return syscall.GetLastError()
	}

	return nil
}

func makeWindowSemiTransparent(mw *walk.MainWindow, alpha byte) error {
	hwnd := mw.Handle()
	if hwnd == 0 {
		return syscall.EINVAL
	}

	exStyle := win.GetWindowLong(hwnd, win.GWL_EXSTYLE)
	win.SetWindowLong(hwnd, win.GWL_EXSTYLE, exStyle|win.WS_EX_LAYERED)

	return setLayeredWindowAttributes(hwnd, win.COLORREF(0), alpha, lwaAlpha)
}

func setLayeredWindowAttributes(hwnd win.HWND, key win.COLORREF, alpha byte, flags uint32) error {
	r1, _, err := procSetLayeredWindowAttributes.Call(
		uintptr(hwnd),
		uintptr(key),
		uintptr(alpha),
		uintptr(flags),
	)
	if r1 == 0 {
		if errno, ok := err.(syscall.Errno); ok && errno != 0 {
			return errno
		}
		return syscall.EINVAL
	}

	return nil
}

func beginWindowDrag(mw *walk.MainWindow) {
	hwnd := mw.Handle()
	if hwnd == 0 {
		return
	}

	win.ReleaseCapture()
	win.SendMessage(hwnd, win.WM_NCLBUTTONDOWN, uintptr(win.HTCAPTION), 0)
}

func hideFromTaskbar(mw *walk.MainWindow) error {
	if mw == nil {
		return syscall.EINVAL
	}

	hwnd := win.HWND(mw.Handle())
	if hwnd == 0 {
		return syscall.EINVAL
	}

	if err := ensureOwnerWindow(); err != nil {
		return err
	}

	if err := setWindowLongPtrWithError(hwnd, win.GWL_HWNDPARENT, uintptr(hwndOwner)); err != nil {
		return err
	}

	exStyle := uint32(win.GetWindowLong(hwnd, win.GWL_EXSTYLE))
	exStyle &^= wsExAppWindow
	exStyle |= wsExToolWindow

	if err := setWindowLongWithError(hwnd, win.GWL_EXSTYLE, int32(exStyle)); err != nil {
		return err
	}

	if !win.SetWindowPos(hwnd, 0, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOZORDER|win.SWP_FRAMECHANGED|win.SWP_NOACTIVATE) {
		return syscall.GetLastError()
	}

	return nil
}

func ensureOwnerWindow() error {
	if hwndOwner != 0 {
		return nil
	}

	className, _ := syscall.UTF16PtrFromString("STATIC")

	r, _, err := procCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		0,
		0,
		0, 0, 0, 0,
		uintptr(win.HWND_MESSAGE),
		0,
		uintptr(win.GetModuleHandle(nil)),
		0,
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

func setWindowLongPtrWithError(hwnd win.HWND, index int, value uintptr) error {
	win.SetLastError(0)
	prev := win.SetWindowLongPtr(hwnd, index, value)
	if prev == 0 {
		if err := syscall.GetLastError(); err != syscall.Errno(0) {
			return err
		}
	}
	return nil
}

func setWindowLongWithError(hwnd win.HWND, index int32, value int32) error {
	win.SetLastError(0)
	prev := win.SetWindowLong(hwnd, index, value)
	if prev == 0 {
		if err := syscall.GetLastError(); err != syscall.Errno(0) {
			return err
		}
	}
	return nil
}
