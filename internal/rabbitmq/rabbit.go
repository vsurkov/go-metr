package rabbitmq

import "github.com/streadway/amqp"

type Rabbit struct {
	Cfg Config
}

type Config struct {
	Uri          string
	Exchange     string
	ExchangeType string
	//Queue        string
	BindingKey  string
	ConsumerTag string
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}
