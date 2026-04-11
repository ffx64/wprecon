package engine

import (
	"fmt"
	"time"

	"github.com/ffx64/wprecon/internal/domain"
	"github.com/ffx64/wprecon/internal/executor"
	"github.com/ffx64/wprecon/internal/pipeline"
	"github.com/ffx64/wprecon/internal/templates"
)

type ScanTask struct {
	Target   string
	Template *domain.Template
}

type ScanResult struct {
	Target   string
	Findings []domain.Finding
	Error    error
}

type Scanner struct {
	templateLoader *templates.Loader
	pipeline       *pipeline.Pipeline
	workerPool     *WorkerPool
	maxWorkers     int
	requestsPerSec int
}

func NewScanner(templateDir string, maxWorkers int, requestsPerSec int) (*Scanner, error) {
	loader, err := templates.NewLoader(templateDir)
	if err != nil {
		return nil, err
	}

	p := pipeline.NewPipeline()

	scanner := &Scanner{
		templateLoader: loader,
		pipeline:       p,
		maxWorkers:     maxWorkers,
		requestsPerSec: requestsPerSec,
	}

	return scanner, nil
}

func (s *Scanner) Scan(targets []string, templateIDs []string) ([]ScanResult, error) {
	return s.ScanWithOptions(targets, templateIDs, DefaultScanOptions())
}

func (s *Scanner) GetTemplates() ([]*domain.Template, error) {
	return s.templateLoader.ListTemplates()
}

func (s *Scanner) ScanWithOptions(targets []string, templateIDs []string, options *ScanOptions) ([]ScanResult, error) {
	if options == nil {
		options = DefaultScanOptions()
	}

	if options.Stealthy {
		options.Mode = ModePassive
		if options.Sleep == 0 {
			options.Sleep = 2 * time.Second
		}
		if options.Concurrency > 2 {
			options.Concurrency = 2
		}
		if options.RateLimit == 0 || options.RateLimit > 5 {
			options.RateLimit = 5
		}
		options.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}

	logger := NewLogger(options.Verbose)

	loadedTemplates, err := s.templateLoader.LoadTemplates(templateIDs)
	if err != nil {
		return nil, err
	}

	for _, tmpl := range loadedTemplates {
		logger.TemplateLoaded(tmpl.ID, tmpl.Info.Name, tmpl.Info.Severity, tmpl.Info.Tags)
	}
	logger.Info(fmt.Sprintf("Loaded %d templates", len(loadedTemplates)))

	beforeFilter := len(loadedTemplates)
	loadedTemplates = s.filterTemplatesByTags(loadedTemplates, options)
	if diff := beforeFilter - len(loadedTemplates); diff > 0 {
		logger.Info(fmt.Sprintf("Filtered %d templates by tags (remaining: %d)", diff, len(loadedTemplates)))
	}

	beforeMode := len(loadedTemplates)
	loadedTemplates = s.filterTemplatesByMode(loadedTemplates, options)
	if diff := beforeMode - len(loadedTemplates); diff > 0 {
		logger.Info(fmt.Sprintf("Filtered %d templates by mode (remaining: %d)", diff, len(loadedTemplates)))
	}

	loadedTemplates = s.prioritizeWAFTemplates(loadedTemplates)

	execOpts := &executor.HTTPExecutorOptions{
		Proxy:           options.Proxy,
		Timeout:         options.Timeout,
		UserAgent:       options.UserAgent,
		FollowRedirects: options.FollowRedirects,
		VerifyTLS:       options.VerifyTLS,
		RandomizeUA:     options.RandomizeUA,
	}
	scanPipeline := pipeline.NewPipelineWithOptions(execOpts)
	scanPipeline.SetLogger(logger)

	logger.Info(fmt.Sprintf("Starting scan: %d targets × %d templates = %d tasks, concurrency=%d",
		len(targets), len(loadedTemplates), len(targets)*len(loadedTemplates), options.Concurrency))

	scanStart := time.Now()

	var discoveryFindings []domain.Finding
	if options.Discovery {
		for _, target := range targets {
			discovery, err := DiscoverFromHTML(target, scanPipeline.GetExecutor(), logger)
			if err != nil {
				logger.Info(fmt.Sprintf("Discovery failed for %s: %v", target, err))
				continue
			}
			discoveryFindings = append(discoveryFindings, DiscoveryToFindings(target, discovery)...)
		}
	}

	rateLimiter := NewRateLimiter(options.RateLimit)

	wp := NewWorkerPool(options.Concurrency, options.Concurrency*2)
	wp.Start(func(task Task) interface{} {
		return s.executeScanTaskWithOptions(task.(*ScanTask), options, rateLimiter, scanPipeline, logger)
	})

	go func() {
		for _, target := range targets {
			for _, tmpl := range loadedTemplates {
				wp.Submit(&ScanTask{
					Target:   target,
					Template: tmpl,
				})
			}
		}
		wp.CloseInput()
	}()

	var results []ScanResult
	totalFindings := 0
	for result := range wp.Results() {
		if scanResult, ok := result.(*ScanResult); ok {
			results = append(results, *scanResult)
			totalFindings += len(scanResult.Findings)
		}
	}

	logger.Summary(len(loadedTemplates), totalFindings, time.Since(scanStart))

	if len(discoveryFindings) > 0 {
		if len(results) > 0 {
			results[0].Findings = append(discoveryFindings, results[0].Findings...)
		} else {
			results = append(results, ScanResult{
				Target:   targets[0],
				Findings: discoveryFindings,
			})
		}
	}

	return results, nil
}

func (s *Scanner) executeScanTaskWithOptions(task *ScanTask, options *ScanOptions, rateLimiter RateLimiter, scanPipeline *pipeline.Pipeline, logger *Logger) *ScanResult {
	rateLimiter.Take()

	if options.Sleep > 0 {
		time.Sleep(options.Sleep)
	}

	logger.TemplateStart(task.Template.ID, task.Target)

	findings, err := scanPipeline.Execute(task.Target, task.Template)
	if err != nil {
		return &ScanResult{
			Target:   task.Target,
			Findings: nil,
			Error:    err,
		}
	}

	return &ScanResult{
		Target:   task.Target,
		Findings: findings,
		Error:    nil,
	}
}

func (s *Scanner) filterTemplatesByTags(templates []*domain.Template, options *ScanOptions) []*domain.Template {
	if len(options.TemplatesTags) == 0 && len(options.SkipTags) == 0 {
		return templates
	}

	filtered := make([]*domain.Template, 0)
	for _, tmpl := range templates {
		if len(options.TemplatesTags) > 0 && !tmpl.HasAnyTag(options.TemplatesTags) {
			continue
		}

		if len(options.SkipTags) > 0 && !tmpl.HasNoneTag(options.SkipTags) {
			continue
		}

		filtered = append(filtered, tmpl)
	}

	return filtered
}

func (s *Scanner) filterTemplatesByMode(templates []*domain.Template, options *ScanOptions) []*domain.Template {
	if options.Mode == ModeNormal {
		return templates
	}

	filtered := make([]*domain.Template, 0)
	for _, tmpl := range templates {
		if options.Mode == ModePassive && tmpl.Mode == "aggressive" {
			continue
		}

		if options.Mode == ModeAggressive && tmpl.Mode == "passive" {
			fmt.Println("[WARN] Skipping passive template in aggressive mode:", tmpl.ID)
			continue
		}

		filtered = append(filtered, tmpl)
	}

	return filtered
}

func (s *Scanner) prioritizeWAFTemplates(templates []*domain.Template) []*domain.Template {
	wafTemplates := make([]*domain.Template, 0)
	otherTemplates := make([]*domain.Template, 0)

	for _, tmpl := range templates {
		if tmpl.HasTag("waf") {
			wafTemplates = append(wafTemplates, tmpl)
		} else {
			otherTemplates = append(otherTemplates, tmpl)
		}
	}

	result := make([]*domain.Template, 0, len(templates))
	result = append(result, wafTemplates...)
	result = append(result, otherTemplates...)

	return result
}
