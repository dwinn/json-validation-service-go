# JSON Validation Service in Go

### Introduction

This service is designed to validate a JSON document against a JSON schema that can both be uploaded by the user.

This is my first attempt at a Go application, having had no prior experience with the language. I followed a few tutorials online on how to create REST endpoints in Go before beginning.

### Dependencies

In order to run this application, you must first install the dependencies. These include:

**Gorilla Mux.** This is used for routing incoming requests.

`go get github.com/gorilla/mux`

**JSON Schema.** This is used for validating a JSON document against a JSON schema.

`go get github.com/santhosh-tekuri/jsonschema`

**Testify Assert.** This makes unit tests more readable.

`github.com/stretchr/testify/assert`

### Running the Application

`go run main.go`

Then try out the following three curl requests:
```
curl http://localhost:8080/schema/2 -X POST -d @test-resources/test-schema.json
curl http://localhost:8080/schema/2 -X GET
curl http://localhost:8080/validate/2 -X POST -d @test-resources/test-config.json
```

### Unit Testing

For running unit tests, you will need to install Testify.

`go get github.com/stretchr/testify`

To run the unit tests:

`go test`