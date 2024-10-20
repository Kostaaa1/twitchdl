package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Kostaaa1/twitchdl/web/handlers"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	handler := handlers.New()

	r.HandleFunc("/", handler.HandleHome).Methods("GET")
	r.HandleFunc("/download", handler.HandleDownload).Methods("POST")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	log.Println("Listening on port :8080")
	log.Fatal(srv.ListenAndServe())
}
