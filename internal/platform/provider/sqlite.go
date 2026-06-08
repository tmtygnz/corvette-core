package provider

import (
	"database/sql"
	"log/slog"

	_ "github.com/ncruces/go-sqlite3/driver"
)

type SQLiteProvider struct {
	Conn *sql.DB
}

func CreateSQLiteProvider() *SQLiteProvider {
	db, err := sql.Open("sqlite3", "./corvette.db")
	if err != nil {
		slog.Error("Failed to load corvette.db.", "err", err.Error())
		panic(err)
	}

	slog.Info("LibSQL provider created.")

	return &SQLiteProvider{
		Conn: db,
	}
}

func (lsp *SQLiteProvider) Close() {
	slog.Info("Closing LibSQL provider.")
	lsp.Conn.Close()
}
