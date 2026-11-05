package catalog

import (
	"banner-fingerprint/internal/fingerprint"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type Entry struct {
	Name       string  `json:"name"`
	Protocol   string  `json:"protocol"`
	Product    string  `json:"product"`
	Confidence float64 `json:"confidence"`
}

type Catalog struct {
	TotalRules     int      `json:"total_rules"`
	TotalPortHints int      `json:"total_port_hints"`
	Protocols      []string `json:"protocols"`
	Entries        []Entry  `json:"entries"`
}

// Inspect validates compilable expressions and unique rule identities before listing the catalog.
func Inspect(path string) (Catalog, error) {
	result := Catalog{Protocols: make([]string, 0), Entries: make([]Entry, 0)}
	if _, err := fingerprint.Load(path); err != nil {
		return result, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return result, err
	}
	var file struct {
		Rules     []Entry                    `json:"rules"`
		PortHints map[string]json.RawMessage `json:"port_hints"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return result, err
	}
	names := make(map[string]bool)
	protocols := make(map[string]bool)
	for _, entry := range file.Rules {
		if entry.Name == "" {
			return result, fmt.Errorf("rule name cannot be empty")
		}
		if names[entry.Name] {
			return result, fmt.Errorf("duplicate rule name %q", entry.Name)
		}
		names[entry.Name] = true
		protocols[entry.Protocol] = true
		result.Entries = append(result.Entries, entry)
	}
	for protocol := range protocols {
		result.Protocols = append(result.Protocols, protocol)
	}
	sort.Strings(result.Protocols)
	sort.Slice(result.Entries, func(i, j int) bool { return result.Entries[i].Name < result.Entries[j].Name })
	result.TotalRules = len(result.Entries)
	result.TotalPortHints = len(file.PortHints)
	return result, nil
}
