package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"log"
)

type Event struct {
	SessionID string
	Project   string
	Page      string
	LoadTime  int64
}

func main() {
	app := fiber.New(fiber.Config{
		AppName: "Go-Metr v0.0.0",
	})

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).SendString(app.Config().AppName)
	})
	app.Get("/status", healthCheck)
	app.Get("/metrics", monitor.New(monitor.Config{Title: "MyService Metrics Page"}))

	eventApp := app.Group("/event")
	eventApp.Get("", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusForbidden)
	})
	eventApp.Post("", createEvent)

	log.Fatal(app.Listen(":3000"))
}

func createEvent(ctx *fiber.Ctx) error {
	body := new(Event)
	err := ctx.BodyParser(body)

	if err != nil {
		err := ctx.Status(fiber.StatusBadRequest).SendString(err.Error())
		if err != nil {
			return err
		}
		return err
	}

	event := Event{
		SessionID: body.SessionID,
		Project:   body.Project,
		Page:      body.Page,
		LoadTime:  body.LoadTime,
	}

	return ctx.Status(fiber.StatusOK).JSON(event)
}

// TODO реализовать самодиагностику, посмотри на https://github.com/mackerelio/go-osstat
func healthCheck(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).SendString("OK")
}
