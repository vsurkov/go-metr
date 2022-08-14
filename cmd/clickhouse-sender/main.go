//// Command pubsub is an example of a fanout exchange with dynamic reliable
//// membership, reading from stdin, writing to stdout.
////
//// This example shows how to implement reconnect logic independent from a
//// publish/subscribe loop with bridges to application types.
//
//package main
//
//import (
//	"bufio"
//	"crypto/sha1"
//	"flag"
//	"fmt"
//	"github.com/gofiber/fiber/v2"
//	"github.com/gofiber/fiber/v2/middleware/logger"
//	"github.com/gofiber/fiber/v2/middleware/requestid"
//	"io"
//	"log"
//	"os"
//	"runtime"
//
//	"github.com/streadway/amqp"
//	"golang.org/x/net/context"
//)
//
//var url = flag.String("url", "amqp:///", "AMQP url for both the publisher and subscriber")
//
//// exchange binds the publishers to the subscribers
//const exchange = "pubsub"
//
//// message is the application type for a message.  This can contain identity,
//// or a reference to the recevier chan for further demuxing.
//type message []byte
//
//// session composes an amqp.Connection with an amqp.Channel
//type session struct {
//	*amqp.Connection
//	*amqp.Channel
//}
//
//// Close tears the connection down, taking the channel with it.
//func (s session) Close() error {
//	if s.Connection == nil {
//		return nil
//	}
//	return s.Connection.Close()
//}
//
//// redial continually connects to the URL, exiting the program when no longer possible
//func redial(ctx context.Context, url string) chan chan session {
//	sessions := make(chan chan session)
//
//	go func() {
//		sess := make(chan session)
//		defer close(sessions)
//
//		for {
//			select {
//			case sessions <- sess:
//			case <-ctx.Done():
//				log.Println("shutting down session factory")
//				return
//			}
//
//			conn, err := amqp.Dial(url)
//			if err != nil {
//				log.Fatalf("cannot (re)dial: %v: %q", err, url)
//			}
//
//			ch, err := conn.Channel()
//			if err != nil {
//				log.Fatalf("cannot create channel: %v", err)
//			}
//
//			if err := ch.ExchangeDeclare(exchange, "fanout", false, true, false, false, nil); err != nil {
//				log.Fatalf("cannot declare fanout exchange: %v", err)
//			}
//
//			select {
//			case sess <- session{conn, ch}:
//			case <-ctx.Done():
//				log.Println("shutting down new session")
//				return
//			}
//		}
//	}()
//
//	return sessions
//}
//
//// publish publishes messages to a reconnecting session to a fanout exchange.
//// It receives from the application specific source of messages.
//func publish(sessions chan chan session, messages <-chan message) {
//	for session := range sessions {
//		var (
//			running bool
//			reading = messages
//			pending = make(chan message, 1)
//			confirm = make(chan amqp.Confirmation, 1)
//		)
//
//		pub := <-session
//
//		// publisher confirms for this channel/connection
//		if err := pub.Confirm(false); err != nil {
//			log.Printf("publisher confirms not supported")
//			close(confirm) // confirms not supported, simulate by always nacking
//		} else {
//			pub.NotifyPublish(confirm)
//		}
//
//		log.Printf("publishing...")
//
//	Publish:
//		for {
//			var body message
//			select {
//			case confirmed, ok := <-confirm:
//				if !ok {
//					break Publish
//				}
//				if !confirmed.Ack {
//					log.Printf("nack message %d, body: %q", confirmed.DeliveryTag, string(body))
//				}
//				reading = messages
//
//			case body = <-pending:
//				routingKey := "ignored for fanout exchanges, application dependent for other exchanges"
//				err := pub.Publish(exchange, routingKey, false, false, amqp.Publishing{
//					Body: body,
//				})
//				// Retry failed delivery on the next session
//				if err != nil {
//					pending <- body
//					pub.Close()
//					break Publish
//				}
//
//			case body, running = <-reading:
//				// all messages consumed
//				if !running {
//					return
//				}
//				// work on pending delivery until ack'd
//				pending <- body
//				reading = nil
//			}
//		}
//	}
//}
//
//// identity returns the same host/process unique string for the lifetime of
//// this process so that subscriber reconnections reuse the same queue name.
//func identity() string {
//	hostname, err := os.Hostname()
//	h := sha1.New()
//	fmt.Fprint(h, hostname)
//	fmt.Fprint(h, err)
//	fmt.Fprint(h, os.Getpid())
//	return fmt.Sprintf("%x", h.Sum(nil))
//}
//
//// subscribe consumes deliveries from an exclusive queue from a fanout exchange and sends to the application specific messages chan.
//func subscribe(sessions chan chan session, messages chan<- message) {
//	queue := identity()
//
//	for session := range sessions {
//		sub := <-session
//
//		if _, err := sub.QueueDeclare(queue, false, true, true, false, nil); err != nil {
//			log.Printf("cannot consume from exclusive queue: %q, %v", queue, err)
//			return
//		}
//
//		routingKey := "application specific routing key for fancy toplogies"
//		if err := sub.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
//			log.Printf("cannot consume without a binding to exchange: %q, %v", exchange, err)
//			return
//		}
//
//		deliveries, err := sub.Consume(queue, "", false, true, false, false, nil)
//		if err != nil {
//			log.Printf("cannot consume from: %q, %v", queue, err)
//			return
//		}
//
//		log.Printf("subscribed...")
//
//		for msg := range deliveries {
//			messages <- message(msg.Body)
//			sub.Ack(msg.DeliveryTag, false)
//		}
//	}
//}
//
//// read is this application's translation to the message format, scanning from
//// stdin.
//func read(r io.Reader) <-chan message {
//	lines := make(chan message)
//	go func() {
//		defer close(lines)
//		scan := bufio.NewScanner(r)
//		for scan.Scan() {
//			lines <- message(scan.Bytes())
//		}
//	}()
//	return lines
//}
//
//// write is this application's subscriber of application messages, printing to
//// stdout.
//func write(w io.Writer) chan<- message {
//	lines := make(chan message)
//	go func() {
//		for line := range lines {
//			fmt.Fprintln(w, string(line))
//		}
//	}()
//	return lines
//}
//
//func main() {
//	flag.Parse()
//
//	ctx, done := context.WithCancel(context.Background())
//
//	go func() {
//		publish(redial(ctx, *url), read(os.Stdin))
//		done()
//	}()
//
//	go func() {
//		subscribe(redial(ctx, *url), write(os.Stdout))
//		done()
//	}()
//
//	<-ctx.Done()
//
//	var listeners = runtime.NumCPU() * 8
//
//	fmt.Print(serve(count, listeners))
//
//	a := fiber.New(fiber.Config{
//		AppName: fmt.Sprintf("go-metr"),
//	})
//
//	// Fiber middleware configuration
//	a.Use(logger.New())
//	a.Use(requestid.New())
//
//	// Fiber endpoints configuration
//	a.Get("/", func(ctx *fiber.Ctx) error {
//		return ctx.Status(fiber.StatusOK).SendString(a.Config().AppName)
//	})
//	a.Post("/event", receiveEventHandler)
//
//	// Start Fiber server on port
//	log.Fatal(a.Listen(":3000"))
//
//}

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
	queue        = flag.String("queue", "98dd60853a72.events.queue", "Ephemeral AMQP queue name")
	bindingKey   = flag.String("key", "test-key", "AMQP binding key")
	consumerTag  = flag.String("consumer-tag", "simple-consumer", "AMQP consumer tag (should not be blank)")
	lifetime     = flag.Duration("lifetime", 5*time.Second, "lifetime of process before shutdown (0s=infinite)")
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

	c, err := NewConsumer(*uri, *exchange, *exchangeType, *queue, *bindingKey, *consumerTag)
	if err != nil {
		log.Fatalf("%s", err)
	}

	//if *lifetime > 0 {
	//	log.Printf("running for %s", *lifetime)
	//	time.Sleep(*lifetime)
	//} else {
	//	log.Printf("running forever")
	//	select {}
	//}
	//
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
