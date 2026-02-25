package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		dsn = "postgres://sentinel:sentinel@localhost:5432/sentinel?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}
	users := []struct{ E, R string }{{"admin@sentinel.local", "admin"}, {"analyst@sentinel.local", "analyst"}, {"viewer@sentinel.local", "viewer"}}
	for _, u := range users {
		h, _ := bcrypt.GenerateFromPassword([]byte("Sentinel#123"), bcrypt.DefaultCost)
		_, err = pool.Exec(context.Background(), `INSERT INTO users(email,password_hash,role) VALUES($1,$2,$3) ON CONFLICT (email) DO UPDATE SET password_hash=$2, role=$3`, u.E, string(h), u.R)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("users seeded")
}
