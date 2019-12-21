package env

import "github.com/jmoiron/sqlx"

// Env provides context to handlers
type Env struct {
	DB *sqlx.DB
}

// New returns a new environment
func New(db *sqlx.DB) *Env {
	return &Env{DB: db}
}
