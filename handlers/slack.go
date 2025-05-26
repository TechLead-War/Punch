package handlers

import (
	"Punch/internal/db"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SendReportHandler(w http.ResponseWriter, r *http.Request) {
	totalMonthlyHoursStr := os.Getenv("TOTAL_MONTHLY_HOURS")
	webhook := os.Getenv("SLACK_WEBHOOK_URL")

	totalMonthlyHours, _ := strconv.Atoi(totalMonthlyHoursStr)

	message := buildMonthlyReport(totalMonthlyHours)
	fmt.Println("Generated Message:\n" + message)

	err := sendToSlack(webhook, message)
	if err != nil {
		fmt.Println("Failed to send to Slack:", err)
		http.Error(w, "Slack send failed: "+err.Error(), 500)
		return
	}

	w.Write([]byte(fmt.Sprintf("Slack message sent successfully at %s", time.Now().Format(time.RFC3339))))
}

func buildMonthlyReport(monthlyTotal int) string {
	ctx := context.Background()
	now := time.Now()
	year, month := now.Year(), now.Month()

	// start of current month
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0) // exclusive end of month

	// Fetch punch timestamps from MongoDB
	cursor, err := db.Collection.Find(ctx, bson.M{
		"timestamp": bson.M{"$gte": start, "$lt": end},
	}, options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}}))
	if err != nil {
		return "DB fetch error"
	}
	defer cursor.Close(ctx)

	// Collect punches by day
	dayMap := map[string][]time.Time{}
	for cursor.Next(ctx) {
		var doc struct {
			Timestamp time.Time `bson:"timestamp"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		dayStr := doc.Timestamp.Format("2006-01-02")
		dayMap[dayStr] = append(dayMap[dayStr], doc.Timestamp)
	}

	// Convert days to hours → group by ISO week number
	weekHours := map[int]float64{}
	for dateStr, times := range dayMap {
		if len(times) < 2 {
			continue
		}
		start := times[0]
		end := times[len(times)-1]
		hours := end.Sub(start).Hours()

		date, _ := time.Parse("2006-01-02", dateStr)
		_, week := date.ISOWeek()
		weekHours[week] += hours
	}

	// Build weekly report
	var report []string
	curr := start
	monthUsed := 0.0
	weekIndex := 1

	for curr.Before(end) {
		weekStart := curr
		weekEnd := weekStart.AddDate(0, 0, 6)
		_, weekNum := weekStart.ISOWeek()

		h := weekHours[weekNum]
		monthUsed += h

		report = append(report, fmt.Sprintf("Week%d: %.1f hours (%s - %s)",
			weekIndex,
			h,
			weekStart.Format("2 Jan"),
			weekEnd.Format("2 Jan"),
		))

		curr = weekStart.AddDate(0, 0, 7)
		weekIndex++
	}

	monthLeft := float64(monthlyTotal) - monthUsed

	report = append(report,
		fmt.Sprintf("\nUsed this month: %.1f / %d hours", monthUsed, monthlyTotal),
		fmt.Sprintf("Remaining: %.1f hours", monthLeft),
	)

	return "```" + strings.Join(report, "\n") + "```"
}

func sendToSlack(webhook, message string) error {
	payload := map[string]string{"text": message}
	data, _ := json.Marshal(payload)

	resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
