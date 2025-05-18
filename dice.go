package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ChartPoint struct {
	Value int      `json:"value"`
	Count int      `json:"count"`
	Ideal *float64 `json:"ideal,omitempty"`
}

func (app *application) diceHandler(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, name, sides FROM dice ORDER BY name`
	rows, err := app.db.Query(query)
	defer rows.Close()

	if err != nil {
		app.internalServerError(w, err)
		return
	}

	type Die struct {
		ID    int
		Name  string
		Sides int
	}
	var dice []Die
	for rows.Next() {
		var d Die
		rows.Scan(&d.ID, &d.Name, &d.Sides)
		dice = append(dice, d)
	}

	app.renderPage(w, "./ui/dice_home.tmpl", dice)
}

func (app *application) addDieHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")
	sides, err := strconv.Atoi(r.FormValue("sides"))
	if err != nil {
		app.badRequest(w, "'sides' must be a number")
		return
	}

	_, err = app.db.Exec("INSERT INTO dice (name, sides) VALUES (?, ?)", name, sides)
	if err != nil {
		http.Error(w, "Die already exists. Please pick a different name", http.StatusConflict)
		return
	}

	http.Redirect(w, r, "/dice", http.StatusSeeOther)
}

func (app *application) addRollHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	dieID, err := strconv.Atoi(r.FormValue("die_id"))
	if err != nil {
		app.badRequest(w, "die_id must be a number")
		return
	}
	rollValues := strings.Split(r.FormValue("rolls"), "\n")

	tx, err := app.db.Begin()
	if err != nil {
		app.internalServerError(w, err)
		return
	}
	stmt, err := tx.Prepare("INSERT INTO rolls (die_id, value) VALUES (?, ?)")
	if err != nil {
		app.internalServerError(w, err)
		return
	}
	defer stmt.Close()

	for _, val := range rollValues {
		if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
			stmt.Exec(dieID, n)
		}
	}
	err = tx.Commit()
	if err != nil {
		app.internalServerError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/dice/view-die?id=%v", dieID), http.StatusSeeOther)
}

func (app *application) viewDieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		app.badRequest(w, "id must be an integer")
		return
	}

	var name string
	var sides int
	err = app.db.QueryRow("SELECT name, sides FROM dice WHERE id = ?", id).Scan(&name, &sides)
	if err != nil {
		log.Println(err)
		app.badRequest(w, "cannot find die with that id")
	}

	rows, err := app.db.Query("SELECT value FROM rolls WHERE die_id = ?", id)
	defer rows.Close()
	if err != nil {
		app.internalServerError(w, err)
		return
	}

	counts := make([]int, sides)
	total := 0
	for rows.Next() {
		var v int
		err = rows.Scan(&v)
		if err != nil {
			app.internalServerError(w, err)
			return
		}
		if v >= 1 && v <= sides {
			counts[v-1]++
			total++
		}
	}

	var ideal *float64
	if total > 0 {
		f := float64(total) / float64(sides)
		ideal = &f
	}

	var data []ChartPoint
	for i := range sides {
		p := ChartPoint{Value: i + 1, Count: counts[i]}
		if ideal != nil {
			p.Ideal = ideal
		}
		data = append(data, p)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		app.internalServerError(w, err)
		return
	}

	pageData := struct {
		Name     string
		Sides    int
		ID       int
		JsonData template.JS
	}{
		Name:     name,
		Sides:    sides,
		ID:       id,
		JsonData: template.JS(jsonData),
	}
	app.renderPage(w, "./ui/dice_dist.tmpl", pageData)
}
