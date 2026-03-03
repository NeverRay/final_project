package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"go_final_project/api"
	"go_final_project/pkg/db"
)

func main() {
	port := "7540"
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		port = envPort
	}

	dbFile := "scheduler.db"
	if envFile := os.Getenv("TODO_DBFILE"); envFile != "" {
		dbFile = envFile
	}

	if !filepath.IsAbs(dbFile) {
		var err error
		dbFile, err = filepath.Abs(dbFile)
		if err != nil {
			log.Fatal("Failed to get absolute path:", err)
		}
	}

	log.Printf("Using database file: %s", dbFile)

	if err := db.Init(dbFile); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	handler := api.NewHandler()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/task", handler.TaskHandler)
	mux.HandleFunc("/api/task/done", handler.DoneHandler)
	mux.HandleFunc("/api/tasks", handler.TasksHandler)
	mux.HandleFunc("/api/nextdate", handler.NextDateHandler)

	mux.Handle("/", http.FileServer(http.Dir("./web")))

	httpHandler := enableCORS(mux)

	addr := ":" + port
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, httpHandler); err != nil {
		log.Fatal(err)
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
