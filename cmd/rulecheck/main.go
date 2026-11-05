package main

import (
	"banner-fingerprint/internal/catalog"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	path := flag.String("rules", "rules/rules.json", "rule file to validate and catalog")
	flag.Parse()
	result, err := catalog.Inspect(*path)
	if err == nil {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(result)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
