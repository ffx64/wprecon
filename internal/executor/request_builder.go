package executor

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ffx64/wprecon/internal/domain"
)

type RequestBuilder struct {
	variables *Variables
}

func NewRequestBuilder(vars *Variables) *RequestBuilder {
	return &RequestBuilder{
		variables: vars,
	}
}

func (rb *RequestBuilder) BuildRequest(baseURL string, req *domain.Request) (*domain.HTTPRequest, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL is empty")
	}

	paths := rb.variables.resolvePaths(req.Path)
	if len(paths) == 0 {
		paths = []string{"/"}
	}

	path := paths[0]

	requestURL, err := rb.buildURL(baseURL, path)
	if err != nil {
		return nil, err
	}

	headers := rb.variables.ResolveHeaders(req.Headers)

	body := rb.variables.ResolveBody(req.Body)

	cookies := req.Cookies
	if cookies == nil {
		cookies = make(map[string]string)
	}

	queryParams := req.QueryParams
	if queryParams == nil {
		queryParams = make(map[string]string)
	}

	return &domain.HTTPRequest{
		URL:         requestURL,
		Method:      req.Method,
		Headers:     headers,
		Body:        []byte(body),
		Cookies:     cookies,
		QueryParams: queryParams,
	}, nil
}

func (rb *RequestBuilder) buildURL(baseURL, path string) (string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid baseURL: %w", err)
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if idx := strings.Index(path, "?"); idx >= 0 {
		parsedURL.Path = path[:idx]
		parsedURL.RawQuery = path[idx+1:]
	} else {
		parsedURL.Path = path
	}

	return parsedURL.String(), nil
}
