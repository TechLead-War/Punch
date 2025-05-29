package main

import (
	"Punch/handlers"
	"Punch/internal/db"
	"fmt"
	"log"
	"net/http"
)

func main() {
	db.Connect()

	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/slack/report", handlers.SendReportHandler)
	http.HandleFunc("/monthly", handlers.MonthlyReportHandler)
	http.HandleFunc("/export", handlers.ExportExcelHandler)
	http.HandleFunc("/hello", handlers.HelloHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to the Timestamp Service!")
	})

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
