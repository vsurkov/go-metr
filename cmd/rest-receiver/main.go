package main

import (
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/vsurkov/go-metr/internal/app/instance"
	"github.com/vsurkov/go-metr/internal/common/helpers"
	"github.com/vsurkov/go-metr/internal/platform/database"
	rabbitmq2 "github.com/vsurkov/go-metr/internal/platform/rabbitmq"
	"os"
)

var (
	app     instance.Instance
	Version string
)

func main() {
	configureParams() // Configure viper params and config variables
	// Logger configure
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.Level(viper.GetInt("server.log_level")))

	if viper.GetBool("server.pretty_log") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}) // Pretty console
	}
	log.Info().Msg("logger enabled successfully")

	config := instance.Config{
		Port:     viper.GetInt("server.port"),
		Name:     viper.GetString("server.name"),
		FullName: viper.GetString("server.full_name"),
		Version:  Version,
	}
	err := config.Validate()
	if err != nil {
		helpers.FailOnError(err, helpers.Core, "application config is not valid")
	}

	//viper.WatchConfig()
	//viper.OnConfigChange(func(e fsnotify.Event) {
	//	log.Info().
	//		//Int("level", viper.GetInt("server.log_level")).
	//		Str("Config", e.Name).
	//		Msg("changed logging")
	//	zerolog.SetGlobalLevel(zerolog.Level(int8(viper.GetInt("server.log_level"))))
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
		Host:              viper.GetString("database.host"),
		Port:              viper.GetInt("database.port"),
		Database:          viper.GetString("database.default_database"),
		User:              viper.GetString("database.user"),
		Password:          viper.GetString("database.password"),
		Debug:             viper.GetBool("database.debug"),
		DialTimeout:       viper.GetDuration("database.dial_timeout"),
		MaxOpenConns:      viper.GetInt("database.max_open_conns"),
		MaxIdleConns:      viper.GetInt("database.max_idle_conns"),
		ConnMaxLifetime:   viper.GetDuration("database.conn_max_lifetime"),
		CompressionMethod: clickhouse.CompressionLZ4,
	}
	err = dbConfig.Validate()
	if err != nil {
		helpers.FailOnError(err, helpers.Clickhouse, "database config is not valid")
	}

	clickhouseDB, err := new(database.Database).NewConnection(*dbConfig)
	if err != nil {
		helpers.FailOnError(err, helpers.Clickhouse, "can't create newConnection to DB")
	}
	app.DB = *clickhouseDB

	app.KnownSystems, err = app.DB.GetSystems()
	helpers.FailOnError(err, helpers.Clickhouse, "can't receive systems from default.systems")

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
		helpers.FailOnError(err, helpers.Rabbit, "RabbitMQ config is not valid")
	}
	err = rabbitmq2.NewProducer(&app.RB)
	if err != nil {
		helpers.FailOnError(err, helpers.Rabbit, "can't initialise RabbitMQ producer")
	}

	// Батчи для рэббита в лоб не сделать, можно использовать потом для передачи в массива в бинарной последовательности
	//b := new(buffer.Buffer)
	//app.RB.Buffer = b.NewBuffer(*bufferSize)

	// Fiber configuration
	a := fiber.New(fiber.Config{
		AppName: fmt.Sprintf("go-metr %v v%v", app.Config.FullName, app.Config.Version),
	})

	// Fiber middleware configuration
	if viper.GetBool("server.http_logging") {
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
	a.Post("/event", ReceiveEventHandler)
	a.Get("/event", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusForbidden)
	})

	// Run Fiber server on port
	helpers.FailOnError(a.Listen(fmt.Sprintf(":%v", app.Config.Port)), helpers.Fiber, "can't run Fiber")
}
