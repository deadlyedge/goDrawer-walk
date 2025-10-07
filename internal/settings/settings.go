package settings

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Drawer represents a drawer configuration
type Drawer struct {
	Name string `toml:"name"`
	Path string `toml:"path"`
	Size Size   `toml:"size"`
}

// Size represents the size of a drawer window
type Size struct {
	Width  int `toml:"width"`
	Height int `toml:"height"`
}

// Settings represents the complete configuration
type Settings struct {
	Drawers             []Drawer          `toml:"drawers"`
	WindowPosition      Point             `toml:"window_position"`
	BackgroundColorHSLA []float64         `toml:"background_color_hsla"`
	StartWithWindows    bool              `toml:"start_with_windows"`
	DefaultIconFolder   string            `toml:"default_icon_folder_path"`
	DefaultIconFile     string            `toml:"default_icon_file_theme"`
	DefaultIconUnknown  string            `toml:"default_icon_unknown_theme"`
	ThumbnailSize       Size              `toml:"thumbnail_size"`
	ExtensionIconMap    map[string]string `toml:"extension_icon_map"`
}

// Point represents a 2D point
type Point struct {
	X int `toml:"x"`
	Y int `toml:"y"`
}

// Read reads and parses the drawers-settings.toml file
func Read(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			Init(path)
		}
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings Settings
	if _, err := toml.Decode(string(data), &settings); err != nil {
		return nil, fmt.Errorf("failed to parse TOML file: %w", err)
	}

	return &settings, nil
}

// Update updates the settings file with the provided settings
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

// Init settings file
func Init(path string) {
	// Check if the settings file already exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// If the file doesn't exist, create it with default settings
		defaultSettings := Settings{
			Drawers: []Drawer{
				{
					Name: "Drawer 1",
					Path: "C:\\",
					Size: Size{Width: 800, Height: 600},
				},
			},
			WindowPosition:      Point{X: 100, Y: 100},
			BackgroundColorHSLA: []float64{0, 0, 0, 1},
			StartWithWindows:    false,
			DefaultIconFolder:   "C:\\Program Files\\File Explorer\\assets\\icons\\folder.png",
			DefaultIconFile:     "C:\\Program Files\\File Explorer\\assets\\icons\\file.png",
			DefaultIconUnknown:  "C:\\Program Files\\File Explorer\\assets\\icons\\unknown.png",
			ThumbnailSize:       Size{Width: 128, Height: 128},
			ExtensionIconMap:    map[string]string{},
		}

		// Write the default settings to the file
		file, err := os.Create(path)
		if err != nil {
			fmt.Println("Error creating settings file:", err)
			return
		}
		defer file.Close()

		encoder := toml.NewEncoder(file)
		if err := encoder.Encode(defaultSettings); err != nil {
			fmt.Println("Error encoding default settings:", err)
			return
		}

		fmt.Println("Settings file created successfully.")

	} else if err != nil {
		fmt.Println("Error checking settings file:", err)
	} else {
		fmt.Println("Settings file already exists.")
	}
}

// Print categorizes and prints the settings information
func Print(settings *Settings) {
	fmt.Println("=== Drawers Settings ===")
	fmt.Println()

	// Print Drawers
	fmt.Println("📁 Drawers:")
	for i, drawer := range settings.Drawers {
		fmt.Printf("  %d. %s\n", i+1, drawer.Name)
		fmt.Printf("     Path: %s\n", drawer.Path)
		fmt.Printf("     Size: %dx%d\n", drawer.Size.Width, drawer.Size.Height)
		fmt.Println()
	}

	// Print Window Configuration
	fmt.Println("🪟 Window Configuration:")
	fmt.Printf("  Position: (%d, %d)\n", settings.WindowPosition.X, settings.WindowPosition.Y)
	fmt.Printf("  Background HSLA: %v\n", settings.BackgroundColorHSLA)
	fmt.Printf("  Start with Windows: %t\n", settings.StartWithWindows)
	fmt.Println()

	// Print Icon Settings
	fmt.Println("🖼️  Icon Settings:")
	fmt.Printf("  Default Folder Icon: %s\n", settings.DefaultIconFolder)
	fmt.Printf("  Default File Icon Theme: %s\n", settings.DefaultIconFile)
	fmt.Printf("  Default Unknown Icon: %s\n", settings.DefaultIconUnknown)
	fmt.Printf("  Thumbnail Size: %dx%d\n", settings.ThumbnailSize.Width, settings.ThumbnailSize.Height)
	fmt.Println()

	// Print Extension Icon Map
	fmt.Println("📋 Extension Icon Map:")
	for ext, icon := range settings.ExtensionIconMap {
		fmt.Printf("  %s -> %s\n", ext, icon)
	}
	fmt.Println()
}
