package main

import (
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/spf13/viper"
	"github.com/vsurkov/go-metr/internal/app/instance"
	"github.com/vsurkov/go-metr/internal/common/buffer"
	"github.com/vsurkov/go-metr/internal/common/helpers"
	"github.com/vsurkov/go-metr/internal/platform/database"
	rabbitmq2 "github.com/vsurkov/go-metr/internal/platform/rabbitmq"
	"log"
)

var (
	app     instance.Instance
	Version string
)

func main() {
	configureParams() // Configure viper params and config variables

	config := instance.Config{
		Port:     viper.GetInt("server.port"),
		Name:     viper.GetString("server.name"),
		FullName: viper.GetString("server.full_name"),
		Version:  Version,
	}
	err := config.Validate()
	if err != nil {
		helpers.FailOnError(err, "application config is not valid")
	}

	//viper.WatchConfig()
	//viper.OnConfigChange(func(e fsnotify.Event) {
	//	fmt.Println("Config file changed:", e.Name)
	//})

	//Application configuration
	app = instance.Instance{
		Config:       config,
		DB:           database.Database{},
		RB:           rabbitmq2.Rabbit{},
		KnownSystems: nil,
	}

	// Clickhouse configuration
	dbConfig := &database.Config{
		Host:              viper.GetString("db.host"),
		Port:              viper.GetInt("db.port"),
		Database:          viper.GetString("db.default_database"),
		User:              viper.GetString("db.user"),
		Password:          viper.GetString("db.password"),
		Debug:             viper.GetBool("db.debug"),
		DialTimeout:       viper.GetDuration("db.dial_timeout"),
		MaxOpenConns:      viper.GetInt("db.max_open_conns"),
		MaxIdleConns:      viper.GetInt("db.max_idle_conns"),
		ConnMaxLifetime:   viper.GetDuration("conn_max_lifetime"),
		CompressionMethod: clickhouse.CompressionLZ4,
	}
	err = dbConfig.Validate()
	if err != nil {
		helpers.FailOnError(err, "database config is not valid")
	}

	clickhouseDB, err := new(database.Database).NewConnection(*dbConfig)
	if err != nil {
		log.Fatal(err.Error())
	}
	app.DB = *clickhouseDB

	b := new(buffer.Buffer)
	app.DB.Buffer = b.NewBuffer(viper.GetInt("server.buffer_size"))

	// RabbitMQ configuration
	app.RB.Config = &rabbitmq2.Config{
		Host:         viper.GetString("rabbitmq.host"),
		Port:         viper.GetInt("rabbitmq.port"),
		User:         viper.GetString("rabbitmq.user"),
		Password:     viper.GetString("rabbitmq.password"),
		Exchange:     viper.GetString("rabbitmq.exchange"),
		ExchangeType: viper.GetString("rabbitmq.exchange_type"),
		Queue:        fmt.Sprintf("%v.events.queue", app.Config.Name),
		BindingKey:   viper.GetString("rabbitmq.binding_key"),
		ConsumerTag:  viper.GetString("rabbitmq.consumer_tag"),
	}
	err = app.RB.Config.Validate()
	if err != nil {
		helpers.FailOnError(err, "RabbitMQ config is not valid")
	}
	err = rabbitmq2.InitProducer(&app.RB)
	if err != nil {
		helpers.FailOnError(err, "can' initialise RabbitMQ")
	}

	//	TODO сделать множественные очереди
	cons, err := rabbitmq2.NewConsumer(app.RB.Config, app.RB.Config.Queue, app.DB)
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
		AppName: fmt.Sprintf("go-metr %v v%v", app.Config.FullName, app.Config.Version),
	})

	// Fiber middleware configuration
	if viper.GetBool("server.logging") {
		a.Use(logger.New())
	}
	if viper.GetBool("server.enable_request_id") {
		a.Use(requestid.New())
	}
	if viper.GetBool("server.enable_profiling") {
		a.Use(pprof.New())
	}

	// Fiber endpoints configuration
	a.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).SendString(a.Config().AppName)
	})
	a.Get("/metrics", monitor.New(monitor.Config{Title: "Metrics"}))
	a.Get("/status", HealthCheckHandler)

	// Run Fiber server on port
	log.Fatal(a.Listen(fmt.Sprintf(":%v", app.Config.Port)))
}
