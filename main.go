package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/santhosh-tekuri/jsonschema"
	"io"
	"log"
	"net/http"
	"reflect"
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

	// Seeing what happens when I print all variables.
	for k, v := range mux.Vars(r) {
		log.Printf("key=%v, value=%v", k, v)
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.Write(createErrorResponse(w, vars["schemaid"]))
	}
	fmt.Println(string(requestBody))

	// Clean JSON to remove null values.
	var nullsRemoved = removeNulls(toJsonMap(requestBody))
	cleanJson, err := json.Marshal(nullsRemoved)
	fmt.Println(string(cleanJson))

	// Checks if JSON is valid. Here for now as playing with jsonschema.
	validateJson(w, nullsRemoved, vars["schemaid"])

	if err != nil {
		w.Write(createErrorResponse(w, vars["schemaid"]))
	}

	// Create the success response.
	successResponse := SuccessResponse{"uploadSchema", "config-schema", "success"}
	response, err := json.Marshal(successResponse)

	fmt.Printf(string(response))
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

func removeNulls(m map[string]interface{}) map[string]interface{} {

	// I took this algorithm from https://www.ribice.ba/golang-null-values/ and modified it to return the map.
	val := reflect.ValueOf(m)
	for _, e := range val.MapKeys() {
		v := val.MapIndex(e)
		if v.IsNil() {
			delete(m, e.String())
			continue
		}
		switch t := v.Interface().(type) {
		// If key is a JSON object (Go Map), use recursion to go deeper
		case map[string]interface{}:
			removeNulls(t)
		}
	}

	return m
}

/**
  Convert JSON to a JSON map.
*/
func toJsonMap(requestBody []byte) map[string]interface{} {
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(requestBody, &jsonMap)
	if err != nil {
		panic(err)
	}

	return jsonMap
}

func validateJson(w http.ResponseWriter, nullsRemoved map[string]interface{}, schemaId string) {
	sch, err := jsonschema.Compile("resources/config-schema.json")
	if err != nil {
		w.Write(createErrorResponse(w, schemaId))
	}
	if err = sch.ValidateInterface(nullsRemoved); err != nil {
		w.Write(createErrorResponse(w, schemaId))
	}
}
