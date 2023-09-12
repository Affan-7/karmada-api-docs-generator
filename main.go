package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sort"
	"strings"
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
		Paths   interface{}
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

	paths := swaggerJson.Paths.(map[string]interface{})

	uniqueTags := map[string]struct{}{} // It's a set like data structure for go, used to store unique tags

	for _, pathData := range paths {

		pathDataMap := pathData.(map[string]interface{})
		get := pathDataMap["get"]
		getMap := get.(map[string]interface{})
		tags := getMap["tags"]
		tagsSlice := tags.([]interface{})
		for _, tag := range tagsSlice {
			uniqueTags[tag.(string)] = struct{}{}
		}
	}

	// Convert the unique tags to a slice for sorting
	sortedTags := make([]string, 0, len(uniqueTags))
	for tag := range uniqueTags {
		sortedTags = append(sortedTags, tag)
	}

	// Sort the tags alphabetically
	sort.Strings(sortedTags)

	fileName := "index.md"
	file, err := os.Create(fileName)

	if err != nil {
		log.Fatal(err)
	}

	header := "---\ntitle: Karmada API reference docs\n---\n"

	_, err = file.WriteString(header + "\n")

	if err != nil {
		log.Fatal(err)
	}

	_, err = file.WriteString("Packages:\n\n")

	if err != nil {
		log.Fatal(err)
	}

	for _, tag := range sortedTags {

		_, err = file.WriteString("- [" + tag + "](#" + strings.ToLower(tag) + ")\n")

		if err != nil {
			log.Fatal(err)
		}
	}

	for _, tag := range sortedTags {

		_, err = file.WriteString("\n\n## " + tag)

		if err != nil {
			log.Fatal(err)
		}
		for path, pathData := range paths {

			pathDataMap := pathData.(map[string]interface{})
			get := pathDataMap["get"]
			getMap := get.(map[string]interface{})
			tags := getMap["tags"]
			tagsSlice := tags.([]interface{})

			if tag == tagsSlice[0] {
				_, err = file.WriteString("\n\n" + path)

				if err != nil {
					log.Fatal(err)
				}
			}

		}
	}
}
