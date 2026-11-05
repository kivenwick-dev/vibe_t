package report

import (
	"banner-fingerprint/internal/model"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

type Options struct {
	MinimumConfidence float64
	Protocol          string
	IncludeUnknown    bool
}

func Select(results []model.FingerprintResult, options Options) ([]model.FingerprintResult, error) {
	if options.MinimumConfidence < 0 || options.MinimumConfidence > 1 {
		return nil, fmt.Errorf("minimum confidence must be between 0 and 1")
	}
	selected := make([]model.FingerprintResult, 0, len(results))
	for _, result := range results {
		if result.Confidence < options.MinimumConfidence {
			continue
		}
		if !options.IncludeUnknown && result.Protocol == "unknown" {
			continue
		}
		if options.Protocol != "" && !strings.EqualFold(options.Protocol, result.Protocol) {
			continue
		}
		selected = append(selected, result)
	}
	sort.SliceStable(selected, func(i, j int) bool {
		if selected[i].IP == selected[j].IP {
			return selected[i].Port < selected[j].Port
		}
		return selected[i].IP < selected[j].IP
	})
	return selected, nil
}

func WriteCSV(w io.Writer, results []model.FingerprintResult) error {
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"ip", "port", "protocol", "product", "version", "os_hint", "confidence"}); err != nil {
		return err
	}
	for _, r := range results {
		if err := writer.Write([]string{r.IP, strconv.Itoa(r.Port), r.Protocol, r.Product, r.Version, r.OSHint, strconv.FormatFloat(r.Confidence, 'f', -1, 64)}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func WriteMarkdown(w io.Writer, results []model.FingerprintResult) error {
	// Buffer the full document so a write failure can be returned to the caller.
	var b strings.Builder
	fmt.Fprintf(&b, "# Banner inventory\n\nServices: %d\n\n", len(results))
	counts := make(map[string]int)
	for _, r := range results {
		counts[r.Protocol]++
	}
	protocols := make([]string, 0, len(counts))
	for protocol := range counts {
		protocols = append(protocols, protocol)
	}
	sort.Strings(protocols)
	for _, protocol := range protocols {
		fmt.Fprintf(&b, "- %s: %d\n", escape(protocol), counts[protocol])
	}
	b.WriteString("\n| IP | Port | Protocol | Product | Version | OS | Confidence |\n| --- | ---: | --- | --- | --- | --- | ---: |\n")
	for _, r := range results {
		fmt.Fprintf(&b, "| %s | %d | %s | %s | %s | %s | %.2f |\n", escape(r.IP), r.Port, escape(r.Protocol), escape(r.Product), escape(r.Version), escape(r.OSHint), r.Confidence)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func escape(value string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "|", "&#124;", "\r", " ", "\n", " ").Replace(value)
}
