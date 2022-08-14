package main

import (
	"flag"
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

var (
	app_port     = flag.String("app_port", "4000", "Application port, default http://127.0.0.1:4000")
	uri          = flag.String("uri", "amqp://rabbitmq:rabbitmq@localhost:5672/", "AMQP URI")
	exchange     = flag.String("exchange", "test-exchange", "Durable, non-auto-deleted AMQP exchange name")
	exchangeType = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	queue        = flag.String("queue", "rncb.events.queue", "Ephemeral AMQP queue name")
	bindingKey   = flag.String("key", "test-key", "AMQP binding key")
	consumerTag  = flag.String("consumer-tag", "simple-consumer", "AMQP consumer tag (should not be blank)")
	lifetime     = flag.Duration("lifetime", 5*time.Second, "lifetime of process before shutdown (0s=infinite)")
	bufferSize   = flag.Int("bufferSize", 100, "Buffer size for batch database Writing (default 100 messages)")
)

func init() {
	flag.Parse()
}

func main() {
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
	b := new(Buffer)
	db.buffer = b.newBuffer(*bufferSize)

	c, err := NewConsumer(*uri, *exchange, *exchangeType, *queue, *bindingKey, *consumerTag)
	if err != nil {
		log.Fatalf("%s", err)
	}

	defer func() {
		log.Printf("shutting down")

		if err := c.Shutdown(); err != nil {
			log.Fatalf("error during shutdown: %s", err)
		}
	}()

	// Fiber configuration

	a := fiber.New(fiber.Config{
		AppName: fmt.Sprintf("go-metr v %v, instance %v", app.version, app.id),
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

	// Start Fiber server on port
	log.Fatal(a.Listen(fmt.Sprintf(":%v", *app_port)))
}
