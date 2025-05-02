package main

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
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
	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	r := chi.NewRouter()
	api := api.NewAPI(repository.NewRepository(db))
	r.Post("/api/auth/signup", api.Signup)

	var (
		serverAddr = ":80"
		server     = &http.Server{Addr: serverAddr, Handler: r}
	)

	server.ListenAndServe()
}
