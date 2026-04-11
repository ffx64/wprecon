package matchers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ffx64/wprecon/internal/domain"
)

type StatusMatcher struct {
	expectedCodes []int
}

func (sm *StatusMatcher) Match(response *domain.HTTPResponse, template *domain.Template, request *domain.Request) *domain.MatchResult {
	for _, code := range sm.expectedCodes {
		if response.StatusCode == code {
			return &domain.MatchResult{
				Matched:  true,
				Type:     "status",
				Evidence: map[string]string{"status_code": fmt.Sprintf("%d", response.StatusCode)},
			}
		}
	}
	return &domain.MatchResult{
		Matched:  false,
		Type:     "status",
		Evidence: nil,
	}
}

type WordMatcher struct {
	words []string
}

func (wm *WordMatcher) Match(response *domain.HTTPResponse, template *domain.Template, request *domain.Request) *domain.MatchResult {
	searchable := response.BodyStr
	for key, values := range response.Headers {
		for _, val := range values {
			searchable += "\n" + key + ": " + val
		}
	}

	for _, word := range wm.words {
		if strings.Contains(searchable, word) {
			return &domain.MatchResult{
				Matched:  true,
				Type:     "word",
				Evidence: map[string]string{"word": word},
			}
		}
	}
	return &domain.MatchResult{
		Matched:  false,
		Type:     "word",
		Evidence: nil,
	}
}

type RegexMatcher struct {
	patterns []*regexp.Regexp
}

func (rm *RegexMatcher) Match(response *domain.HTTPResponse, template *domain.Template, request *domain.Request) *domain.MatchResult {
	for _, pattern := range rm.patterns {
		matches := pattern.FindStringSubmatch(response.BodyStr)
		if len(matches) > 0 {
			evidence := make(map[string]string)
			for i, match := range matches {
				evidence[fmt.Sprintf("match_%d", i)] = match
			}
			return &domain.MatchResult{
				Matched:  true,
				Type:     "regex",
				Evidence: evidence,
			}
		}
	}
	return &domain.MatchResult{
		Matched:  false,
		Type:     "regex",
		Evidence: nil,
	}
}

type HeaderMatcher struct {
	headers     map[string]string
	headerNames []string
}

func (hm *HeaderMatcher) Match(response *domain.HTTPResponse, template *domain.Template, request *domain.Request) *domain.MatchResult {
	evidence := make(map[string]string)

	if len(hm.headerNames) > 0 {
		for _, name := range hm.headerNames {
			values, exists := response.Headers[name]
			if exists && len(values) > 0 {
				evidence[name] = values[0]
				return &domain.MatchResult{
					Matched:  true,
					Type:     "header",
					Evidence: evidence,
				}
			}
		}
		return &domain.MatchResult{
			Matched:  false,
			Type:     "header",
			Evidence: nil,
		}
	}

	matched := true
	for key, expectedValue := range hm.headers {
		headerValues, exists := response.Headers[key]
		if !exists {
			matched = false
			break
		}

		if expectedValue == "" {
			if len(headerValues) > 0 {
				evidence[key] = headerValues[0]
			}
			continue
		}

		found := false
		for _, val := range headerValues {
			if strings.Contains(val, expectedValue) {
				found = true
				evidence[key] = val
				break
			}
		}
		if !found {
			matched = false
			break
		}
	}

	return &domain.MatchResult{
		Matched:  matched,
		Type:     "header",
		Evidence: evidence,
	}
}

type SizeMatcher struct {
	min int
	max int
}

func (sm *SizeMatcher) Match(response *domain.HTTPResponse, template *domain.Template, request *domain.Request) *domain.MatchResult {
	size := len(response.Body)

	matched := true
	if sm.min > 0 && size < sm.min {
		matched = false
	}
	if sm.max > 0 && size > sm.max {
		matched = false
	}

	evidence := map[string]string{"size": fmt.Sprintf("%d", size)}

	return &domain.MatchResult{
		Matched:  matched,
		Type:     "size",
		Evidence: evidence,
	}
}

type NegativeMatcher struct {
	inner domain.Matcher
}

func (nm *NegativeMatcher) Match(response *domain.HTTPResponse, template *domain.Template, request *domain.Request) *domain.MatchResult {
	result := nm.inner.Match(response, template, request)
	result.Matched = !result.Matched
	return result
}

type MatcherFactory struct{}

func (mf *MatcherFactory) CreateMatcher(config domain.MatcherConfig) (domain.Matcher, error) {
	var matcher domain.Matcher
	var err error

	switch config.Type {
	case "status":
		matcher = &StatusMatcher{expectedCodes: config.Status}
	case "word":
		matcher = &WordMatcher{words: config.Words}
	case "regex":
		patterns := make([]*regexp.Regexp, len(config.Regex))
		for i, pattern := range config.Regex {
			compiled, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regex pattern: %w", err)
			}
			patterns[i] = compiled
		}
		matcher = &RegexMatcher{patterns: patterns}
	case "header":
		matcher = &HeaderMatcher{headers: config.Header, headerNames: config.Headers}
	case "size":
		matcher = &SizeMatcher{min: config.Size.Min, max: config.Size.Max}
	default:
		return nil, fmt.Errorf("unknown matcher type: %s", config.Type)
	}

	if err != nil {
		return nil, err
	}

	if config.Negative {
		matcher = &NegativeMatcher{inner: matcher}
	}

	return matcher, nil
}

func NewMatcherFactory() *MatcherFactory {
	return &MatcherFactory{}
}
