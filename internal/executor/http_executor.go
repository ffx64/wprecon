package executor

import (
	"bytes"
	"crypto/tls"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ffx64/wprecon/internal/domain"
)

var randomUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0",
	"Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0",
}

type HTTPExecutor struct {
	client      *http.Client
	timeout     time.Duration
	retries     int
	userAgent   string
	randomizeUA bool
}

func NewHTTPExecutor(timeout int, retries int) *HTTPExecutor {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     10,
		},
	}

	return &HTTPExecutor{
		client:    client,
		timeout:   time.Duration(timeout) * time.Second,
		retries:   retries,
		userAgent: "wprecon/1.0",
	}
}

type HTTPExecutorOptions struct {
	Proxy           string
	Timeout         time.Duration
	UserAgent       string
	FollowRedirects bool
	VerifyTLS       bool
	RandomizeUA     bool
}

func NewHTTPExecutorWithOptions(opts *HTTPExecutorOptions) *HTTPExecutor {
	if opts == nil {
		return NewHTTPExecutor(10, 2)
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     10,
	}

	if opts.Proxy != "" {
		proxyURL, err := ParseProxyURL(opts.Proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	if !opts.VerifyTLS {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	userAgent := opts.UserAgent
	if userAgent == "" {
		userAgent = "wprecon/1.0"
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !opts.FollowRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	return &HTTPExecutor{
		client:      client,
		timeout:     timeout,
		retries:     2,
		userAgent:   userAgent,
		randomizeUA: opts.RandomizeUA,
	}
}

func ParseProxyURL(proxyStr string) (*url.URL, error) {
	return url.Parse(proxyStr)
}

func (he *HTTPExecutor) Execute(req *domain.HTTPRequest) (*domain.HTTPResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= he.retries; attempt++ {
		resp, err := he.executeOnce(req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}

	return nil, lastErr
}

func (he *HTTPExecutor) executeOnce(req *domain.HTTPRequest) (*domain.HTTPResponse, error) {
	start := time.Now()

	httpReq, err := http.NewRequest(req.Method, req.URL, nil)
	if err != nil {
		return nil, err
	}

	if req.Headers != nil {
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}
	}

	if httpReq.Header.Get("User-Agent") == "" {
		if he.randomizeUA {
			httpReq.Header.Set("User-Agent", randomUserAgents[rand.Intn(len(randomUserAgents))])
		} else {
			httpReq.Header.Set("User-Agent", he.userAgent)
		}
	}

	if len(req.Body) > 0 {
		httpReq.Body = io.NopCloser(bytes.NewReader(req.Body))
	}

	if req.Cookies != nil {
		for key, value := range req.Cookies {
			httpReq.AddCookie(&http.Cookie{
				Name:  key,
				Value: value,
			})
		}
	}

	httpResp, err := he.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	elapsed := time.Since(start).Seconds() * 1000 // em ms

	return &domain.HTTPResponse{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		Body:       bodyBytes,
		BodyStr:    string(bodyBytes),
		URL:        req.URL,
		Time:       elapsed,
	}, nil
}

func (he *HTTPExecutor) SetUserAgent(ua string) {
	he.userAgent = ua
}

func (he *HTTPExecutor) SetTimeout(timeout time.Duration) {
	he.timeout = timeout
	he.client.Timeout = timeout
}

func IsNetworkError(err error) bool {
	if _, ok := err.(net.Error); ok {
		return true
	}
	return false
}
