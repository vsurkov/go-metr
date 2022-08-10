package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"log"
	"sync"
	//amqp "github.com/rabbitmq/amqp091-go"
)

var (
	rb Rabbit
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

func publishEvent(msg []byte, id uuid.UUID) error {
	uri := fmt.Sprintf("amqp://%v:%v@%v:%d",
		rb.cfg.username,
		rb.cfg.password,
		rb.cfg.host,
		rb.cfg.port)
	conn, err := amqp.Dial(uri)
	if err != nil {
		return err
	}
	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			log.Println("RabbitMQ defer closing connection error")
		}
	}(conn)

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			log.Println("RabbitMQ defer closing channel error")
		}
	}(ch)

	q, err := ch.QueueDeclare(
		fmt.Sprintf("%v.events.queue", app.id),
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}
	// Публикация сообщения в очередь
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() error {
		defer wg.Done()
		err = ch.Publish(
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{
				DeliveryMode: 2,
				MessageId:    id.String(),
				ContentType:  "app/json",
				Body:         msg,
			})
		if err != nil {
			return err
		}
		return nil
	}()
	wg.Wait()
	return nil
}
