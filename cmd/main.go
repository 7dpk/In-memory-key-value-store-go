package main

import (
	"log"
	"net/http"

	"github.com/7dpk/keyvaluestore/database"
	"github.com/7dpk/keyvaluestore/handlers"
	"github.com/gorilla/mux"
)

func main() {
	database := database.NewDatabase()
	handler := &handlers.HTTPHandler{
		Database: database,
	}

	router := mux.NewRouter()
	router.HandleFunc("/", handler.HandleRequest).Methods("POST")

	log.Println("Server started")
	log.Fatal(http.ListenAndServe(":8080", router))
}
