package database

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := map[string]Config{
		"database host is required": {
			Port:     9000,
			User:     "root",
			Password: "",
			Database: "database",
		},
		"database port is required": {
			Host:     "localhost",
			User:     "root",
			Password: "",
			Database: "database",
		},
		"database user is required": {
			Host:     "localhost",
			Port:     9000,
			Password: "",
			Database: "database",
		},
		"database name is required": {
			Host:     "localhost",
			Port:     9000,
			User:     "root",
			Password: "",
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
		Host:     "127.0.0.1",
		Port:     9000,
		User:     "default",
		Password: "",
		Database: "database",
	}

	addres := config.URI()
	assert.Equal(t, "127.0.0.1:9000", addres[0])
}
