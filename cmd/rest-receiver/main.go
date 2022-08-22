package main

import (
	"flag"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/vsurkov/go-metr/internal/app/instance"
	"github.com/vsurkov/go-metr/internal/common/helpers"
	"github.com/vsurkov/go-metr/internal/platform/database"
	rabbitmq2 "github.com/vsurkov/go-metr/internal/platform/rabbitmq"
	"log"
	"time"
)

var (
	app_port              = flag.Int("app_port", 3000, "Application port, default 3000")
	app_full_name         = flag.String("app_full_name", "rest-application", "Application name, default 'rest-application'")
	app_name              = flag.String("app_name", "rncb", "Application name, will be used in Queue prefix, and DB name default 'rncb'")
	app_server_logging    = flag.Bool("app_server_logging", false, "Enabling Fiber http server requests logging, default 'false'")
	app_enable_profiling  = flag.Bool("abb_enable_profiling", false, "Enabling runtime profiling data in the format expected by the pprof visualization tool on http://host:port/debug/pprof/profile, default 'false'")
	app_enable_request_id = flag.Bool("", false, "Enabling server middleware that adds an identifier to the response, default 'false'")

	rabbit_host         = flag.String("rabbit_host", "localhost", "IP address or hostname for connecting to the RabbitMQ, default 'localhost'")
	rabbit_port         = flag.Int("rabbit_port", 5672, "IP address or hostname for connecting to the RabbitMQ, default '5672'")
	rabbit_user         = flag.String("rabbit_user", "rabbitmq", "User name for connecting to the RabbitMQ, default 'rabbitmq'")
	rabbit_password     = flag.String("rabbit_password", "rabbitmq", "Password for connecting to the RabbitMQ, default 'rabbitmq'")
	rabbit_exchange     = flag.String("rabbit_exchange", "exchange", "Durable, non-auto-deleted AMQP exchange name, default 'exchange'")
	rabbit_exchangeType = flag.String("rabbit_exchangeType", "direct", "Exchange type - direct|fanout|topic|x-custom, default direct")
	rabbit_bindingKey   = flag.String("rabbit_bindingKey", "key", "AMQP binding key")
	rabbit_consumerTag  = flag.String("rabbit_consumerTag", "simple-consumer", "AMQP consumer tag (should not be blank)")

	db_host              = flag.String("db_host", "localhost", "IP address or hostname for connecting to the Clickhouse, default 'localhost'")
	db_port              = flag.Int("db_port", 9000, "IP address or hostname for connecting to the Clickhouse, default '9000'")
	db_default_database  = flag.String("db_default_database", "default", "Default database name into Clickhouse, default 'default'")
	db_user              = flag.String("db_user", "default", "User name for connecting to the Clickhouse, default 'default'")
	db_password          = flag.String("db_password", "", "Password for connecting to the Clickhouse, default ''")
	db_debug             = flag.Bool("db_debug", false, "Enable debug mode for database, default false")
	db_dial_timeout      = flag.Duration("db_dial_timeout", time.Second, "Dial database timeout, default 1 second")
	db_max_open_conns    = flag.Int("db_max_open_conns", 10, "Maximum open connections to database, default db_max_idle_conns +5")
	db_max_idle_conns    = flag.Int("db_max_idle_conns", 5, "Maximum idle connections to database, default 5")
	db_conn_max_lifetime = flag.Duration("db_conn_max_lifetime", time.Hour, "Maximum lifetime for connections to database, default 1 hour")

	//bufferSize = flag.Int("bufferSize", 2, "Buffer size for batch database Writing (default 100 messages)")
)

func init() {
	flag.Parse()
}

var (
	app instance.Instance
)

func main() {
	//viper.AddConfigPath("./configs")
	//viper.SetConfigName("config") // Register config file name (no extension)
	//viper.SetConfigType("yaml")   // Look for specific type
	//viper.ReadInConfig()
	//
	//fmt.Println(viper.Get("PORT"))
	//
	//viper.WatchConfig()
	//viper.OnConfigChange(func(e fsnotify.Event) {
	//	fmt.Println("Config file changed:", e.Name)
	//})

	//Application configuration
	app = instance.Instance{
		Config: instance.Config{
			Port:     *app_port,
			Name:     *app_name,
			FullName: *app_full_name,
			Version:  "0.0.1",
		},
		DB:           database.Database{},
		RB:           rabbitmq2.Rabbit{},
		KnownSystems: nil,
	}

	// Clickhouse configuration
	clickhouseDB, err := new(database.Database).NewConnection(database.Config{
		Host:              *db_host,
		Port:              *db_port,
		Database:          *db_default_database,
		User:              *db_user,
		Password:          *db_password,
		Debug:             *db_debug,
		DialTimeout:       *db_dial_timeout,
		MaxOpenConns:      *db_max_open_conns,
		MaxIdleConns:      *db_max_idle_conns,
		ConnMaxLifetime:   *db_conn_max_lifetime,
		CompressionMethod: clickhouse.CompressionLZ4,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	app.DB = *clickhouseDB

	app.KnownSystems, err = app.DB.GetSystems()
	helpers.FailOnError(err, "Can't receive systems")

	// RabbitMQ configuration
	app.RB.Config = &rabbitmq2.Config{
		Host:         *rabbit_host,
		Port:         *rabbit_port,
		User:         *rabbit_user,
		Password:     *rabbit_password,
		Exchange:     *rabbit_exchange,
		ExchangeType: *rabbit_exchangeType,
		Queue:        fmt.Sprintf("%v.events.queue", app.Config.Name),
		BindingKey:   *rabbit_bindingKey,
		ConsumerTag:  *rabbit_consumerTag,
	}
	err = rabbitmq2.InitProducer(&app.RB)
	if err != nil {
		helpers.FailOnError(err, "Can' initialise RabbitMQ")
	}

	// Батчи для рэббита в лоб не сделать, можно использовать потом для передачи в массива в бинарной последовательности
	//b := new(buffer.Buffer)
	//app.RB.Buffer = b.NewBuffer(*bufferSize)

	// Fiber configuration
	a := fiber.New(fiber.Config{
		AppName: fmt.Sprintf("go-metr v %v, app %v", app.Config.Version, app.Config.FullName),
	})

	// Fiber middleware configuration
	if *app_server_logging {
		a.Use(logger.New())
	}
	if *app_enable_request_id {
		a.Use(requestid.New())
	}
	if *app_enable_profiling {
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
	log.Fatal(a.Listen(fmt.Sprintf(":%v", *app_port)))
}
