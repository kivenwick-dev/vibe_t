package fingerprint

import (
	"banner-fingerprint/internal/model"
	"strings"
)

// Candidate records each matching rule in descending confidence order.
type Candidate struct {
	Rule   string                  `json:"rule"`
	Result model.FingerprintResult `json:"result"`
}

type Explanation struct {
	Source       string                  `json:"source"`
	SelectedRule string                  `json:"selected_rule,omitempty"`
	Result       model.FingerprintResult `json:"result"`
	Candidates   []Candidate             `json:"candidates"`
}

func (e *Engine) Explain(input model.BannerInput) Explanation {
	unknown := model.FingerprintResult{IP: input.IP, Port: input.Port, Protocol: "unknown"}
	explanation := Explanation{Source: "unknown", Result: unknown, Candidates: make([]Candidate, 0)}
	banner := strings.TrimSpace(input.Banner)
	for _, rule := range e.rules {
		if len(rule.Ports) > 0 && !containsPort(rule.Ports, input.Port) {
			continue
		}
		if !rule.match.MatchString(banner) {
			continue
		}
		result := unknown
		result.Protocol = rule.Protocol
		result.Product = rule.Product
		result.Confidence = clamp(rule.Confidence)
		if rule.versionMatch != nil {
			result.Version = capture(rule.versionMatch, banner, rule.VersionGroup)
		}
		if rule.osMatch != nil {
			result.OSHint = capture(rule.osMatch, banner, rule.OSGroup)
		}
		explanation.Candidates = append(explanation.Candidates, Candidate{Rule: rule.Name, Result: result})
	}
	if len(explanation.Candidates) > 0 {
		explanation.Source = "rule"
		explanation.SelectedRule = explanation.Candidates[0].Rule
		explanation.Result = explanation.Candidates[0].Result
	} else if hint, ok := e.hints[input.Port]; ok {
		explanation.Source = "port_hint"
		explanation.Result.Protocol = hint.Protocol
		explanation.Result.Product = hint.Product
		explanation.Result.Confidence = clamp(hint.Confidence)
	}
	return explanation
}
