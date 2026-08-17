package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// initMySQL opens the connection pool and ensures the schema exists.
// Connection settings can be overridden via MYSQL_HOST, MYSQL_PORT, MYSQL_USER,
// MYSQL_PASSWORD and MYSQL_DATABASE environment variables (or a .env file).
func initMySQL() error {
	host := getenvDefault("MYSQL_HOST", "127.0.0.1")
	port := getenvDefault("MYSQL_PORT", "3306")
	user := getenvDefault("MYSQL_USER", "root")
	password := getenvDefault("MYSQL_PASSWORD", "password")
	database := getenvDefault("MYSQL_DATABASE", "calculadora_cdi")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4", user, password, host, port, database)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	if err := conn.Ping(); err != nil {
		return err
	}

	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS simulacoes (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			request JSON NOT NULL,
			response JSON NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	db = conn
	return nil
}

// saveSimulationMySQL persists a request/response pair. Only called when the
// user explicitly asks to save a simulation (POST /save), not automatically
// after every calculation.
func saveSimulationMySQL(req simulateRequest, resp simulateResponse) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("mysql: conexão não inicializada")
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}
	respJSON, err := json.Marshal(resp)
	if err != nil {
		return 0, err
	}

	result, err := db.Exec(
		"INSERT INTO simulacoes (request, response) VALUES (?, ?)",
		reqJSON, respJSON,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func logMySQLUnavailable(err error) {
	log.Printf("mysql indisponível, salvamento manual de simulações desabilitado: %v", err)
}
