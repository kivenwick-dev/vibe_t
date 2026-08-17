package fingerprint

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"banner-fingerprint/internal/model"
)

type ruleFile struct {
	PortHints map[string]PortHint `json:"port_hints"`
	Rules     []Rule              `json:"rules"`
}

type PortHint struct {
	Protocol   string  `json:"protocol"`
	Product    string  `json:"product"`
	Confidence float64 `json:"confidence"`
}

type Rule struct {
	Name           string  `json:"name"`
	Protocol       string  `json:"protocol"`
	Product        string  `json:"product"`
	Pattern        string  `json:"pattern"`
	VersionPattern string  `json:"version_pattern"`
	VersionGroup   int     `json:"version_group"`
	OSPattern      string  `json:"os_pattern"`
	OSGroup        int     `json:"os_group"`
	Confidence     float64 `json:"confidence"`
	Ports          []int   `json:"ports"`

	match        *regexp.Regexp
	versionMatch *regexp.Regexp
	osMatch      *regexp.Regexp
}

type Engine struct {
	hints map[int]PortHint
	rules []Rule
}

func Load(path string) (*Engine, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rules: %w", err)
	}

	var config ruleFile
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("decode rules: %w", err)
	}

	hints := make(map[int]PortHint, len(config.PortHints))
	for port, hint := range config.PortHints {
		var portNumber int
		if _, err := fmt.Sscanf(port, "%d", &portNumber); err != nil || portNumber < 1 || portNumber > 65535 {
			return nil, fmt.Errorf("invalid port hint %q", port)
		}
		hints[portNumber] = hint
	}

	for i := range config.Rules {
		rule := &config.Rules[i]
		if rule.Pattern == "" || rule.Protocol == "" {
			return nil, fmt.Errorf("rule %q must define pattern and protocol", rule.Name)
		}
		if rule.Confidence < 0 || rule.Confidence > 1 {
			return nil, fmt.Errorf("rule %q confidence must be between 0 and 1", rule.Name)
		}
		for _, port := range rule.Ports {
			if port < 1 || port > 65535 {
				return nil, fmt.Errorf("rule %q has invalid port %d", rule.Name, port)
			}
		}
		if rule.match, err = regexp.Compile(rule.Pattern); err != nil {
			return nil, fmt.Errorf("compile rule %q: %w", rule.Name, err)
		}
		if rule.VersionPattern != "" {
			if rule.versionMatch, err = regexp.Compile(rule.VersionPattern); err != nil {
				return nil, fmt.Errorf("compile version rule %q: %w", rule.Name, err)
			}
		}
		if rule.OSPattern != "" {
			if rule.osMatch, err = regexp.Compile(rule.OSPattern); err != nil {
				return nil, fmt.Errorf("compile os rule %q: %w", rule.Name, err)
			}
		}
	}
	sort.SliceStable(config.Rules, func(i, j int) bool {
		return config.Rules[i].Confidence > config.Rules[j].Confidence
	})

	return &Engine{hints: hints, rules: config.Rules}, nil
}

func (e *Engine) Identify(input model.BannerInput) model.FingerprintResult {
	result := model.FingerprintResult{
		IP:         input.IP,
		Port:       input.Port,
		Protocol:   "unknown",
		Confidence: 0,
	}
	banner := strings.TrimSpace(input.Banner)
	for _, rule := range e.rules {
		if len(rule.Ports) > 0 && !containsPort(rule.Ports, input.Port) {
			continue
		}
		if !rule.match.MatchString(banner) {
			continue
		}
		result.Protocol = rule.Protocol
		result.Product = rule.Product
		result.Confidence = clamp(rule.Confidence)
		if rule.versionMatch != nil {
			result.Version = capture(rule.versionMatch, banner, rule.VersionGroup)
		}
		if rule.osMatch != nil {
			result.OSHint = capture(rule.osMatch, banner, rule.OSGroup)
		}
		return result
	}

	if hint, ok := e.hints[input.Port]; ok {
		result.Protocol = hint.Protocol
		result.Product = hint.Product
		result.Confidence = clamp(hint.Confidence)
	}
	return result
}

func containsPort(ports []int, wanted int) bool {
	for _, port := range ports {
		if port == wanted {
			return true
		}
	}
	return false
}

func capture(pattern *regexp.Regexp, value string, group int) string {
	matches := pattern.FindStringSubmatch(value)
	if group > 0 && group < len(matches) {
		return matches[group]
	}
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

func clamp(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
