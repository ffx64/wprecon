package enumerators

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

type UserEnumerator struct {
	client  *http.Client
	timeout time.Duration
}

func NewUserEnumerator(timeout time.Duration) *UserEnumerator {
	return &UserEnumerator{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

func (ue *UserEnumerator) EnumerateViaRestAPI(target string) []UserInfo {
	users := []UserInfo{}
	url := fmt.Sprintf("%s/wp-json/wp/v2/users?per_page=100", target)

	resp, err := ue.client.Get(url)
	if err != nil {
		return users
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return users
	}

	body, _ := io.ReadAll(resp.Body)

	var result []map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return users
	}

	for _, user := range result {
		info := UserInfo{
			Method: "rest-api",
		}

		if id, ok := user["id"]; ok {
			info.ID = int(id.(float64))
		}
		if username, ok := user["username"]; ok {
			info.Username = username.(string)
		}
		if name, ok := user["name"]; ok {
			if nameStr, ok := name.(string); ok && nameStr != "" {
				info.Username = nameStr
			}
		}

		users = append(users, info)
	}

	return users
}

func (ue *UserEnumerator) EnumerateViaAuthorPages(target string, maxID int) []UserInfo {
	users := []UserInfo{}

	authorRegex := regexp.MustCompile(`by\s+<a[^>]*>([^<]+)</a>|wp-author-([a-z0-9-]+)|author=([a-z0-9-]+)`)

	for i := 1; i <= maxID; i++ {
		url := fmt.Sprintf("%s/?author=%d", target, i)

		resp, err := ue.client.Get(url)
		if err != nil {
			continue
		}

		if resp.StatusCode == 404 || resp.StatusCode == 301 || resp.StatusCode == 302 {
			resp.Body.Close()
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		matches := authorRegex.FindAllStringSubmatch(string(body), -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				users = append(users, UserInfo{
					Username: match[1],
					ID:       i,
					Method:   "author-pages",
				})
				break
			}
		}
	}

	return users
}

func (ue *UserEnumerator) EnumerateViaHTML(target string) []UserInfo {
	users := []UserInfo{}

	resp, err := ue.client.Get(target)
	if err != nil {
		return users
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	patterns := []string{
		`by\s+<a[^>]*>([^<]+)</a>`,
		`href=['"]/author/([a-z0-9-]+)['"]\s*>([^<]+)</a>`,
		`author=['"']?([a-z0-9-]+)['"']`,
		`(?:Posted|Written)\s+by\s+([a-zA-Z0-9_\s]+)[<\n]`,
	}

	foundAuthors := make(map[string]bool)

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(string(body), -1)

		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				author := match[1]
				if !foundAuthors[author] {
					users = append(users, UserInfo{
						Username: author,
						Method:   "html",
					})
					foundAuthors[author] = true
				}
			}
		}
	}

	return users
}

func (ue *UserEnumerator) EnumerateViaRSS(target string) []UserInfo {
	users := []UserInfo{}

	feedPaths := []string{"/feed/", "/index.php/feed/", "/rss/"}
	creatorRegex := regexp.MustCompile(`<creator>([^<]+)</creator>`)
	authorRegex := regexp.MustCompile(`<author>([^<]+)</author>`)

	foundAuthors := make(map[string]bool)

	for _, path := range feedPaths {
		url := fmt.Sprintf("%s%s", target, path)

		resp, err := ue.client.Get(url)
		if err != nil {
			continue
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		matches := creatorRegex.FindAllStringSubmatch(string(body), -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				author := match[1]
				if !foundAuthors[author] {
					users = append(users, UserInfo{
						Username: author,
						Method:   "rss",
					})
					foundAuthors[author] = true
				}
			}
		}

		matches = authorRegex.FindAllStringSubmatch(string(body), -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				author := match[1]
				if !foundAuthors[author] {
					users = append(users, UserInfo{
						Username: author,
						Method:   "rss",
					})
					foundAuthors[author] = true
				}
			}
		}
	}

	return users
}

func (ue *UserEnumerator) EnumerateAll(target string, maxAuthorID int) []UserInfo {
	allUsers := make(map[string]UserInfo)

	restUsers := ue.EnumerateViaRestAPI(target)
	for _, user := range restUsers {
		key := user.Username
		allUsers[key] = user
	}

	authorUsers := ue.EnumerateViaAuthorPages(target, maxAuthorID)
	for _, user := range authorUsers {
		key := user.Username
		if existing, ok := allUsers[key]; ok {
			if user.ID != 0 {
				existing.ID = user.ID
			}
			allUsers[key] = existing
		} else {
			allUsers[key] = user
		}
	}

	htmlUsers := ue.EnumerateViaHTML(target)
	for _, user := range htmlUsers {
		key := user.Username
		if _, ok := allUsers[key]; !ok {
			allUsers[key] = user
		}
	}

	rssUsers := ue.EnumerateViaRSS(target)
	for _, user := range rssUsers {
		key := user.Username
		if _, ok := allUsers[key]; !ok {
			allUsers[key] = user
		}
	}

	result := make([]UserInfo, 0, len(allUsers))
	for _, user := range allUsers {
		result = append(result, user)
	}

	return result
}
