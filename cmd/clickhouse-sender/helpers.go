package main

import (
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

//
//func identity() string {
//	hostname, err := os.Hostname()
//	h := sha1.New()
//	fmt.Fprint(h, hostname)
//	fmt.Fprint(h, err)
//	fmt.Fprint(h, os.Getpid())
//	return fmt.Sprintf("%x", h.Sum(nil))
//}
