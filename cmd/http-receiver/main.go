package main

import (
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"log"
	"time"
)

func main() {
	// Clickhouse configuration
	clickhouseDB, err := new(Database).Connect(dbConfig{
		host:              "127.0.0.1",
		port:              9000,
		database:          "default",
		username:          "default",
		password:          "",
		debug:             false,
		dialTimeout:       time.Second,
		maxOpenConns:      10,
		maxIdleConns:      5,
		connMaxLifetime:   time.Hour,
		compressionMethod: clickhouse.CompressionLZ4,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	db = *clickhouseDB

	// RabbitMQ configuration
	rb.cfg = rabbitConfig{
		username: "rabbitmq",
		password: "rabbitmq",
		host:     "127.0.0.1",
		port:     5672,
	}
	//rabbit, err := new(Rabbit).Connect(cfg)
	//if err != nil {
	//	log.Fatal(err.Error())
	//}
	//rb = *rabbit

	// Fiber configuration
	appVersion := "0.0.1"
	app := fiber.New(fiber.Config{
		AppName: fmt.Sprintf("go-metr v %v", appVersion),
	})

	// Fiber middleware configuration
	app.Use(logger.New())
	app.Use(requestid.New())

	// Fiber endpoints configuration
	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).SendString(app.Config().AppName)
	})
	app.Get("/metrics", monitor.New(monitor.Config{Title: "Metrics"}))
	app.Get("/status", healthCheckHandler)
	app.Post("/event", receiveEventHandler)
	app.Get("/event", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusForbidden)
	})

	// Start Fiber server on port
	log.Fatal(app.Listen(":3000"))
}
