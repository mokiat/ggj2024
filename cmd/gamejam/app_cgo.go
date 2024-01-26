//go:build !js

package main

import (
	"fmt"

	gameui "github.com/mokiat/ggj2024/internal/ui"
	"github.com/mokiat/ggj2024/resources"
	glapp "github.com/mokiat/lacking-native/app"
	glgame "github.com/mokiat/lacking-native/game"
	glui "github.com/mokiat/lacking-native/ui"
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/util/resource"
)

func runApplication() error {
	registry, err := asset.NewDirRegistry(".")
	if err != nil {
		return fmt.Errorf("failed to initialize registry: %w", err)
	}
	locator := ui.WrappedLocator(resource.NewFSLocator(resources.UI))

	gameController := game.NewController(registry, glgame.NewShaderCollection())
	uiController := ui.NewController(locator, glui.NewShaderCollection(), func(w *ui.Window) {
		gameui.BootstrapApplication(w, gameController)
	})

	cfg := glapp.NewConfig("GGJ", 1280, 800)
	cfg.SetFullscreen(false)
	cfg.SetMaximized(false)
	cfg.SetMinSize(1280, 800)
	cfg.SetVSync(true)
	cfg.SetIcon("ui/images/icon.png")
	cfg.SetLocator(locator)
	return glapp.Run(cfg, app.NewLayeredController(gameController, uiController))
}
