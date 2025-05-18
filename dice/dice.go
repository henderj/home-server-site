package dice

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"joshhend.dev/home-server/utils"

	_ "github.com/mattn/go-sqlite3"
)

func AddRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /dice", diceHandler)
	mux.HandleFunc("POST /dice/add-die", addDieHandler)
	mux.HandleFunc("POST /dice/add-roll", addRollHandler)
	mux.HandleFunc("GET /dice/view-die", viewDieHandler)
}

var db *sql.DB

func doDbMigrations() error {
	query := `
	CREATE TABLE IF NOT EXISTS dice (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		sides INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS rolls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		die_id INTEGER,
		value INTEGER,
		FOREIGN KEY (die_id) REFERENCES dice(id)
	)
	`

	_, err := db.Exec(query)
	return err
}

func getDB() (*sql.DB, error) {
	if db != nil {
		return db, nil
	}
	var err error
	db, err = sql.Open("sqlite3", "./dice/dice.db")
	if err != nil {
		return nil, err
	}
	err = doDbMigrations()
	if err != nil {
		return nil, err
	}
	log.Println("'dice' db connected successfully")
	return db, nil
}

type ChartPoint struct {
	Value int      `json:"value"`
	Count int      `json:"count"`
	Ideal *float64 `json:"ideal,omitempty"`
}

func diceHandler(w http.ResponseWriter, r *http.Request) {
	db, err := getDB()
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	query := `SELECT id, name, sides FROM dice ORDER BY name`
	rows, err := db.Query(query)
	defer rows.Close()

	if err != nil {
		utils.InternalServerError(w, err)
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

	utils.RenderPage(w, "./ui/dice_home.tmpl", dice)
}

func addDieHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")
	sides, err := strconv.Atoi(r.FormValue("sides"))
	if err != nil {
		utils.BadRequest(w, "'sides' must be a number")
		return
	}

	db, err := getDB()
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	_, err = db.Exec("INSERT INTO dice (name, sides) VALUES (?, ?)", name, sides)
	if err != nil {
		http.Error(w, "Die already exists. Please pick a different name", http.StatusConflict)
		return
	}

	http.Redirect(w, r, "/dice", http.StatusSeeOther)
}

func addRollHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	dieID, err := strconv.Atoi(r.FormValue("die_id"))
	if err != nil {
		utils.BadRequest(w, "die_id must be a number")
		return
	}
	rollValues := strings.Split(r.FormValue("rolls"), "\n")

	db, err := getDB()
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	stmt, err := tx.Prepare("INSERT INTO rolls (die_id, value) VALUES (?, ?)")
	if err != nil {
		utils.InternalServerError(w, err)
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
		utils.InternalServerError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/dice/view-die?id=%v", dieID), http.StatusSeeOther)
}

func viewDieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		utils.BadRequest(w, "id must be an integer")
		return
	}

	db, err := getDB()
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	var name string
	var sides int
	err = db.QueryRow("SELECT name, sides FROM dice WHERE id = ?", id).Scan(&name, &sides)
	if err != nil {
		log.Println(err)
		utils.BadRequest(w, "cannot find die with that id")
	}

	rows, err := db.Query("SELECT value FROM rolls WHERE die_id = ?", id)
	defer rows.Close()
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	counts := make([]int, sides)
	total := 0
	for rows.Next() {
		var v int
		err = rows.Scan(&v)
		if err != nil {
			utils.InternalServerError(w, err)
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
		utils.InternalServerError(w, err)
		return
	}

	pageData := struct{
		Name string
		Sides int
		ID int
		JsonData template.JS
	}{
		Name: name,
		Sides: sides,
		ID: id,
		JsonData: template.JS(jsonData),
	}
	utils.RenderPage(w, "./ui/dice_dist.tmpl", pageData)
}
