package main

import (
	"banner-fingerprint/internal/fingerprint"
	"banner-fingerprint/internal/model"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	rules := flag.String("rules", "rules/rules.json", "rule file")
	input := flag.String("input", "testdata/sample.json", "input JSON array")
	flag.Parse()
	if err := run(*rules, *input); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(rules, input string) error {
	engine, err := fingerprint.Load(rules)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	var inputs []model.BannerInput
	if err := json.Unmarshal(data, &inputs); err != nil {
		return err
	}
	explanations := make([]fingerprint.Explanation, 0, len(inputs))
	for _, item := range inputs {
		explanations = append(explanations, engine.Explain(item))
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(explanations)
}
