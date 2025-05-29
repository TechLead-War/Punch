package handlers

import (
	"Punch/internal/db"
	"fmt"
	"net/http"
	"time"

	"github.com/xuri/excelize/v2"
)

func ExportExcelHandler(w http.ResponseWriter, r *http.Request) {
	// Load the Excel template
	f, err := excelize.OpenFile("templates/Punch Template.xlsx")
	if err != nil {
		http.Error(w, "Template load error: "+err.Error(), 500)
		return
	}

	loc := time.Now().Location()
	now := time.Now().In(loc)
	year, month := now.Year(), now.Month()

	punches, err := db.GetPunchesByMonth(year, month)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), 500)
		return
	}

	// Group punches by day
	dailyMap := make(map[string][]time.Time)
	for _, ts := range punches {
		dayKey := ts.In(loc).Format("2006-01-02") // e.g. "2025-05-29"
		dailyMap[dayKey] = append(dailyMap[dayKey], ts.In(loc))
	}

	// Sort and fill from row 2
	sheet := "Sheet1"
	row := 2

	for d := 1; d <= 31; d++ {
		date := time.Date(year, month, d, 0, 0, 0, 0, loc)
		dayStr := date.Format("2006-01-02")
		times := dailyMap[dayStr]

		if len(times) >= 2 {
			// sort timestamps
			min, max := times[0], times[0]
			for _, t := range times {
				if t.Before(min) {
					min = t
				}
				if t.After(max) {
					max = t
				}
			}
			duration := max.Sub(min).Hours()

			// Fill: Date | Day | Hours
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), date.Format("02 Jan"))
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), date.Weekday().String())
			f.SetCellValue(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("%.2f", duration))
		} else {
			// No valid data, fill just date/day
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), date.Format("02 Jan"))
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), date.Weekday().String())
			f.SetCellValue(sheet, fmt.Sprintf("E%d", row), "0.00")
		}

		row++
	}

	// Serve the file
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=punch_report.xlsx")
	_ = f.Write(w)
}
