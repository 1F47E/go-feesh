package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/1F47E/go-feesh/pkg/core"
	"github.com/1F47E/go-feesh/pkg/logger"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	flogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/gofiber/websocket/v2"
)

type Api struct {
	app     *fiber.App
	core    *core.Core
	wsMsgCh chan []byte
}

func NewApi(core *core.Core) *Api {
	log := logger.Log.WithField("scope", "api.new")
	app := fiber.New(
		fiber.Config{
			BodyLimit: 1024 * 1024 * 100, // 100MB
		})
	app.Use(cors.New())
	app.Use(flogger.New())
	app.Use(recover.New())

	// Set log format to JSON

	// Middleware function
	app.Use(func(c *fiber.Ctx) error {
		customLogger := logger.LoggerEntry{Entry: *logger.Log.WithField("path", c.Path())}
		c.Locals("logger", customLogger)
		// c.Locals("logger", logger.Log.WithField("path", c.Path()))
		return c.Next()
	})

	wsMsgCh := make(chan []byte)
	a := Api{app, core, wsMsgCh}

	// setup routes
	api := a.app.Group("/v0")
	api.Get("/swagger/*", swagger.HandlerDefault) // default
	api.Get("/monitor", monitor.New())
	api.Get("/stats", a.Stats)
	api.Get("/info", a.NodeInfo)
	api.Get("/pool", a.Pool)

	// websockets
	// Optional middleware
	app.Use("/ws", func(c *fiber.Ctx) error {
		if c.Get("host") == "localhost:3000" {
			c.Locals("Host", "Localhost:3000")
			return c.Next()
		}
		return c.Status(403).SendString("Request origin not allowed")
	})

	// Upgraded websocket request
	api.Get("/ws", websocket.New(func(c *websocket.Conn) {
		fmt.Println(c.Locals("Host"))
		// on connection say hello
		err := c.WriteMessage(websocket.TextMessage, []byte("Hello, Client!"))
		if err != nil {
			log.Println("write:", err)
		}
		for msg := range wsMsgCh {
			err := c.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Errorf("ws write error: %v", err)
				break
			}
		}
	}))

	// demo ws msg
	go func() {
		cnt := 0
		for {
			msg := fmt.Sprintf("Hello, Client! %d", cnt)
			wsMsgCh <- []byte(msg)
			log.Debugf("ws msg sent: %s", msg)
			cnt++
			time.Sleep(1 * time.Second)
		}
	}()

	return &a
}

// will block
func (a *Api) Listen() error {
	log := logger.Log.WithField("scope", "api.listen")
	log.Info("Starting server...")
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
