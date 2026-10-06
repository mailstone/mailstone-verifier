package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend
var assets embed.FS

// Linux has no bundle to carry an icon. The window gets the one we hand
// over at startup — 256 px at most: X11 silently drops a 512 px icon as
// "too large" for its property limit, and the window then shows a dark
// placeholder. Launchers and the GNOME dock look up a .desktop entry named
// after the program name instead (the Wayland app id / X11 WM_CLASS) and
// take the icon from the theme; build/linux/install.sh installs that entry
// and the hicolor icons, and ProgramName below must match its file name.
//
//go:embed build/linux/appicon.png
var linuxIcon []byte

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "MailStone Verifier",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       false,
			},
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Linux: &linux.Options{
			Icon:        linuxIcon,
			ProgramName: "mailstone-verifier", // = build/linux/mailstone-verifier.desktop
		},
	})

	if err != nil {
		log.Fatal("Error:", err.Error())
	}
}
