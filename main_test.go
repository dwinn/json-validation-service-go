package main

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var a = App{}

func TestUploadSchemaReturnsSuccess(t *testing.T) {
	file, _ := ioutil.ReadFile("test-resources/success.json")

	r, _ := http.NewRequest("POST", "/schema/{schemaid}", bytes.NewBuffer(file))

	var vars = map[string]string{
		"schemaid": "test-config",
	}
	r = mux.SetURLVars(r, vars)

	// Call our method.
	w := httptest.NewRecorder()
	a.uploadSchema(w, r)

	expected := `{"Action":"uploadSchema","ID":"test-config","Status":"success"}`

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, expected, w.Body.String())
	assert.FileExists(t, "json-uploads/test-config.json")
}

func TestUploadSchemaReturnsErrorIfJsonInvalid(t *testing.T) {
	file, _ := ioutil.ReadFile("test-resources/failure.json")

	r, _ := http.NewRequest("POST", "/schema/{schemaid}", bytes.NewBuffer(file))

	var vars = map[string]string{
		"schemaid": "test-config",
	}
	r = mux.SetURLVars(r, vars)

	// Call our method.
	w := httptest.NewRecorder()
	a.uploadSchema(w, r)

	expected := `{"Action":"uploadSchema","ID":"test-config","Status":"error","Message":"Invalid JSON"}`

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, expected, w.Body.String())
}

func TestDownloadSchemaReturnsSchema(t *testing.T) {
	file, _ := ioutil.ReadFile("json-uploads/test-schema.json")

	r, _ := http.NewRequest("GET", "/schema/{schemaid}", nil)

	var vars = map[string]string{
		"schemaid": "test-schema",
	}
	r = mux.SetURLVars(r, vars)

	// Call our method.
	w := httptest.NewRecorder()
	a.downloadSchema(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, string(file), w.Body.String())
	assert.FileExists(t, "json-uploads/test-schema.json")
}

func TestDownloadSchemaReturnsErrorIfErrorGettingFile(t *testing.T) {
	r, _ := http.NewRequest("GET", "/schema/{schemaid}", nil)

	var vars = map[string]string{
		"schemaid": "test-schemaid-not-exists",
	}
	r = mux.SetURLVars(r, vars)

	// Call our method.
	w := httptest.NewRecorder()
	a.downloadSchema(w, r)

	expected := `{"Action":"uploadSchema","ID":"test-schemaid-not-exists","Status":"error","Message":"Invalid JSON"}`

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, expected, w.Body.String())
}

func TestValidateDocumentReturnsSuccess(t *testing.T) {
	file, _ := ioutil.ReadFile("test-resources/test-config.json")

	r, _ := http.NewRequest("POST", "/validate/{schemaid}", bytes.NewBuffer(file))

	var vars = map[string]string{
		"schemaid": "test-config",
	}
	r = mux.SetURLVars(r, vars)

	// Call our method.
	w := httptest.NewRecorder()
	a.validateDocument(w, r)

	expected := `{"Action":"validateDocument","ID":"test-config","Status":"success"}`

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, expected, w.Body.String())
	assert.FileExists(t, "json-uploads/test-config.json")
}

func TestValidateDocumentReturnsErrorIfJsonIsInvalid(t *testing.T) {
	file, _ := ioutil.ReadFile("test-resources/test-config.json")

	r, _ := http.NewRequest("POST", "/validate/{schemaid}", bytes.NewBuffer(file))

	var vars = map[string]string{
		"schemaid": "test-schemaid-not-exists",
	}
	r = mux.SetURLVars(r, vars)

	// Call our method.
	w := httptest.NewRecorder()
	a.validateDocument(w, r)

	expected := `{"Action":"uploadSchema","ID":"test-schemaid-not-exists","Status":"error","Message":"Invalid JSON"}`

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, expected, w.Body.String())
	assert.FileExists(t, "json-uploads/test-config.json")
}

func TestCreateSuccessResponseReturnsSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	createSuccessResponse(w, "uploadSchema", "test-config")

	expected := `{"Action":"uploadSchema","ID":"test-config","Status":"success"}`

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, expected, w.Body.String())
}

func TestCreateErrorResponseReturnsError(t *testing.T) {
	w := httptest.NewRecorder()
	createErrorResponse(w, "test-schemaid-not-exists")

	expected := `{"Action":"uploadSchema","ID":"test-schemaid-not-exists","Status":"error","Message":"Invalid JSON"}`

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, expected, w.Body.String())
}

func TestRemoveNullsRemovesNulls(t *testing.T) {
	file, _ := ioutil.ReadFile("test-resources/test-config.json")
	fileNoNulls, _ := ioutil.ReadFile("test-resources/test-config-no-nulls.json")

	var result = removeNulls(toJsonMap(file))
	assert.Equal(t, toJsonMap(fileNoNulls), result)
}

func TestToJsonMapConvertsToJsonMap(t *testing.T) {
	file, _ := ioutil.ReadFile("test-resources/test-config.json")

	jsonMap := make(map[string]interface{})
	_ = json.Unmarshal(file, &jsonMap)

	assert.Equal(t, jsonMap, toJsonMap(file))
}

func TestToJsonMapReturnsNullIfInvalidJson(t *testing.T) {
	file, _ := ioutil.ReadFile("test-resources/test-config-not-exists.json")

	assert.Nil(t, toJsonMap(file))
}

func TestValidateJsonSchemaReturnsTrueIfDocumentMatches(t *testing.T) {
	file, _ := ioutil.ReadFile("test-resources/test-config.json")

	assert.True(t, validateJsonSchema(removeNulls(toJsonMap(file)), "test-schema"))
}

func TestValidateJsonSchemaReturnsFalseIfSchemaCannotBeFound(t *testing.T) {
	file, _ := ioutil.ReadFile("test-resources/test-config-not-exists.json")

	assert.False(t, validateJsonSchema(toJsonMap(file), "test-schema"))
}

func TestValidateJsonSchemaReturnsFalseIfDocumentDoesNotMatch(t *testing.T) {
	file, _ := ioutil.ReadFile("test-resources/test-config.json")

	assert.False(t, validateJsonSchema(toJsonMap(file), "test-schema"))
}
