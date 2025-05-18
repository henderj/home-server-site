package utils

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func RenderPage(w http.ResponseWriter, pagePath string, pageData any) error {
	tmpl, err := template.ParseFiles("./ui/base.tmpl", pagePath)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "base", pageData)
}

func InternalServerError(w http.ResponseWriter, err error) {
	log.Printf("[ERROR] %v\n", err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func BadRequest(w http.ResponseWriter, msg string) {
	http.Error(w, fmt.Sprintf("Bad Request: %v", msg), http.StatusBadRequest)
}
