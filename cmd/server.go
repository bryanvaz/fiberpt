package main

import (
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New(fiber.Config{
		Prefork:       true,
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Fiber",
		AppName:       "Test App v1.0.1",
	})

	// Generic noop endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Ping-pong endpoint
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	// Ping-pong endpoint with JSON input
	app.Post("/ping", func(c *fiber.Ctx) error {
		type PingRequest struct {
			Message string `json:"message"`
		}
		req := new(PingRequest)
		if err := c.BodyParser(req); err != nil {
			return err
		}

		if req.Message == "ping" {
			return c.JSON(fiber.Map{
				"message": "pong",
			})
		}
		return c.JSON(fiber.Map{
			"message": req.Message,
		})
	})

	// Start server
	err := app.Listen("localhost:3000")
	if err != nil {
		panic(err)
	}
}
