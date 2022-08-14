package buffer

import (
	"log"
	"sync"
)

type BufferedWriter interface {
	writeBatch(msg []Event) error
}

type Buffer struct {
	buffer []Event
	cap    int
	mux    sync.Mutex
}

func (b *Buffer) newBuffer(cap int) *Buffer {
	return &Buffer{
		buffer: make([]Event, cap),
		cap:    cap,
		mux:    sync.Mutex{},
	}
}

func (b *Buffer) buffWrite(bf *Buffer, msg *Event, bw BufferedWriter) {
	if len(bf.buffer) < bf.cap {
		bf.buffer = append(bf.buffer, *msg)
	} else {

		bf.mux.Lock()
		err := bw.writeBatch(bf.buffer)
		//err := db.WriteBatch(buffer)
		if err != nil {
			log.Printf("Error on writing batch of Event to database: %v\n", err)
		}
		log.Printf("Flush buffered %v messages", bf.cap)
		bf.buffer = nil
		bf.mux.Unlock()
	}

}
