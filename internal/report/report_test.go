package report

import (
	"banner-fingerprint/internal/model"
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
)

func TestSelectFiltersAndSortsWithoutMutatingInput(t *testing.T) {
	input := []model.FingerprintResult{
		{IP: "b", Port: 80, Protocol: "HTTP", Confidence: 0.9},
		{IP: "a", Port: 443, Protocol: "HTTP", Confidence: 0.8},
		{IP: "a", Port: 80, Protocol: "HTTP", Confidence: 0.7},
		{IP: "c", Port: 9999, Protocol: "unknown"},
	}
	got, err := Select(input, Options{MinimumConfidence: 0.7, Protocol: "http"})
	if err != nil || len(got) != 3 || got[0].Port != 80 || got[1].Port != 443 || input[0].IP != "b" {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err := Select(input, Options{MinimumConfidence: 1.1}); err == nil {
		t.Fatal("accepted invalid confidence")
	}
	all, err := Select(input, Options{IncludeUnknown: true})
	if err != nil || len(all) != 4 {
		t.Fatalf("%+v %v", all, err)
	}
}

func TestReportEscapesAndCSVValues(t *testing.T) {
	input := []model.FingerprintResult{{IP: "192.0.2.1", Port: 80, Protocol: "HTTP", Product: "<test>|quoted,\nvalue", Confidence: 0.9}}
	var markdown bytes.Buffer
	if err := WriteMarkdown(&markdown, input); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(markdown.String(), "&lt;test&gt;&#124;quoted, value") || !strings.Contains(markdown.String(), "Services: 1") {
		t.Fatal(markdown.String())
	}
	var data bytes.Buffer
	if err := WriteCSV(&data, input); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&data).ReadAll()
	if err != nil || len(rows) != 2 || rows[1][3] != input[0].Product {
		t.Fatalf("%v %v", rows, err)
	}
}
