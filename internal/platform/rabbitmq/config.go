package rabbitmq

import (
	"errors"
	"fmt"
)

type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	Exchange     string
	ExchangeType string
	Queue        string
	BindingKey   string
	ConsumerTag  string
}

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if c.Host == "" {
		return errors.New("host is required")
	}
	if c.Port == 0 {
		return errors.New("port is required")
	}
	if c.User == "" {
		return errors.New("user is required")
	}
	if c.Exchange == "" {
		return errors.New("exchange is required")
	}
	if c.ExchangeType == "" {
		return errors.New("exchangeType is required")
	}
	if c.Queue == "" {
		return errors.New("queue is required")
	}
	if c.BindingKey == "" {
		return errors.New("bindKey is required")
	}
	if c.ConsumerTag == "" {
		return errors.New("consumerTag is required")
	}

	return nil
}

// URI returns a Database driver compatible data source name.
func (c Config) URI() string {
	return fmt.Sprintf("amqp://%v:%v@%v:%d/", c.User, c.Password, c.Host, c.Port)
}
