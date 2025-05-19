package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

type ChartPoint struct {
	Value int      `json:"value"`
	Count int      `json:"count"`
	Ideal *float64 `json:"ideal,omitempty"`
}

type DiceSet struct {
	ID   int
	Name string
	Dice []Die
}

type Die struct {
	ID    int
	Sides int
}

var DieSides []int = []int{4, 6, 8, 10, 12, 20}

func getDiceInSet(db *sql.DB, setID int) ([]Die, error) {
	query := `SELECT id, sides FROM dice WHERE set_id = ? ORDER BY sides`
	rows, err := db.Query(query, setID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dice []Die
	for rows.Next() {
		var d Die
		rows.Scan(&d.ID, &d.Sides)
		dice = append(dice, d)
	}

	return dice, nil
}

func getDiceSets(db *sql.DB) ([]DiceSet, error) {
	query := `SELECT id, name FROM dice_set ORDER BY name`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []DiceSet
	for rows.Next() {
		var set DiceSet
		rows.Scan(&set.ID, &set.Name)
		sets = append(sets, set)
	}

	for _, set := range sets {
		dice, err := getDiceInSet(db, set.ID)
		if err != nil {
			return nil, err
		}
		set.Dice = dice
	}

	return sets, nil
}

func (app *application) diceHandler(w http.ResponseWriter, r *http.Request) {
	diceSets, err := getDiceSets(app.db)
	if err != nil {
		app.internalServerError(w, err)
		return
	}

	err = app.renderPage(w, "./ui/dice_home.tmpl", diceSets)
	if err != nil {
		app.internalServerError(w, err)
	}
}

func (app *application) addDiceSetHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")

	tx, err := app.db.Begin()
	if err != nil {
		app.internalServerError(w, err)
		return
	}

	query := `INSERT INTO dice_set (name) VALUES (?) RETURNING id`
	res, err := app.db.Exec(query, name)
	if err != nil {
		log.Println(err)
		http.Error(w, "Dice set already exists. Please pick a different name", http.StatusConflict)
		tx.Rollback()
		return
	}
	setID, err := res.LastInsertId()
	if err != nil {
		app.internalServerError(w, err)
		tx.Rollback()
		return
	}

	stmtSql := `INSERT INTO dice (sides, set_id) VALUES (?, ?)`
	stmt, err := tx.Prepare(stmtSql)
	if err != nil {
		app.internalServerError(w, err)
		tx.Rollback()
		return
	}
	defer stmt.Close()

	for _, sides := range DieSides {
		_, err = stmt.Exec(sides, setID)
		if err != nil {
			app.internalServerError(w, err)
			tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		app.internalServerError(w, err)
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

	query := `SELECT set_id FROM dice WHERE id = ?`
	var setID int
	err = app.db.QueryRow(query, dieID).Scan(&setID)
	if err != nil {
		app.internalServerError(w, err)
		return
	}

	tx, err := app.db.Begin()
	if err != nil {
		app.internalServerError(w, err)
		return
	}
	stmt, err := tx.Prepare("INSERT INTO roll (die_id, value) VALUES (?, ?)")
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

	http.Redirect(w, r, fmt.Sprintf("/dice/view-set?set_id=%v", setID), http.StatusSeeOther)
}

func (app *application) viewSetHandler(w http.ResponseWriter, r *http.Request) {
	setID, err := strconv.Atoi(r.URL.Query().Get("set_id"))
	if err != nil {
		app.badRequest(w, "set_id must be an integer")
		return
	}

	selectedDieSides, err := strconv.Atoi(r.URL.Query().Get("selected_die"))
	if err != nil {
		selectedDieSides = 12
	}

	if !slices.Contains(DieSides, selectedDieSides) {
		app.badRequest(w, "invalid number of sides")
		return
	}

	var set DiceSet
	setQuery := `SELECT name FROM dice_set WHERE id = ?`
	err = app.db.QueryRow(setQuery, setID).Scan(&set.Name)
	if err != nil {
		log.Println(err)
		app.badRequest(w, "cannot find set with that id")
		return
	}
	set.ID = setID
	set.Dice = make([]Die, len(DieSides))
	for i, sides := range DieSides {
		set.Dice[i] = Die{Sides: sides}
	}

	var dieID int
	query := `SELECT id FROM dice WHERE set_id = ? AND sides = ?`
	err = app.db.QueryRow(query, setID, selectedDieSides).Scan(&dieID)
	if err != nil {
		log.Println(err)
		app.badRequest(w, "cannot find selected die")
		return
	}

	query2 := `SELECT value FROM roll WHERE die_id = ?`
	rows, err := app.db.Query(query2, dieID)
	defer rows.Close()
	if err != nil {
		app.internalServerError(w, err)
		return
	}

	counts := make([]int, selectedDieSides)
	total := 0
	for rows.Next() {
		var v int
		err = rows.Scan(&v)
		if err != nil {
			app.internalServerError(w, err)
			return
		}
		if v >= 1 && v <= selectedDieSides {
			counts[v-1]++
			total++
		}
	}

	var ideal *float64
	if total > 0 {
		f := float64(total) / float64(selectedDieSides)
		ideal = &f
	}

	var data []ChartPoint
	for i := range selectedDieSides {
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

	type DieSelected struct {
		Die
		Selected bool
	}

	diceSelected := make([]DieSelected, len(set.Dice))
	for i, die := range set.Dice {
		selected := die.Sides == selectedDieSides
		diceSelected[i] = DieSelected{Die: die, Selected: selected}
	}

	pageData := struct {
		SetName  string
		SetID int
		Dice     []DieSelected
		Die      Die
		JsonData template.JS
	}{
		SetName:  set.Name,
		SetID: setID,
		Dice: diceSelected,
		Die: Die{ID: dieID, Sides: selectedDieSides},
		JsonData: template.JS(jsonData),
	}
	err = app.renderPage(w, "./ui/dice_dist.tmpl", pageData)
	if err != nil {
		app.internalServerError(w, err)
	}
}
