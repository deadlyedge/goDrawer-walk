# AI Development Plan

## 1. Theme & Resources
- Extract the color palette and sizing constants from the design mock.
- Load the custom `Lobster-Regular.ttf` font at runtime and register it with Walk.
- Centralise theme resources (fonts, brushes, icons) so each window can share them.

## 2. Main Window Shell
- Replace the prototype slider UI with the branded header + container layout.
- Render the drawer list from `settings.Settings.Drawers`, applying themed rows and hover states.
- Wire the `Add Drawer` button to open a folder selector, append to settings, persist via `settings.Update`, and refresh the view.

## 3. Drawer Browsing Window
- Build a tiled or detail view that lists folder contents with the provided icons.
- Support double-click to open subdirectories or launch the selected file via ShellExecute.
- Provide breadcrumb / path label and a close/back control consistent with the theme.

## 4. Settings Window
- Implement the HSLA sliders with live preview and output string, following the tested layered window style.
- Add toggles for autostart / lock window, persisting updates to `goDrawer-settings.toml`.
- Apply theme changes immediately to open windows where feasible.

## 5. Polishing & Verification
- Ensure all windows remain borderless, draggable, and hidden from the taskbar using the reference code.
- Sanity check builds on Windows (`go build`) and exercise the UI manually.
- Prepare incremental Git commits per milestone to keep history readable.
