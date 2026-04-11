package domain

type ExtractionResult struct {
	Name   string
	Values []string
}

type Extractor interface {
	Extract(response *HTTPResponse, template *Template, request *Request) []ExtractionResult
}
