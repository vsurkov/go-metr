package helpers

import (
	"github.com/rs/zerolog/log"
)

func FailOnError(err error, service string, msg string) {
	if err != nil {
		log.Fatal().
			Err(err).
			Str("service", service).
			Msgf("%s", msg)
	}
}
