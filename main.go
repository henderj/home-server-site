package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"joshhend.dev/home-server/dice"
	"joshhend.dev/home-server/utils"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("PORT not set. using 8080 as default")
		port = "8080"
	}

	dbDsn := os.Getenv("DB_DSN")
	if dbDsn == "" {
		log.Println("DB_DSB not set. using ./database.db as default")
		dbDsn = "./database.db"
	}

	db, err := sql.Open("sqlite3", dbDsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v\n", err)
	}
	defer db.Close()

	// TODO: setup db tables

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("GET /{$}", homeHandler)

	dice.AddRoutes(mux)

	log.Printf("Server listening on port %v\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
	if err != nil {
		log.Fatalf("Server failed: %v\n", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderPage(w, "./ui/home.tmpl", nil)
}
