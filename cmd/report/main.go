package main

import (
	"banner-fingerprint/internal/fingerprint"
	"banner-fingerprint/internal/model"
	"banner-fingerprint/internal/report"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	input := flag.String("input", "testdata/sample.json", "input JSON array")
	rules := flag.String("rules", "rules/rules.json", "rule file")
	format := flag.String("format", "markdown", "markdown or csv")
	minimum := flag.Float64("min-confidence", 0, "minimum confidence, 0 through 1")
	protocol := flag.String("protocol", "", "optional protocol filter")
	unknown := flag.Bool("include-unknown", true, "include unrecognized services")
	flag.Parse()
	err := run(*input, *rules, *format, report.Options{MinimumConfidence: *minimum, Protocol: *protocol, IncludeUnknown: *unknown})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(input, rules, format string, options report.Options) error {
	if format != "markdown" && format != "csv" {
		return fmt.Errorf("unsupported format %q", format)
	}
	if _, err := report.Select(nil, options); err != nil {
		return err
	}
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
	results := make([]model.FingerprintResult, 0, len(inputs))
	for _, item := range inputs {
		results = append(results, engine.Identify(item))
	}
	selected, err := report.Select(results, options)
	if err != nil {
		return err
	}
	if format == "csv" {
		return report.WriteCSV(os.Stdout, selected)
	}
	return report.WriteMarkdown(os.Stdout, selected)
}
