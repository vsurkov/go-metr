package rabbitmq

import (
	"github.com/streadway/amqp"
	"github.com/vsurkov/go-metr/internal/app/event"
	"github.com/vsurkov/go-metr/internal/common/batch"
)

type Rabbit struct {
	Config     *Config
	Buffer     *batch.Batch
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

type RabbitMsg struct {
	//QueueName string
	Message *event.Event
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}
