package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go_final_project/pkg/db"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Title is required"})
		return
	}

	if err := checkDate(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid date format"})
		return
	}

	if req.Repeat != "" && req.Repeat != " " {
		if err := validateRepeat(req.Repeat); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid repeat format"})
			return
		}
	}

	task := &db.Task{
		Date:    req.Date,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	}

	id, err := db.AddTask(task)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": itoa(id)})
}
