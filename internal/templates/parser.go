package templates

import (
	"fmt"
	"os"

	"github.com/ffx64/wprecon/internal/domain"
	"gopkg.in/yaml.v3"
)

type Parser struct{}

func (p *Parser) Parse(filePath string) (*domain.Template, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var template domain.Template
	err = yaml.Unmarshal(content, &template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validateTemplate(&template); err != nil {
		return nil, err
	}

	return &template, nil
}

func validateTemplate(t *domain.Template) error {
	if t.ID == "" {
		return fmt.Errorf("template missing required field: id")
	}

	if t.Info.Name == "" {
		return fmt.Errorf("template missing required field: info.name")
	}

	if len(t.Requests) == 0 {
		return fmt.Errorf("template must have at least one request")
	}

	for i, req := range t.Requests {
		if req.Method == "" {
			req.Method = "GET"
			t.Requests[i] = req
		}
		if len(req.Path) == 0 {
			req.Path = []string{"/"}
			t.Requests[i] = req
		}
	}

	return nil
}

func NewParser() *Parser {
	return &Parser{}
}
