package codec

import (
	"banner-fingerprint/internal/model"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
)

func TestDecodeFormats(t *testing.T) {
	for format, input := range map[string]string{
		"json":  `[{"ip":"192.0.2.1","port":22,"banner":"SSH"}]`,
		"jsonl": "{\"ip\":\"192.0.2.1\",\"port\":22,\"banner\":\"SSH\"}\n",
		"csv":   "ip,port,banner\n192.0.2.1,22,SSH\n",
	} {
		got, err := Decode(strings.NewReader(input), format)
		if err != nil || len(got) != 1 || got[0].Port != 22 || got[0].Banner != "SSH" {
			t.Fatalf("%s: %+v %v", format, got, err)
		}
	}
}

func TestInvalidInputs(t *testing.T) {
	for _, tc := range []struct{ format, input string }{
		{"csv", "ip,port,banner\nx,no,b\n"}, {"csv", "ip,port,banner\nx,65536,b\n"},
		{"csv", "wrong,header\n"}, {"json", "[] []"}, {"jsonl", "{"}, {"xml", ""},
	} {
		if _, err := Decode(strings.NewReader(tc.input), tc.format); err == nil {
			t.Errorf("accepted %+v", tc)
		}
	}
}

func TestExportsPreserveEscaping(t *testing.T) {
	result := model.FingerprintResult{IP: "192.0.2.1", Port: 80, Protocol: "HTTP", Product: "quoted,\"product\"\nline", Confidence: 0.9}
	for _, format := range []string{"json", "jsonl", "csv"} {
		var b bytes.Buffer
		if err := Encode(&b, []model.FingerprintResult{result}, format, true); err != nil {
			t.Fatal(err)
		}
		if format == "csv" {
			rows, err := csv.NewReader(&b).ReadAll()
			if err != nil || len(rows) != 2 || rows[1][3] != result.Product {
				t.Fatalf("%v %v", rows, err)
			}
		} else if format == "json" {
			var rows []model.FingerprintResult
			if err := json.Unmarshal(b.Bytes(), &rows); err != nil || len(rows) != 1 || rows[0] != result {
				t.Fatalf("%v %v", rows, err)
			}
		} else {
			var row model.FingerprintResult
			if err := json.Unmarshal(b.Bytes(), &row); err != nil || row != result {
				t.Fatalf("%v %v", row, err)
			}
		}
	}
}
