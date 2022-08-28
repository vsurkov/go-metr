package instance

import (
	"errors"
)

type Config struct {
	Port     int
	Name     string
	FullName string
	Version  string
}

func (c Config) Validate() error {

	if c.Port == 0 {
		return errors.New("port is required")
	}

	if c.Name == "" {
		return errors.New("name is required")
	}

	if c.FullName == "" {
		return errors.New("full name is required")
	}

	if c.Version == "" {
		return errors.New("version is required")
	}
	return nil
}
