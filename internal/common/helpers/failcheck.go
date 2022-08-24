package helpers

import (
	"log"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%v: %v", msg, err)
	}
}
