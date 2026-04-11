package domain

type MatchResult struct {
	Matched  bool
	Type     string
	Evidence map[string]string
}

type Matcher interface {
	Match(response *HTTPResponse, template *Template, request *Request) *MatchResult
}
