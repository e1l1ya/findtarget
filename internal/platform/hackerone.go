package platform

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/e1l1ya/findtarget/pkg/types"
)

const hackerOneBaseURL = "https://api.hackerone.com/v1/hackers/programs"

// HackerOne fetches data from the HackerOne API and processes it based on the configuration.
func HackerOne(config *types.Config, env *types.EnvConfig) error {
	client := &http.Client{}
	headers := map[string][]string{
		"Accept": {"application/json"},
	}

	// Check if the config has an "Include" array
	if len(config.FindTarget.HackerOne.Include) > 0 {
		// Define a regex to extract the handle from the URL
		handleRegex := regexp.MustCompile(`https://hackerone\.com/([^?]+)`)

		for _, includeURL := range config.FindTarget.HackerOne.Include {
			// Extract the handle using the regex
			matches := handleRegex.FindStringSubmatch(includeURL)
			if len(matches) < 2 {
				fmt.Printf("Invalid HackerOne URL: %s\n", includeURL)
				continue
			}
			handle := matches[1]

			fmt.Println(handle)

			// Process the program using the extracted handle
			hasHost, err := processHackerOneProgram(client, handle, headers, env, config)
			if err != nil {
				return err
			}

			// If MaxPrograms is set, limit the number of processed programs
			if config.FindTarget.HackerOne.MaxPrograms != 0 && hasHost {
				return nil
			}
		}
		return nil // Skip the rest of the logic if "Include" is used
	}

	baseURL := hackerOneBaseURL
	var limit int8 = 0

	for baseURL != "" {
		// Fetch programs from HackerOne
		result, nextURL, err := fetchHackerOnePrograms(client, baseURL, headers, env)
		if err != nil {
			return err
		}
		baseURL = nextURL

		// Process each program
		for _, program := range result.Data {
			if config.FindTarget.HackerOne.MaxPrograms != 0 && limit >= config.FindTarget.HackerOne.MaxPrograms {
				return nil
			}

			hasHost, err := processHackerOneProgram(client, program.Attributes.Handle, headers, env, config)
			if err != nil {
				return err
			}

			if config.FindTarget.HackerOne.MaxPrograms != 0 && hasHost {
				limit++
			}
		}
	}

	return nil
}

// fetchHackerOnePrograms fetches a list of programs from HackerOne.
func fetchHackerOnePrograms(client *http.Client, baseURL string, headers map[string][]string, env *types.EnvConfig) (types.HackerOneResponse, string, error) {
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return types.HackerOneResponse{}, "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header = headers
	req.SetBasicAuth(env.H1Username, env.H1APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return types.HackerOneResponse{}, "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.HackerOneResponse{}, "", fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.HackerOneResponse{}, "", fmt.Errorf("error reading response body: %v", err)
	}

	var result types.HackerOneResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return types.HackerOneResponse{}, "", fmt.Errorf("error parsing JSON: %v", err)
	}

	return result, result.Links.Next, nil
}

// processHackerOneProgram processes a single HackerOne program and its scopes.
func processHackerOneProgram(client *http.Client, handle string, headers map[string][]string, env *types.EnvConfig, config *types.Config) (bool, error) {
	programURL := fmt.Sprintf("%s/%s/structured_scopes?page[size]=100", hackerOneBaseURL, handle)

	req, err := http.NewRequest("GET", programURL, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %v", err)
	}
	req.Header = headers
	req.SetBasicAuth(env.H1Username, env.H1APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("error reading response body: %v", err)
	}

	// Check if the response contains "URL" or "WILDCARD"
	if !strings.Contains(string(body), "URL") && !strings.Contains(string(body), "WILDCARD") {
		return false, nil
	}

	var programResult types.H1ProgramStruct
	if err := json.Unmarshal(body, &programResult); err != nil {
		return false, fmt.Errorf("error parsing JSON: %v", err)
	}

	return processScopes(programResult.Data, config), nil
}

// processScopes processes the scopes of a HackerOne program.
func processScopes(scopes []types.H1ScopeData, config *types.Config) bool {
	hasHost := false

	for _, scopeData := range scopes {
		assetType := strings.ToLower(scopeData.Attributes.AssetType)
		assetIdentifier := scopeData.Attributes.AssetIdentifier

		if config.FindTarget.HackerOne.Scope == "wide" && assetType == "wildcard" {
			hasHost = processWildcardScope(assetIdentifier) || hasHost
		} else if config.FindTarget.HackerOne.Scope == "narrow" && assetType == "url" {
			hasHost = processURLScope(assetIdentifier) || hasHost
		} else if config.FindTarget.HackerOne.Scope == "all" {
			hasHost = processWildcardScope(assetIdentifier) || hasHost
		}
	}

	return hasHost
}

// processWildcardScope processes a wildcard scope and prints the host.
func processWildcardScope(assetIdentifier string) bool {
	if strings.Contains(assetIdentifier, ",") {
		hosts := strings.Split(assetIdentifier, ",")
		host := strings.Replace(hosts[0], "*.", "", 1)
		fmt.Println(host)
		return true
	}

	if strings.Contains(assetIdentifier, "*") {
		host := strings.Replace(assetIdentifier, "*.", "", 1)
		fmt.Println(host)
		return true
	}

	return false
}

// processURLScope processes a URL scope and prints the host.
func processURLScope(assetIdentifier string) bool {
	if strings.Contains(assetIdentifier, ",") {
		hosts := strings.Split(assetIdentifier, ",")
		for _, host := range hosts {
			if !strings.Contains(host, "*") {
				fmt.Println(host)
			}
		}
		return true
	}

	if !strings.Contains(assetIdentifier, "*") {
		fmt.Println(assetIdentifier)
		return true
	}

	return false
}
