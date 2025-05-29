package handlers

import (
	"Punch/internal/db"
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

type monthlyData struct {
	Month  string
	Year   int
	DaysJS template.JS
	Total  float64
	Left   float64
}

func MonthlyReportHandler(w http.ResponseWriter, r *http.Request) {
	loc := time.Now().Location()
	now := time.Now().In(loc)
	year, month := now.Year(), now.Month()

	punches, err := db.GetPunchesByMonth(year, month)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	byDay := map[int][]time.Time{}
	for _, ts := range punches {
		d := ts.In(loc).Day()
		byDay[d] = append(byDay[d], ts.In(loc))
	}

	// at top of MonthlyReportHandler, after year/month:
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	lastDay := firstOfMonth.AddDate(0, 1, -1).Day() // e.g. 31 for May

	// replace your old `days := make([]float64, now.Day())` with:
	days := make([]float64, lastDay)
	for d := 1; d <= now.Day(); d++ {
		times := byDay[d]
		if len(times) >= 2 {
			sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
			days[d-1] = times[len(times)-1].Sub(times[0]).Hours()
		}
	}

	var total float64
	for _, h := range days {
		total += h
	}

	// override from env if set, else fallback
	var target float64
	if s := os.Getenv("TOTAL_MONTHLY_HOURS"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			target = v
		} else {
			target = 8.0 * float64(now.Day())
		}
	} else {
		target = 8.0 * float64(now.Day())
	}
	left := target - total

	daysJSON, _ := json.Marshal(days)
	tpl := template.Must(template.ParseFiles("templates/monthly.html"))
	tpl.Execute(w, monthlyData{
		Month:  month.String(),
		Year:   year,
		DaysJS: template.JS(daysJSON),
		Total:  total,
		Left:   left,
	})
}
