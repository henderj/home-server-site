package utils

import (
	"html/template"
	"net/http"
)

func RenderPage(w http.ResponseWriter, pagePath string, pageData any) error {
	tmpl, err := template.ParseFiles("./base.tmpl", pagePath)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "base", pageData)
}
