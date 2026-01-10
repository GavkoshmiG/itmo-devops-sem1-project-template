package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"project_sem/internal/app"
)

func main() {
	db, err := app.OpenDB()
	if err != nil {
		log.Printf("db open: %v", err)
		return
	}
	defer db.Close()

	if err := app.EnsureSchema(db); err != nil {
		log.Printf("db schema: %v", err)
		return
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v0/prices", app.PostPricesHandler(db)).Methods(http.MethodPost)
	router.HandleFunc("/api/v0/prices", app.GetPricesHandler(db)).Methods(http.MethodGet)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Println("listening on :8080")
	log.Fatal(server.ListenAndServe())
}
