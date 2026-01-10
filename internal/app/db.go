package app

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

const (
	defaultDBHost     = "localhost"
	defaultDBPort     = "5432"
	defaultDBUser     = "validator"
	defaultDBPassword = "val1dat0r"
	defaultDBName     = "project-sem-1"
)

type statsResponse struct {
	TotalItems      int     `json:"total_items"`
	TotalCategories int     `json:"total_categories"`
	TotalPrice      float64 `json:"total_price"`
}

func OpenDB() (*sql.DB, error) {
	host := getEnv("DB_HOST", defaultDBHost)
	port := getEnv("DB_PORT", defaultDBPort)
	user := getEnv("DB_USER", defaultDBUser)
	password := getEnv("DB_PASSWORD", defaultDBPassword)
	name := getEnv("DB_NAME", defaultDBName)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func EnsureSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS prices (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			category VARCHAR(255) NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			create_date TIMESTAMP NOT NULL
		);
	`)
	return err
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
