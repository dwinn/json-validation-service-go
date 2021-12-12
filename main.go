package main

import (
	"encoding/json"
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

// SuccessResponse Struct for storing a successful response.
type SuccessResponse struct {
	Action string `json:"Action"`
	ID     string `json:"ID"`
	Status string `json:"Status"`
}

// ErrorResponse Struct for storing a successful response.
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

// initializeRoutes Initialize our end points.
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/schema/{schemaid}", a.uploadSchema).Methods("POST")
	a.Router.HandleFunc("/schema/{schemaid}", a.downloadSchema).Methods("GET")
	a.Router.HandleFunc("/validate/{schemaid}", a.validateDocument).Methods("POST")
}

// uploadSchema Endpoint for the upload of a JSON schema.
//
// Returns a success or error response.
func (a *App) uploadSchema(responseWriter http.ResponseWriter, request *http.Request) {
	schemaId := mux.Vars(request)["schemaid"]
	requestBody, _ := io.ReadAll(request.Body)

	// Check if uploaded JSON is valid.
	if toJsonMap(requestBody) == nil {
		createErrorResponse(responseWriter, schemaId)
		return
	}

	// Save the JSON to the disk.
	err := ioutil.WriteFile("json-uploads/"+schemaId+".json", requestBody, os.ModePerm)
	if err != nil {
		createErrorResponse(responseWriter, schemaId)
	}

	// No errors... Create the success response.
	createSuccessResponse(responseWriter, "uploadSchema", schemaId)
}

// downloadSchema Endpoint for the download of a JSON schema.
//
// Returns a success or error response.
func (a *App) downloadSchema(responseWriter http.ResponseWriter, request *http.Request) {

	schemaId := mux.Vars(request)["schemaid"]

	// Get the JSON file from the disk.
	file, err := ioutil.ReadFile("json-uploads/" + schemaId + ".json")
	if err != nil {
		createErrorResponse(responseWriter, schemaId)
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)
	responseWriter.Write(file)
}

// validateDocument Endpoint for the validation of a JSON document against a JSON schema.
//
// Returns a success or error response.
func (a *App) validateDocument(responseWriter http.ResponseWriter, request *http.Request) {
	schemaId := mux.Vars(request)["schemaid"]
	requestBody, _ := io.ReadAll(request.Body)

	// Clean JSON to remove null values.
	var nullsRemoved = removeNulls(toJsonMap(requestBody))

	// Check if uploaded JSON is valid.
	cleanJson, _ := json.Marshal(nullsRemoved)
	if validateJsonSchema(toJsonMap(cleanJson), schemaId) != true {
		createErrorResponse(responseWriter, schemaId)
		return
	}

	// No errors... Create the success response.
	createSuccessResponse(responseWriter, "validateDocument", schemaId)
}

// createSuccessResponse Helper method to create a success response.
func createSuccessResponse(responseWriter http.ResponseWriter, action string, schemaId string) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusCreated)

	successResponse := SuccessResponse{action, schemaId, "success"}

	response, _ := json.Marshal(successResponse)

	responseWriter.Write(response)
}

// createErrorResponse Helper method to create an error response.
func createErrorResponse(responseWriter http.ResponseWriter, schemaId string) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusInternalServerError)

	errorResponse := ErrorResponse{"uploadSchema", schemaId, "error", "Invalid JSON"}

	response, _ := json.Marshal(errorResponse)

	responseWriter.Write(response)
}

// removeNulls Remove all nulls from a JSON map.
func removeNulls(jsonMap map[string]interface{}) map[string]interface{} {

	// This bit was borrowed from a tutorial on https://www.ribice.ba/golang-null-values/
	// I modified it slightly to return the map.
	val := reflect.ValueOf(jsonMap)
	for _, e := range val.MapKeys() {
		v := val.MapIndex(e)
		if v.IsNil() {
			delete(jsonMap, e.String())
			continue
		}
		switch t := v.Interface().(type) {
		// If key is a JSON object (Go Map), use recursion to go deeper
		case map[string]interface{}:
			removeNulls(t)
		}
	}

	return jsonMap
}

// toJsonMap Convert JSON to a JSON map. This will return null if the JSON is invalid.
func toJsonMap(requestBody []byte) map[string]interface{} {
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(requestBody, &jsonMap)

	if err != nil {
		return nil
	}

	return jsonMap
}

// validateJsonSchema Uses the jsonschema library to validate a JSON map against a JSON schema.
//
// Returns true if the JSON is valid, or false if not.
func validateJsonSchema(jsonMap map[string]interface{}, schemaId string) bool {
	sch, err := jsonschema.Compile("json-uploads/" + schemaId + ".json")
	if err != nil {
		return false
	}

	if err = sch.ValidateInterface(jsonMap); err != nil {
		return false
	}

	return true
}
