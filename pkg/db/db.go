package db

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT DEFAULT '',
    repeat VARCHAR(128) DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := err != nil

	DB, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	if err := DB.Ping(); err != nil {
		DB.Close()
		return err
	}

	if install {
		_, err = DB.Exec(schema)
		if err != nil {
			DB.Close()
			return err
		}
	}

	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
