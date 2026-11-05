package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"banner-fingerprint/internal/codec"
	"banner-fingerprint/internal/model"
)

func main() {
	server := flag.String("server", "http://localhost:8080", "fingerprint server URL")
	inputPath := flag.String("input", "-", "input JSON file, or - for stdin")
	pretty := flag.Bool("pretty", true, "pretty-print response")
	inputFormat := flag.String("input-format", "json", "input format: json, jsonl, csv")
	outputFormat := flag.String("output-format", "json", "output format: json, jsonl, csv")
	flag.Parse()
	switch *outputFormat {
	case "json", "jsonl", "csv":
	default:
		fatal(fmt.Errorf("unsupported output format %q", *outputFormat))
	}

	data, err := readInput(*inputPath)
	if err != nil {
		fatal(err)
	}
	inputs, err := codec.Decode(bytes.NewReader(data), *inputFormat)
	if err != nil {
		fatal(fmt.Errorf("decode input: %w", err))
	}
	payload, _ := json.Marshal(inputs)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	endpoint := strings.TrimRight(*server, "/") + "/fingerprint"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		fatal(fmt.Errorf("request server: %w", err))
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fatal(fmt.Errorf("read response: %w", err))
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		fatal(fmt.Errorf("server returned %s: %s", response.Status, bytes.TrimSpace(body)))
	}
	var results []model.FingerprintResult
	if err := json.Unmarshal(body, &results); err != nil {
		fatal(fmt.Errorf("decode response: %w", err))
	}
	if err := codec.Encode(os.Stdout, results, *outputFormat, *pretty); err != nil {
		fatal(err)
	}
}

func readInput(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
