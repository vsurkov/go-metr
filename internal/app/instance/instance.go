package instance

import (
	"github.com/vsurkov/go-metr/internal/platform/database"
	"github.com/vsurkov/go-metr/internal/platform/rabbitmq"
)

type Instance struct {
	Config       Config
	DB           database.Database
	RB           rabbitmq.Rabbit
	KnownSystems map[string]string
}

func (i Instance) NewInstance(c Config) *Instance {
	return &Instance{
		Config: Config{
			Port:     c.Port,
			Name:     c.Name,
			FullName: c.FullName,
			Version:  c.Version,
		},
		//DB:           database.Database{},
		//RB:           rabbitmq.Rabbit{},
		//KnownSystems: nil,
	}
}
