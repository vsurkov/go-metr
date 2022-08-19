package main

import (
	"flag"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/vsurkov/go-metr/internal/db"
	"github.com/vsurkov/go-metr/internal/helpers"
	"github.com/vsurkov/go-metr/internal/instance"
	"github.com/vsurkov/go-metr/internal/rabbitmq"
	"log"
	"time"
)

var (
	app_port = flag.String("app_port", "3000", "Application port, default http://127.0.0.1:3000")
	app_name = flag.String("app_name", "rest-application", "Application name, default rest-application")
	app_id   = flag.String("app_id", "rncb", "Instance name, will be used in Queue prefix, default rncb")
	uri      = flag.String("uri", "amqp://rabbitmq:rabbitmq@localhost:5672/", "AMQP URI")
	//exchange     = flag.String("exchange", "test-exchange", "Durable, non-auto-deleted AMQP exchange name")
	//exchangeType = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	//queue        = flag.String("queue", "rncb.events.queue", "Ephemeral AMQP queue name")
	//bindingKey   = flag.String("key", "test-key", "AMQP binding key")
	//consumerTag  = flag.String("consumer-tag", "simple-consumer", "AMQP consumer tag (should not be blank)")
	bufferSize = flag.Int("bufferSize", 2, "Buffer size for batch database Writing (default 100 messages)")
)

func init() {
	flag.Parse()
}

var app = new(instance.Instance)

func main() {
	app.FullName = *app_name
	app.Name = *app_id
	app.Version = "0.0.1"

	// Clickhouse configuration
	clickhouseDB, err := new(db.Database).Connect(db.DBConfig{
		Host:              "127.0.0.1",
		Port:              9000,
		Database:          "default",
		Username:          "default",
		Password:          "",
		Debug:             false,
		DialTimeout:       time.Second,
		MaxOpenConns:      10,
		MaxIdleConns:      5,
		ConnMaxLifetime:   time.Hour,
		CompressionMethod: clickhouse.CompressionLZ4,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	app.DB = *clickhouseDB

	app.KnownSystems, err = app.DB.GetSystems()
	helpers.FailOnError(err, "Can't receive systems")

	// RabbitMQ configuration
	app.RB.Cfg = &rabbitmq.Config{
		Uri:   *uri,
		Queue: fmt.Sprintf("%v.events.queue", app.Name),
	}
	err = rabbitmq.InitProducer(&app.RB)
	if err != nil {
		helpers.FailOnError(err, "Can' initialise RabbitMQ")
	}

	// Батчи для рэббита в лоб не сделать, можно использовать потом для передачи в массива в бинарной последовательности
	//b := new(buffer.Buffer)
	//app.RB.Buffer = b.NewBuffer(*bufferSize)

	// Fiber configuration
	a := fiber.New(fiber.Config{
		AppName: fmt.Sprintf("go-metr %v v%v, %v", app.FullName, app.Version, app.Name),
	})

	// Fiber middleware configuration
	a.Use(logger.New())
	//a.Use(requestid.New())
	a.Use(pprof.New())
	// Fiber endpoints configuration
	a.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).SendString(a.Config().AppName)
	})
	a.Get("/metrics", monitor.New(monitor.Config{Title: "Metrics"}))
	a.Get("/status", HealthCheckHandler)
	a.Post("/event", ReceiveEventHandler)
	a.Get("/event", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusForbidden)
	})

	// Run Fiber server on port
	log.Fatal(a.Listen(fmt.Sprintf(":%v", *app_port)))
}
