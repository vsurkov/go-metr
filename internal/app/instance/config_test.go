package instance

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := map[string]Config{
		"port is required": {
			Name:     "rncb",
			FullName: "rest-application",
			Version:  "0.0.1",
		},
		"name is required": {
			Port:     3000,
			FullName: "rest-application",
			Version:  "0.0.1",
		},
		"full name is required": {
			Port:    3000,
			Name:    "rncb",
			Version: "0.0.1",
		},
		"version is required": {
			Port:     3000,
			Name:     "rncb",
			FullName: "rest-application",
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
