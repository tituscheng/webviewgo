package webview

// MenuItem represents a single entry in a menu or tray.
type MenuItem struct {
	Label     string
	Shortcut  string // platform-specific shortcut notation
	Action    func()
	Items     []MenuItem // sub-menu
	Separator bool
	Checked   bool
	Disabled  bool
}

// Tray represents a system tray icon and menu.
type Tray struct {
	Icon    []byte
	Menu    []MenuItem
	Tooltip string
}
