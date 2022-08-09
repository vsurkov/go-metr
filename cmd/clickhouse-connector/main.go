// This example declares a durable Exchange, and publishes a single message to
// that Exchange with a given routing key.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)
	
var (
	uri          = flag.String("uri", "amqp://rabbitmq:rabbitmq@localhost:5672/", "AMQP URI")
	exchangeName = flag.String("exchange", "test-exchange", "Durable AMQP exchange name")
	exchangeType = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	routingKey   = flag.String("key", "test-key", "AMQP routing key")
	body         = flag.String("body", "foobar", "Body of message")
	reliable     = flag.Bool("reliable", true, "Wait for the publisher confirmation before exiting")
)

func init() {
	flag.Parse()
}

func main() {
	if err := publish(*uri, *exchangeName, *exchangeType, *routingKey, *body, *reliable); err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("published %dB OK", len(*body))
}

func publish(amqpURI, exchange, exchangeType, routingKey, body string, reliable bool) error {

	// This function dials, connects, declares, publishes, and tears down,
	// all in one go. In a real service, you probably want to maintain a
	// long-lived connection as state, and publish against that.

	log.Printf("dialing %q", amqpURI)
	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}
	defer connection.Close()

	log.Printf("got Connection, getting Channel")
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}

	log.Printf("got Channel, declaring %q Exchange (%q)", exchangeType, exchange)
	if err := channel.ExchangeDeclare(
		exchange,     // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	// Reliable publisher confirms require confirm.select support from the
	// connection.
	if reliable {
		log.Printf("enabling publishing confirms.")
		if err := channel.Confirm(false); err != nil {
			return fmt.Errorf("Channel could not be put into confirm mode: %s", err)
		}

		confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

		defer confirmOne(confirms)
	}

	log.Printf("declared Exchange, publishing %dB body (%q)", len(body), body)
	if err = channel.Publish(
		exchange,   // publish to an exchange
		routingKey, // routing to 0 or more queues
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		return fmt.Errorf("Exchange Publish: %s", err)
	}

	return nil
}

// One would typically keep a channel of publishings, a sequence number, and a
// set of unacknowledged sequence numbers and loop until the publishing channel
// is closed.
func confirmOne(confirms <-chan amqp.Confirmation) {
	log.Printf("waiting for confirmation of one publishing")

	if confirmed := <-confirms; confirmed.Ack {
		log.Printf("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
	} else {
		log.Printf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
	}
}

//package main
//
//import (
//	"context"
//	"fmt"
//	"github.com/ClickHouse/clickhouse-go/v2"
//	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
//	"github.com/gofiber/fiber/v2"
//	"github.com/gofiber/fiber/v2/middleware/logger"
//	"github.com/gofiber/fiber/v2/middleware/monitor"
//	"github.com/gofiber/fiber/v2/middleware/requestid"
//	"github.com/google/uuid"
//	"log"
//	"time"
//)
//
//type Event struct {
//	Date         time.Time `json:"date"`
//	SystemId     uuid.UUID `json:"systemId"`
//	SessionId    uuid.UUID `json:"sessionId"`
//	TotalLoading float64   `json:"totalLoading"`
//	DomLoading   float64   `json:"domLoading"`
//	Uri          string    `json:"uri"`
//	UserAgent    string    `json:"userAgent"`
//}
//
//var (
//	dbConn driver.Conn
//	dbCtx  context.Context
//)
//
//func main() {
//
//	err := initDB()
//	if err != nil {
//		log.Fatal(err.Error())
//	}
//
//	app := fiber.New(fiber.Config{
//		//TODO билдить релиз
//		AppName: "go-metr v0.0.0",
//	})
//
//	app.Use(logger.New())
//	app.Use(requestid.New())
//
//	app.Get("/", func(ctx *fiber.Ctx) error {
//		return ctx.Status(fiber.StatusOK).SendString(app.Config().AppName)
//	})
//	app.Get("/status", healthCheck)
//	app.Get("/metrics", monitor.New(monitor.Config{Title: "Metrics"}))
//
//	eventApp := app.Group("/event")
//	eventApp.Get("", func(ctx *fiber.Ctx) error {
//		return ctx.SendStatus(fiber.StatusForbidden)
//	})
//
//	//TODO добавить таймауты и вынести в параметр
//	eventApp.Post("", createEvent)
//
//	log.Fatal(app.Listen(":3000"))
//}
//
//func createEvent(ctx *fiber.Ctx) error {
//	body := new(Event)
//	err := ctx.BodyParser(body)
//
//	if err != nil {
//		err := ctx.Status(fiber.StatusBadRequest).SendString(err.Error())
//		if err != nil {
//			return err
//		}
//		return err
//	}
//
//	if err := dbConn.Ping(dbCtx); err != nil {
//		if exception, ok := err.(*clickhouse.Exception); ok {
//			log.Printf("Catch exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
//		}
//		return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
//	}
//
//	// TODO нужны валидаторы!
//
//	insertQuery := fmt.Sprintf("INSERT INTO events (date, systemId, sessionId, totalLoading, domLoading, uri, userAgent) VALUES ('%v', '%v', '%v', '%v', '%v', '%v', '%v')",
//
//		time.Now().Format("20060102150405"),
//		body.SystemId,
//		body.SessionId,
//		body.TotalLoading,
//		body.DomLoading,
//		body.Uri,
//		body.UserAgent)
//
//	err = dbConn.Exec(dbCtx, insertQuery)
//
//	if err != nil {
//		return err
//	}
//
//	return ctx.SendStatus(fiber.StatusOK)
//}
//
//// TODO реализовать самодиагностику, посмотри на https://github.com/mackerelio/go-osstat
//func healthCheck(ctx *fiber.Ctx) error {
//	return ctx.Status(fiber.StatusOK).SendString("OK")
//}
//
//func initDB() error {
//	conn, err := clickhouse.Open(&clickhouse.Options{
//		Addr: []string{"127.0.0.1:9000"},
//		Auth: clickhouse.Auth{
//			Database: "default",
//			Username: "default",
//			Password: "",
//		},
//		//Debug:           true,
//		DialTimeout:     time.Second,
//		MaxOpenConns:    10,
//		MaxIdleConns:    5,
//		ConnMaxLifetime: time.Hour,
//		Compression: &clickhouse.Compression{
//			Method: clickhouse.CompressionLZ4,
//		},
//	})
//	if err != nil {
//		return err
//	}
//	dbConn = conn
//
//	ctx := clickhouse.Context(context.Background(), clickhouse.WithSettings(clickhouse.Settings{
//		"max_block_size": 10,
//	}), clickhouse.WithProgress(func(p *clickhouse.Progress) {
//		log.Println("progress: ", p)
//	}), clickhouse.WithProfileInfo(func(p *clickhouse.ProfileInfo) {
//		log.Println("profile info: ", p)
//	}))
//	if err := conn.Ping(ctx); err != nil {
//		if exception, ok := err.(*clickhouse.Exception); ok {
//			log.Printf("Catch exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
//		}
//		return err
//	}
//	dbCtx = ctx
//
//	if err := conn.Exec(ctx, `DROP TABLE IF EXISTS events`); err != nil {
//		return err
//	}
//	err = conn.Exec(ctx, `
//		CREATE TABLE IF NOT EXISTS events (
//			date DateTime,
//			systemId UUID,
//			sessionId UUID,
//			totalLoading Float64,
//			domLoading Float64,
//			uri String,
//			userAgent String
//		) engine=Log
//	`)
//
//	if err != nil {
//		return err
//	}
//	return err
//}
//
////func createTables() error {
////	//CREATE TABLE IF NOT EXISTS systems (
////	//	systemId UUID,
////	//	systemName String
////	//) engine=Log
////}
