package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"log"
	"time"
)

type Event struct {
	Date         time.Time `json:"date"`
	SystemId     uuid.UUID `json:"systemId"`
	SessionId    uuid.UUID `json:"sessionId"`
	TotalLoading float64   `json:"totalLoading"`
	DomLoading   float64   `json:"domLoading"`
	Uri          string    `json:"uri"`
	UserAgent    string    `json:"userAgent"`
}

func receiveEventHandler(ctx *fiber.Ctx) error {
	body := new(Event)
	err := ctx.BodyParser(body)

	if err != nil {
		err := ctx.Status(fiber.StatusBadRequest).SendString(err.Error())
		if err != nil {
			return err
		}
		return err
	}

	// TODO проверка списка систем
	ok, err := db.HasSystem(body.SystemId)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("SystemId %v not found", body.SystemId))
	}

	// Публикуем сообщение в очередь events.queue RabbitMQ
	err = publishEvent(body)
	if err != nil {
		log.Printf("%v %v", body.SessionId, err.Error())
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.SendStatus(fiber.StatusOK)
}
