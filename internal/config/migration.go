package config

// CurrentVersion is the active schema version of the configuration
const CurrentVersion = "1"

// migrate0to1 handles migration from version 0 (or unspecified version) to version 1.
// It fills in default values for fields/sections introduced in version 1 if they are missing.
func migrate0to1(data map[string]interface{}) map[string]interface{} {
	if _, ok := data["tui"]; !ok {
		data["tui"] = map[string]interface{}{
			"theme":              "auto",
			"streaming_delay_ms": 50,
			"keybindings": map[string]interface{}{
				"navigate_down": []interface{}{"j", "down"},
				"navigate_up":   []interface{}{"k", "up"},
				"select":        []interface{}{"enter"},
				"back":          []interface{}{"esc", "q"},
				"copy":          []interface{}{"c", "ctrl+c"},
				"quit":          []interface{}{"ctrl+q"},
			},
		}
	}
	if _, ok := data["llm"]; !ok {
		data["llm"] = map[string]interface{}{
			"max_concurrent": 1,
			"retry_attempts": 3,
			"retry_backoff":  []interface{}{1000, 2000, 4000},
		}
	}
	return data
}

// migrate takes a raw maprepresentation of JSON config, detects its version,
// and applies sequential migrations up to CurrentVersion.
func migrate(data map[string]interface{}) map[string]interface{} {
	versionRaw, ok := data["version"]
	var version string
	if !ok {
		version = "0"
	} else {
		if v, ok := versionRaw.(string); ok {
			version = v
		} else {
			version = "0"
		}
	}

	// Order-based migration dispatcher.
	if version == "0" {
		data = migrate0to1(data)
		version = "1"
	}

	// Scaffolding for future migrations:
	// if version == "1" {
	//     data = migrate1to2(data)
	//     version = "2"
	// }

	data["version"] = CurrentVersion
	return data
}
