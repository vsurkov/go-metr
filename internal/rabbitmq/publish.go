package rabbitmq

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"log"
	"sync"
)

func PublishEvent(cfg Config, qName string, msg []byte, id uuid.UUID) error {
	conn, err := amqp.Dial(cfg.Uri)
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
		fmt.Sprintf("%v.events.queue", qName),
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
