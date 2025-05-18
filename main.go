package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	db *sql.DB
}

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

	db := setupDB(dbDsn, "setup.sql")
	mux := http.NewServeMux()

	app := application{
		db: db,
	}

	app.routes(mux)

	log.Printf("Server listening on port %v\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
	if err != nil {
		log.Fatalf("Server failed: %v\n", err)
	}
}

func (app *application) routes(mux *http.ServeMux) {
	fs := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("GET /{$}", app.homeHandler)

	mux.HandleFunc("GET /dice", app.diceHandler)
	mux.HandleFunc("POST /dice/add-die", app.addDieHandler)
	mux.HandleFunc("POST /dice/add-roll", app.addRollHandler)
	mux.HandleFunc("GET /dice/view-die", app.viewDieHandler)
}

func (app *application) homeHandler(w http.ResponseWriter, r *http.Request) {
	app.renderPage(w, "./ui/home.tmpl", nil)
}

// Connects to and sets up DB.
// Will exit process if connection or setup fails
func setupDB(dbDsn, setupFile string) *sql.DB {
	db, err := sql.Open("sqlite3", dbDsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v\n", err)
	}
	defer db.Close()

	dbSetupFile, err := os.ReadFile(setupFile)
	if err != nil {
		log.Fatalf("Failed to read sql setup file: %v\n", err)
	}
	_, err = db.Exec(string(dbSetupFile))
	if err != nil {
		log.Fatalf("Failed to execute sql setup file: %v\n", err)
	}
	log.Println("DB setup successfully")
	return db
}

func (*application) renderPage(w http.ResponseWriter, pagePath string, pageData any) error {
	tmpl, err := template.ParseFiles("./ui/base.tmpl", pagePath)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "base", pageData)
}

func (*application) internalServerError(w http.ResponseWriter, err error) {
	log.Printf("[ERROR] %v\n", err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func (*application) badRequest(w http.ResponseWriter, msg string) {
	http.Error(w, fmt.Sprintf("Bad Request: %v", msg), http.StatusBadRequest)
}
