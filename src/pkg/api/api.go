package api

import (
	"log"
	"os"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Api struct {
	app *fiber.App
}

func NewApi() *Api {
	app := fiber.New(
		fiber.Config{
			BodyLimit: 1024 * 1024 * 100, // 100MB
		})
	app.Use(cors.New())
	app.Use(logger.New())

	// RECOVERY
	app.Use(recover.New())

	a := Api{app}

	// setup routes
	api := a.app.Group("/v0")
	api.Get("/ping", a.Ping)
	return &a
}

// will block
func (a *Api) Listen() error {
	log.Println("Starting server...")
	err := a.app.Listen(os.Getenv("API_HOST"))
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// shutdown
func (a *Api) Shutdown() error {
	log.Println("Shutting down server...")
	return a.app.Shutdown()
}
