package event

import (
	json2 "encoding/json"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vsurkov/go-metr/internal/common/helpers"
)

type Event struct {
	Timestamp    int64     `json:"Timestamp"`
	MessageID    uuid.UUID `json:"MessageID"`
	SystemId     uuid.UUID `json:"SystemId"`
	SessionId    uuid.UUID `json:"SessionId"`
	TotalLoading float64   `json:"TotalLoading"`
	DomLoading   float64   `json:"DomLoading"`
	Uri          string    `json:"URI"`
	UserAgent    string    `json:"UserAgent"`
	Body         []byte
}

func (evt Event) Unmarshal(b []byte) *Event {
	msg := new(Event)
	err := json2.Unmarshal(b, &msg)
	if err != nil {
		log.Error().
			Str("service", helpers.Core).
			Str("method", "Unmarshal").
			Dict("dict", zerolog.Dict().
				Str("SessionID", "").
				Str("Action", "unmarshalling").
				Str("MessageID", "").
				Str("Raw", string(b)).
				Err(err),
			).Msg("unmarshalling error")
	}
	return msg
}
