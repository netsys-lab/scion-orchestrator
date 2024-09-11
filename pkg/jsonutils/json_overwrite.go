package jsonutils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

/**
 * Update a JSON file with new values.
 * @param jsonPath The path to the JSON file to update.
 * @param updates A list of updates to apply in the format "field.path=value".
 * @param outputPath The path to write the updated JSON to.
 * @return An error if the operation failed.
 * Mainly used to serve a different topology file to the endhosts in case of e.g. NAT inside of the AS.
 */
func OverwriteJSON(jsonPath string, updates []string, outputPath string) error {
	// Read JSON from file
	jsonData, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the JSON into a generic map
	var data map[string]interface{}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Apply updates
	for _, update := range updates {
		parts := strings.Split(update, "=")
		if len(parts) != 2 {
			return fmt.Errorf("invalid format: %s", update)
		}

		fieldPath := parts[0]
		value := parts[1]

		keys := strings.Split(fieldPath, ".")
		updateNestedMap(data, keys, value)
	}

	// Marshal the updated JSON back to a byte array
	updatedJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated JSON: %w", err)
	}

	// Write the updated JSON to a new file
	err = ioutil.WriteFile(outputPath, updatedJSON, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func updateNestedMap(data map[string]interface{}, keys []string, value string) {
	for i := 0; i < len(keys)-1; i++ {
		key := keys[i]

		if _, ok := data[key].(map[string]interface{}); !ok {
			data[key] = make(map[string]interface{})
		}

		data = data[key].(map[string]interface{})
	}

	data[keys[len(keys)-1]] = value
}
