package rabbitmq

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
	"github.com/vsurkov/go-metr/internal/common/helpers"
	"sync"
)

func NewProducer(r *Rabbit) error {
	cfg := r.Config
	log.Info().
		Str("service", helpers.Rabbit).
		Str("method", "NewProducer").
		Dict("dict", zerolog.Dict().
			Str("URI", cfg.URI()).
			Str("User", cfg.User),
		).Msg("dialing")

	conn, err := amqp.Dial(cfg.URI())
	if err != nil {
		log.Error().
			Str("service", helpers.Rabbit).
			Str("method", "NewProducer").
			Dict("dict", zerolog.Dict().
				Str("URI", cfg.URI()).
				Err(err),
			).Msg("AMQP dialling error")
		return err
	}

	log.Info().
		Str("service", helpers.Rabbit).
		Str("method", "NewProducer").
		Dict("dict", zerolog.Dict().
			Str("URI", cfg.URI()),
		).Msg("connected, getting Channel")

	ch, err := conn.Channel()
	if err != nil {
		log.Error().
			Str("service", helpers.Rabbit).
			Str("method", "NewProducer").
			Dict("dict", zerolog.Dict().
				Str("URI", cfg.URI()).
				Err(err),
			).Msg("open channel error")
		return err
	}
	r.Channel = ch

	log.Info().
		Str("service", helpers.Rabbit).
		Str("method", "NewProducer").
		Dict("dict", zerolog.Dict().
			Str("URI", cfg.URI()).
			Str("Exchange", cfg.Exchange).
			Str("ExchangeType", cfg.ExchangeType),
		).Msg("channel upped, declaring Exchange")

	if err = ch.ExchangeDeclare(
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
			Str("method", "NewProducer").
			Dict("dict", zerolog.Dict().
				Str("URI", cfg.URI()).
				Str("Exchange", cfg.Exchange).
				Str("ExchangeType", cfg.ExchangeType).
				Err(err),
			).Msg("exchange declare error")
		return err
	}

	log.Info().
		Str("service", helpers.Rabbit).
		Str("method", "NewProducer").
		Dict("dict", zerolog.Dict().
			Str("URI", cfg.URI()).
			Str("Queue", cfg.Queue),
		).Msg("declared Exchange, declaring Queue")

	_, err = ch.QueueDeclare(
		cfg.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error().
			Str("service", helpers.Rabbit).
			Str("method", "NewProducer").
			Dict("dict", zerolog.Dict().
				Str("URI", cfg.URI()).
				Str("Queue", cfg.Queue).
				Err(err),
			).Msg("queue declare error")
		return err
	}
	return nil
}

func PublishEvent(r Rabbit, rbm *RabbitMsg) error {
	msg := rbm.Message
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
				MessageId:    msg.MessageID.String(),
				ContentType:  "app/json",
				Body:         msg.Body,
			})
		if err != nil {
			log.Error().
				Str("service", helpers.Rabbit).
				Str("method", "PublishEvent").
				Dict("dict", zerolog.Dict().
					Str("Queue", r.Config.Queue).
					Str("MessageId", msg.MessageID.String()).
					Str("ContentType", "app/json").
					Str("Body", "msg.Message.Body").
					Err(err),
				).Msg("publishing message to Queue error")
			helpers.FailOnError(err, helpers.Rabbit, "error on publish message to Queue")
		}
		log.Debug().
			Str("service", helpers.Rabbit).
			Str("method", "PublishEvent").
			Dict("dict", zerolog.Dict().
				Str("Queue", r.Config.Queue).
				Str("MessageId", msg.MessageID.String()).
				Str("ContentType", "app/json").
				Str("Body", string(msg.Body)),
			).Msg("message published to Queue")
	}()
	wg.Wait()
	return nil
}
