package api

import (
	"net/http"

	"github.com/1F47E/go-feesh/pkg/core"
	"github.com/1F47E/go-feesh/pkg/logger"
	"github.com/1F47E/go-feesh/pkg/notificator"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	flogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/gofiber/websocket/v2"
)

type Api struct {
	app         *fiber.App
	core        *core.Core
	notificator *notificator.Notificator
}

func NewApi(core *core.Core, notificator *notificator.Notificator) *Api {
	log := logger.Log.WithField("scope", "api.new")
	app := fiber.New(
		fiber.Config{
			BodyLimit: 1024 * 1024 * 100, // 100MB
		})
	app.Use(cors.New())
	app.Use(flogger.New())
	app.Use(recover.New())

	// Middleware function
	app.Use(func(c *fiber.Ctx) error {
		customLogger := logger.LoggerEntry{Entry: *logger.Log.WithField("path", c.Path())}
		c.Locals("logger", customLogger)
		return c.Next()
	})

	a := Api{app, core, notificator}

	// setup routes
	api := a.app.Group("/v0")
	api.Get("/swagger/*", swagger.HandlerDefault) // default
	api.Get("/monitor", monitor.New())
	api.Get("/stats", a.Stats)
	api.Get("/info", a.NodeInfo)
	api.Get("/ping", a.Ping)
	api.Get("/version", a.Version)
	api.Get("/pool", a.Pool)

	// websockets
	api.Get("/ws", websocket.New(func(c *websocket.Conn) {
		defer func() {
			a.notificator.UnregisterCh <- c
			c.Close()
		}()

		// register new client
		a.notificator.RegisterCh <- c

		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("read error:", err)
				}

				return // Calls the deferred function, i.e. closes the connection on error
			}

			if messageType == websocket.TextMessage {
				log.Debugf("ws msg received: %s", message)
				// TODO: receive pings
			} else {
				log.Error("websocket message received of type", messageType)
			}
		}
	}))

	return &a
}

// will block
func (a *Api) Listen() error {
	log := logger.Log.WithField("scope", "api.listen")

	log.Info("Starting WS service...")
	a.notificator.Start()

	log.Info("Starting http server...")
	err := a.app.Listen(a.core.Cfg.ApiHost)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// shutdown
func (a *Api) Shutdown() error {
	logger.Log.Info("Shutting down server...")
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
