package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres"
	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/service"
)

func main() {
	var (
		name     = flag.String("name", "Admin", "Nome do usuario")
		email    = flag.String("email", "", "Email do usuario")
		password = flag.String("password", "", "Senha em texto puro")
		role     = flag.String("role", "admin", "Papel (admin/operator)")
		active   = flag.Bool("active", true, "Usuario ativo")
	)
	flag.Parse()

	if *email == "" || *password == "" {
		log.Fatal("email and password are required")
	}
	roleValue := domain.UserRole(*role)
	if !roleValue.IsValid() {
		log.Fatal("invalid role: use admin or operator")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	hash, err := service.HashPassword(*password)
	if err != nil {
		log.Fatalf("hash password failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect db failed: %v", err)
	}
	defer pool.Close()

	const query = `
INSERT INTO users (name, email, password_hash, role, active)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (email) DO UPDATE
SET name = EXCLUDED.name,
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    active = EXCLUDED.active,
    updated_at = now()
`

	if _, err := pool.Exec(ctx, query, *name, *email, hash, string(roleValue), *active); err != nil {
		log.Fatalf("insert user failed: %v", err)
	}

	log.Printf("user created/updated: %s (%s)", *email, *role)
}
