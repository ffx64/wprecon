package enumerators

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

type VersionEnumerator struct {
	client  *http.Client
	timeout time.Duration
}

func NewVersionEnumerator(timeout time.Duration) *VersionEnumerator {
	return &VersionEnumerator{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

func (ve *VersionEnumerator) DetectWordPressVersion(target string) string {
	version := ve.getVersionFromReadme(target)
	if version != "" {
		return version
	}

	version = ve.getVersionFromVersionFile(target)
	if version != "" {
		return version
	}

	version = ve.getVersionFromHeader(target)
	return version
}

func (ve *VersionEnumerator) getVersionFromReadme(target string) string {
	url := fmt.Sprintf("%s/readme.html", target)

	resp, err := ve.client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	regex := regexp.MustCompile(`Version\s+([0-9.]+)`)
	matches := regex.FindStringSubmatch(string(body))

	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (ve *VersionEnumerator) getVersionFromVersionFile(target string) string {
	url := fmt.Sprintf("%s/wp-includes/version.php", target)

	resp, err := ve.client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	regex := regexp.MustCompile(`\$wp_version\s*=\s*['"']([0-9.]+)['"']`)
	matches := regex.FindStringSubmatch(string(body))

	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (ve *VersionEnumerator) getVersionFromHeader(target string) string {
	resp, err := ve.client.Get(target)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if poweredBy := resp.Header.Get("X-Powered-By"); poweredBy != "" {
		regex := regexp.MustCompile(`WordPress[/\s]+([0-9.]+)`)
		matches := regex.FindStringSubmatch(poweredBy)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

func (ve *VersionEnumerator) DetectPluginVersion(target, pluginName string) string {
	// Tenta readme.txt
	version := ve.getPluginVersionFromReadme(target, pluginName)
	if version != "" {
		return version
	}

	version = ve.getPluginVersionFromMain(target, pluginName)
	return version
}

func (ve *VersionEnumerator) getPluginVersionFromReadme(target, pluginName string) string {
	url := fmt.Sprintf("%s/wp-content/plugins/%s/readme.txt", target, pluginName)

	resp, err := ve.client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	regex := regexp.MustCompile(`Stable\s+tag:\s+([0-9.]+)`)
	matches := regex.FindStringSubmatch(string(body))

	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (ve *VersionEnumerator) getPluginVersionFromMain(target, pluginName string) string {
	url := fmt.Sprintf("%s/wp-content/plugins/%s/%s.php", target, pluginName, pluginName)

	resp, err := ve.client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	regex := regexp.MustCompile(`Version:\s+([0-9.]+)`)
	matches := regex.FindStringSubmatch(string(body))

	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (ve *VersionEnumerator) DetectThemeVersion(target, themeName string) string {
	url := fmt.Sprintf("%s/wp-content/themes/%s/readme.txt", target, themeName)

	resp, err := ve.client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	regex := regexp.MustCompile(`Stable\s+tag:\s+([0-9.]+)`)
	matches := regex.FindStringSubmatch(string(body))

	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (ve *VersionEnumerator) ExtractVersionsFromChangelog(target, pluginName string) []string {
	versions := []string{}

	url := fmt.Sprintf("%s/wp-content/plugins/%s/CHANGELOG.md", target, pluginName)

	resp, err := ve.client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return versions
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	patterns := []string{
		`##\s+([0-9.]+)`,
		`-\s+([0-9.]+)`,
		`\(([0-9.]+)\)`,
	}

	foundVersions := make(map[string]bool)

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(string(body), -1)

		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				version := match[1]
				if !foundVersions[version] {
					versions = append(versions, version)
					foundVersions[version] = true
				}
			}
		}
	}

	return versions
}
