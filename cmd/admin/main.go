package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"sup-anapa/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	username := flag.String("username", "", "admin username")
	password := flag.String("password", "", "admin password")
	flag.Parse()

	*username = strings.TrimSpace(*username)
	*password = strings.TrimSpace(*password)

	if *username == "" {
		log.Fatal("username is required. Usage: go run ./cmd/admin -username admin -password 'secret'")
	}

	if *password == "" {
		log.Fatal("password is required. Usage: go run ./cmd/admin -username admin -password 'secret'")
	}

	if len(*password) < 8 {
		log.Fatal("password must be at least 8 characters")
	}

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	var id int

	err = db.QueryRow(ctx, `
		INSERT INTO admins (username, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (username)
		DO UPDATE SET password_hash = EXCLUDED.password_hash
		RETURNING id
	`, *username, string(hash)).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Fatal("admin was not created or updated")
		}

		fmt.Fprintln(os.Stderr, "failed SQL:")
		fmt.Fprintln(os.Stderr, "INSERT INTO admins (username, password_hash) VALUES (...) ON CONFLICT (username) DO UPDATE ...")
		log.Fatalf("failed to create/update admin: %v", err)
	}

	fmt.Printf("Admin user is ready: username=%s id=%d\n", *username, id)
}
