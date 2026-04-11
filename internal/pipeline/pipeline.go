package pipeline

import (
	"fmt"
	"time"

	"github.com/ffx64/wprecon/internal/domain"
	"github.com/ffx64/wprecon/internal/executor"
	"github.com/ffx64/wprecon/internal/extractors"
	"github.com/ffx64/wprecon/internal/matchers"
	"github.com/google/uuid"
)

type Pipeline struct {
	httpExecutor     *executor.HTTPExecutor
	requestBuilder   *executor.RequestBuilder
	matcherFactory   *matchers.MatcherFactory
	extractorFactory *extractors.ExtractorFactory
	logger           PipelineLogger
}

type PipelineLogger interface {
	IsVerbose() bool
	RequestStart(method, url string)
	RequestDone(method, url string, status int, timeMs float64, bodyLen int)
	RequestError(method, url string, err error)
	RequestRetry(method, url string, attempt int, err error)
	MatcherResult(matcherType string, matched bool, evidence map[string]string)
	ExtractorResult(name string, values []string)
	FindingDetected(templateID, severity, url string, status int)
	NoMatch(templateID, url string)
}

type noopLogger struct{}

func (n *noopLogger) IsVerbose() bool                                                            { return false }
func (n *noopLogger) RequestStart(method, url string)                                            {}
func (n *noopLogger) RequestDone(method, url string, status int, timeMs float64, bodyLen int)    {}
func (n *noopLogger) RequestError(method, url string, err error)                                 {}
func (n *noopLogger) RequestRetry(method, url string, attempt int, err error)                    {}
func (n *noopLogger) MatcherResult(matcherType string, matched bool, evidence map[string]string) {}
func (n *noopLogger) ExtractorResult(name string, values []string)                               {}
func (n *noopLogger) FindingDetected(templateID, severity, url string, status int)               {}
func (n *noopLogger) NoMatch(templateID, url string)                                             {}

func NewPipeline() *Pipeline {
	return &Pipeline{
		httpExecutor:     executor.NewHTTPExecutor(10, 2),
		requestBuilder:   executor.NewRequestBuilder(executor.NewVariables()),
		matcherFactory:   matchers.NewMatcherFactory(),
		extractorFactory: extractors.NewExtractorFactory(),
		logger:           &noopLogger{},
	}
}

func NewPipelineWithOptions(opts *executor.HTTPExecutorOptions) *Pipeline {
	return &Pipeline{
		httpExecutor:     executor.NewHTTPExecutorWithOptions(opts),
		requestBuilder:   executor.NewRequestBuilder(executor.NewVariables()),
		matcherFactory:   matchers.NewMatcherFactory(),
		extractorFactory: extractors.NewExtractorFactory(),
		logger:           &noopLogger{},
	}
}

func (p *Pipeline) SetLogger(logger PipelineLogger) {
	if logger != nil {
		p.logger = logger
	}
}

func (p *Pipeline) GetExecutor() *executor.HTTPExecutor {
	return p.httpExecutor
}

func (p *Pipeline) Execute(target string, template *domain.Template) ([]domain.Finding, error) {
	var findings []domain.Finding

	vars := executor.NewVariables()
	vars.InitializeForTarget(target, time.Now())
	requestBuilder := executor.NewRequestBuilder(vars)

	for _, req := range template.Requests {
		paths := req.Path
		if len(paths) == 0 {
			paths = []string{"/"}
		}

		for _, path := range paths {
			singlePathReq := req
			singlePathReq.Path = []string{path}

			httpReq, err := requestBuilder.BuildRequest(target, &singlePathReq)
			if err != nil {
				p.logger.RequestError(singlePathReq.Method, target+path, err)
				continue
			}

			p.logger.RequestStart(httpReq.Method, httpReq.URL)

			httpResp, err := p.httpExecutor.Execute(httpReq)
			if err != nil {
				p.logger.RequestError(httpReq.Method, httpReq.URL, err)
				continue
			}

			p.logger.RequestDone(httpReq.Method, httpReq.URL, httpResp.StatusCode, httpResp.Time, len(httpResp.Body))

			matched := p.checkMatchers(&singlePathReq, httpResp, template)
			if !matched {
				p.logger.NoMatch(template.ID, httpResp.URL)
				continue
			}

			evidence := p.extractEvidence(&singlePathReq, httpResp, template)

			finding := domain.Finding{
				ID:         generateFindingID(),
				TemplateID: template.ID,
				Name:       template.Info.Name,
				Severity:   template.Info.Severity,
				Target:     target,
				MatchedURL: httpResp.URL,
				Method:     singlePathReq.Method,
				StatusCode: httpResp.StatusCode,
				Evidence:   evidence,
				Timestamp:  time.Now(),
			}

			p.logger.FindingDetected(template.ID, template.Info.Severity, httpResp.URL, httpResp.StatusCode)
			findings = append(findings, finding)
		}
	}

	return findings, nil
}

func (p *Pipeline) checkMatchers(req *domain.Request, resp *domain.HTTPResponse, template *domain.Template) bool {
	if len(req.Matchers) == 0 {
		return resp.StatusCode < 500
	}

	for _, matcherConfig := range req.Matchers {
		matcher, err := p.matcherFactory.CreateMatcher(matcherConfig)
		if err != nil {
			continue
		}

		result := matcher.Match(resp, template, req)
		p.logger.MatcherResult(matcherConfig.Type, result.Matched, result.Evidence)
		if !result.Matched {
			return false
		}
	}

	return true
}

func (p *Pipeline) extractEvidence(req *domain.Request, resp *domain.HTTPResponse, template *domain.Template) map[string]string {
	evidence := make(map[string]string)

	if len(req.Extractors) == 0 {
		return evidence
	}

	for _, extractorConfig := range req.Extractors {
		extractor, err := p.extractorFactory.CreateExtractor(extractorConfig)
		if err != nil {
			continue
		}

		results := extractor.Extract(resp, template, req)
		for _, result := range results {
			if len(result.Values) > 0 {
				evidence[result.Name] = result.Values[0]
				p.logger.ExtractorResult(result.Name, result.Values)
			}
		}
	}

	return evidence
}

func generateFindingID() string {
	return fmt.Sprintf("finding_%d_%s", time.Now().UnixNano(), uuid.New().String())
}
