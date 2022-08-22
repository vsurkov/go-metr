package main

import (
	json2 "encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/vsurkov/go-metr/internal/app/event"
	rabbitmq2 "github.com/vsurkov/go-metr/internal/platform/rabbitmq"
	"log"
	"time"
)

func isExist(id uuid.UUID) (bool, error) {
	// TODO добавить инвалидацию по таймауту
	if _, ok := app.KnownSystems[id.String()]; ok {
		return true, nil
	}
	return false, nil
}

func ReceiveEventHandler(ctx *fiber.Ctx) error {
	// Парсинг body к Event
	msgID := uuid.New()
	body := &event.Event{}
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
	// Маршалинг body в msg
	msg, err := json2.Marshal(body)
	if err != nil {
		log.Printf("SessionId: %v on %v messageID: %v, %v", body.SessionId, "marshalling", msgID, err.Error())
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	// Заполняем данные по конкретному сообщению
	body.Timestamp = time.Now().Unix()
	body.MessageID = uuid.New()
	body.Body = msg

	// Публикация в очередь
	err = rabbitmq2.PublishEvent(app.RB, &rabbitmq2.RabbitMsg{
		Message: body,
	})
	//err = app.RB.Buffer.BuffWrite(app.RB.Buffer, body, app.RB)

	if err != nil {
		log.Printf("SessionId: %v on %v messageID: %v, %v", body.SessionId, "publishing", msgID, err.Error())
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.SendStatus(fiber.StatusOK)
}
