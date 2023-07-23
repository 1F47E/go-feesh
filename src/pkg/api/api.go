package api

import (
	"go-btc-scan/src/pkg/core"
	log "go-btc-scan/src/pkg/logger"
	"os"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Api struct {
	app  *fiber.App
	core *core.Core
}

func NewApi(core *core.Core) *Api {
	app := fiber.New(
		fiber.Config{
			BodyLimit: 1024 * 1024 * 100, // 100MB
		})
	app.Use(cors.New())
	app.Use(logger.New())

	// RECOVERY
	app.Use(recover.New())

	a := Api{app, core}

	// setup routes
	api := a.app.Group("/v0")
	api.Get("/monitor", monitor.New())
	api.Get("/ping", a.Ping)
	api.Get("/info", a.NodeInfo)
	api.Get("/pool", a.Pool)
	return &a
}

// will block
func (a *Api) Listen() error {
	log.Log.Info("Starting server...")
	err := a.app.Listen(os.Getenv("API_HOST"))
	if err != nil {
		log.Log.Fatal(err)
	}
	return nil
}

// shutdown
func (a *Api) Shutdown() error {
	log.Log.Info("Shutting down server...")
	return a.app.Shutdown()
}
