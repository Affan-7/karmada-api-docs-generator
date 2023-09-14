package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"
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

	// Load the json into swaggerJson struct
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

				_, err = file.WriteString("#### " + strings.ToUpper(method) + " " + path + "\n\n")

				if err != nil {
					log.Fatal(err)
				}

				pathData := paths[path]
				pathDataMap := pathData.(map[string]interface{})
				methodData := pathDataMap[method]
				methodDataMap := methodData.(map[string]interface{})
				description := methodDataMap["description"]
				descriptionStr, ok := description.(string)
				if !ok {
					log.Fatal(`Can't do type assertion for methodDataMap["description"]`)
				}

				capitalizeFirstLetter(&descriptionStr)

				_, err = file.WriteString(descriptionStr + "\n\n")

				if err != nil {
					log.Fatal(err)
				}

				if pathDataMap["parameters"] != nil {
					_, err = file.WriteString("| Parameter | Description |\n")

					if err != nil {
						log.Fatal(err)
					}

					_, err = file.WriteString("|---|---|\n")

					if err != nil {
						log.Fatal(err)
					}

					writeTableToFile(file, pathDataMap)
				} else if methodDataMap["parameters"] != nil {
					_, err = file.WriteString("| Parameter | Description |\n")

					if err != nil {
						log.Fatal(err)
					}

					_, err = file.WriteString("|:---|:---|\n")

					if err != nil {
						log.Fatal(err)
					}

					writeTableToFile(file, methodDataMap)
				}
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

func capitalizeFirstLetter(s *string) {
	if s == nil || len(*s) == 0 {
		return
	}

	runeSlice := []rune(*s)
	runeSlice[0] = unicode.ToUpper(runeSlice[0])

	*s = string(runeSlice)
}

func replaceNewlinesWithBR(s *string) {
	// Replace all occurrences of "\n" with "<br></br>"
	*s = strings.ReplaceAll(*s, "\n", "<br></br>")
}

func writeTableToFile(file *os.File, data map[string]interface{}) {
	var err error

	if data["parameters"] != nil {

		parametersData := data["parameters"]
		parametersSlice := parametersData.([]interface{})
		for _, parameter := range parametersSlice {
			parameterMap := parameter.(map[string]interface{})

			parameterName, ok := parameterMap["name"].(string)
			if !ok {
				log.Fatal(`Can't do type assertion for parameterMap["name"]`)
			}

			parameterType, ok := parameterMap["type"].(string)
			if !ok {
				log.Fatal(`Can't do type assertion for parameterType`)
			}

			parameterIn, ok := parameterMap["in"].(string)
			if !ok {
				log.Fatal(`Can't do type assertion for parameterType`)
			}

			if parameterName != "body" {
				if parameterMap["description"] != nil {
					parameterDescription, ok := parameterMap["description"].(string)
					if !ok {
						log.Fatal(`Can't do type assertion for parameterMap["description"]`)
					}

					replaceNewlinesWithBR(&parameterDescription)

					_, err = file.WriteString("|**" + parameterName + "**<br></br>" + parameterType + "<br></br>*(" + parameterIn + ")*|" + parameterDescription + "|\n")

					if err != nil {
						log.Fatal(err)
					}
				} else {
					parameterDescription := ""
					_, err = file.WriteString("|**" + parameterName + "**<br></br>" + parameterType + "<br></br>*(" + parameterIn + ")*|" + parameterDescription + "|\n")

					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
	}
}
