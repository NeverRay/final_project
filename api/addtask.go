package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go_final_project/pkg/db"
)

type TaskRequest struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

func checkDate(task *TaskRequest) error {
	now := time.Now()
	today := now.Format("20060102")

	if task.Date == "" || task.Date == " " {
		task.Date = today
		return nil
	}

	if task.Date == "today" {
		task.Date = today
		return nil
	}

	t, err := time.Parse("20060102", task.Date)
	if err != nil {
		return err
	}

	taskDateStr := t.Format("20060102")
	if taskDateStr < today {
		if task.Repeat == "" || task.Repeat == " " {
			task.Date = today
		} else {
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil || next == "" {
				task.Date = today
			} else {
				task.Date = next
			}
		}
	}

	return nil
}

func validateRepeat(repeat string) error {
	if repeat == "" || repeat == " " {
		return nil
	}

	_, err := NextDate(time.Now(), time.Now().Format("20060102"), repeat)
	return err
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
