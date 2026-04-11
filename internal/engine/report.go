package engine

import (
	"github.com/ffx64/wprecon/internal/enumerators"
)

type EnhancedScanResult struct {
	Target           string                   `json:"target"`
	Status           string                   `json:"status"` // success, error, partial
	Findings         []ScanFindingJSON        `json:"findings"`
	WordPressVersion string                   `json:"wordpress_version,omitempty"`
	Plugins          []enumerators.PluginInfo `json:"plugins,omitempty"`
	Themes           []enumerators.ThemeInfo  `json:"themes,omitempty"`
	Users            []enumerators.UserInfo   `json:"users,omitempty"`
	Error            string                   `json:"error,omitempty"`
	EnumerationTime  int64                    `json:"enumeration_time_ms,omitempty"`
	ScanStatistics   *ScanStatistics          `json:"statistics,omitempty"`
}

type ScanFindingJSON struct {
	TemplateID string            `json:"template_id"`
	Name       string            `json:"name"`
	Severity   string            `json:"severity"`
	MatchedURL string            `json:"matched_url"`
	StatusCode int               `json:"status_code"`
	Evidence   map[string]string `json:"evidence,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
}

type ScanStatistics struct {
	TotalTemplates      int   `json:"total_templates"`
	TemplatesExecuted   int   `json:"templates_executed"`
	TotalFindings       int   `json:"total_findings"`
	HighSeverityCount   int   `json:"high_severity"`
	MediumSeverityCount int   `json:"medium_severity"`
	LowSeverityCount    int   `json:"low_severity"`
	InfoCount           int   `json:"info_count"`
	ScanTimeMS          int64 `json:"scan_time_ms"`
}

type Report struct {
	Timestamp int64                `json:"timestamp"`
	ScanMode  string               `json:"mode"`
	Results   []EnhancedScanResult `json:"results"`
	Summary   ReportSummary        `json:"summary"`
	TotalTime int64                `json:"total_time_ms"`
}

type ReportSummary struct {
	TotalTargets         int `json:"total_targets"`
	SuccessfulScans      int `json:"successful_scans"`
	FailedScans          int `json:"failed_scans"`
	TotalFindings        int `json:"total_findings"`
	TotalUsersDiscovered int `json:"total_users_discovered"`
	TotalPluginsFound    int `json:"total_plugins_found"`
	TotalThemesFound     int `json:"total_themes_found"`
	UniqueVersionsFound  int `json:"unique_versions_found"`
}
