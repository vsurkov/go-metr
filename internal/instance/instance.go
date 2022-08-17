package instance

import (
	"github.com/vsurkov/go-metr/internal/db"
	"github.com/vsurkov/go-metr/internal/rabbitmq"
)

type Instance struct {
	Name         string
	Version      string
	ID           string
	DB           db.Database
	RB           rabbitmq.Rabbit
	KnownSystems map[string]string
}
