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

type Instance struct {
	name         string
	version      string
	id           string
	knownSystems map[string]string
}

var app = new(Instance)

func main() {
	app.name = "rest-receiver"
	app.version = "0.0.1"
	//app.hostIP, app.hostMAC = getNetInfo()
	//app.id = strings.ReplaceAll(app.hostMAC.String(), ":", "")
	app.id = "rncb"

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

	app.knownSystems, err = db.GetSystems()
	failOnError(err, "Can't receive systems")

	// RabbitMQ configuration
	rb.cfg = rabbitConfig{
		username: "rabbitmq",
		password: "rabbitmq",
		host:     "127.0.0.1",
		port:     5672,
	}

	// Fiber configuration

	a := fiber.New(fiber.Config{
		AppName: fmt.Sprintf("go-metr %v v%v, %v", app.name, app.version, app.id),
	})

	// Fiber middleware configuration
	a.Use(logger.New())
	a.Use(requestid.New())

	// Fiber endpoints configuration
	a.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).SendString(a.Config().AppName)
	})
	a.Get("/metrics", monitor.New(monitor.Config{Title: "Metrics"}))
	a.Get("/status", healthCheckHandler)
	a.Post("/event", receiveEventHandler)
	a.Get("/event", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusForbidden)
	})

	// Start Fiber server on port
	log.Fatal(a.Listen(":3000"))
}
