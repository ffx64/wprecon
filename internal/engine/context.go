package engine

import (
	"context"
	"net/url"
	"time"
)

type ScanMode string

const (
	ModePassive    ScanMode = "passive"
	ModeAggressive ScanMode = "aggressive"
	ModeNormal     ScanMode = "normal"
)

type ScanOptions struct {
	Mode ScanMode

	RateLimit   int
	Sleep       time.Duration
	Concurrency int

	Proxy     string
	Timeout   time.Duration
	UserAgent string

	FollowRedirects bool
	VerifyTLS       bool
	Verbose         bool
	RandomizeUA     bool
	Stealthy        bool
	Discovery       bool

	TemplatesTags []string
	SkipTags      []string
}

func DefaultScanOptions() *ScanOptions {
	return &ScanOptions{
		Mode:            ModeNormal,
		RateLimit:       50,
		Sleep:           0,
		Concurrency:     10,
		Timeout:         10 * time.Second,
		UserAgent:       "wprecon/1.0",
		FollowRedirects: true,
		VerifyTLS:       true,
	}
}

type ScanContext struct {
	ctx    context.Context
	cancel context.CancelFunc

	Variables map[string]string
	Options   *ScanOptions
}

func NewScanContext(timeout int, options *ScanOptions) *ScanContext {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)

	if options == nil {
		options = DefaultScanOptions()
	}

	return &ScanContext{
		ctx:       ctx,
		cancel:    cancel,
		Variables: make(map[string]string),
		Options:   options,
	}
}

func (sc *ScanContext) Context() context.Context {
	return sc.ctx
}

func (sc *ScanContext) Done() {
	sc.cancel()
}

func (sc *ScanContext) SetVariable(key, value string) {
	sc.Variables[key] = value
}

func (sc *ScanContext) GetVariable(key string) string {
	return sc.Variables[key]
}

func (sc *ScanContext) IsPassiveMode() bool {
	return sc.Options.Mode == ModePassive
}

func (sc *ScanContext) IsAggressiveMode() bool {
	return sc.Options.Mode == ModeAggressive
}

func (sc *ScanContext) GetProxyURL() (*url.URL, error) {
	if sc.Options.Proxy == "" {
		return nil, nil
	}
	return url.Parse(sc.Options.Proxy)
}
