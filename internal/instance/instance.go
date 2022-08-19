package instance

import (
	"github.com/vsurkov/go-metr/internal/db"
	"github.com/vsurkov/go-metr/internal/rabbitmq"
)

type Instance struct {
	FullName     string
	Version      string
	Name         string
	DB           db.Database
	RB           rabbitmq.Rabbit
	KnownSystems map[string]string
}
