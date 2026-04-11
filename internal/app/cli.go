package app

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ffx64/wprecon/internal/domain"
	"github.com/ffx64/wprecon/internal/engine"
)

type CLIApp struct {
	scanner *engine.Scanner
}

func NewCLIApp(templateDir string) (*CLIApp, error) {
	scanner, err := engine.NewScanner(templateDir, 10, 50)
	if err != nil {
		return nil, err
	}

	return &CLIApp{
		scanner: scanner,
	}, nil
}

func (ca *CLIApp) Run(args []string) error {
	if len(args) < 2 {
		return ca.printHelp()
	}

	command := args[1]

	switch command {
	case "scan":
		return ca.runScan(args[2:])
	case "brute":
		return ca.runBrute(args[2:])
	case "list-templates":
		return ca.listTemplates()
	case "help":
		return ca.printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		return ca.printHelp()
	}
}

func (ca *CLIApp) runScan(args []string) error {
	if len(args) < 1 {
		fmt.Println("Usage: wprecon scan <target> [options]")
		fmt.Println("\nOptions:")
		fmt.Println("  -t, --templates       Comma-separated template IDs")
		fmt.Println("  -o, --output          Output format (json, table)")
		fmt.Println("  --json                Shortcut for --output json")
		fmt.Println("\nAdvanced Options:")
		fmt.Println("  --passive             Run in passive mode (no aggressive probes)")
		fmt.Println("  --aggressive          Run in aggressive mode (all templates)")
		fmt.Println("  --rate-limit N        Requests per second (default: unlimited)")
		fmt.Println("  --sleep DURATION      Sleep between requests (e.g., 500ms)")
		fmt.Println("  --concurrency N       Number of concurrent workers (default: 10)")
		fmt.Println("  --proxy URL           Proxy URL (http://host:port or socks5://host:port)")
		fmt.Println("  --timeout DURATION    Request timeout (e.g., 10s)")
		fmt.Println("  --user-agent STRING   Custom user agent")
		fmt.Println("  --no-redirects        Don't follow HTTP redirects")
		fmt.Println("  --insecure            Skip TLS certificate verification")
		fmt.Println("  -v, --verbose         Show detailed output for every request")
		fmt.Println("  --random-ua           Randomize User-Agent per request")
		fmt.Println("  --stealthy            Stealthy mode (passive, low rate, delays)")
		fmt.Println("  --discovery           Discover plugins/themes from HTML source")
		fmt.Println("  --output-file FILE    Write results to file (JSON)")
		fmt.Println("  --tags TAG1,TAG2      Run templates with these tags")
		fmt.Println("  --skip-tags TAG1      Skip templates with these tags")
		return nil
	}

	target := args[0]
	templates := []string{}
	outputFormat := "table"
	outputFile := ""

	options := engine.DefaultScanOptions()

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-t", "--templates":
			if i+1 < len(args) {
				templates = strings.Split(args[i+1], ",")
				for j := range templates {
					templates[j] = strings.TrimSpace(templates[j])
				}
				i++
			}
		case "-o", "--output":
			if i+1 < len(args) {
				outputFormat = args[i+1]
				i++
			}
		case "--json":
			outputFormat = "json"
		case "--passive":
			options.Mode = engine.ModePassive
		case "--aggressive":
			options.Mode = engine.ModeAggressive
		case "--rate-limit":
			if i+1 < len(args) {
				if rateLimit, err := strconv.Atoi(args[i+1]); err == nil {
					options.RateLimit = rateLimit
				}
				i++
			}
		case "--sleep":
			if i+1 < len(args) {
				if duration, err := time.ParseDuration(args[i+1]); err == nil {
					options.Sleep = duration
				}
				i++
			}
		case "--concurrency":
			if i+1 < len(args) {
				if concurrency, err := strconv.Atoi(args[i+1]); err == nil {
					options.Concurrency = concurrency
				}
				i++
			}
		case "--proxy":
			if i+1 < len(args) {
				options.Proxy = args[i+1]
				i++
			}
		case "--timeout":
			if i+1 < len(args) {
				if timeout, err := time.ParseDuration(args[i+1]); err == nil {
					options.Timeout = timeout
				}
				i++
			}
		case "--user-agent":
			if i+1 < len(args) {
				options.UserAgent = args[i+1]
				i++
			}
		case "--no-redirects":
			options.FollowRedirects = false
		case "--insecure":
			options.VerifyTLS = false
		case "-v", "--verbose":
			options.Verbose = true
		case "--random-ua":
			options.RandomizeUA = true
		case "--stealthy":
			options.Stealthy = true
		case "--discovery":
			options.Discovery = true
		case "--output-file":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		case "--tags":
			if i+1 < len(args) {
				options.TemplatesTags = strings.Split(args[i+1], ",")
				for j := range options.TemplatesTags {
					options.TemplatesTags[j] = strings.TrimSpace(options.TemplatesTags[j])
				}
				i++
			}
		case "--skip-tags":
			if i+1 < len(args) {
				options.SkipTags = strings.Split(args[i+1], ",")
				for j := range options.SkipTags {
					options.SkipTags[j] = strings.TrimSpace(options.SkipTags[j])
				}
				i++
			}
		}
	}

	fmt.Printf("Scanning: %s\n", target)
	if len(templates) > 0 {
		fmt.Printf("Templates: %v\n", templates)
	}
	fmt.Printf("Mode: %s\n", options.Mode)
	if options.RateLimit > 0 {
		fmt.Printf("Rate Limit: %d req/s\n", options.RateLimit)
	}
	if options.Concurrency > 0 {
		fmt.Printf("Concurrency: %d\n", options.Concurrency)
	}
	if options.Proxy != "" {
		fmt.Printf("Proxy: %s\n", options.Proxy)
	}
	if len(options.TemplatesTags) > 0 {
		fmt.Printf("Filter Tags: %v\n", options.TemplatesTags)
	}
	if len(options.SkipTags) > 0 {
		fmt.Printf("Skip Tags: %v\n", options.SkipTags)
	}
	if options.Verbose {
		fmt.Println("Verbose: enabled")
	}
	if options.RandomizeUA {
		fmt.Println("Random User-Agent: enabled")
	}
	if options.Stealthy {
		fmt.Println("Stealthy: enabled (passive, low rate, delays)")
	}
	if options.Discovery {
		fmt.Println("Discovery: enabled (extract plugins/themes from HTML)")
	}
	if outputFile != "" {
		fmt.Printf("Output File: %s\n", outputFile)
	}
	fmt.Println()

	results, err := ca.scanner.ScanWithOptions([]string{target}, templates, options)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if outputFile != "" {
		if err := ca.writeResultsToFile(results, outputFormat, outputFile); err != nil {
			fmt.Printf("Warning: failed to write output file: %v\n", err)
		} else {
			fmt.Printf("\nResults written to %s\n", outputFile)
		}
	}
	return ca.printResults(results, outputFormat)
}

func (ca *CLIApp) printResults(results []engine.ScanResult, format string) error {
	switch format {
	case "json":
		return ca.printJSON(results)
	default:
		return ca.printTable(results)
	}
}

func (ca *CLIApp) printJSON(results []engine.ScanResult) error {
	type jsonFinding struct {
		TemplateID string            `json:"template_id"`
		Name       string            `json:"name"`
		Severity   string            `json:"severity"`
		Target     string            `json:"target"`
		URL        string            `json:"matched_url"`
		Method     string            `json:"method"`
		StatusCode int               `json:"status_code"`
		Evidence   map[string]string `json:"evidence,omitempty"`
	}

	type jsonOutput struct {
		Target   string        `json:"target"`
		Findings []jsonFinding `json:"findings"`
		Error    string        `json:"error,omitempty"`
	}

	var output []jsonOutput
	for _, result := range results {
		entry := jsonOutput{
			Target: result.Target,
		}
		if result.Error != nil {
			entry.Error = result.Error.Error()
		}
		for _, finding := range result.Findings {
			entry.Findings = append(entry.Findings, jsonFinding{
				TemplateID: finding.TemplateID,
				Name:       finding.Name,
				Severity:   finding.Severity,
				Target:     finding.Target,
				URL:        finding.MatchedURL,
				Method:     finding.Method,
				StatusCode: finding.StatusCode,
				Evidence:   finding.Evidence,
			})
		}
		output = append(output, entry)
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func (ca *CLIApp) writeResultsToFile(results []engine.ScanResult, format string, filePath string) error {
	type jsonFinding struct {
		TemplateID string            `json:"template_id"`
		Name       string            `json:"name"`
		Severity   string            `json:"severity"`
		Target     string            `json:"target"`
		URL        string            `json:"matched_url"`
		Method     string            `json:"method"`
		StatusCode int               `json:"status_code"`
		Evidence   map[string]string `json:"evidence,omitempty"`
	}
	type jsonOutput struct {
		Target   string        `json:"target"`
		Findings []jsonFinding `json:"findings"`
		Error    string        `json:"error,omitempty"`
	}

	var output []jsonOutput
	for _, result := range results {
		entry := jsonOutput{Target: result.Target}
		if result.Error != nil {
			entry.Error = result.Error.Error()
		}
		for _, finding := range result.Findings {
			entry.Findings = append(entry.Findings, jsonFinding{
				TemplateID: finding.TemplateID,
				Name:       finding.Name,
				Severity:   finding.Severity,
				Target:     finding.Target,
				URL:        finding.MatchedURL,
				Method:     finding.Method,
				StatusCode: finding.StatusCode,
				Evidence:   finding.Evidence,
			})
		}
		output = append(output, entry)
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func (ca *CLIApp) runBrute(args []string) error {
	if len(args) < 1 {
		fmt.Println("Usage: wprecon brute <target> [options]")
		fmt.Println("\nOptions:")
		fmt.Println("  --method METHOD       Brute method: login, xmlrpc (default: login)")
		fmt.Println("  --usernames LIST      Comma-separated usernames or @file")
		fmt.Println("  --passwords FILE      Password wordlist file path")
		fmt.Println("  --threads N           Concurrent threads (default: 5)")
		fmt.Println("  --proxy URL           Proxy URL")
		fmt.Println("  --timeout DURATION    Request timeout (default: 10s)")
		fmt.Println("  --insecure            Skip TLS verification")
		fmt.Println("  -v, --verbose         Verbose output")
		return nil
	}

	target := args[0]
	opts := &engine.BruteOptions{
		Target:    target,
		Method:    engine.BruteMethodLogin,
		Threads:   5,
		Timeout:   10 * time.Second,
		VerifyTLS: true,
	}

	var usernamesArg string
	var passwordsFile string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--method":
			if i+1 < len(args) {
				switch args[i+1] {
				case "xmlrpc":
					opts.Method = engine.BruteMethodXMLRPC
				default:
					opts.Method = engine.BruteMethodLogin
				}
				i++
			}
		case "--usernames":
			if i+1 < len(args) {
				usernamesArg = args[i+1]
				i++
			}
		case "--passwords":
			if i+1 < len(args) {
				passwordsFile = args[i+1]
				i++
			}
		case "--threads":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil {
					opts.Threads = n
				}
				i++
			}
		case "--proxy":
			if i+1 < len(args) {
				opts.Proxy = args[i+1]
				i++
			}
		case "--timeout":
			if i+1 < len(args) {
				if d, err := time.ParseDuration(args[i+1]); err == nil {
					opts.Timeout = d
				}
				i++
			}
		case "--insecure":
			opts.VerifyTLS = false
		case "--user-agent":
			if i+1 < len(args) {
				opts.UserAgent = args[i+1]
				i++
			}
		case "-v", "--verbose":
			opts.Verbose = true
		}
	}

	if usernamesArg == "" {
		return fmt.Errorf("--usernames is required")
	}
	if strings.HasPrefix(usernamesArg, "@") {
		data, err := os.ReadFile(usernamesArg[1:])
		if err != nil {
			return fmt.Errorf("failed to read usernames file: %w", err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				opts.Usernames = append(opts.Usernames, line)
			}
		}
	} else {
		opts.Usernames = strings.Split(usernamesArg, ",")
		for j := range opts.Usernames {
			opts.Usernames[j] = strings.TrimSpace(opts.Usernames[j])
		}
	}

	if passwordsFile == "" {
		return fmt.Errorf("--passwords is required")
	}
	data, err := os.ReadFile(passwordsFile)
	if err != nil {
		return fmt.Errorf("failed to read passwords file: %w", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			opts.Passwords = append(opts.Passwords, line)
		}
	}

	if len(opts.Passwords) == 0 {
		return fmt.Errorf("no passwords loaded from %s", passwordsFile)
	}

	fmt.Printf("🔑 Brute-force: %s\n", target)
	fmt.Printf("Method: %s\n", opts.Method)
	fmt.Printf("Usernames: %d\n", len(opts.Usernames))
	fmt.Printf("Passwords: %d\n", len(opts.Passwords))
	fmt.Printf("Threads: %d\n", opts.Threads)
	fmt.Println()

	logger := engine.NewLogger(opts.Verbose)
	bf := engine.NewBruteForcer(opts, logger)
	results, err := bf.Run()
	if err != nil {
		return fmt.Errorf("brute-force failed: %w", err)
	}

	fmt.Println()
	found := 0
	for _, r := range results {
		if r.Success {
			found++
			fmt.Printf("✅ FOUND: %s:%s\n", r.Username, r.Password)
		}
	}

	if found == 0 {
		fmt.Println("❌ No valid credentials found.")
	} else {
		fmt.Printf("\n🔑 Found %d valid credential(s)\n", found)
	}

	return nil
}

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[1;31m"
	colorYellow  = "\033[1;33m"
	colorGreen   = "\033[1;32m"
	colorCyan    = "\033[1;36m"
	colorBlue    = "\033[1;34m"
	colorWhite   = "\033[1;37m"
	colorGray    = "\033[0;90m"
	colorMagenta = "\033[1;35m"
)

func severityOrder(s string) int {
	switch strings.ToLower(s) {
	case "critical":
		return 0
	case "high":
		return 1
	case "medium":
		return 2
	case "low":
		return 3
	case "info":
		return 4
	default:
		return 5
	}
}

func severityColor(s string) string {
	switch strings.ToLower(s) {
	case "critical":
		return colorMagenta
	case "high":
		return colorRed
	case "medium":
		return colorYellow
	case "low":
		return colorGreen
	case "info":
		return colorCyan
	default:
		return colorWhite
	}
}

func severityTag(s string) string {
	upper := strings.ToUpper(s)
	return fmt.Sprintf("%s[%s]%s", severityColor(s), upper, colorReset)
}

func (ca *CLIApp) printTable(results []engine.ScanResult) error {

	type targetResult struct {
		Target   string
		Findings []domain.Finding
		Errors   []error
	}

	targetMap := make(map[string]*targetResult)
	var targetOrder []string

	for _, result := range results {
		tr, exists := targetMap[result.Target]
		if !exists {
			tr = &targetResult{Target: result.Target}
			targetMap[result.Target] = tr
			targetOrder = append(targetOrder, result.Target)
		}
		if result.Error != nil {
			tr.Errors = append(tr.Errors, result.Error)
		}
		tr.Findings = append(tr.Findings, result.Findings...)
	}

	for _, target := range targetOrder {
		tr := targetMap[target]

		for _, err := range tr.Errors {
			fmt.Printf("%s✗%s Error: %v\n", colorRed, colorReset, err)
		}

		if len(tr.Findings) == 0 {
			fmt.Printf("\n%s[*]%s No findings for %s\n", colorGray, colorReset, target)
			continue
		}

		sort.Slice(tr.Findings, func(i, j int) bool {
			return severityOrder(tr.Findings[i].Severity) < severityOrder(tr.Findings[j].Severity)
		})

		counts := map[string]int{}
		for _, f := range tr.Findings {
			counts[strings.ToLower(f.Severity)]++
		}

		fmt.Printf("\n%s┌─ Target: %s%s\n", colorWhite, target, colorReset)

		var parts []string
		for _, sev := range []string{"critical", "high", "medium", "low", "info"} {
			if c, ok := counts[sev]; ok && c > 0 {
				parts = append(parts, fmt.Sprintf("%s%d %s%s", severityColor(sev), c, sev, colorReset))
			}
		}
		fmt.Printf("%s│%s  %s%d findings%s (%s)\n", colorWhite, colorReset, colorWhite, len(tr.Findings), colorReset, strings.Join(parts, ", "))
		fmt.Printf("%s│%s\n", colorWhite, colorReset)

		for i, finding := range tr.Findings {
			prefix := "├──"
			childPrefix := "│  "
			if i == len(tr.Findings)-1 {
				prefix = "└──"
				childPrefix = "   "
			}

			fmt.Printf("%s%s%s %s %s%s%s\n", colorWhite, prefix, colorReset,
				severityTag(finding.Severity), colorWhite, finding.Name, colorReset)

			fmt.Printf("%s%s%s %s%s%s %s[%d]%s\n", colorWhite, childPrefix, colorReset,
				colorGray, finding.MatchedURL, colorReset,
				statusColor(finding.StatusCode), finding.StatusCode, colorReset)

			if len(finding.Evidence) > 0 {
				for k, v := range finding.Evidence {
					displayVal := v
					if len(displayVal) > 120 {
						displayVal = displayVal[:117] + "..."
					}
					fmt.Printf("%s%s%s %s%s:%s %s\n", colorWhite, childPrefix, colorReset,
						colorCyan, k, colorReset, displayVal)
				}
			}
		}
	}

	totalFindings := 0
	for _, tr := range targetMap {
		totalFindings += len(tr.Findings)
	}
	if totalFindings > 0 {
		fmt.Printf("\n%s[✓]%s Scan complete: %s%d%s finding(s) across %d target(s)\n",
			colorGreen, colorReset, colorWhite, totalFindings, colorReset, len(targetOrder))
	} else {
		fmt.Printf("\n%s[✓]%s Scan complete: no findings\n", colorGreen, colorReset)
	}

	return nil
}

func statusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return colorGreen
	case code >= 300 && code < 400:
		return colorYellow
	case code >= 400:
		return colorRed
	default:
		return colorWhite
	}
}

func (ca *CLIApp) listTemplates() error {
	templates, err := ca.scanner.GetTemplates()
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	type category struct {
		name      string
		templates []*domain.Template
	}

	catMap := make(map[string]*category)
	var catOrder []string

	for _, tmpl := range templates {
		catName := extractCategory(tmpl.ID)
		cat, exists := catMap[catName]
		if !exists {
			cat = &category{name: catName}
			catMap[catName] = cat
			catOrder = append(catOrder, catName)
		}
		cat.templates = append(cat.templates, tmpl)
	}

	sort.Strings(catOrder)

	maxIDLen := 0
	for _, tmpl := range templates {
		if len(tmpl.ID) > maxIDLen {
			maxIDLen = len(tmpl.ID)
		}
	}
	if maxIDLen > 45 {
		maxIDLen = 45
	}

	sevCounts := map[string]int{}
	for _, tmpl := range templates {
		sevCounts[strings.ToLower(tmpl.Info.Severity)]++
	}

	fmt.Printf("\n%sAvailable Templates: %d%s", colorWhite, len(templates), colorReset)
	var sevParts []string
	for _, sev := range []string{"critical", "high", "medium", "low", "info"} {
		if c, ok := sevCounts[sev]; ok && c > 0 {
			sevParts = append(sevParts, fmt.Sprintf("%s%d %s%s", severityColor(sev), c, sev, colorReset))
		}
	}
	if len(sevParts) > 0 {
		fmt.Printf(" (%s)", strings.Join(sevParts, ", "))
	}
	fmt.Println()
	fmt.Println(strings.Repeat("─", 80))

	for _, catName := range catOrder {
		cat := catMap[catName]

		sort.Slice(cat.templates, func(i, j int) bool {
			return severityOrder(cat.templates[i].Info.Severity) < severityOrder(cat.templates[j].Info.Severity)
		})

		fmt.Printf("\n%s[%s]%s %s(%d templates)%s\n",
			colorWhite, strings.ToUpper(catName), colorReset,
			colorGray, len(cat.templates), colorReset)

		for _, tmpl := range cat.templates {
			id := tmpl.ID
			if len(id) > maxIDLen {
				id = id[:maxIDLen-2] + ".."
			}

			tags := ""
			if len(tmpl.Info.Tags) > 0 {
				tags = fmt.Sprintf("%s[%s]%s", colorGray, strings.Join(tmpl.Info.Tags, ", "), colorReset)
			}

			fmt.Printf("  %s%-*s%s %s %s\n",
				colorBlue, maxIDLen, id, colorReset,
				severityTag(tmpl.Info.Severity),
				tmpl.Info.Name)

			if tags != "" {
				fmt.Printf("  %s%s\n", strings.Repeat(" ", maxIDLen+2), tags)
			}
		}
	}

	fmt.Println()
	return nil
}

func extractCategory(id string) string {
	prefixes := []string{"common-", "cve-", "wordpress-", "waf-", "exposure-"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(id, prefix) {
			return strings.TrimSuffix(prefix, "-")
		}
	}

	if idx := strings.Index(id, "-"); idx > 0 {
		return id[:idx]
	}
	return "other"
}

func (ca *CLIApp) printHelp() error {
	fmt.Println(`
wprecon - WordPress Reconnaissance Scanner

Usage:
  wprecon <command> [options]

Commands:
  scan                Scan target(s)
  brute               Password brute-force attack
  list-templates      List available templates
  help                Show this help

Scan Options:
  -t, --templates     Comma-separated template IDs (default: all)
  -o, --output        Output format: json, table (default: table)
  --json              Shortcut for --output json
  --output-file FILE  Write results to file (JSON format)

Advanced Scan Options:
  --passive           Passive mode (safe templates only)
  --aggressive        Aggressive mode (all templates)
  --stealthy          Stealthy mode (passive, low rate, browser UA, delays)
  --rate-limit N      Requests per second (default: unlimited)
  --sleep DURATION    Sleep between requests (e.g., 500ms, 1s)
  --concurrency N     Concurrent workers (default: 10)
  --proxy URL         Proxy URL (http://host:port, socks5://host:port)
  --timeout DURATION  Request timeout (e.g., 10s, 5s)
  --user-agent STRING Custom user agent string
  --random-ua         Randomize User-Agent per request (WAF evasion)
  --discovery         Discover plugins/themes from HTML source code
  --no-redirects      Don't follow HTTP redirects
  --insecure          Skip TLS certificate verification
  -v, --verbose       Show detailed output for every request and matcher
  --tags TAG1,TAG2    Run only templates with these tags
  --skip-tags TAG1    Skip templates with these tags

Brute-Force Options:
  wprecon brute <url> --usernames admin --passwords wordlist.txt
  --method            Brute-force method: login, xmlrpc (default: login)
  --usernames         Comma-separated usernames or @file
  --passwords         Password wordlist file path
  --threads N         Concurrent brute-force threads (default: 5)

Examples:
  wprecon scan example.com
  wprecon scan example.com -t wordpress-version,wordpress-plugins
  wprecon scan example.com --passive --rate-limit 5
  wprecon scan example.com --stealthy
  wprecon scan example.com --aggressive --random-ua --output-file results.json
  wprecon brute example.com --usernames admin --passwords rockyou.txt
  wprecon brute example.com --usernames admin --passwords rockyou.txt --method xmlrpc
  wprecon list-templates`)
	return nil
}
