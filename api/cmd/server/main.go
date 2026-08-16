package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"kylistran-api/internal/auth"
	"kylistran-api/internal/db"
	"kylistran-api/internal/handlers"
)

func main() {
	loadDotEnv(".env")

	dbPath := getenv("DB_PATH", "./kylistran.db")
	port := getenv("PORT", "8080")
	jwtSecret := mustEnv("JWT_SECRET")
	origins := strings.Split(mustEnv("ALLOWED_ORIGINS"), ",")
	adminUsername := mustEnv("ADMIN_USERNAME")
	adminPassword := mustEnv("ADMIN_PASSWORD")

	conn, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	authSvc := auth.New(jwtSecret, 30*24*time.Hour)

	adminHash, err := authSvc.HashPassword(adminPassword)
	if err != nil {
		log.Fatalf("hash admin password: %v", err)
	}
	if err := db.SeedAdmin(conn, adminUsername, adminHash); err != nil {
		log.Fatalf("seed admin: %v", err)
	}

	h := handlers.New(conn, authSvc)
	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("kylistran-api listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, auth.CORS(origins)(mux)))
}

// loadDotEnv populates process env vars from a simple KEY=VALUE file, one per
// line, for local development. Vars already set in the environment win, so
// this never overrides Docker's env_file or a real deployment's env.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		os.Setenv(key, value)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required environment variable %s", key)
	}
	return v
}
