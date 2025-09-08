package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
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
	Name  string
	Sides int
}

var DieSides []int = []int{4, 6, 8, 10, 10, 12, 20}

func getDiceInSet(db *sql.DB, setID int) ([]Die, error) {
	query := `SELECT id, name, sides FROM dice WHERE set_id = ? ORDER BY sides`
	rows, err := db.Query(query, setID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dice []Die
	for rows.Next() {
		var d Die
		rows.Scan(&d.ID, &d.Name, &d.Sides)
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

	app.renderPage(w, r, "./ui/dice_home.tmpl", diceSets)
}

func (app *application) deleteDiceSetHandler(w http.ResponseWriter, r *http.Request) {
	setID, err := strconv.Atoi(r.URL.Query().Get("set_id"))
	if err != nil {
		app.badRequest(w, "set_id must be an integer")
		return
	}

	err = app.deleteDiceSet(setID)
	if err != nil {
		app.internalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *application) deleteDiceSet(id int) error {
	query := `DELETE FROM dice_set WHERE id = ?`
	_, err := app.db.Exec(query, id)
	return err
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

	stmtSql := `INSERT INTO dice (sides, set_id, name) VALUES (?, ?, ?)`
	stmt, err := tx.Prepare(stmtSql)
	if err != nil {
		app.internalServerError(w, err)
		tx.Rollback()
		return
	}
	defer stmt.Close()

	d10Count := 0
	for _, sides := range DieSides {
		name := fmt.Sprintf("d%v", sides)
		if sides == 10 {
			d10Count++
			name = fmt.Sprintf("d10 (%v)", d10Count)
		}
		_, err = stmt.Exec(sides, setID, name)
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
	setID, err := strconv.Atoi(r.FormValue("set_id"))
	if err != nil {
		app.badRequest(w, "set_id must be a number")
		return
	}
	rollValues := strings.Split(r.FormValue("rolls"), "\n")

  // TODO: make sure die exists
	// query := `SELECT id FROM dice WHERE die_id = ?`
	// err = app.db.QueryRow(query, dieID).Scan(&dieID)
	// if err != nil {
	// 	if errors.Is(err, sql.ErrNoRows) {
	// 		app.badRequest(w, "dice does not exist")
	// 	} else {
	// 		app.internalServerError(w, err)
	// 	}
	// 	return
	// }

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

  http.Redirect(w, r, fmt.Sprintf("/dice/view-set?set_id=%v&selected_die=%v", setID, dieID), http.StatusSeeOther)
}

func (app *application) viewSetHandler(w http.ResponseWriter, r *http.Request) {
	setID, err := strconv.Atoi(r.URL.Query().Get("set_id"))
	if err != nil {
		app.badRequest(w, "set_id must be an integer")
		return
	}

	selectedDieID, err := strconv.Atoi(r.URL.Query().Get("selected_die"))
	if err != nil {
		selectedDieID = -1
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
	setDice, err := getDiceInSet(app.db, setID)
	if err != nil {
		app.internalServerError(w, err)
		return
	}
	set.Dice = setDice

	if selectedDieID == -1 && len(set.Dice) > 0 {
		selectedDieID = set.Dice[0].ID
	}

	selectedDieIndex := slices.IndexFunc(set.Dice, func(d Die) bool { return d.ID == selectedDieID })
	if selectedDieIndex == -1 {
		app.badRequest(w, "cannot find selected die")
		return
	}
	selectedDie := set.Dice[selectedDieIndex]

	query2 := `SELECT value FROM roll WHERE die_id = ?`
	rows, err := app.db.Query(query2, selectedDieID)
	defer rows.Close()
	if err != nil {
		app.internalServerError(w, err)
		return
	}

	counts := make([]int, selectedDie.Sides)
	total := 0
	for rows.Next() {
		var v int
		err = rows.Scan(&v)
		if err != nil {
			app.internalServerError(w, err)
			return
		}
		if v >= 1 && v <= selectedDie.Sides {
			counts[v-1]++
			total++
		}
	}

	var ideal *float64
	if total > 0 {
		f := float64(total) / float64(selectedDie.Sides)
		f = math.Ceil(f)
		ideal = &f
	}

	var data []ChartPoint
	for i := range selectedDie.Sides {
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
		selected := i == selectedDieIndex
		diceSelected[i] = DieSelected{Die: die, Selected: selected}
	}

	pageData := struct {
		Set        DiceSet
		Dice       []DieSelected
		Die        Die
		TotalRolls int
		JsonData   template.JS
	}{
		Set:        set,
		Dice:       diceSelected,
		Die:        selectedDie,
		TotalRolls: total,
		JsonData:   template.JS(jsonData),
	}
	app.renderPage(w, r, "./ui/dice_dist.tmpl", pageData)
}
