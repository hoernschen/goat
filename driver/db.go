package db

import (
	"database/sql"
	"log"
	"os"

	sqlite "github.com/mattn/go-sqlite3"
)

func NewDB() *DB {
	sql.Register("sqlite3_custom", &sqlite.SQLiteDriver{
		ConnectHook: func(conn *sqlite.SQLiteConn) error {
			return nil
		},
	})

	os.Remove("./goat.db")
	db, err := sql.Open("sqlite3", "./goat.db")

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE users (id text PRIMARY KEY, name text NOT NULL)")
	if err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	_, err = db.Exec("CREATE TABLE rooms (id text PRIMARY KEY, name text NOT NULL)")
	if err != nil {
		log.Fatal("Failed to create rooms table:", err)
	}

	//defer db.Close()
	return db
}
