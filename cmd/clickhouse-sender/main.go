package main

import (
	"flag"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/vsurkov/go-metr/internal/buffer"
	"github.com/vsurkov/go-metr/internal/db"
	"github.com/vsurkov/go-metr/internal/instance"
	"github.com/vsurkov/go-metr/internal/rabbitmq"
	"log"
	"time"
)

var app = new(instance.Instance)

var (
	app_port     = flag.String("app_port", "4000", "Application port, default http://127.0.0.1:4000")
	uri          = flag.String("uri", "amqp://rabbitmq:rabbitmq@localhost:5672/", "AMQP URI")
	exchange     = flag.String("exchange", "test-exchange", "Durable, non-auto-deleted AMQP exchange name")
	exchangeType = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	queue        = flag.String("queue", "rncb.events.queue", "Ephemeral AMQP queue name")
	bindingKey   = flag.String("key", "test-key", "AMQP binding key")
	consumerTag  = flag.String("consumer-tag", "simple-consumer", "AMQP consumer tag (should not be blank)")
	bufferSize   = flag.Int("bufferSize", 100, "Buffer size for batch database Writing (default 100 messages)")
)

func init() {
	flag.Parse()
}

func main() {
	app.Version = "0.0.1"
	app.Name = "rncb"

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
	b := new(buffer.Buffer)
	app.DB.Buffer = b.NewBuffer(*bufferSize)

	// RabbitMQ configuration
	app.RB.Cfg = rabbitmq.Config{
		Uri:          *uri,
		Exchange:     *exchange,
		ExchangeType: *exchangeType,
		BindingKey:   *bindingKey,
		ConsumerTag:  *consumerTag,
	}

	//	TODO сделать множественные очереди
	cons, err := rabbitmq.NewConsumer(app.RB.Cfg, *queue, app.DB)
	if err != nil {
		log.Fatalf("%s", err)
	}

	defer func() {
		log.Printf("shutting down")

		if err := cons.Shutdown(); err != nil {
			log.Fatalf("error during shutdown: %s", err)
		}
	}()

	// Fiber configuration

	a := fiber.New(fiber.Config{
		AppName: fmt.Sprintf("go-metr v %v, app %v", app.Version, app.Name),
	})

	// Fiber middleware configuration
	a.Use(logger.New())
	a.Use(requestid.New())

	// Fiber endpoints configuration
	a.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).SendString(a.Config().AppName)
	})
	a.Get("/metrics", monitor.New(monitor.Config{Title: "Metrics"}))
	a.Get("/status", HealthCheckHandler)

	// Start Fiber server on port
	log.Fatal(a.Listen(fmt.Sprintf(":%v", *app_port)))
}
