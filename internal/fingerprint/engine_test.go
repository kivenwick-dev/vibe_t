package fingerprint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"banner-fingerprint/internal/model"
)

func TestIdentifyExamples(t *testing.T) {
	engine, err := Load(filepath.Join("..", "..", "rules", "rules.json"))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name     string
		input    model.BannerInput
		protocol string
		product  string
		version  string
		osHint   string
	}{
		{"ssh", model.BannerInput{Port: 22, Banner: "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.6"}, "SSH", "OpenSSH", "8.9p1", "Ubuntu"},
		{"nginx", model.BannerInput{Port: 80, Banner: "HTTP/1.1 200 OK\r\nServer: nginx/1.24.0"}, "HTTP", "nginx", "1.24.0", ""},
		{"apache", model.BannerInput{Port: 443, Banner: "HTTP/2 200\r\nserver: Apache/2.4.57"}, "HTTP", "Apache", "2.4.57", ""},
		{"jetty", model.BannerInput{Port: 8080, Banner: "HTTP/1.1 200 OK\r\nServer: Jetty(9.4.51.v20230217)"}, "HTTP", "Jetty", "9.4.51", ""},
		{"mysql", model.BannerInput{Port: 3306, Banner: "8.0.32 MySQL Community Server"}, "MySQL", "MySQL", "8.0.32", ""},
		{"redis", model.BannerInput{Port: 6379, Banner: "+PONG"}, "Redis", "Redis", "", ""},
		{"ftp", model.BannerInput{Port: 21, Banner: "220 ProFTPD 1.3.7 Server"}, "FTP", "ProFTPD", "1.3.7", ""},
		{"unknown", model.BannerInput{Port: 1234, Banner: "opaque service"}, "unknown", "", "", ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := engine.Identify(test.input)
			if got.Protocol != test.protocol || got.Product != test.product || got.Version != test.version || got.OSHint != test.osHint {
				t.Fatalf("got %#v", got)
			}
		})
	}
}

func TestLoadMissingRules(t *testing.T) {
	_, err := Load(os.DevNull + "-missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSelfTestBatch(t *testing.T) {
	engine, err := Load(filepath.Join("..", "..", "rules", "rules.json"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "self-test.json"))
	if err != nil {
		t.Fatal(err)
	}
	var inputs []model.BannerInput
	if err := json.Unmarshal(data, &inputs); err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 20 {
		t.Fatalf("input count = %d", len(inputs))
	}
	expected := map[int]string{
		22: "SSH", 80: "HTTP", 443: "HTTP", 3306: "MySQL",
		6379: "Redis", 21: "FTP", 8080: "HTTP", 8443: "HTTP", 8888: "HTTP",
	}
	for _, input := range inputs {
		result := engine.Identify(input)
		if input.Port == 9999 || input.Port == 12345 {
			if result.Protocol != "unknown" {
				t.Fatalf("port %d: expected unknown, got %#v", input.Port, result)
			}
			continue
		}
		if result.Protocol != expected[input.Port] {
			t.Fatalf("port %d: expected %q, got %#v", input.Port, expected[input.Port], result)
		}
		if result.Confidence < 0 || result.Confidence > 1 {
			t.Fatalf("port %d: invalid confidence %v", input.Port, result.Confidence)
		}
		t.Logf("%s:%d => protocol=%s product=%s version=%s os_hint=%s confidence=%.2f",
			input.IP, input.Port, result.Protocol, result.Product, result.Version, result.OSHint, result.Confidence)
	}
}
