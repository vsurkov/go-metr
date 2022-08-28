package rabbitmq

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := map[string]Config{
		"host is required": {
			Port:         5672,
			Host:         "",
			User:         "rabbitmq",
			Password:     "",
			Exchange:     "test-exchange",
			ExchangeType: "direct",
			Queue:        "events.queue",
			BindingKey:   "bind-key",
			ConsumerTag:  "simple-consumer",
		},
		"port is required": {
			Host:         "localhost",
			User:         "rabbitmq",
			Password:     "",
			Exchange:     "test-exchange",
			ExchangeType: "direct",
			Queue:        "events.queue",
			BindingKey:   "bind-key",
			ConsumerTag:  "simple-consumer",
		},
		"user is required": {
			Port:         5672,
			Host:         "localhost",
			Password:     "",
			Exchange:     "test-exchange",
			ExchangeType: "direct",
			Queue:        "events.queue",
			BindingKey:   "bind-key",
			ConsumerTag:  "simple-consumer",
		},
		"exchange is required": {
			Port:         5672,
			Host:         "localhost",
			User:         "rabbitmq",
			Password:     "",
			ExchangeType: "direct",
			Queue:        "events.queue",
			BindingKey:   "bind-key",
			ConsumerTag:  "simple-consumer",
		},
		"exchangeType is required": {
			Port:        5672,
			Host:        "localhost",
			User:        "rabbitmq",
			Password:    "",
			Exchange:    "test-exchange",
			Queue:       "events.queue",
			BindingKey:  "bind-key",
			ConsumerTag: "simple-consumer",
		},
		"queue is required": {
			Port:         5672,
			Host:         "localhost",
			User:         "rabbitmq",
			Password:     "",
			Exchange:     "test-exchange",
			ExchangeType: "direct",
			BindingKey:   "bind-key",
			ConsumerTag:  "simple-consumer",
		},
		"bindKey is required": {
			Port:         5672,
			Host:         "localhost",
			User:         "rabbitmq",
			Password:     "",
			Exchange:     "test-exchange",
			ExchangeType: "direct",
			Queue:        "events.queue",
			ConsumerTag:  "simple-consumer",
		},
		"consumerTag is required": {
			Port:         5672,
			Host:         "localhost",
			User:         "rabbitmq",
			Password:     "",
			Exchange:     "test-exchange",
			ExchangeType: "direct",
			Queue:        "events.queue",
			BindingKey:   "bind-key",
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			err := test.Validate()

			assert.EqualError(t, err, name)
		})
	}
}

func TestConfig_Addr(t *testing.T) {
	config := Config{
		Host:     "localhost",
		Port:     5672,
		User:     "rabbitmq",
		Password: "rabbitmq",
	}

	assert.Equal(t, "amqp://rabbitmq:rabbitmq@localhost:5672/", config.URI())
}
