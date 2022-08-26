package main

import (
	json2 "encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vsurkov/go-metr/internal/app/event"
	"github.com/vsurkov/go-metr/internal/common/helpers"
	rabbitmq2 "github.com/vsurkov/go-metr/internal/platform/rabbitmq"
	"time"
)

func isExist(id uuid.UUID) bool {
	// TODO добавить инвалидацию по таймауту
	if _, ok := app.KnownSystems[id.String()]; ok {
		return true
	}
	return false
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
	ok := isExist(body.SystemId)
	if !ok {
		log.Warn().
			Str("service", helpers.Core).
			Str("method", "ReceiveEventHandler").
			Dict("dict", zerolog.Dict().
				Str("MessageID", body.MessageID.String()).
				Str("SystemId", body.SystemId.String()),
			).Msg("systemID from received message not found")
		return ctx.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("SystemId %v not found", body.SystemId))
	}

	// Заполняем данные по конкретному сообщению
	body.Timestamp = time.Now().Unix()
	body.MessageID = msgID

	// Маршалинг body в msg
	msg, err := json2.Marshal(body)
	if err != nil {
		log.Error().
			Str("service", helpers.Core).
			Str("method", "ReceiveEventHandler").
			Dict("dict", zerolog.Dict().
				Str("SessionID", body.SessionId.String()).
				Str("Action", "marshalling").
				Str("MessageID", msgID.String()).
				Err(err),
			).Msg("marshalling message error")
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	body.Body = msg

	// Публикация в очередь
	err = rabbitmq2.PublishEvent(app.RB, &rabbitmq2.RabbitMsg{
		Message: body,
	})
	//err = app.RB.Buffer.BuffWrite(app.RB.Buffer, body, app.RB)

	if err != nil {
		log.Error().
			Str("service", helpers.Rabbit).
			Str("method", "ReceiveEventHandler").
			Dict("dict", zerolog.Dict().
				Str("SessionID", body.SessionId.String()).
				Str("Action", "publishing").
				Str("MessageID", msgID.String()).
				Err(err),
			).Msg("publishing error")
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.SendStatus(fiber.StatusOK)
}
