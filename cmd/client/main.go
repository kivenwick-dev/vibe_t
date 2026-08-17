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
	"time"

	"banner-fingerprint/internal/model"
)

func main() {
	server := flag.String("server", "http://localhost:8080", "fingerprint server URL")
	inputPath := flag.String("input", "-", "input JSON file, or - for stdin")
	pretty := flag.Bool("pretty", true, "pretty-print response")
	flag.Parse()

	data, err := readInput(*inputPath)
	if err != nil {
		fatal(err)
	}
	var inputs []model.BannerInput
	if err := json.Unmarshal(data, &inputs); err != nil {
		fatal(fmt.Errorf("decode input JSON: %w", err))
	}
	payload, _ := json.Marshal(inputs)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, *server+"/fingerprint", bytes.NewReader(payload))
	if err != nil {
		fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		fatal(fmt.Errorf("request server: %w", err))
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		fatal(fmt.Errorf("server returned %s: %s", response.Status, bytes.TrimSpace(body)))
	}
	if *pretty {
		var output bytes.Buffer
		if err := json.Indent(&output, body, "", "  "); err == nil {
			body = output.Bytes()
		}
	}
	fmt.Println(string(bytes.TrimSpace(body)))
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
