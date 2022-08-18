package rabbitmq

import (
	"github.com/streadway/amqp"
	"github.com/vsurkov/go-metr/internal/buffer"
)

type Rabbit struct {
	Cfg    Config
	Buffer *buffer.Buffer
}

type Config struct {
	Uri          string
	Exchange     string
	ExchangeType string
	Queue        string
	BindingKey   string
	ConsumerTag  string
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}
