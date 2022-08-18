package event

import (
	json2 "encoding/json"
	"github.com/google/uuid"
	"log"
)

type Event struct {
	Timestamp    int64     `json:"Timestamp"`
	MessageID    uuid.UUID `json:"MessageID"`
	SystemId     uuid.UUID `json:"SystemId"`
	SessionId    uuid.UUID `json:"SessionId"`
	TotalLoading float64   `json:"TotalLoading"`
	DomLoading   float64   `json:"DomLoading"`
	Uri          string    `json:"Uri"`
	UserAgent    string    `json:"UserAgent"`
	Body         []byte
}

func (evt Event) Unmarshal(b []byte) *Event {
	msg := new(Event)
	err := json2.Unmarshal(b, &msg)
	if err != nil {
		log.Printf("Unmarshalling error, %v\n", err)
	}
	return msg
}
