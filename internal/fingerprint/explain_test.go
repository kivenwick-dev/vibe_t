package fingerprint

import (
	"banner-fingerprint/internal/model"
	"path/filepath"
	"testing"
)

func TestExplainSourcesAndRanking(t *testing.T) {
	engine, err := Load(filepath.Join("..", "..", "rules", "rules.json"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		input        model.BannerInput
		source, rule string
		candidates   int
	}{
		{model.BannerInput{Port: 80, Banner: "HTTP/1.1 200 OK\r\nServer: nginx/1.24.0"}, "rule", "nginx", 2},
		{model.BannerInput{Port: 22, Banner: "opaque"}, "port_hint", "", 0},
		{model.BannerInput{Port: 9999, Banner: "opaque"}, "unknown", "", 0},
	}
	for _, tc := range cases {
		got := engine.Explain(tc.input)
		if got.Source != tc.source || got.SelectedRule != tc.rule || len(got.Candidates) != tc.candidates {
			t.Fatalf("%+v", got)
		}
		if got.Result != engine.Identify(tc.input) {
			t.Fatal("Identify and Explain differ")
		}
		for i := 1; i < len(got.Candidates); i++ {
			if got.Candidates[i].Result.Confidence > got.Candidates[i-1].Result.Confidence {
				t.Fatal("candidates are not ranked")
			}
		}
	}
}

func TestExplainExtractsFields(t *testing.T) {
	engine, err := Load(filepath.Join("..", "..", "rules", "rules.json"))
	if err != nil {
		t.Fatal(err)
	}
	got := engine.Explain(model.BannerInput{IP: "192.0.2.1", Port: 22, Banner: "  SSH-2.0-OpenSSH_9.6 Ubuntu  "})
	if got.Result.Version != "9.6" || got.Result.OSHint != "Ubuntu" || got.Result.IP != "192.0.2.1" {
		t.Fatalf("%+v", got)
	}
}
