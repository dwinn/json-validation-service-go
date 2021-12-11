package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/santhosh-tekuri/jsonschema"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

/**
Endpoint for the upload of a JSON schema.
*/
func (a *App) uploadSchema(responseWriter http.ResponseWriter, request *http.Request) {

	schemaId := mux.Vars(request)["schemaid"]

	// Check if JSON exists in the body.
	requestBody, err := io.ReadAll(request.Body)
	if err != nil {
		createErrorResponse(responseWriter, schemaId)
		return
	}

	// Check if uploaded JSON is valid.
	if toJsonMap(requestBody) == nil {
		createErrorResponse(responseWriter, schemaId)
		return
	}

	// Save the JSON to the disk.
	ioutil.WriteFile("json-uploads/"+schemaId+".json", requestBody, os.ModePerm)

	// No errors... Create the success response.
	createSuccessResponse(responseWriter, schemaId)
}

func (a *App) downloadSchema(responseWriter http.ResponseWriter, request *http.Request) {

	schemaId := mux.Vars(request)["schemaid"]

	// Get the JSON file from the disk.
	file, err := ioutil.ReadFile("json-uploads/" + schemaId + ".json")

	if err != nil {
		createErrorResponse(responseWriter, schemaId)
	}

	responseWriter.Write(file)
}

func (a *App) validateDocument(responseWriter http.ResponseWriter, request *http.Request) {

	schemaId := mux.Vars(request)["schemaid"]

	// Check if JSON exists in the body.
	requestBody, err := io.ReadAll(request.Body)
	if err != nil {
		createErrorResponse(responseWriter, schemaId)
		return
	}

	// Clean JSON to remove null values.
	var nullsRemoved = removeNulls(toJsonMap(requestBody))
	cleanJson, err := json.Marshal(nullsRemoved)
	fmt.Println(string(cleanJson))

	// Check if uploaded JSON is valid.
	if validateJsonSchema(toJsonMap(cleanJson)) != true {
		createErrorResponse(responseWriter, schemaId)
		return
	}

	// No errors... Create the success response.
	createSuccessResponse(responseWriter, schemaId)
}

func createErrorResponse(responseWriter http.ResponseWriter, schemaId string) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusInternalServerError)

	errorResponse := ErrorResponse{"uploadSchema", schemaId, "error", "Invalid JSON"}

	response, err := json.Marshal(errorResponse)
	if err != nil {
		createErrorResponse(responseWriter, schemaId)
	}

	responseWriter.Write(response)
}

func createSuccessResponse(responseWriter http.ResponseWriter, schemaId string) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusCreated)

	successResponse := SuccessResponse{"uploadSchema", schemaId, "success"}

	response, err := json.Marshal(successResponse)
	if err != nil {
		createErrorResponse(responseWriter, schemaId)
	}

	responseWriter.Write(response)
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
  Convert JSON to a JSON map. This will return null if the JSON is invalid.
*/
func toJsonMap(requestBody []byte) map[string]interface{} {
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(requestBody, &jsonMap)

	if err != nil {
		return nil
	}

	return jsonMap
}

func validateJsonSchema(nullsRemoved map[string]interface{}) bool {
	sch, err := jsonschema.Compile("resources/config-schema.json")
	if err != nil {
		return false
	}

	if err = sch.ValidateInterface(nullsRemoved); err != nil {
		return false
	}

	return true
}
