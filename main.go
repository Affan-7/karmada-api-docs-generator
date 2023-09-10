package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {

	type license struct {
		Name string
		Url  string
	}

	type info struct {
		Description string
		Title       string
		License     license
		Version     string
	}

	type swagger struct {
		Swagger string
		Info    info
	}

	swaggerFile, err := os.Open("swagger.json")

	if err != nil {
		log.Fatal(err)
	}

	defer swaggerFile.Close()

	byteResult, err := io.ReadAll(swaggerFile)

	if err != nil {
		log.Fatal(err)
	}

	var swaggerJson swagger
	json.Unmarshal([]byte(byteResult), &swaggerJson)

	fmt.Println(swaggerJson.Info.Version)
}
