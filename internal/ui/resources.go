package ui

import (
	"image/color"
	"math"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"

	"github.com/deadlyedge/goDrawer/internal/settings"
	"github.com/lxn/walk"
)

const (
	frPrivate = 0x10
)

var (
	gdi32                 = syscall.NewLazyDLL("gdi32.dll")
	procAddFontResourceEx = gdi32.NewProc("AddFontResourceExW")
	brandFontOnce         sync.Once
	brandFontErr          error
	brandFontFamily       = "Lobster"
)

// palette defines the computed color set used by the UI.
type palette struct {
	Accent        walk.Color
	AccentLight   walk.Color
	AccentDark    walk.Color
	Background    walk.Color
	Surface       walk.Color
	SurfaceLight  walk.Color
	TextPrimary   walk.Color
	TextSecondary walk.Color
	Alpha         byte
}

func ensureBrandFont(fontPath string) error {
	brandFontOnce.Do(func() {
		absPath, err := filepath.Abs(fontPath)
		if err != nil {
			brandFontErr = err
			return
		}

		ptr, err := syscall.UTF16PtrFromString(absPath)
		if err != nil {
			brandFontErr = err
			return
		}

		r, _, callErr := procAddFontResourceEx.Call(
			uintptr(unsafe.Pointer(ptr)),
			frPrivate,
			0,
		)
		if r == 0 {
			if errno, ok := callErr.(syscall.Errno); ok && errno != 0 {
				brandFontErr = errno
			} else {
				brandFontErr = syscall.EINVAL
			}
		}
	})

	return brandFontErr
}

func buildPalette(theme settings.Theme) palette {
	h := clampFloat(float64(theme.Hue)/360.0, 0, 1)
	s := clampFloat(float64(theme.Saturation)/100.0, 0, 1)
	l := clampFloat(float64(theme.Lightness)/100.0, 0, 1)

	base := hslaToRGBA(h, s, l, 1)
	darker := hslaToRGBA(h, s, clampFloat(l-0.18, 0, 1), 1)
	lighter := hslaToRGBA(h, s, clampFloat(l+0.18, 0, 1), 1)
	background := hslaToRGBA(h, clampFloat(s*0.40, 0, 1), clampFloat(l-0.18, 0, 1), 1)
	surface := hslaToRGBA(h, clampFloat(s*0.45, 0, 1), clampFloat(l-0.05, 0, 1), 1)
	surfaceLight := hslaToRGBA(h, clampFloat(s*0.35, 0, 1), clampFloat(l+0.07, 0, 1), 1)

	return palette{
		Accent:        toWalkColor(base),
		AccentLight:   toWalkColor(lighter),
		AccentDark:    toWalkColor(darker),
		Background:    toWalkColor(background),
		Surface:       toWalkColor(surface),
		SurfaceLight:  toWalkColor(surfaceLight),
		TextPrimary:   walk.RGB(245, 245, 245),
		TextSecondary: walk.RGB(215, 215, 215),
		Alpha:         byte(clampFloat(float64(theme.Alpha)/100.0, 0, 1) * 255),
	}
}

func hslaToRGBA(h, s, l, a float64) color.RGBA {
	var r, g, b float64

	if s == 0 {
		r, g, b = l, l, l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q
		r = hueToRGB(p, q, h+1.0/3.0)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-1.0/3.0)
	}

	return color.RGBA{
		R: uint8(clampFloat(r, 0, 1) * 255),
		G: uint8(clampFloat(g, 0, 1) * 255),
		B: uint8(clampFloat(b, 0, 1) * 255),
		A: uint8(clampFloat(a, 0, 1) * 255),
	}
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	switch {
	case t < 1.0/6.0:
		return p + (q-p)*6*t
	case t < 1.0/2.0:
		return q
	case t < 2.0/3.0:
		return p + (q-p)*(2.0/3.0-t)*6
	default:
		return p
	}
}

func clampFloat(value, min, max float64) float64 {
	return math.Max(min, math.Min(max, value))
}

func toWalkColor(c color.RGBA) walk.Color {
	return walk.RGB(c.R, c.G, c.B)
}
