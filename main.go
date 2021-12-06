package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type App struct {
	Router *mux.Router
}

func main() {
	a := App{}
	a.Initialize()

	a.Run(":8080")
}

func (a *App) Initialize() {
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/schema/{schemaid}", a.getSchema).Methods("GET")
}

func (a *App) getSchema(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	vars := mux.Vars(r)

	var response = "Endpoint " + vars["schemaid"]

	w.Write([]byte(response))
}
