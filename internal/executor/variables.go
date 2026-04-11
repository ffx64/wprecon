package executor

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ffx64/wprecon/internal/domain"
)

type Variables struct {
	vars map[string]string
}

func NewVariables() *Variables {
	return &Variables{
		vars: make(map[string]string),
	}
}

func (v *Variables) Set(key, value string) {
	v.vars[key] = value
}

func (v *Variables) Get(key string) string {
	return v.vars[key]
}

func (v *Variables) Resolve(input string) string {
	result := input

	for key, value := range v.vars {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

func (v *Variables) ResolveURL(baseURL, path string) (string, error) {
	resolvedPath := v.Resolve(path)

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	parsedURL.Path = resolvedPath
	return parsedURL.String(), nil
}

func (v *Variables) ResolveHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return make(map[string]string)
	}

	resolved := make(map[string]string)
	for key, value := range headers {
		resolved[key] = v.Resolve(value)
	}
	return resolved
}

func (v *Variables) ResolveBody(body string) string {
	return v.Resolve(body)
}

func (v *Variables) ResolveRequest(req *domain.Request) *domain.Request {
	return &domain.Request{
		Method:      req.Method,
		Path:        v.resolvePaths(req.Path),
		Headers:     v.ResolveHeaders(req.Headers),
		Body:        v.ResolveBody(req.Body),
		Matchers:    req.Matchers,
		Extractors:  req.Extractors,
		Cookies:     req.Cookies,
		QueryParams: req.QueryParams,
	}
}

func (v *Variables) resolvePaths(paths []string) []string {
	resolved := make([]string, len(paths))
	for i, p := range paths {
		resolved[i] = v.Resolve(p)
	}
	return resolved
}

func (v *Variables) InitializeForTarget(target string, startTime time.Time) {
	v.Set("BaseURL", target)
	v.Set("Timestamp", fmt.Sprintf("%d", startTime.Unix()))
	v.Set("RandomInt", fmt.Sprintf("%d", time.Now().Nanosecond()%10000))
}
