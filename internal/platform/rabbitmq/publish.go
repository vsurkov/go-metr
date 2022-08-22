package rabbitmq

import (
	"github.com/streadway/amqp"
	"github.com/vsurkov/go-metr/internal/common/helpers"
	"log"
	"sync"
)

//func (rb Rabbit) WriteBatch(mss []event.Event) error {
//	buf := new(bytes.Buffer)
//	err := binary.Write(buf, binary.LittleEndian, mss)
//	return err
//}

func InitProducer(r *Rabbit) error {
	conn, err := amqp.Dial(r.Config.URI())
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	r.Channel = ch

	if err = ch.ExchangeDeclare(
		r.Config.Exchange,     // name of the exchange
		r.Config.ExchangeType, // type
		true,                  // durable
		false,                 // delete when complete
		false,                 // internal
		false,                 // noWait
		nil,                   // arguments
	); err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		r.Config.Queue,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}
	return nil
}

func PublishEvent(r Rabbit, msg *RabbitMsg) error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := r.Channel.Publish(
			"",
			r.Config.Queue,
			false,
			false,
			amqp.Publishing{
				DeliveryMode: 2,
				MessageId:    msg.Message.MessageID.String(),
				ContentType:  "app/json",
				Body:         msg.Message.Body,
			})
		if err != nil {
			log.Printf("Error on publishing messages to queue - reconnecting to RabbitMQ\n %v", err)

			// On error try to reconnect
			err = InitProducer(&r)
			if err != nil {
				helpers.FailOnError(err, "Can' initialise RabbitMQ")
			}
		}
	}()
	wg.Wait()
	return nil
}
