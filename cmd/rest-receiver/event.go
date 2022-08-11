package main

import (
	json2 "encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"log"
	"time"
)

type Event struct {
	Date         time.Time `json:"Date"`
	SystemId     uuid.UUID `json:"SystemId"`
	SessionId    uuid.UUID `json:"SessionId"`
	TotalLoading float64   `json:"TotalLoading"`
	DomLoading   float64   `json:"DomLoading"`
	Uri          string    `json:"Uri"`
	UserAgent    string    `json:"UserAgent"`
}

func receiveEventHandler(ctx *fiber.Ctx) error {
	// Парсинг body к Event
	msgID := uuid.New()
	body := new(Event)
	err := ctx.BodyParser(body)
	if err != nil {
		err := ctx.Status(fiber.StatusBadRequest).SendString(err.Error())
		if err != nil {
			return err
		}
		return err
	}

	// Проверка в списке систем по SystemId
	ok, err := isExist(body.SystemId)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("SystemId %v not found", body.SystemId))
	}

	// Публикация сообщения в очередь events.queue RabbitMQ
	// Маршалинк body в json
	json, err := json2.Marshal(body)
	if err != nil {
		log.Printf("SessionId: %v on %v messageID: %v, %v", body.SessionId, "marshalling", msgID, err.Error())
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	// Публикация в очередь

	err = publishEvent(json, msgID)
	if err != nil {
		log.Printf("SessionId: %v on %v messageID: %v, %v", body.SessionId, "publishing", msgID, err.Error())
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func isExist(id uuid.UUID) (bool, error) {
	// TODO добавить инвалидацию по таймауту
	if _, ok := app.knownSystems[id.String()]; ok {
		return true, nil
	}
	return false, nil
}
