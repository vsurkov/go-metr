package rabbitmq

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"github.com/vsurkov/go-metr/internal/app/event"
	"github.com/vsurkov/go-metr/internal/common/helpers"
	"github.com/vsurkov/go-metr/internal/platform/database"
)

func NewConsumer(cfg *Config, queueName string, db *database.Database) (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     cfg.ConsumerTag,
		done:    make(chan error),
	}

	var err error

	log.Info().
		Str("service", helpers.Rabbit).
		Str("method", "NewConsumer").
		Dict("dict", zerolog.Dict().
			Str("URI", cfg.URI()).
			Str("User", cfg.User),
		).Msg("dialing")

	c.conn, err = amqp.Dial(cfg.URI())
	if err != nil {
		log.Error().
			Str("service", helpers.Rabbit).
			Str("method", "NewConsumer").
			Dict("dict", zerolog.Dict().
				Str("URI", cfg.URI()).
				Err(err),
			).Msg("AMQP dialling error")
		return nil, err
	}

	go func() {
		log.Error().
			Str("service", helpers.Rabbit).
			Str("method", "NewConsumer").
			Dict("dict", zerolog.Dict().
				Str("URI", cfg.URI()).
				Err(<-c.conn.NotifyClose(make(chan *amqp.Error))),
			).Msg("closing connection")
	}()

	log.Info().
		Str("service", helpers.Rabbit).
		Str("method", "NewConsumer").
		Dict("dict", zerolog.Dict().
			Str("URI", cfg.URI()),
		).Msg("connected, getting Channel")

	c.channel, err = c.conn.Channel()
	if err != nil {
		log.Error().
			Str("service", helpers.Rabbit).
			Str("method", "NewConsumer").
			Dict("dict", zerolog.Dict().
				Str("URI", cfg.URI()).
				Err(err),
			).Msg("open channel error")
		return nil, err
	}

	log.Info().
		Str("service", helpers.Rabbit).
		Str("method", "NewConsumer").
		Dict("dict", zerolog.Dict().
			Str("URI", cfg.URI()).
			Str("Exchange", cfg.Exchange).
			Str("ExchangeType", cfg.ExchangeType),
		).Msg("channel upped, declaring Exchange")

	if err = c.channel.ExchangeDeclare(
		cfg.Exchange,     // name of the exchange
		cfg.ExchangeType, // type
		true,             // durable
		false,            // delete when complete
		false,            // internal
		false,            // noWait
		nil,              // arguments
	); err != nil {
		log.Error().
			Str("service", helpers.Rabbit).
			Str("method", "NewConsumer").
			Dict("dict", zerolog.Dict().
				Str("URI", cfg.URI()).
				Str("Exchange", cfg.Exchange).
				Str("ExchangeType", cfg.ExchangeType).
				Err(err),
			).Msg("exchange declare error")
		return nil, err
	}

	log.Info().
		Str("service", helpers.Rabbit).
		Str("method", "NewConsumer").
		Dict("dict", zerolog.Dict().
			Str("URI", cfg.URI()).
			Str("Queue", queueName),
		).Msg("declared Exchange, declaring Queue")

	queue, err := c.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		log.Error().
			Str("service", helpers.Rabbit).
			Str("method", "NewConsumer").
			Dict("dict", zerolog.Dict().
				Str("URI", cfg.URI()).
				Str("Queue", cfg.Queue).
				Err(err),
			).Msg("queue declare error")
		return nil, err
	}

	log.Info().
		Str("service", helpers.Rabbit).
		Str("method", "NewConsumer").
		Dict("dict", zerolog.Dict().
			Str("Queue", queue.Name).
			Int("Messages", queue.Messages).
			Int("Consumers", queue.Consumers).
			Str("BindingKey", cfg.BindingKey),
		).Msg("declared Exchange, binding to Exchange")

	if err = c.channel.QueueBind(
		queue.Name,     // name of the queue
		cfg.BindingKey, // bindingKey
		cfg.Exchange,   // sourceExchange
		false,          // noWait
		nil,            // arguments
	); err != nil {
		log.Error().
			Str("service", helpers.Rabbit).
			Str("method", "NewConsumer").
			Dict("dict", zerolog.Dict().
				Str("Queue", queue.Name).
				Int("Messages", queue.Messages).
				Int("Consumers", queue.Consumers).
				Str("BindingKey", cfg.BindingKey).
				Err(err),
			).Msg("binding to Exchange error")
		return nil, err
	}

	log.Info().
		Str("service", helpers.Rabbit).
		Str("method", "NewConsumer").
		Dict("dict", zerolog.Dict().
			Str("ConsumerTag", c.tag),
		).Msg("queue bound to Exchange, starting Consumer")

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
		log.Error().
			Str("service", helpers.Rabbit).
			Str("method", "NewConsumer").
			Dict("dict", zerolog.Dict().
				Str("ConsumerTag", c.tag).
				Err(err),
			).Msg("starting Consumer error")
		return nil, err
	}

	go messageStoreHandler(deliveries, c.done, db)
	return c, nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		log.Error().
			Str("service", helpers.Rabbit).
			Str("method", "Shutdown").
			Dict("dict", zerolog.Dict().
				Str("ConsumerTag", c.tag).
				Err(err),
			).Msg("cancel channel error")
		return err
	}

	if err := c.conn.Close(); err != nil {
		log.Error().
			Str("service", helpers.Rabbit).
			Str("method", "Shutdown").
			Dict("dict", zerolog.Dict().
				Str("ConsumerTag", c.tag).
				Err(err),
			).Msg("close connection error")
		return err
	}

	defer func() {
		log.Info().
			Str("service", helpers.Rabbit).
			Str("method", "Shutdown").
			Dict("dict", zerolog.Dict().
				Str("ConsumerTag", c.tag),
			).Msg("AMQP shutdown OK")
	}()

	// wait for messageStoreHandler() to exit
	return <-c.done
}

func messageStoreHandler(deliveries <-chan amqp.Delivery, done chan error, db *database.Database) {
	for d := range deliveries {
		// Working with handled message, save to database
		var msg event.Event
		msg = *msg.Unmarshal(d.Body)

		err := db.Batch.Write(&msg)
		if err != nil {
			err = d.Nack(false, true)
			if err != nil {
				log.Error().
					Str("service", helpers.Rabbit).
					Str("method", "messageStoreHandler").
					Dict("dict", zerolog.Dict().
						Err(err),
					).Msg("error on sending Nack")
			}
			log.Info().
				Str("service", helpers.Rabbit).
				Str("method", "messageStoreHandler").
				Dict("dict", zerolog.Dict().
					Err(err),
				).Msg("Nack sent successfully")
			helpers.FailOnError(err, helpers.Clickhouse, "error on writing batch of Events to database")
		}

		log.Debug().
			Str("service", helpers.Rabbit).
			Str("method", "messageStoreHandler").
			Dict("dict", zerolog.Dict().
				Str("MessageId", msg.MessageID.String()).
				Str("Body", string(msg.Body)),
			).Msg("message written to buffer")

		err = d.Ack(false)
		if err != nil {
			log.Error().
				Str("service", helpers.Rabbit).
				Str("method", "messageStoreHandler").
				Dict("dict", zerolog.Dict().
					Err(err),
				).Msg("error on sending Ack")
		}
	}
	log.Info().
		Str("service", helpers.Rabbit).
		Str("method", "messageStoreHandler").
		Dict("dict", zerolog.Dict()).Msg("AMQP deliveries channel closed")
	done <- nil
}
