package domain

import "time"

type Finding struct {
	ID          string            `json:"id"`
	TemplateID  string            `json:"template_id"`
	Name        string            `json:"name"`
	Severity    string            `json:"severity"`
	Target      string            `json:"target"`
	MatchedURL  string            `json:"matched_url"`
	Method      string            `json:"method"`
	StatusCode  int               `json:"status_code"`
	Evidence    map[string]string `json:"evidence"`
	Timestamp   time.Time         `json:"timestamp"`
	Description string            `json:"description,omitempty"`
}

type Report struct {
	Timestamp time.Time `json:"timestamp"`
	Targets   []string  `json:"targets"`
	Duration  float64   `json:"duration_seconds"`
	Findings  []Finding `json:"findings"`
	Summary   Summary   `json:"summary"`
}

type Summary struct {
	Total    int `json:"total"`
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Info     int `json:"info"`
}
