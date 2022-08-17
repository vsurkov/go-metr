package buffer

import (
	"github.com/vsurkov/go-metr/internal/event"
	"log"
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
		Buffer: make([]event.Event, cap),
		cap:    cap,
		mux:    sync.Mutex{},
	}
}

func (b *Buffer) BuffWrite(bf *Buffer, msg *event.Event, bw BufferedWriter) {
	if len(bf.Buffer) < bf.cap {
		bf.Buffer = append(bf.Buffer, *msg)
	} else {

		bf.mux.Lock()
		err := bw.WriteBatch(bf.Buffer)
		//err := db.WriteBatch(buffer)
		if err != nil {
			log.Printf("Error on writing batch of Event to database: %v\n", err)
		}
		log.Printf("Flush buffered %v messages", bf.cap)
		bf.Buffer = nil
		bf.mux.Unlock()
	}

}
