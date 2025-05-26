package handlers

import (
	"Punch/internal/db"
	"context"
	"fmt"
	"net/http"
	"time"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	timestamp := time.Now()

	_, err := db.Collection.InsertOne(context.Background(), map[string]interface{}{
		"timestamp": timestamp,
	})
	if err != nil {
		http.Error(w, "DB insert failed", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Timestamp recorded:", timestamp.Format(time.RFC3339))
}
