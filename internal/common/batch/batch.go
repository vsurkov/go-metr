package batch

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vsurkov/go-metr/internal/app/event"
	"github.com/vsurkov/go-metr/internal/common/helpers"
	"time"
)

type Writer interface {
	Write(msg []event.Event) error
}

type Batch struct {
	MsgChan    chan<- event.Event
	MaxItems   int
	MaxTimeout time.Duration
}

func (b *Batch) NewBatchWriter(ch chan event.Event, maxItems int, maxTimeout time.Duration, out Writer) *Batch {
	go listenAndWrite(&ch, &maxItems, &maxTimeout, out)
	return &Batch{
		MsgChan:    ch,
		MaxItems:   maxItems,
		MaxTimeout: maxTimeout,
	}
}

func (b *Batch) Write(msg *event.Event) error {
	b.MsgChan <- *msg
	return nil // TODO добавить обратный канал, возвращающий статус записи в базу для возврата ошибки и отката
}

func listenAndWrite(ch *chan event.Event, maxItems *int, maxTimeout *time.Duration, out Writer) {
	batches := CollectBatches(ch, maxItems, maxTimeout)
	for {
		for b := range batches {
			log.Info().
				Str("service", helpers.Batch).
				Str("method", "listenAndWrite").
				Dict("dict", zerolog.Dict().
					Int("Batch size", len(b)).
					Int("Cap", cap(b)).
					Int("MaxItems", *maxItems).
					Dur("MaxTimeout (milliseconds)", *maxTimeout),
				).Msg("flushed from batch")
			err := out.Write(b)
			if err != nil {
				log.Error().
					Str("service", helpers.Batch).
					Str("method", "listenAndWrite").
					Dict("dict", zerolog.Dict().
						Int("Batch size", len(b)).
						Int("Cap", cap(b)).
						Err(err),
					).Msg("on writing to Writer")
			}
		}
	}
}

func CollectBatches(values *chan event.Event, maxItems *int, maxTimeout *time.Duration) chan []event.Event {
	batches := make(chan []event.Event)

	go func() {
		defer close(batches)

		for keepGoing := true; keepGoing; {
			var batch []event.Event
			expire := time.After(*maxTimeout)
			for {
				select {
				case value, ok := <-*values:
					if !ok {
						keepGoing = false
						goto done
					}

					batch = append(batch, value)
					if len(batch) == *maxItems {
						goto done
					}

				case <-expire:
					goto done
				}
			}

		done:
			if len(batch) > 0 {
				batches <- batch
			}
		}
	}()

	return batches
}
