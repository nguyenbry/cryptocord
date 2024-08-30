package main

import (
	"context"
	"errors"
	"fmt"
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

type appConfig struct {
	jobsSvc *db.JobsService
	discord *discord.Discord
}

func main() {
	env, ok := os.LookupEnv("ENV")

	if !ok {
		log.Fatal("ENV environment variable is required to start app")
	}

	// load from .env if DEV env
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
		log.Fatal("DB_URI missing from env")
	}

	cmc, ok := os.LookupEnv(CMC_ENV_KEY)

	if !ok {
		log.Fatalf("%v missing from env", CMC_ENV_KEY)
	}

	mainCtx, cancelMain := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancelMain()

	myDb, err := db.New(mainCtx, uri)
	if err != nil {
		log.Fatalf("could not build DB: %v", err)
	}
	err = myDb.Ping(mainCtx)
	if err != nil {
		log.Fatalf("ping db failed: %v", err)
	}

	log.Println("db connected")

	// no more fatals after this point because this needs to be called
	defer func() {
		myDb.Close()
		fmt.Println("closing db")
	}()

	// build app config
	cfg := &appConfig{
		jobsSvc: db.NewJobsService(myDb),
		discord: discord.New(&http.Client{}),
	}

	loop := NewLoop(cfg.jobsSvc, crypto.New(cmc), cfg.discord)

	srv := server.New(cfg.discord, cfg.jobsSvc)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func(w *sync.WaitGroup) {
		defer w.Done()

		<-mainCtx.Done()

		// listen for kill signals and give the server some time to die
		ctxWaitShutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctxWaitShutdown); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server shutdown failed: %v", err)
		} else {
			log.Println("graceful exit")
		}
	}(&wg)

	go func(w *sync.WaitGroup) {
		defer w.Done()
		<-loop.Start(mainCtx)

		fmt.Println("yay I am done")
	}(&wg)

	srv.ApplyRoutes()
	err = srv.Start(":4000") // blocking, will continue when server stops listening

	// just print out why the server stopped
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("server failed due to some other reason: %v", err)
	} else {
		log.Println("graceful exit")
	}

	// when the server stops listening, tell the other goroutine about it
	cancelMain()
	wg.Wait()
}
