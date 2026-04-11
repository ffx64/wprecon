package engine

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ffx64/wprecon/internal/domain"
	"github.com/ffx64/wprecon/internal/executor"
)

type DiscoveryResult struct {
	Plugins []string
	Themes  []string
}

var (
	pluginRegex = regexp.MustCompile(`/wp-content/plugins/([a-zA-Z0-9_-]+)/`)
	themeRegex  = regexp.MustCompile(`/wp-content/themes/([a-zA-Z0-9_-]+)/`)
)

func DiscoverFromHTML(target string, exec *executor.HTTPExecutor, logger *Logger) (*DiscoveryResult, error) {
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}

	logger.Info("Starting dynamic plugin/theme discovery from HTML...")

	req := &domain.HTTPRequest{
		URL:    target,
		Method: "GET",
	}

	logger.RequestStart("GET", target)
	start := time.Now()
	resp, err := exec.Execute(req)
	if err != nil {
		logger.RequestError("GET", target, err)
		return nil, fmt.Errorf("failed to fetch homepage: %w", err)
	}
	logger.RequestDone("GET", target, resp.StatusCode, time.Since(start).Seconds()*1000, len(resp.Body))

	body := string(resp.Body)

	pluginSet := make(map[string]bool)
	for _, match := range pluginRegex.FindAllStringSubmatch(body, -1) {
		if len(match) > 1 {
			slug := strings.ToLower(match[1])
			pluginSet[slug] = true
		}
	}

	themeSet := make(map[string]bool)
	for _, match := range themeRegex.FindAllStringSubmatch(body, -1) {
		if len(match) > 1 {
			slug := strings.ToLower(match[1])
			themeSet[slug] = true
		}
	}

	result := &DiscoveryResult{}
	for slug := range pluginSet {
		result.Plugins = append(result.Plugins, slug)
	}
	sort.Strings(result.Plugins)

	for slug := range themeSet {
		result.Themes = append(result.Themes, slug)
	}
	sort.Strings(result.Themes)

	logger.Info(fmt.Sprintf("Discovery: found %d plugins, %d themes", len(result.Plugins), len(result.Themes)))
	for _, p := range result.Plugins {
		logger.Info(fmt.Sprintf("  Plugin: %s", p))
	}
	for _, t := range result.Themes {
		logger.Info(fmt.Sprintf("  Theme: %s", t))
	}

	return result, nil
}

func DiscoveryToFindings(target string, discovery *DiscoveryResult) []domain.Finding {
	var findings []domain.Finding
	target = strings.TrimRight(target, "/")

	for _, plugin := range discovery.Plugins {
		findings = append(findings, domain.Finding{
			TemplateID: "dynamic-plugin-discovery",
			Name:       fmt.Sprintf("Plugin Detected: %s", plugin),
			Severity:   "info",
			Target:     target,
			MatchedURL: fmt.Sprintf("%s/wp-content/plugins/%s/", target, plugin),
			Method:     "GET",
			StatusCode: 200,
			Evidence:   map[string]string{"plugin": plugin},
			Timestamp:  time.Now(),
		})
	}

	for _, theme := range discovery.Themes {
		findings = append(findings, domain.Finding{
			TemplateID: "dynamic-theme-discovery",
			Name:       fmt.Sprintf("Theme Detected: %s", theme),
			Severity:   "info",
			Target:     target,
			MatchedURL: fmt.Sprintf("%s/wp-content/themes/%s/", target, theme),
			Method:     "GET",
			StatusCode: 200,
			Evidence:   map[string]string{"theme": theme},
			Timestamp:  time.Now(),
		})
	}

	return findings
}
