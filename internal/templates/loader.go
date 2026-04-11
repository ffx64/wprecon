package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ffx64/wprecon/internal/domain"
)

type Loader struct {
	templateDir string
	parser      *Parser
	cache       map[string]*domain.Template
}

func NewLoader(templateDir string) (*Loader, error) {
	_, err := os.ReadDir(templateDir)
	if err != nil {
		return nil, fmt.Errorf("invalid template directory: %w", err)
	}

	return &Loader{
		templateDir: templateDir,
		parser:      NewParser(),
		cache:       make(map[string]*domain.Template),
	}, nil
}

func (l *Loader) LoadTemplates(templateIDs []string) ([]*domain.Template, error) {
	templates := []*domain.Template{}

	if len(templateIDs) == 0 {
		return l.loadAll()
	}

	for _, id := range templateIDs {
		tmpl, err := l.LoadTemplate(id)
		if err != nil {
			return nil, err
		}
		templates = append(templates, tmpl)
	}

	return templates, nil
}

func (l *Loader) LoadTemplate(id string) (*domain.Template, error) {
	if tmpl, ok := l.cache[id]; ok {
		return tmpl, nil
	}

	filePath, err := l.findTemplateFile(id)
	if err != nil {
		return nil, err
	}

	tmpl, err := l.parser.Parse(filePath)
	if err != nil {
		return nil, err
	}

	l.cache[id] = tmpl

	return tmpl, nil
}

func (l *Loader) findTemplateFile(id string) (string, error) {
	var found string
	err := filepath.Walk(l.templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		parser := NewParser()
		tmpl, err := parser.Parse(path)
		if err != nil {
			return nil
		}

		if tmpl.ID == id {
			found = path
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil && err != filepath.SkipDir {
		return "", fmt.Errorf("error searching for template: %w", err)
	}

	if found == "" {
		return "", fmt.Errorf("template not found: %s", id)
	}

	return found, nil
}

func (l *Loader) ListTemplates() ([]*domain.Template, error) {
	return l.loadAll()
}

func (l *Loader) loadAll() ([]*domain.Template, error) {
	templates := []*domain.Template{}

	err := filepath.Walk(l.templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		tmpl, err := l.parser.Parse(path)
		if err != nil {
			return nil
		}

		templates = append(templates, tmpl)
		return nil
	})

	return templates, err
}

func (l *Loader) ClearCache() {
	l.cache = make(map[string]*domain.Template)
}
