package platform

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"

	"github.com/e1l1ya/findtarget/pkg/types"
	"golang.org/x/net/html"
	"golang.org/x/net/proxy"
)

const bugcrowdBaseURL = "https://bugcrowd.com"

// isURL checks if a given string is a valid URL.
func isURL(str string) bool {
	parsedURL, err := url.Parse(str)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}

// constructBaseURL builds the base URL with query parameters.
func constructBaseURL(config *types.Config) string {
	baseURL := bugcrowdBaseURL + "/engagements.json?&page=%d"

	if config.FindTarget.BugCrowd.Category != "" {
		baseURL += fmt.Sprintf("&target_categories=%s", config.FindTarget.BugCrowd.Category)
	}

	if config.FindTarget.BugCrowd.Reward != "" {
		if config.FindTarget.BugCrowd.Reward == "points" {
			baseURL += "&category=vdp"
		} else {
			baseURL += fmt.Sprintf("&category=bug_bounty&rewards_operator=gte&rewards_amount=%s", config.FindTarget.BugCrowd.Reward)
		}
	}

	return baseURL
}

// createHTTPClient creates an HTTP client with optional SOCKS5 proxy support.
func createHTTPClient(proxyURL string) (*http.Client, error) {
	if proxyURL == "" {
		return &http.Client{}, nil
	}

	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %v", err)
	}

	dialer, err := proxy.SOCKS5("tcp", parsedProxyURL.Host, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 proxy dialer: %v", err)
	}

	transport := &http.Transport{Dial: dialer.Dial}
	return &http.Client{Transport: transport}, nil
}

// fetchBriefVersionDocument fetches the brief version document URL from a program page.
func fetchBriefVersionDocument(client *http.Client, programURL string) (string, error) {
	resp, err := client.Get(programURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch program page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response from program page: %d", resp.StatusCode)
	}

	tokenizer := html.NewTokenizer(resp.Body)
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return "", fmt.Errorf("data-api-endpoints attribute not found")
		case html.StartTagToken:
			token := tokenizer.Token()
			if token.Data == "div" {
				for _, attr := range token.Attr {
					if attr.Key == "data-api-endpoints" {
						var apiEndpoints map[string]map[string]string
						if err := json.Unmarshal([]byte(attr.Val), &apiEndpoints); err != nil {
							return "", fmt.Errorf("failed to parse data-api-endpoints JSON: %v", err)
						}
						if engagementBrief, exists := apiEndpoints["engagementBriefApi"]; exists {
							if briefDoc, found := engagementBrief["getBriefVersionDocument"]; found {
								return bugcrowdBaseURL + briefDoc, nil
							}
						}
						return "", fmt.Errorf("getBriefVersionDocument not found in API endpoints")
					}
				}
			}
		}
	}
}

// fetchScopeItems fetches and parses the scope items from the brief document JSON.
func fetchScopeItems(client *http.Client, briefDocumentURL string) ([]types.ScopeItem, error) {
	resp, err := client.Get(briefDocumentURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch brief version document: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response from brief version document: %d", resp.StatusCode)
	}

	var briefData struct {
		Data struct {
			Scope []types.ScopeItem `json:"scope"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&briefData); err != nil {
		return nil, fmt.Errorf("failed to parse brief document JSON: %v", err)
	}

	return briefData.Data.Scope, nil
}

func extractURL(text string) string {
	parts := strings.Split(text, " ")
	if len(parts) > 0 {
		return parts[0]
	}
	return text
}

// processTarget processes and prints target names based on the configuration.
func processTarget(config *types.Config, target types.Target) bool {
	name := extractURL(target.Name)
	uri := extractURL(target.URI)

	if config.FindTarget.BugCrowd.Scope == "narrow" {
		if isURL(name) && !strings.HasPrefix(name, "*") {
			fmt.Println(name)
			return true
		} else if isURL(uri) && !strings.HasPrefix(name, "*") {
			fmt.Println(uri)
			return true
		}
	} else if config.FindTarget.BugCrowd.Scope == "wide" && strings.HasPrefix(name, "*.") {
		widescope := strings.TrimPrefix(name, "*.")
		if !strings.Contains(widescope, "*") {
			fmt.Println(widescope)
			return true
		}
	} else if config.FindTarget.BugCrowd.Scope == "all" {
		if isURL(name) && !strings.HasPrefix(name, "*") {
			fmt.Println(name)
			return true
		} else if isURL(uri) && !strings.HasPrefix(name, "*") {
			fmt.Println(uri)
			return true
		} else if strings.HasPrefix(name, "*.") {
			widescope := strings.TrimPrefix(name, "*.")
			if !strings.Contains(widescope, "*") {
				fmt.Println(widescope)
				return true
			}
		}
	}

	return false
}

// Bugcrowd fetches Bugcrowd data, extracts engagement details, and fetches additional API endpoints.
func Bugcrowd(config *types.Config) error {
	baseURL := constructBaseURL(config)

	client, err := createHTTPClient(config.Proxy)
	if err != nil {
		return err
	}

	// Check if the config has an "Include" array
	if len(config.FindTarget.BugCrowd.Include) > 0 {
		for _, includeURL := range config.FindTarget.BugCrowd.Include {
			// Fetch the brief version document for each URL in the "Include" array
			briefVersionDocument, err := fetchBriefVersionDocument(client, includeURL)
			if err != nil {
				fmt.Printf("Failed to fetch brief version document for %s: %v\n", includeURL, err)
				continue
			}

			// Fetch and process scope items
			scopeItems, err := fetchScopeItems(client, briefVersionDocument+".json")
			if err != nil {
				fmt.Printf("Failed to fetch scope items for %s: %v\n", includeURL, err)
				continue
			}

			for _, item := range scopeItems {
				for _, target := range item.Targets {
					if config.FindTarget.BugCrowd.Category == "" || config.FindTarget.BugCrowd.Category == target.Category {
						processTarget(config, target)
					}
				}
			}
		}
		return nil // Skip the paginated section if "Include" is used
	}

	// Paginated section
	page := 1
	totalPages := 1
	limit := int16(0)

	for page <= totalPages {
		url := fmt.Sprintf(baseURL, page)

		resp, err := client.Get(url)
		if err != nil {
			return fmt.Errorf("failed to fetch data from Bugcrowd: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected response from Bugcrowd: %d", resp.StatusCode)
		}

		var data types.APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return fmt.Errorf("failed to parse Bugcrowd JSON: %v", err)
		}

		if page == 1 && data.PaginationMeta.Limit > 0 {
			totalPages = int(math.Ceil(float64(data.PaginationMeta.TotalCount) / float64(data.PaginationMeta.Limit)))
		}

		for _, engagement := range data.Engagements {
			programURL := bugcrowdBaseURL + engagement.BriefURL
			briefVersionDocument, err := fetchBriefVersionDocument(client, programURL)

			if config.FindTarget.BugCrowd.MaxPrograms != 0 && limit >= config.FindTarget.BugCrowd.MaxPrograms {
				return nil
			}

			if err != nil {
				continue
			}

			scopeItems, err := fetchScopeItems(client, briefVersionDocument+".json")
			if err != nil {
				continue
			}

			for _, item := range scopeItems {
				for _, target := range item.Targets {
					if config.FindTarget.BugCrowd.Category == "" || config.FindTarget.BugCrowd.Category == target.Category {
						if processTarget(config, target) {
							limit++
							if config.FindTarget.BugCrowd.MaxPrograms != 0 && limit >= config.FindTarget.BugCrowd.MaxPrograms {
								return nil
							}
						}
					}
				}
			}
		}

		page++
	}

	return nil
}
