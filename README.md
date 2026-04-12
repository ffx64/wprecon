<div align="center">

# WPRecon

**A powerful, template-driven WordPress reconnaissance and vulnerability scanning tool**

[![Go 1.17+](https://img.shields.io/badge/Go-1.17+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-Welcome-brightgreen?style=for-the-badge)](CONTRIBUTING.md)

[Features](#-features) • [Installation](#-installation) • [Usage](#-usage) • [Architecture](#-architecture) • [Templates](#-templates)

</div>

---

## 📋 Overview

**WPRecon** is a modern WordPress security reconnaissance tool built with Go that helps security engineers identify vulnerabilities, misconfigurations, and information disclosure issues in WordPress installations. 

Unlike traditional vulnerability scanners with hardcoded checks, WPRecon uses a **YAML-driven template architecture**, enabling:

- ✅ Easy extensibility without recompilation
- ✅ Team collaboration via shared templates
- ✅ Rapid deployment as a single binary
- ✅ High-performance parallel scanning
- ✅ Flexible matching and extraction capabilities

### Key Statistics

- **Single Binary** - No runtime dependencies
- **150+ Templates** - Comprehensive WordPress coverage
- **Parallel Scanning** - Configurable worker pools
- **Minimal Code** - ~1,200 lines of production code
- **Production Ready** - Battle-tested architecture

---

## ⚡ Features

### Core Capabilities

| Feature | Description |
|---------|-------------|
| **Parallel Scanning** | Configurable worker pool (default: 10 workers) for concurrent requests |
| **Rate Limiting** | Built-in request throttling (default: 50 req/s) to be network-friendly |
| **Auto Retry** | Automatic retry logic for failed HTTP requests |
| **Hot Reload Templates** | No recompilation needed - add templates and scan |
| **Variable Resolution** | Dynamic variables: `{{BaseURL}}`, `{{Timestamp}}`, `{{RandomInt}}` |
| **Multi-Format Output** | Human-readable tables and structured JSON |
| **CLI & API Modes** | Both command-line and REST API interfaces |
| **Severity Levels** | Findings classified as info, low, medium, high, critical |

### HTTP Operations

- All standard HTTP methods (GET, POST, PUT, DELETE)
- Custom headers and cookies
- Request body parameters
- SSL/TLS support
- Configurable timeouts

### Matchers (5 Types)

Identify positive hits in responses:

1. **Status** - HTTP status code matching (200, 404, 500, etc.)
2. **Word** - Substring presence in response body
3. **Regex** - Pattern matching with capture groups
4. **Header** - Response header validation
5. **Size** - Response body size checks

### Extractors (2 Types)

Extract actionable data from matched responses:

1. **Regex** - Extract capture groups using regular expressions
2. **JSON** - Simple JSON path extraction

### Template Library

150+ production-ready templates covering:

- **WordPress Detection** - Core detection and version fingerprinting
- **Plugin Detection** - Popular plugin discovery and version detection
- **User Enumeration** - WordPress user discovery
- **CVE Checks** - Known vulnerability detection
- **Exposures** - Information disclosure (backups, .git, .env, etc.)
- **WAF Detection** - Cloudflare, AWS WAF, ModSecurity, Sucuri, Wordfence
- **Security Analysis** - Misconfiguration detection and security headers
- **Database** - Database exposure and repair detection

---

## 📦 Installation

### Requirements

- Go 1.17 or higher
- Linux, macOS, or Windows

### From Source

```bash
# Clone the repository
git clone https://github.com/ffx64/wprecon.git
cd wprecon

# Build the binary
go build -o wprecon ./cmd/wprecon/main.go

# Move to your PATH (optional)
sudo mv wprecon /usr/local/bin/
```

### Install via Go

```bash
go install github.com/ffx64/wprecon/cmd/wprecon@latest
```

### Docker (Optional)

```bash
docker build -t wprecon .
docker run wprecon scan https://example.com
```

---

## 🚀 Usage

### Command-Line Interface

#### Basic Scan

Scan a WordPress website with all available templates:

```bash
wprecon scan https://example.com
```

#### Scan with Specific Templates

Run only selected templates:

```bash
wprecon scan https://example.com -t wordpress-detection,plugin-detection,user-enumeration
```

#### Custom Output Format

```bash
# JSON output for automation
wprecon scan https://example.com --output json

# Pretty-printed table (default)
wprecon scan https://example.com --output table
```

#### Control Parallelization and Rate Limiting

```bash
# Use 20 concurrent workers and 100 requests/sec
wprecon scan https://example.com --workers 20 --rate-limit 100
```

#### List Available Templates

```bash
wprecon list-templates
```

#### Show Help

```bash
wprecon --help
```

---

## 📋 Example Output

When running a scan, WPRecon produces detailed findings:

**Table Format:**

```
ID                    | Template                | Name                      | Severity | Target              | Evidence
---------------------|------------------------|--------------------------|----------|---------------------|----------
finding_001           | wordpress-version      | WordPress Version Found  | info     | example.com         | 6.2.1
finding_002           | plugin-detection       | Plugin Detected          | low      | example.com         | Yoast SEO 16.0
finding_003           | security-headers       | Missing Security Headers | medium   | example.com         | X-Frame-Options
```

**JSON Format:**

```json
{
  "scan_id": "scan_uuid_001",
  "timestamp": "2026-04-10T14:37:00Z",
  "target": "https://example.com",
  "total_findings": 15,
  "findings": [
    {
      "id": "finding_uuid_001",
      "template_id": "wordpress-detection",
      "name": "WordPress Installation Detected",
      "description": "WordPress CMS installation identified on target",
      "severity": "info",
      "cvss_score": 0.0,
      "matched_url": "https://example.com/wp-admin/",
      "http_method": "GET",
      "status_code": 200,
      "response_time_ms": 245,
      "evidence": {
        "version": "6.2.1",
        "wp_version_header": "6.2.1"
      },
      "timestamp": "2026-04-10T14:37:00Z",
      "remediation": "Keep WordPress updated to the latest version"
    }
  ]
}
```

---

## 🏗️ Architecture

### Project Structure

```
wprecon/
├── cmd/wprecon/
│   └── main.go                    # CLI entrypoint
├── internal/
│   ├── app/
│   │   ├── cli.go                 # CLI command handlers
│   │   └── api.go                 # REST API handlers
│   ├── engine/
│   │   ├── scanner.go             # Scan orchestration
│   │   ├── worker_pool.go         # Parallelization engine
│   │   ├── context.go             # Scan context management
│   │   ├── logger.go              # Logging utilities
│   │   └── rate_limit.go          # Rate limiting
│   ├── executor/
│   │   ├── http_executor.go       # HTTP client
│   │   ├── request_builder.go     # Request construction
│   │   └── variables.go           # Variable resolution
│   ├── matchers/
│   │   └── matchers.go            # Response matching logic
│   ├── extractors/
│   │   └── extractors.go          # Data extraction
│   ├── templates/
│   │   ├── loader.go              # Template discovery
│   │   └── parser.go              # YAML parsing
│   ├── pipeline/
│   │   └── pipeline.go            # Scan workflow
│   └── domain/
│       ├── template.go            # Data models
│       ├── finding.go
│       ├── http.go
│       ├── matcher.go
│       └── extractor.go
├── templates/                      # YAML template library
│   ├── wordpress/                  # WordPress-specific
│   ├── cves/                       # CVE detection
│   ├── exposures/                  # Info disclosure
│   ├── common/                     # Generic checks
│   └── waf/                        # WAF detection
└── go.mod                          # Dependencies
```

### Architecture Diagram

```
┌─────────────────────────────────────────────┐
│         User Interface (CLI/API)            │
└─────────────────────────────────────────────┘
          │
          ├─→ Template Loader
          ├─→ Rate Limiter
          └─→ Worker Pool Manager
          │
          ▼
┌─────────────────────────────────────────────┐
│          Scan Pipeline Engine               │
│  ┌────────→ Request Builder                 │
│  │ ┌────────→ HTTP Executor                 │
│  │ │ ┌────────→ Matcher (5 types)           │
│  │ │ │ ┌────────→ Extractor (2 types)      │
│  │ │ │ │ ┌────────→ Finding Report         │
│  │ │ │ │ │                                 │
│  ▼ ▼ ▼ ▼ ▼                                 │
│ Template → Request → Response → Finding    │
└─────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────┐
│       Output Formatters & Storage           │
│    (JSON, Table, Report Generation)         │
└─────────────────────────────────────────────┘
```

### Design Principles

1. **Single Responsibility** - Each module has clear, focused purpose
2. **Template-Driven** - All scanning logic lives in YAML, not code
3. **Horizontal Scalability** - Modular design allows easy extension
4. **Factory Pattern** - Simple addition of new matchers and extractors
5. **No External Dependencies** - Only UUID and YAML parsing libraries
6. **Thread-Safe** - Safe concurrent scanning with worker pools

---

## 📚 Templates

Templates are YAML files that define how WPRecon should scan for specific vulnerabilities or information.

### Template Structure

```yaml
id: wordpress-version-detection
name: WordPress Version Detection
description: Detects and extracts WordPress version from readme.html
severity: info
author: WPRecon Team

requests:
  - url: "{{BaseURL}}/readme.html"
    method: GET
    
    matchers:
      - type: status
        status:
          - 200
      
      - type: word
        words:
          - "WordPress"
    
    extractors:
      - type: regex
        regex:
          - 'Version (\d+\.\d+\.\d+)'
```

### Template Categories

#### WordPress Core

| Template | Purpose |
|----------|---------|
| `wordpress-detection` | Detect WordPress installation |
| `wordpress-version` | Fingerprint WordPress version |
| `plugins.yaml` | Detect installed plugins |
| `themes.yaml` | Detect installed themes |
| `users.yaml` | Enumerate WordPress users |

#### Vulnerabilities

- CVE detection for WordPress plugins
- Known vulnerability checks
- Security misconfiguration detection

#### Information Disclosure

- Backup file exposure (wp-config.php.bak, .sql)
- Git repository exposure (.git/)
- Environment file exposure (.env)
- S3 bucket misconfiguration

#### WAF Detection

- Cloudflare
- AWS WAF
- ModSecurity
- Sucuri
- Wordfence
- Generic WAF detection

#### Security Analysis

- Missing security headers
- Weak authentication
- CORS misconfiguration
- Debug mode detection

### Adding Custom Templates

1. Create a new YAML file in `templates/custom/`
2. Follow the template structure (see example above)
3. Run WPRecon - templates are auto-discovered

Example:

```bash
# Create custom template
cat > templates/custom/my-check.yaml << 'EOF'
id: my-custom-check
name: My Custom Check
description: Detects custom vulnerability
severity: medium
author: Your Name

requests:
  - url: "{{BaseURL}}/vulnerable-endpoint"
    method: GET
    
    matchers:
      - type: status
        status:
          - 200
      
      - type: word
        words:
          - "vulnerable"
EOF

# Scan with custom template
wprecon scan https://example.com -t my-custom-check
```

---

## 🔧 Configuration

### Environment Variables

```bash
# HTTP timeout (default: 10s)
export WPRECON_TIMEOUT=15

# Number of workers (default: 10)
export WPRECON_WORKERS=20

# Requests per second (default: 50)
export WPRECON_RATE_LIMIT=100

# Log level (default: info)
export WPRECON_LOG_LEVEL=debug

# Template directories (default: ./templates)
export WPRECON_TEMPLATE_DIRS=/path/to/templates:/path/to/more/templates
```

### CLI Flags

```bash
wprecon scan <target> \
  --templates-dir ./custom-templates \
  --workers 20 \
  --rate-limit 100 \
  --timeout 15 \
  --output json \
  --log-level debug
```

---

## 💼 Use Cases

### Security Assessments

Comprehensive security evaluation of WordPress installations:

```bash
wprecon scan https://example.com --output json > assessment_report.json
```

### CI/CD Integration

Automated vulnerability scanning in deployment pipelines:

```yaml
# GitHub Actions example
- name: WordPress Security Scan
  run: |
    wprecon scan ${{ secrets.STAGING_URL }} \
      --output json \
      --templates wordpress-detection,plugin-detection,cve-checks
```

### Mass Reconnaissance

Scan multiple targets efficiently:

```bash
cat targets.txt | while read target; do
  wprecon scan "$target" --output json >> results.json
done
```

---

## 🛠️ Development

### Prerequisites

- Go 1.17+

### Building

```bash
# Build binary
go build -o wprecon ./cmd/wprecon/main.go

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o wprecon-linux ./cmd/wprecon/main.go
GOOS=darwin GOARCH=amd64 go build -o wprecon-darwin ./cmd/wprecon/main.go
GOOS=windows GOARCH=amd64 go build -o wprecon-windows.exe ./cmd/wprecon/main.go
```

### Project Dependencies

```text
google/uuid v1.6.0   - UUID generation for findings
gopkg.in/yaml.v3     - YAML template parsing
```

### Code Structure

- **Minimal Dependencies** - Only essential libraries
- **Clean Architecture** - Layered design for maintainability
- **Extensible** - Easy to add matchers, extractors, and templates
- **Well-Documented** - Clear code comments and documentation

---

## 🤝 Contributing Templates

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-template`)
3. Add your template to `templates/`
4. Test thoroughly: `wprecon scan https://test.site -t your-template`
5. Submit a Pull Request

---

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with ❤️ for Matheus (ffx64)**

[⬆ back to top](#-wprecon)

</div>
