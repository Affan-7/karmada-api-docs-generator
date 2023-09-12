package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {

	// structs to load json
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
		Paths   map[string]interface{}
	}

	// Read the swagger.json file
	swaggerFile, err := os.Open("swagger.json")

	if err != nil {
		log.Fatal(err)
	}

	defer swaggerFile.Close()

	byteResult, err := io.ReadAll(swaggerFile)

	if err != nil {
		log.Fatal(err)
	}

	// Load the json into structs
	var swaggerJson swagger
	json.Unmarshal([]byte(byteResult), &swaggerJson)

	// Get the unique tags form the swagger.json
	paths := swaggerJson.Paths

	uniqueTags := map[string]struct{}{}

	for _, pathData := range paths {

		pathDataMap := pathData.(map[string]interface{})
		for _, method := range []string{"get", "post", "put", "delete", "patch"} {
			if methodData, ok := pathDataMap[method]; ok {
				methodMap := methodData.(map[string]interface{})
				tags := methodMap["tags"].([]interface{})
				for _, tag := range tags {
					uniqueTags[tag.(string)] = struct{}{}
				}
			}
		}
	}

	// sort the unique tags
	sortedTags := make([]string, 0, len(uniqueTags))
	for tag := range uniqueTags {
		sortedTags = append(sortedTags, tag)
	}

	sort.Strings(sortedTags)

	// Create index.md with the header
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

	// Create index of packages with links
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
		_, err = file.WriteString("\n## " + tag + "\n\n")

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("\n## " + tag + "\n")

		sortedPathWithMethod := make([]map[string]string, 0)

		for path, pathData := range paths {
			pathDataMap := pathData.(map[string]interface{})
			for method, methodData := range pathDataMap {
				if method != "parameters" {
					methodMap := methodData.(map[string]interface{})
					tags := methodMap["tags"].([]interface{})
					for _, tag1 := range tags {
						if tag == tag1 {
							sortedPathWithMethod = append(sortedPathWithMethod, map[string]string{path: method})
						}
					}
				}
			}
		}
		sortData(&sortedPathWithMethod)

		for _, myMap := range sortedPathWithMethod {
			for path, method := range myMap {
				_, err = file.WriteString(strings.ToUpper(method) + " " + path + "\n\n")

				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(strings.ToUpper(method) + " " + path)
			}
		}

	}
}

func sortData(data *[]map[string]string) {
	// A custom sorting function

	valuePriority := map[string]int{
		"get":     1,
		"put":     2,
		"post":    3,
		"delete":  4,
		"options": 5,
		"head":    6,
		"patch":   7,
	}

	sort.Slice(*data, func(i, j int) bool {
		keysI := make([]string, 0, len((*data)[i]))
		keysJ := make([]string, 0, len((*data)[j]))
		valuesI := make([]string, 0, len((*data)[i]))
		valuesJ := make([]string, 0, len((*data)[j]))

		// Extract keys from maps
		for key, value := range (*data)[i] {
			keysI = append(keysI, key)
			valuesI = append(valuesI, value)
		}
		for key, value := range (*data)[j] {
			keysJ = append(keysJ, key)
			valuesJ = append(valuesJ, value)
		}

		if keysI[0] == keysJ[0] {
			valuePriorityI := valuePriority[valuesI[0]]
			valuePriorityJ := valuePriority[valuesJ[0]]
			return valuePriorityI < valuePriorityJ
		} else {
			return keysI[0] < keysJ[0]
		}
	})
}
