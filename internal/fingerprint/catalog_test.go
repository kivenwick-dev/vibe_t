package fingerprint

import (
	"banner-fingerprint/internal/model"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExtendedCatalogFixtures(t *testing.T) {
	root := filepath.Join("..", "..")
	engine, err := Load(filepath.Join(root, "rules", "rules.json"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "testdata", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	var inputs []model.BannerInput
	if err := json.Unmarshal(data, &inputs); err != nil {
		t.Fatal(err)
	}
	products := []string{"Postfix", "Dovecot", "Dovecot", "Memcached", "Mosquitto", "RabbitMQ", "Caddy", "lighttpd"}
	if len(inputs) != len(products) {
		t.Fatal("fixture count mismatch")
	}
	for i, input := range inputs {
		got := engine.Identify(input)
		if got.Product != products[i] || got.Confidence != 0.92 {
			t.Fatalf("fixture %d: %+v", i, got)
		}
	}
}
