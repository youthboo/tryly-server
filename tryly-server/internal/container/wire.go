//go:build wireinject
// +build wireinject

package container

import (
	"github.com/google/wire"
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/api"
)

// InitializeApp initializes and returns the Fiber application with all dependencies
// Wire will auto-generate the dependency graph at compile time
func InitializeApp() (*fiber.App, error) {
	wire.Build(
		Set,
		api.SetupRoutes,
	)
	return &fiber.App{}, nil
}
