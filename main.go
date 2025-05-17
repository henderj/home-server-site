package main

import (
	"log"
	"net/http"

	"joshhend.dev/home-server/dice"
	"joshhend.dev/home-server/utils"
)

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("GET /{$}", homeHandler)

	dice.AddRoutes(mux)

	log.Println("Server listening on port 8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderPage(w, "./home.tmpl", nil)
}
