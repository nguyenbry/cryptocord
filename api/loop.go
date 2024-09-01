package main

import (
	"context"
	"log"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/nguyenbry/crypto-reports/crypto"
	"github.com/nguyenbry/crypto-reports/db"
	"github.com/nguyenbry/crypto-reports/discord"
)

type loop struct {
	jobsSvc *db.JobsService
	crypto  *crypto.Crypto
	discord *discord.Discord

	doneChan chan struct{}
}

func NewLoop(j *db.JobsService, c *crypto.Crypto, d *discord.Discord) *loop {
	return &loop{jobsSvc: j, crypto: c, discord: d}
}

func (l *loop) Start(ctx context.Context, d time.Duration) {
	if l.doneChan != nil {
		// already started
		return
	}

	l.doneChan = make(chan struct{})

	go func() {
		defer func() {
			// signal to caller we are done
			l.doneChan <- struct{}{}
		}()

		t := time.NewTicker(d)

		for {
			select {
			case <-ctx.Done():
				// caller tells us to stop
				return
			case <-t.C:
				err := l.singleLoop(ctx)

				if err != nil {
					log.Printf("loop failed: %v", err)
				} else {
					log.Println("loop finished")
				}

			}
		}
	}()
}

func (l *loop) singleLoop(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// get all the jobs
	jobs, err := l.jobsSvc.All(ctx)

	if err != nil {
		return err
	}

	quote, err := l.crypto.Bitcoin(ctx)

	if err != nil {
		return err
	}

	intValue := int(math.Round(quote.Price))
	asStr := strconv.Itoa(intValue)

	wg := sync.WaitGroup{}

	for _, x := range jobs {
		wg.Add(1)

		go func(w *sync.WaitGroup) {
			defer w.Done()

			myCtx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			// fire and forget for now
			l.discord.Send(myCtx, x.Url, asStr)
		}(&wg)

	}

	wg.Wait()

	return nil
}

func (l *loop) Done() <-chan struct{} {
	if l.doneChan == nil {
		panic("can not call done on a nil")
	}

	return l.doneChan
}
