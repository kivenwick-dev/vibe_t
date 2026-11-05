package api

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"

	"banner-fingerprint/internal/model"
)

type protocolSummary struct {
	Protocol          string  `json:"protocol"`
	Count             int     `json:"count"`
	AverageConfidence float64 `json:"average_confidence"`
}

type batchSummary struct {
	Total       int                       `json:"total"`
	Unknown     int                       `json:"unknown"`
	UniqueHosts int                       `json:"unique_hosts"`
	Protocols   []protocolSummary         `json:"protocols"`
	Results     []model.FingerprintResult `json:"results"`
}

func (s *Server) summary(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	decoder := json.NewDecoder(r.Body)
	var inputs []model.BannerInput
	if err := decoder.Decode(&inputs); err != nil || inputs == nil {
		writeError(w, http.StatusBadRequest, "request body must be a JSON array")
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(w, http.StatusBadRequest, "request body must contain exactly one JSON array")
		return
	}
	if len(inputs) > 10000 {
		writeError(w, http.StatusRequestEntityTooLarge, "maximum batch size is 10000")
		return
	}
	result := batchSummary{Total: len(inputs), Protocols: make([]protocolSummary, 0), Results: make([]model.FingerprintResult, 0, len(inputs))}
	hosts := make(map[string]bool)
	counts := make(map[string]int)
	confidence := make(map[string]float64)
	for _, input := range inputs {
		if input.Port < 1 || input.Port > 65535 || input.IP == "" {
			writeError(w, http.StatusBadRequest, "each input requires an IP and port between 1 and 65535")
			return
		}
		item := s.engine.Identify(input)
		result.Results = append(result.Results, item)
		hosts[input.IP] = true
		counts[item.Protocol]++
		confidence[item.Protocol] += item.Confidence
		if item.Protocol == "unknown" {
			result.Unknown++
		}
	}
	result.UniqueHosts = len(hosts)
	for protocol, count := range counts {
		result.Protocols = append(result.Protocols, protocolSummary{protocol, count, confidence[protocol] / float64(count)})
	}
	sort.Slice(result.Protocols, func(i, j int) bool { return result.Protocols[i].Protocol < result.Protocols[j].Protocol })
	writeJSON(w, http.StatusOK, result)
}
