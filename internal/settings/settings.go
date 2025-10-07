package settings

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Drawer represents a drawer configuration.
type Drawer struct {
	Name string `toml:"name"`
	Path string `toml:"path"`
	Size Size   `toml:"size"`
}

// Size represents the size of a drawer window.
type Size struct {
	Width  int `toml:"width"`
	Height int `toml:"height"`
}

// Startup represents startup configuration.
type Startup struct {
	StartWithWindows bool `toml:"start_with_windows"`
	WindowLocked     bool `toml:"window_locked"`
}

// Theme captures HSLA values for the UI layer.
type Theme struct {
	Hue        int `toml:"h"`
	Saturation int `toml:"s"`
	Lightness  int `toml:"l"`
	Alpha      int `toml:"a"`
}

// Settings represents the complete configuration.
type Settings struct {
	Startup          Startup           `toml:"startup"`
	Drawers          []Drawer          `toml:"drawers"`
	WindowPosition   Point             `toml:"window_position"`
	ThumbnailSize    Size              `toml:"thumbnail_size"`
	Theme            Theme             `toml:"theme"`
	ExtensionIconMap map[string]string `toml:"extension_icon_map"`
	Deprecated       map[string]string `toml:"deprecated,omitempty"`
}

// Point represents a 2D point.
type Point struct {
	X int `toml:"x"`
	Y int `toml:"y"`
}

// DefaultTheme returns the base theme reflected in the design reference.
func DefaultTheme() Theme {
	return Theme{Hue: 192, Saturation: 40, Lightness: 36, Alpha: 80}
}

// Read reads and parses the goDrawer-settings.toml file.
func Read(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			Init(path)
			data, err = os.ReadFile(path)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read settings file: %w", err)
		}
	}

	var settings Settings
	if _, err := toml.Decode(string(data), &settings); err != nil {
		return nil, fmt.Errorf("failed to parse TOML file: %w", err)
	}

	settings.applyDefaults()
	return &settings, nil
}

// Update updates the settings file with the provided settings.
func Update(path string, settings *Settings) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to open settings file for writing: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(settings); err != nil {
		return fmt.Errorf("failed to encode settings: %w", err)
	}

	return nil
}

// Init creates the settings file using default values if it does not exist.
func Init(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		defaultSettings := Settings{
			Startup: Startup{
				StartWithWindows: false,
				WindowLocked:     false,
			},
			Drawers: []Drawer{
				{Name: "Drawer 1", Path: "C:\\", Size: Size{Width: 800, Height: 600}},
			},
			WindowPosition:   Point{X: 100, Y: 100},
			ThumbnailSize:    Size{Width: 96, Height: 96},
			Theme:            DefaultTheme(),
			ExtensionIconMap: map[string]string{},
			Deprecated:       map[string]string{},
		}

		defaultSettings.applyDefaults()

		file, createErr := os.Create(path)
		if createErr != nil {
			fmt.Println("Error creating settings file:", createErr)
			return
		}
		defer file.Close()

		encoder := toml.NewEncoder(file)
		if encodeErr := encoder.Encode(defaultSettings); encodeErr != nil {
			fmt.Println("Error encoding default settings:", encodeErr)
			return
		}

		fmt.Println("Settings file created successfully.")
	} else if err != nil {
		fmt.Println("Error checking settings file:", err)
	} else {
		fmt.Println("Settings file already exists.")
	}
}

// Print categorizes and prints the settings information.
func Print(settings *Settings) {
	fmt.Println("=== Drawers Settings ===")
	fmt.Println()

	fmt.Println(":: Startup ::")
	fmt.Printf("  Start with Windows: %t\n", settings.Startup.StartWithWindows)
	fmt.Printf("  Window Locked: %t\n", settings.Startup.WindowLocked)
	fmt.Println()

	fmt.Println(":: Drawers ::")
	for i, drawer := range settings.Drawers {
		fmt.Printf("  %d. %s\n", i+1, drawer.Name)
		fmt.Printf("     Path: %s\n", drawer.Path)
		fmt.Printf("     Size: %dx%d\n", drawer.Size.Width, drawer.Size.Height)
		fmt.Println()
	}

	fmt.Println(":: Window Position ::")
	fmt.Printf("  Position: (%d, %d)\n", settings.WindowPosition.X, settings.WindowPosition.Y)
	fmt.Println()

	fmt.Println(":: Thumbnail ::")
	fmt.Printf("  Size: %dx%d\n", settings.ThumbnailSize.Width, settings.ThumbnailSize.Height)
	fmt.Println()

	fmt.Println(":: Theme (HSLA) ::")
	fmt.Printf("  H: %d\n", settings.Theme.Hue)
	fmt.Printf("  S: %d\n", settings.Theme.Saturation)
	fmt.Printf("  L: %d\n", settings.Theme.Lightness)
	fmt.Printf("  A: %d\n", settings.Theme.Alpha)
	fmt.Println()

	fmt.Println(":: Extension Icon Map ::")
	for ext, icon := range settings.ExtensionIconMap {
		fmt.Printf("  %s -> %s\n", ext, icon)
	}
	fmt.Println()
}

func (s *Settings) applyDefaults() {
	if s.Drawers == nil {
		s.Drawers = []Drawer{}
	}
	if s.ExtensionIconMap == nil {
		s.ExtensionIconMap = map[string]string{}
	}
	if s.Deprecated == nil {
		s.Deprecated = map[string]string{}
	}
	if s.ThumbnailSize.Width == 0 {
		s.ThumbnailSize.Width = 96
	}
	if s.ThumbnailSize.Height == 0 {
		s.ThumbnailSize.Height = 96
	}
	if s.Theme == (Theme{}) {
		s.Theme = DefaultTheme()
	}
}
