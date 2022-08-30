package batch

import (
	"github.com/vsurkov/go-metr/internal/app/event"
	"time"
)

type Writer interface {
	Write(msg []event.Event) error
}

type Batch struct {
	Batches    chan []event.Event
	MsgChan    chan<- event.Event
	MaxItems   int
	MaxTimeout time.Duration
	Ch         chan string
}

func (b *Batch) NewBatch(values chan<- event.Event, maxItems int, maxTimeout time.Duration) *Batch {
	return &Batch{
		Batches:    make(chan []event.Event),
		MsgChan:    values,
		MaxItems:   maxItems,
		MaxTimeout: maxTimeout,
	}
}

//func (b *Batch) Write(msg *event.Event) {
//	b.MsgChan <- *msg
//}

func (b *Batch) Write(msg *event.Event) error {
	b.Ch <- msg.MessageID.String()
	//log.Print(msg.MessageID.String())
	return nil
}

func (b *Batch) BatchMessages(values <-chan event.Event) chan []event.Event {
	go func() {
		//defer close(b.Batches)
		for keepGoing := true; keepGoing; {
			var batch []event.Event
			expire := time.After(b.MaxTimeout)
			for {
				select {
				case value, ok := <-values:
					if !ok {
						keepGoing = false
						goto done
					}

					batch = append(batch, value)
					if len(batch) == b.MaxItems {
						goto done
					}

				case <-expire:
					goto done
				}
			}

		done:
			if len(batch) > 0 {
				b.Batches <- batch
			}
		}
	}()

	return b.Batches
}
