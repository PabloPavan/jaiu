package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres"
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
		log.Fatal("email e password sao obrigatorios")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL nao definido")
	}

	hash, err := service.HashPassword(*password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
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

	if _, err := pool.Exec(ctx, query, *name, *email, hash, *role, *active); err != nil {
		log.Fatalf("insert user: %v", err)
	}

	log.Printf("usuario criado/atualizado: %s (%s)", *email, *role)
}
