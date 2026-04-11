package extractors

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/ffx64/wprecon/internal/domain"
)

type RegexExtractor struct {
	name     string
	patterns []*regexp.Regexp
}

func (re *RegexExtractor) Extract(response *domain.HTTPResponse, template *domain.Template, request *domain.Request) []domain.ExtractionResult {
	var results []domain.ExtractionResult

	for _, pattern := range re.patterns {
		matches := pattern.FindAllStringSubmatch(response.BodyStr, -1)
		if len(matches) > 0 {
			values := make([]string, 0)
			for _, match := range matches {
				if len(match) > 1 {
					values = append(values, match[1])
				} else if len(match) > 0 {
					values = append(values, match[0])
				}
			}

			if len(values) > 0 {
				results = append(results, domain.ExtractionResult{
					Name:   re.name,
					Values: values,
				})
			}
		}
	}

	return results
}

type JSONExtractor struct {
	name  string
	paths []string
}

func (je *JSONExtractor) Extract(response *domain.HTTPResponse, template *domain.Template, request *domain.Request) []domain.ExtractionResult {
	var results []domain.ExtractionResult

	var jsonData map[string]interface{}
	err := json.Unmarshal(response.Body, &jsonData)
	if err != nil {
		return results
	}

	for _, path := range je.paths {
		value := extractJSONValue(jsonData, path)
		if value != "" {
			results = append(results, domain.ExtractionResult{
				Name:   je.name,
				Values: []string{value},
			})
		}
	}

	return results
}

func extractJSONValue(data map[string]interface{}, path string) string {
	value, exists := data[path]
	if !exists {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

type ExtractorFactory struct{}

func (ef *ExtractorFactory) CreateExtractor(config domain.ExtractorConfig) (domain.Extractor, error) {
	switch config.Type {
	case "regex":
		patterns := make([]*regexp.Regexp, len(config.Regex))
		for i, pattern := range config.Regex {
			compiled, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regex pattern: %w", err)
			}
			patterns[i] = compiled
		}
		return &RegexExtractor{name: config.Name, patterns: patterns}, nil

	case "json":
		return &JSONExtractor{name: config.Name, paths: config.JSON}, nil

	default:
		return nil, fmt.Errorf("unknown extractor type: %s", config.Type)
	}
}

func NewExtractorFactory() *ExtractorFactory {
	return &ExtractorFactory{}
}
