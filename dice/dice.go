package dice

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"joshhend.dev/home-server/utils"
)

func AddRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/update", diceUpdateHandler)
	mux.HandleFunc("/dice", diceHandler)
}

type ChartPoint struct {
	Value int     `json:"value"`
	Count int     `json:"count"`
	Ideal *float64 `json:"ideal,omitempty"`
}

func diceHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderPage(w, "./dice/dice_dist.tmpl", nil)
}

func diceUpdateHandler(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	rollsInput := r.FormValue("rolls")
	die := r.FormValue("die")
	showIdeal := r.FormValue("showIdeal") != ""

	sides, err := strconv.Atoi(strings.TrimPrefix(die, "d"))
	if err != nil || sides < 1 {
		http.Error(w, "Invalid die type", http.StatusBadRequest)
		return
	}

	// Parse rolls
	lines := strings.Split(rollsInput, "\n")
	rolls := []int{}
	for _, line := range lines {
		num, err := strconv.Atoi(strings.TrimSpace(line))
		if err == nil && num >= 1 && num <= sides {
			rolls = append(rolls, num)
		}
	}

	// Build distribution
	counts := make([]int, sides)
	for _, roll := range rolls {
		counts[roll-1]++
	}

	// Build chart data
	var chartData []ChartPoint
	var ideal *float64
	if showIdeal && len(rolls) > 0 {
		v := float64(len(rolls)) / float64(sides)
		ideal = &v
	}

	for i := 0; i < sides; i++ {
		point := ChartPoint{
			Value: i + 1,
			Count: counts[i],
		}
		if ideal != nil {
			point.Ideal = ideal
		}
		chartData = append(chartData, point)
	}

	jsonData, _ := json.Marshal(chartData)

	// Return HTML fragment
	w.Header().Set("Content-Type", "text/html")
	tmpl := `
<div id="chart-container" class="card">
  <canvas id="rollChart" width="600" height="400"></canvas>
  <script type="application/json" id="chart-data">{{.}}</script>
</div>`
	t := template.Must(template.New("chart").Parse(tmpl))
	t.Execute(w, template.JS(jsonData)) // safe JSON injection
}

