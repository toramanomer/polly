package main

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/toramanomer/polly/api"
	"github.com/toramanomer/polly/repository"
)

func init() {
	envFile, err := os.Open(".env")
	if err != nil {
		log.Fatalf("Failed to read .env file: %v", err)
	}
	defer envFile.Close()

	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		name, value, found := strings.Cut(line, "=")
		if !found {
			log.Fatalf("Invalid line in .env file: %s\n", line)
		}

		if err := os.Setenv(name, value); err != nil {
			log.Fatalf("Failed to set %s environment variable: %v\n", name, err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error scanning .env file: %v\n", err)
	}
}

func main() {
	// -------------------- DB Setup
	connString := os.Getenv("DATABASE_URL")
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Error parsing database URL: %v", err)
	}

	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}
	// --------------------

	// -------------------- API Setup
	var (
		r = chi.NewRouter()
		a = api.NewAPI(repository.NewRepository(db))
	)

	r.Use(middleware.Logger)
	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", a.Signup)
			r.Post("/signin", a.Signin)
			r.Post("/signout", a.Signout)
			r.Get("/me", a.Me)
		})
		r.Route("/polls", func(r chi.Router) {
			withAuth := r.With(api.AuthMiddleware)
			withAuth.Post("/", a.CreatePoll)
			withAuth.Get("/", a.GetUserPolls)
			withAuth.Delete("/{pollID}", a.DeletePoll)

			r.Get("/{pollID}", a.GetPollByID)
			r.With(api.WithTurnstileProtection).Post("/{pollID}/vote", a.VoteOnPoll)
		})
	})

	var (
		serverAddr = ":8000"
		server     = &http.Server{Addr: serverAddr, Handler: r}
	)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error starting the HTTP server: %v", err)
	}
}
