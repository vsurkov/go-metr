package rabbitmq

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/vsurkov/go-metr/internal/event"
	"log"
	"sync"
)

func (rb Rabbit) WriteBatch(mss []event.Event) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, mss)
	return err
}

func PublishEvent(cfg Config, evt *event.Event) error {
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
		fmt.Sprintf("%v.events.queue", cfg.Queue),
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
	go func() {
		defer wg.Done()
		err = ch.Publish(
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{
				DeliveryMode: 2,
				MessageId:    evt.MessageID.String(),
				ContentType:  "app/json",
				Body:         evt.Body,
			})
		if err != nil {
			log.Printf("Error on publishing message %v to queue\n %v", evt.MessageID.String(), err)
		}
	}()
	wg.Wait()
	return nil
}
