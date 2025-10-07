package ui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"github.com/deadlyedge/goDrawer/internal/settings"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

type drawerWindow struct {
	app         *App
	drawer      settings.Drawer
	window      *walk.MainWindow
	header      *walk.Composite
	pathLabel   *walk.Label
	tableView   *walk.TableView
	model       *fileTableModel
	currentPath string
}

type fileItem struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

type fileTableModel struct {
	walk.TableModelBase
	items []fileItem
}

func (m *fileTableModel) RowCount() int {
	return len(m.items)
}

func (m *fileTableModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(m.items) {
		return nil
	}

	item := m.items[row]
	switch col {
	case 0:
		return item.Name
	case 1:
		if item.IsDir {
			return "Folder"
		}
		return fmt.Sprintf("%d KB", item.Size/1024)
	case 2:
		return item.ModTime.Format("2006-01-02 15:04")
	default:
		return ""
	}
}

func (m *fileTableModel) Sort(col int, order walk.SortOrder) error {
	less := func(i, j int) bool {
		lhs := m.items[i]
		rhs := m.items[j]

		switch col {
		case 0:
			if lhs.IsDir != rhs.IsDir {
				return lhs.IsDir
			}
			if order == walk.SortAscending {
				return lhs.Name < rhs.Name
			}
			return lhs.Name > rhs.Name
		case 1:
			if lhs.IsDir != rhs.IsDir {
				return lhs.IsDir
			}
			if order == walk.SortAscending {
				return lhs.Size < rhs.Size
			}
			return lhs.Size > rhs.Size
		case 2:
			if order == walk.SortAscending {
				return lhs.ModTime.Before(rhs.ModTime)
			}
			return lhs.ModTime.After(rhs.ModTime)
		default:
			return true
		}
	}

	sort.SliceStable(m.items, less)
	m.PublishRowsReset()
	return nil
}

func (m *fileTableModel) Reset(items []fileItem) {
	m.items = items
	m.PublishRowsReset()
}

func (a *App) openDrawer(drawer settings.Drawer) {
	dw := &drawerWindow{
		app:    a,
		drawer: drawer,
		model:  &fileTableModel{},
	}

	if err := dw.open(); err != nil {
		log.Printf("failed to open drawer %s: %v", drawer.Name, err)
		return
	}

	a.drawerWindows = append(a.drawerWindows, dw)
}

func (dw *drawerWindow) open() error {
	dw.currentPath = dw.drawer.Path

	dragHandler := func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton && dw.window != nil {
			beginWindowDrag(dw.window)
		}
	}

	mwDef := declarative.MainWindow{
		AssignTo:    &dw.window,
		Title:       fmt.Sprintf("%s - goDrawer", dw.drawer.Name),
		MinSize:     declarative.Size{Width: dw.drawer.Size.Width, Height: dw.drawer.Size.Height},
		Size:        declarative.Size{Width: dw.drawer.Size.Width, Height: dw.drawer.Size.Height},
		Layout:      declarative.VBox{MarginsZero: true, Spacing: 0},
		OnMouseDown: dragHandler,
		Children: []declarative.Widget{
			declarative.Composite{
				AssignTo:    &dw.header,
				OnMouseDown: dragHandler,
				Layout: declarative.HBox{
					Margins: declarative.Margins{Left: 16, Top: 12, Right: 16, Bottom: 8},
				},
				Children: []declarative.Widget{
					declarative.PushButton{
						Text:      "Up",
						MinSize:   declarative.Size{Width: 48, Height: 30},
						OnClicked: func() { dw.goUp() },
					},
					declarative.HSpacer{},
					declarative.Label{
						AssignTo:    &dw.pathLabel,
						Text:        dw.currentPath,
						OnMouseDown: dragHandler,
					},
				},
			},
			declarative.TableView{
				AssignTo:            &dw.tableView,
				Columns:             []declarative.TableViewColumn{{Title: "Name", Width: 220}, {Title: "Info", Width: 90}, {Title: "Modified", Width: 140}},
				LastColumnStretched: true,
				OnItemActivated:     func() { dw.openSelected() },
			},
		},
	}

	if err := mwDef.Create(); err != nil {
		return err
	}

	dw.tableView.SetModel(dw.model)

	if err := makeWindowBorderless(dw.window); err != nil {
		return err
	}
	if err := makeWindowSemiTransparent(dw.window, dw.app.palette.Alpha); err != nil {
		return err
	}
	if err := hideFromTaskbar(dw.window); err != nil {
		return err
	}

	dw.applyTheme()

	if err := dw.loadDirectory(dw.currentPath); err != nil {
		return err
	}

	dw.window.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		dw.saveSize()
		dw.app.persistDrawerSettings(dw.drawer)
	})
	dw.window.Disposing().Attach(func() {
		dw.app.unregisterDrawer(dw)
	})

	dw.window.Show()
	return nil
}

func (dw *drawerWindow) applyTheme() {
	if dw.window != nil && dw.app.brushes.Background != nil {
		dw.window.SetBackground(dw.app.brushes.Background)
	}
	if dw.header != nil && dw.app.brushes.AccentDark != nil {
		dw.header.SetBackground(dw.app.brushes.AccentDark)
	}
	if dw.pathLabel != nil {
		dw.pathLabel.SetTextColor(dw.app.palette.TextPrimary)
	}
	if dw.tableView != nil && dw.app.brushes.Surface != nil {
		dw.tableView.SetBackground(dw.app.brushes.Surface)
	}
}

func (dw *drawerWindow) loadDirectory(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var items []fileItem
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		items = append(items, fileItem{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir
		}
		return items[i].Name < items[j].Name
	})

	dw.model.Reset(items)
	dw.currentPath = path
	if dw.pathLabel != nil {
		dw.pathLabel.SetText(path)
	}

	return nil
}

func (dw *drawerWindow) goUp() {
	parent := filepath.Dir(dw.currentPath)
	if parent == dw.currentPath || parent == "" {
		return
	}
	if err := dw.loadDirectory(parent); err != nil {
		log.Printf("failed to navigate to parent: %v", err)
	}
}

func (dw *drawerWindow) openSelected() {
	if dw.tableView == nil {
		return
	}
	index := dw.tableView.CurrentIndex()
	if index < 0 || index >= len(dw.model.items) {
		return
	}
	item := dw.model.items[index]
	if item.IsDir {
		if err := dw.loadDirectory(item.Path); err != nil {
			log.Printf("failed to open folder %s: %v", item.Path, err)
		}
		return
	}

	if err := shellOpen(item.Path); err != nil {
		log.Printf("failed to open file %s: %v", item.Path, err)
	}
}

func (dw *drawerWindow) applySettings(size settings.Size) {
	if dw.window == nil {
		return
	}

	dw.window.SetSize(walk.Size{Width: size.Width, Height: size.Height})
}

func (dw *drawerWindow) close() {
	if dw.window != nil {
		dw.window.Close()
	}
}

func (dw *drawerWindow) saveSize() {
	if dw.window == nil {
		return
	}
	size := dw.window.Size()
	dw.drawer.Size = settings.Size{Width: size.Width, Height: size.Height}
}

func (a *App) unregisterDrawer(dw *drawerWindow) {
	for i, existing := range a.drawerWindows {
		if existing == dw {
			a.drawerWindows = append(a.drawerWindows[:i], a.drawerWindows[i+1:]...)
			break
		}
	}
}

func shellOpen(path string) error {
	filePtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	if !win.ShellExecute(0, nil, filePtr, nil, nil, win.SW_SHOWNORMAL) {
		return fmt.Errorf("shell execute failed for %s", path)
	}
	return nil
}
