# JSON Validation Service in Go

This is my first attempt at a Go application, having had no prior experience with the language.

In order to run this application, you must first install Gorilla Mux. This is used for routing incoming requests.

`go get github.com/gorilla/mux`

You must also install JSON Schema. This is used for validating a JSON document against a JSON schema.

`go get github.com/santhosh-tekuri/jsonschema`

To run the application:

`go run main.go`