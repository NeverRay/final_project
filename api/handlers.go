package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go_final_project/pkg/db"
)

const (
	dateFormat = "20060102"
	tasksLimit = 50
)

type TaskRequest struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskResponse struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func checkDate(task *TaskRequest) error {
	now := time.Now()
	today := now.Format(dateFormat)

	if task.Date == "" || task.Date == " " {
		task.Date = today
		return nil
	}

	if task.Date == "today" {
		task.Date = today
		return nil
	}

	t, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return err
	}

	taskDateStr := t.Format(dateFormat)
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

	_, err := NextDate(time.Now(), time.Now().Format(dateFormat), repeat)
	return err
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	tasks, err := db.GetTasks(tasksLimit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	taskList := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		taskList[i] = TaskResponse{
			ID:      strconv.FormatInt(t.ID, 10),
			Date:    t.Date,
			Title:   t.Title,
			Comment: t.Comment,
			Repeat:  t.Repeat,
		}
	}

	writeJSON(w, http.StatusOK, TasksResponse{Tasks: taskList})
}

func (h *Handler) TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Task not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"id":      strconv.FormatInt(task.ID, 10),
		"date":    task.Date,
		"title":   task.Title,
		"comment": task.Comment,
		"repeat":  task.Repeat,
	})
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.ID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
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
		ID:      id,
		Date:    req.Date,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	}

	if err := db.UpdateTask(task); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Task not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{})
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Task not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{})
}

func (h *Handler) DoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Task not found"})
		return
	}

	if task.Repeat == "" || task.Repeat == " " {
		if err := db.DeleteTask(id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	} else {
		nextDate := CalculateNextDate(task.Date, task.Repeat)
		if nextDate != "" {
			err := db.UpdateTask(&db.Task{
				ID:      id,
				Date:    nextDate,
				Title:   task.Title,
				Comment: task.Comment,
				Repeat:  task.Repeat,
			})
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{})
}

func (h *Handler) TasksHandler(w http.ResponseWriter, r *http.Request) {
	tasksHandler(w, r)
}

func (h *Handler) NextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeatStr := r.URL.Query().Get("repeat")

	var now time.Time
	if nowStr != "" {
		var err error
		now, err = time.Parse(dateFormat, nowStr)
		if err != nil {
			now = time.Now()
		}
	} else {
		now = time.Now()
	}

	nextDate, err := NextDate(now, dateStr, repeatStr)
	if err != nil {
		nextDate = ""
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}
