package domain

type Template struct {
	ID       string    `yaml:"id"`
	Info     Info      `yaml:"info"`
	Requests []Request `yaml:"requests"`
	Mode     string    `yaml:"mode,omitempty"`
}

type Info struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Severity    string   `yaml:"severity"`
	Tags        []string `yaml:"tags,omitempty"`
	Author      string   `yaml:"author,omitempty"`
}

func (t *Template) HasTag(tag string) bool {
	for _, tagVal := range t.Info.Tags {
		if tagVal == tag {
			return true
		}
	}
	return false
}

func (t *Template) HasAnyTag(tags []string) bool {
	if len(tags) == 0 {
		return true
	}
	for _, reqTag := range tags {
		if t.HasTag(reqTag) {
			return true
		}
	}
	return false
}

func (t *Template) HasNoneTag(excludeTags []string) bool {
	for _, excludeTag := range excludeTags {
		if t.HasTag(excludeTag) {
			return false
		}
	}
	return true
}

type Request struct {
	Method      string            `yaml:"method"`
	Path        []string          `yaml:"path,omitempty"`
	Headers     map[string]string `yaml:"headers,omitempty"`
	Body        string            `yaml:"body,omitempty"`
	Matchers    []MatcherConfig   `yaml:"matchers,omitempty"`
	Extractors  []ExtractorConfig `yaml:"extractors,omitempty"`
	Cookies     map[string]string `yaml:"cookies,omitempty"`
	QueryParams map[string]string `yaml:"query_params,omitempty"`
}

type MatcherConfig struct {
	Type     string            `yaml:"type"`
	Negative bool              `yaml:"negative,omitempty"`
	Status   []int             `yaml:"status,omitempty"`
	Words    []string          `yaml:"words,omitempty"`
	Regex    []string          `yaml:"regex,omitempty"`
	Header   map[string]string `yaml:"header,omitempty"`
	Headers  []string          `yaml:"headers,omitempty"`
	Size     SizeMatcher       `yaml:"size,omitempty"`
	DSL      []string          `yaml:"dsl,omitempty"`
}

type SizeMatcher struct {
	Min int `yaml:"min,omitempty"`
	Max int `yaml:"max,omitempty"`
}

type ExtractorConfig struct {
	Type  string   `yaml:"type"`
	Name  string   `yaml:"name"`
	Regex []string `yaml:"regex,omitempty"`
	JSON  []string `yaml:"json,omitempty"`
}
