package platform

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"

	"github.com/e1l1ya/findtarget/pkg/types"

	"golang.org/x/net/proxy"

	"golang.org/x/net/html"
)

// isURL checks if a given string is a valid URL.
func isURL(str string) bool {
	parsedURL, err := url.Parse(str)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}

// fetchBriefVersionDocument fetches a program page and extracts the engagementBriefApi.getBriefVersionDocument value.
func fetchBriefVersionDocument(client *http.Client, programURL string) (string, error) {
	// Make the request to fetch the HTML content
	resp, err := client.Get(programURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch program page: %v", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response from program page: %d", resp.StatusCode)
	}

	// Parse the HTML and extract the data-api-endpoints value
	tokenizer := html.NewTokenizer(resp.Body)

	inTargetContainer := false
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			// End of document
			return "", fmt.Errorf("data-api-endpoints attribute not found")

		case html.StartTagToken:
			token := tokenizer.Token()

			// Check if we found the container with class "rp-researcher-engagement-brief__grid-container"
			if token.Data == "div" {
				for _, attr := range token.Attr {
					if attr.Key == "class" && strings.Contains(attr.Val, "rp-researcher-engagement-brief__grid-container") {
						inTargetContainer = true
						break
					}
				}
			}

			// If we're inside the container, find the first div with data-api-endpoints
			if inTargetContainer && token.Data == "div" {
				for _, attr := range token.Attr {
					if attr.Key == "data-api-endpoints" {
						// Extract and parse JSON from the attribute
						var apiEndpoints map[string]map[string]string
						if err := json.Unmarshal([]byte(attr.Val), &apiEndpoints); err != nil {
							return "", fmt.Errorf("failed to parse data-api-endpoints JSON: %v", err)
						}

						// Retrieve the engagementBriefApi.getBriefVersionDocument value
						if engagementBrief, exists := apiEndpoints["engagementBriefApi"]; exists {
							if briefDoc, found := engagementBrief["getBriefVersionDocument"]; found {
								return "https://bugcrowd.com" + briefDoc, nil
							}
						}

						return "", fmt.Errorf("getBriefVersionDocument not found in API endpoints")
					}
				}
			}
		}
	}
}

// fetchScopeItems fetches and parses the brief version document JSON to extract scope items.
func fetchScopeItems(client *http.Client, briefDocumentURL string) ([]types.ScopeItem, error) {
	// Make the request to fetch the JSON content
	resp, err := client.Get(briefDocumentURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch brief version document: %v", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response from brief version document: %d", resp.StatusCode)
	}

	// Decode the JSON response
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

// FetchBugcrowdData fetches Bugcrowd data, extracts engagement details, and fetches additional API endpoints.
func Bugcrowd(config *types.Config) error {

	baseURL := "https://bugcrowd.com/engagements.json?&page=%d"

	if config.FindTarget.BugCrowd.Category != "" {
		baseURL += fmt.Sprintf("&target_categories=%s", config.FindTarget.BugCrowd.Category)
	}

	if config.FindTarget.BugCrowd.Reward != "" && config.FindTarget.BugCrowd.Reward == "points" {
		baseURL += "&category=vdp"
	} else if config.FindTarget.BugCrowd.Reward != "" {
		baseURL += fmt.Sprintf("&category=bug_bounty&rewards_operator=gte&rewards_amount=%s", config.FindTarget.BugCrowd.Reward)
	}

	// Create an HTTP client with optional SOCKS5 proxy support
	var client *http.Client
	if config.Proxy != "" {
		proxyURL, err := url.Parse(config.Proxy)
		if err != nil {
			return fmt.Errorf("invalid proxy URL: %v", err)
		}

		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, nil, proxy.Direct)
		if err != nil {
			return fmt.Errorf("failed to create SOCKS5 proxy dialer: %v", err)
		}

		transport := &http.Transport{Dial: dialer.Dial}
		client = &http.Client{Transport: transport}
	} else {
		client = &http.Client{}
	}

	// Fetch the first page to determine total pages
	page := 1
	totalPages := 1 // Default to 1; will be updated after the first request
	var limit int16 = 0
	var has_host bool = false

	for page <= totalPages {
		// Construct the paginated URL
		url := fmt.Sprintf(baseURL, page)

		// Make the request
		resp, err := client.Get(url)
		if err != nil {
			return fmt.Errorf("failed to fetch data from Bugcrowd: %v", err)
		}
		defer resp.Body.Close()

		// Check for successful response
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected response from Bugcrowd: %d", resp.StatusCode)
		}

		// Decode the JSON response
		var data types.APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return fmt.Errorf("failed to parse Bugcrowd JSON: %v", err)
		}

		// Set total pages after fetching the first response
		if page == 1 && data.PaginationMeta.Limit > 0 {
			totalPages = int(math.Ceil(float64(data.PaginationMeta.TotalCount) / float64(data.PaginationMeta.Limit)))
		}

		// Iterate over engagements and fetch additional data from their pages
		for _, engagement := range data.Engagements {
			programURL := "https://bugcrowd.com" + engagement.BriefURL
			briefVersionDocument, err := fetchBriefVersionDocument(client, programURL)

			if config.FindTarget.BugCrowd.MaxPrograms != 0 && limit >= config.FindTarget.BugCrowd.MaxPrograms {
				return nil
			}

			if err != nil {
				continue
			}

			// Fetch and parse the brief document JSON
			scopeItems, err := fetchScopeItems(client, briefVersionDocument+".json")
			if err != nil {
				continue
			}

			// Print scope target names
			for _, item := range scopeItems {
				for _, target := range item.Targets {
					if config.FindTarget.BugCrowd.Category != "" && config.FindTarget.BugCrowd.Category == target.Category {
						var name string = extractURL(target.Name)
						var uri string = extractURL(target.URI)
						if config.FindTarget.BugCrowd.Scope != "all" {
							if isURL(name) && config.FindTarget.BugCrowd.Scope == "narrow" && !strings.HasPrefix(name, "*") {
								fmt.Println(name)
								has_host = true
							} else if isURL(uri) && config.FindTarget.BugCrowd.Scope == "narrow" && !strings.HasPrefix(name, "*") {
								fmt.Println(uri)
								has_host = true
							} else if strings.HasPrefix(name, "*.") && config.FindTarget.BugCrowd.Scope == "wide" {
								widescope := strings.TrimPrefix(name, "*.")
								if !strings.Contains(widescope, "*") {
									fmt.Println(widescope)
									has_host = true
								}
							}
						} else {
							if isURL(name) && !strings.HasPrefix(name, "*") {
								fmt.Println(name)
								has_host = true
							} else if isURL(uri) && !strings.HasPrefix(name, "*") {
								fmt.Println(uri)
								has_host = true
							} else if strings.HasPrefix(name, "*.") {
								widescope := strings.TrimPrefix(name, "*.")
								if !strings.Contains(widescope, "*") {
									fmt.Println(widescope)
									has_host = true
								}
							}
						}

					}
				}
			}
			if config.FindTarget.BugCrowd.MaxPrograms != 0 && has_host {
				limit++
				has_host = false
			}
		}

		// Move to the next page
		page++
	}

	return nil
}
