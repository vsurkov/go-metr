package rabbitmq

import (
	"github.com/streadway/amqp"
	"github.com/vsurkov/go-metr/internal/app/event"
	"github.com/vsurkov/go-metr/internal/common/buffer"
)

type Rabbit struct {
	Config     *Config
	Buffer     *buffer.Buffer
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
