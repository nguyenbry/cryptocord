package main

import (
	"context"
	"fmt"
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
}

func NewLoop(j *db.JobsService, c *crypto.Crypto, d *discord.Discord) *loop {
	return &loop{jobsSvc: j, crypto: c, discord: d}
}

func (l *loop) Start(ctx context.Context) <-chan struct{} {
	t := time.NewTicker(time.Second * 7)

	go func() {
		for {
			select {
			case <-t.C:
				fmt.Println("ayee")
				err := l.singleLoop(ctx)

				if err != nil {
					fmt.Printf("loop failed: %v", err)
				} else {
					fmt.Println("loop finished")
				}
			case <-ctx.Done():
				// stop looping
				return
			}
		}
	}()

	return ctx.Done()
}

func (l *loop) singleLoop(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, time.Minute*10)
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
