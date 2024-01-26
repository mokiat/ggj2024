//go:build js

package main

import (
	"fmt"

	gameui "github.com/mokiat/ggj2024/internal/ui"
	"github.com/mokiat/ggj2024/resources"
	jsapp "github.com/mokiat/lacking-js/app"
	jsgame "github.com/mokiat/lacking-js/game"
	jsui "github.com/mokiat/lacking-js/ui"
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/ui"
	"github.com/mokiat/lacking/util/resource"
)

func runApplication() error {
	registry, err := asset.NewWebRegistry(".")
	if err != nil {
		return fmt.Errorf("failed to initialize registry: %w", err)
	}
	resourceLocator := ui.WrappedLocator(resource.NewFSLocator(resources.UI))
	gameController := game.NewController(registry, jsgame.NewShaderCollection())
	uiController := ui.NewController(resourceLocator, jsui.NewShaderCollection(), func(w *ui.Window) {
		gameui.BootstrapApplication(w, gameController)
	})

	cfg := jsapp.NewConfig("screen")
	cfg.AddGLExtension("EXT_color_buffer_float")
	cfg.SetFullscreen(false)
	return jsapp.Run(cfg, app.NewLayeredController(gameController, uiController))
}
