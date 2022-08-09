package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	//amqp "github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	cfg rabbitConfig
}

type rabbitConfig struct {
	username string
	password string
	host     string
	port     int
}

var (
	rb Rabbit
)

//func (rb Rabbit) Connect(c rabbitConfig) (*Rabbit, error) {
//	uri := fmt.Sprintf("amqp://%v:%v@%v:%d",
//		c.username,
//		c.password,
//		c.host,
//		c.port)
//	conn, err := amqp.Dial(uri)
//	//amqp://guest:guest@localhost:5672/
//	failOnError(err, "Failed to connect to RabbitMQ")
//	defer conn.Close()
//	log.Printf("RabbitMQ: connected to instance")
//
//	ch, err := conn.Channel()
//	failOnError(err, "Failed to open a channel")
//	defer ch.Close()
//	log.Printf("RabbitMQ: channel was opened")
//
//	return &Rabbit{
//		ch: ch,
//	}, nil
//}

func publishEvent(msg *Event) error {

	uri := fmt.Sprintf("amqp://%v:%v@%v:%d",
		rb.cfg.username,
		rb.cfg.password,
		rb.cfg.host,
		rb.cfg.port)
	conn, err := amqp.Dial(uri)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	log.Printf("RabbitMQ: connected to instance")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	log.Printf("RabbitMQ: channel was opened")

	q, err := ch.QueueDeclare(
		"events.queue", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)

	failOnError(err, "Failed to declare a queue")

	// TODO раскоментировать реализацию когда сможешь протестировать под нагрузкой
	//wg := sync.WaitGroup{}
	//wg.Add(1)
	//go func() error {
	//	defer wg.Done()
	//	err = rb.Publish(ch, q, fmt.Sprintf("%v", msg))
	//	if err != nil {
	//		return err
	//	}
	//	return nil
	//}()
	//wg.Wait()
	//return nil
	err = rb.Publish(ch, q, fmt.Sprintf("%v", msg))
	if err != nil {
		return err
	}
	return nil
}

func (rb Rabbit) Publish(ch *amqp.Channel, q amqp.Queue, msg string) error {
	//body := "{\n    \"systemId\":     \"3fade182-685e-4a62-a109-3a17ad87cb33\",\n    \"sessionId\":    \"f7364a70-ddec-459c-86fa-afb76b51935c\",\n    \"totalLoading\": 0.2,\n    \"domLoading\":   0.05,\n    \"uri\":          \"foo://example.com:8042/over/there?name=ferret#nose\",\n    \"userAgent\":    \"Mozilla/5.0 (Macintosh; Intel Mac OS X 12_5) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.6 Safari/605.1.15\"\n}"
	err := ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(msg),
		})
	if err != nil {
		return fmt.Errorf("Failed to publishEvent a message")
	}
	log.Printf(" [x] Sent %s\n", msg)
	return nil
}
