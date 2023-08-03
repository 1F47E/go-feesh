package api

import (
	"net/http"

	"github.com/1F47E/go-feesh/pkg/core"
	log "github.com/1F47E/go-feesh/pkg/logger"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
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

	api.Get("/swagger/*", swagger.HandlerDefault) // default

	api.Get("/monitor", monitor.New())
	api.Get("/stats", a.Stats)
	api.Get("/info", a.NodeInfo)
	api.Get("/pool", a.Pool)
	return &a
}

// will block
func (a *Api) Listen() error {
	log.Log.Info("Starting server...")
	err := a.app.Listen(a.core.Cfg.ApiHost)
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

type APISuccess struct {
	Data interface{} `json:"data"`
}

type APIError struct {
	Error       string `json:"error"`
	Description string `json:"description,omitempty"`
	RequestID   string `json:"request_id,omitempty"`
}

func apiSuccess(c *fiber.Ctx, data interface{}) error {
	return c.JSON(APISuccess{Data: data})
}

func apiError(c *fiber.Ctx, httpCode int, messages ...string) error {
	e := APIError{
		Error: http.StatusText(httpCode),
	}

	if len(messages) > 0 {
		e.Description = messages[0]
	}

	if len(messages) > 1 {
		e.RequestID = messages[1]
	}

	return c.Status(httpCode).JSON(e)
}
