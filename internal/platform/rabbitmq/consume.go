package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/vsurkov/go-metr/internal/app/event"
	"github.com/vsurkov/go-metr/internal/platform/database"
	"log"
)

func NewConsumer(cfg *Config, queueName string, db database.Database) (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     cfg.ConsumerTag,
		done:    make(chan error),
	}

	var err error

	log.Printf("dialing %q", cfg.URI())
	c.conn, err = amqp.Dial(cfg.URI())
	if err != nil {
		return nil, fmt.Errorf("dial: %s", err)
	}

	go func() {
		fmt.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	log.Printf("got Channel, declaring Exchange (%q)", cfg.Exchange)
	if err = c.channel.ExchangeDeclare(
		cfg.Exchange,     // name of the exchange
		cfg.ExchangeType, // type
		true,             // durable
		false,            // delete when complete
		false,            // internal
		false,            // noWait
		nil,              // arguments
	); err != nil {
		return nil, fmt.Errorf("exchange Declare: %s", err)
	}

	log.Printf("declared Exchange, declaring Queue %q", queueName)
	queue, err := c.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("queue Declare: %s", err)
	}

	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, cfg.BindingKey)

	if err = c.channel.QueueBind(
		queue.Name,     // name of the queue
		cfg.BindingKey, // bindingKey
		cfg.Exchange,   // sourceExchange
		false,          // noWait
		nil,            // arguments
	); err != nil {
		return nil, fmt.Errorf("queue bind: %s", err)
	}

	log.Printf("queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		false,      // noAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("queue Consume: %s", err)
	}

	go handle(deliveries, c.done, db)

	return c, nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

func handle(deliveries <-chan amqp.Delivery, done chan error, db database.Database) {
	for d := range deliveries {
		// Working with handled message, save to database
		var msg event.Event
		msg = *msg.Unmarshal(d.Body)
		err := db.Buffer.BuffWrite(db.Buffer, &msg, db)
		if err != nil {
			log.Printf("error on writing batch of Event to database: %v\n", err)
		}
		//buffWrite(msg)
		err = d.Ack(false)
		if err != nil {
			log.Printf("error on sending false ack: %v\n", err)
		}
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil
}
