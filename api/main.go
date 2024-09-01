package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/nguyenbry/crypto-reports/crypto"
	"github.com/nguyenbry/crypto-reports/db"
	"github.com/nguyenbry/crypto-reports/discord"
	"github.com/nguyenbry/crypto-reports/server"
)

const DB_ENV_KEY = "DB_URI"
const CMC_ENV_KEY = "CMC_KEY"
const ENV_ENV_KEY = "ENV"

type appConfig struct {
	jobsSvc *db.JobsService
	discord *discord.Discord
}

func main() {
	env, ok := os.LookupEnv(ENV_ENV_KEY)

	if !ok {
		log.Fatalf("%v environment variable is required to start app", ENV_ENV_KEY)
	}

	if env == "DEV" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("loading .env file failed")
		}
		log.Println(".env file loaded in development environment")
	}

	// setup pg pool
	uri, ok := os.LookupEnv(DB_ENV_KEY)

	if !ok {
		log.Fatalf("%v missing from env", DB_ENV_KEY)
	}

	cmc, ok := os.LookupEnv(CMC_ENV_KEY)

	if !ok {
		log.Fatalf("%v missing from env", CMC_ENV_KEY)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	myDb, err := db.New(ctx, uri)
	if err != nil {
		log.Fatalf("could not build db: %v", err)
	}
	err = myDb.Ping(ctx)
	if err != nil {
		log.Fatalf("ping db failed: %v", err)
	}

	// no more fatals after this point because this needs to be called
	defer myDb.Close()

	// build app config
	cfg := &appConfig{
		jobsSvc: db.NewJobsService(myDb),
		discord: discord.New(&http.Client{}),
	}

	loop := NewLoop(cfg.jobsSvc, crypto.New(cmc), cfg.discord)

	srv := server.New(cfg.discord, cfg.jobsSvc)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-ctx.Done()

		// listen for kill signals and give the server some time to die
		ctxWaitShutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctxWaitShutdown); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server shutdown failed: %v", err)
		}
	}()

	loop.Start(ctx, time.Minute*10)

	srv.ApplyRoutes()
	err = srv.Start(":4000") // blocking
	// cancel parent context to stop everything else
	cancel()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("server failed due to some other reason: %v", err)
	}

	wg.Wait()

	// wait for loop to terminate
	<-loop.Done()

}
