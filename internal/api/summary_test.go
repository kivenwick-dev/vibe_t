package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSummaryAggregation(t *testing.T) {
	body := `[{"ip":"192.0.2.1","port":22,"banner":"SSH-2.0-OpenSSH_9.0"},{"ip":"192.0.2.1","port":9999,"banner":"opaque"},{"ip":"192.0.2.2","port":80,"banner":"HTTP/1.1 200 OK"}]`
	response := httptest.NewRecorder()
	testHandler(t).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/fingerprint/summary", strings.NewReader(body)))
	if response.Code != 200 {
		t.Fatalf("%d: %s", response.Code, response.Body)
	}
	var summary batchSummary
	if err := json.Unmarshal(response.Body.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if summary.Total != 3 || summary.UniqueHosts != 2 || summary.Unknown != 1 || len(summary.Results) != 3 {
		t.Fatalf("%+v", summary)
	}
	if len(summary.Protocols) != 3 || summary.Protocols[0].Protocol != "HTTP" || summary.Protocols[1].AverageConfidence != 0.95 {
		t.Fatalf("%+v", summary.Protocols)
	}
}

func TestSummaryValidation(t *testing.T) {
	cases := []struct {
		body   string
		status int
	}{
		{`[]`, 200}, {`null`, 400}, {`{}​​`, 400}, {`[] []`, 400}, {`[] trailing`, 400},
		{`[{"ip":"192.0.2.1","port":0}]`, 400}, {`[{"port":22}]`, 400},
		{`[{"ip":"192.0.2.1","port":65536}]`, 400},
	}
	for _, tc := range cases {
		r := httptest.NewRecorder()
		testHandler(t).ServeHTTP(r, httptest.NewRequest(http.MethodPost, "/fingerprint/summary", strings.NewReader(tc.body)))
		if r.Code != tc.status {
			t.Errorf("%q: got %d want %d", tc.body, r.Code, tc.status)
		}
	}
}
