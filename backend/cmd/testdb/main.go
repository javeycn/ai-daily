package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	fmt.Println("Starting SQLite test...")
	if len(os.Args) < 2 {
		fmt.Println("Usage: testdb <db_path>")
		os.Exit(1)
	}
	dbPath := os.Args[1]
	fmt.Printf("Opening database: %s\n", dbPath)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Printf("Open error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println("Database opened, running query...")
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Articles count: %d\n", count)
	fmt.Println("Done!")
}
