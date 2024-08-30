package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/nguyenbry/crypto-reports/db"
	"github.com/nguyenbry/crypto-reports/discord"
)

type server struct {
	discord *discord.Discord
	router  *chi.Mux
	svr     *http.Server
	jobsSvc *db.JobsService
}

func New(d *discord.Discord, j *db.JobsService) *server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	return &server{discord: d, router: r, jobsSvc: j}
}

func (s *server) ApplyRoutes() {
	s.router.Get("/", s.handleHealthCheck)
	s.router.Post("/jobs", s.handleNewJob)
}

func (s *server) handleNewJob(w http.ResponseWriter, r *http.Request) {
	var data interface{}
	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		log.Printf("decoding body failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m, ok := data.(map[string]interface{})

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	val, ok := m["url"]

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url, ok := val.(string)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rCtx := r.Context()

	ctx, cancel := context.WithTimeout(rCtx, time.Second*5)
	defer cancel()
	// validate url
	err = s.discord.Send(ctx, url, "testing")

	if err != nil {
		log.Printf("posting new job failed, could not send test webhook: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := s.jobsSvc.Create(rCtx, url)

	if err != nil {
		if errors.Is(err, db.ErrUnique) {
			log.Printf("already exists: %v", err)
		} else {
			log.Printf("creating failed: %v", err)
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(id.String()))
}

func (s *server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// test, get random str
	testUrl := uuid.New().String()
	id, err := s.jobsSvc.Create(r.Context(), testUrl)

	// ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	// defer cancel()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Write([]byte(id.String()))
}

func (s *server) Start(addr string) error {
	outer := chi.NewRouter()

	outer.Mount("/api", s.router)

	srv := http.Server{Addr: addr, Handler: outer}
	s.svr = &srv
	return srv.ListenAndServe()
}

func (s *server) Shutdown(c context.Context) error {
	return s.svr.Shutdown(c)
}
