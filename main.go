package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/santhosh-tekuri/jsonschema"
	"log"
	"net/http"
	"os"
)

type App struct {
	Router *mux.Router
}

type SuccessResponse struct {
	Action string `json:"Action"`
	ID     string `json:"ID"`
	Status string `json:"Status"`
}

type ErrorResponse struct {
	Action  string `json:"Action"`
	ID      string `json:"ID"`
	Status  string `json:"Status"`
	Message string `json:"Message"`
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
	a.Router.HandleFunc("/schema/{schemaid}", a.uploadSchema).Methods("POST")
	a.Router.HandleFunc("/schema/{schemaid}", a.downloadSchema).Methods("GET")
	a.Router.HandleFunc("/validate/{schemaid}", a.validateDocument).Methods("POST")
}

func (a *App) uploadSchema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	vars := mux.Vars(r)

	// To be removed/changed.
	_, err := jsonschema.Compile("resources/config-schema.json")

	if err != nil {
		w.Write(createErrorResponse(w, vars["schemaid"]))
	}

	f, err := os.Open("resources/config-schema.json")
	if err != nil {
		w.Write(createErrorResponse(w, vars["schemaid"]))
	}
	defer f.Close()

	// Create the success response.
	successResponse := SuccessResponse{"uploadSchema", vars["schemaid"], "success"}
	response, err := json.Marshal(successResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(response)
}

func (a *App) downloadSchema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	vars := mux.Vars(r)

	var response = "Endpoint " + vars["schemaid"]

	w.Write([]byte(response))
}

func (a *App) validateDocument(w http.ResponseWriter, r *http.Request) {
}

func createErrorResponse(w http.ResponseWriter, schemaId string) []byte {

	errorResponse := ErrorResponse{"uploadSchema", schemaId, "error", "Invalid JSON"}

	response, err := json.Marshal(errorResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	return response
}
