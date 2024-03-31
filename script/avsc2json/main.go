package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hamba/avro/v2"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: avsc_to_json <input.avsc>")
		return
	}

	avscFilename := os.Args[1]

	avscBytes, err := os.ReadFile(avscFilename)
	if err != nil {
		fmt.Println("Error reading AVSC file:", err)
		return
	}

	// Load the schema from a file or some other source
	schema, err := avro.Parse(string(avscBytes))
	if err != nil {
		fmt.Println("Error parsing schema:", err)
		return
	}
	output := map[string]interface{}{
		"schema": schema.String(),
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}
	fmt.Println(string(jsonBytes))

	//		jsonFilename := strings.TrimSuffix(avscFilename, ".avsc") + ".json"
	//		err = os.WriteFile(jsonFilename, jsonBytes, 0644) //nolint:gosec
	//		if err != nil {
	//			fmt.Println("Error writing JSON file:", err)
	//			return
	//		}
}
