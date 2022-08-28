package buffer

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vsurkov/go-metr/internal/app/event"
	"github.com/vsurkov/go-metr/internal/common/helpers"
	"sync"
)

type BufferedWriter interface {
	WriteBatch(msg []event.Event) error
}

type Buffer struct {
	Buffer []event.Event
	cap    int
	mux    sync.Mutex
}

func (b *Buffer) NewBuffer(cap int) *Buffer {
	return &Buffer{
		Buffer: make([]event.Event, 0),
		cap:    cap,
		mux:    sync.Mutex{},
	}
}

func (b *Buffer) BuffWrite(bf *Buffer, msg *event.Event, bw BufferedWriter) error {
	if len(bf.Buffer) < bf.cap {
		bf.Buffer = append(bf.Buffer, *msg)
	} else {
		bf.mux.Lock()
		err := bw.WriteBatch(bf.Buffer)
		if err != nil {
			return err
		}
		log.Printf("Flush buffered %v messages", bf.cap)
		log.Info().
			Str("service", helpers.Clickhouse).
			Str("method", "BuffWrite").
			Dict("dict", zerolog.Dict().
				Int("BufferSize", bf.cap),
			).Msg("Flushed successfully")
		bf.Buffer = nil
		bf.mux.Unlock()
	}
	return nil
}
