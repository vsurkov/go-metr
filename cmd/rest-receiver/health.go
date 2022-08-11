package main

import "github.com/gofiber/fiber/v2"

// TODO реализовать самодиагностику, посмотри на https://github.com/mackerelio/go-osstat
func healthCheckHandler(ctx *fiber.Ctx) error {

	// Проверка доступности Clickhouse - без базы не сможем принимать метрики (проверка на наличие проекта)
	if ok, _ := db.Ping(); !ok {
		return ctx.Status(fiber.StatusInternalServerError).SendString("ERROR")
	}

	// TODO Проверка доступности RabbitMQ
	// rabbitmq-plugins enable rabbitmq_prometheus
	// {hostname}:15692/metrics
	return ctx.Status(fiber.StatusOK).SendString("OK")
}
