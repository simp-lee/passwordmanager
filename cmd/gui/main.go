package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/simp-lee/passwordmanager/backend"
)

//go:embed frontend
var assets embed.FS

func main() {
	app := backend.NewApp()

	// Create app window
	err := wails.Run(&options.App{
		Title:            "password manager",
		Width:            1024,
		Height:           768,
		MinWidth:         800,
		MinHeight:        600,
		DisableResize:    false,
		Fullscreen:       false,
		WindowStartState: options.Normal,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.SetContext,
		Bind: []interface{}{
			app,
		},

		// Windows 特定选项
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			// 可以设置应用图标
			// Icon:                 "path/to/icon.ico",
		},

		// Mac 特定选项
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarDefault(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "password manager",
				Message: "安全的本地密码管理工具",
				// 可以设置应用图标
				// Icon:    "path/to/icon.png",
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
