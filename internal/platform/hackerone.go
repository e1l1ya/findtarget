package platform

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/e1l1ya/findtarget/pkg/types"
)

func HackerOne(config *types.Config, env *types.EnvConfig) error {
	baseURL := "https://api.hackerone.com/v1/hackers/programs"

	headers := map[string][]string{
		"Accept": {"application/json"},
	}

	client := &http.Client{}
	var limit int8 = 0
	var has_host bool = false

	for baseURL != "" {
		req, err := http.NewRequest("GET", baseURL, nil)
		if err != nil {
			fmt.Println("Error creating request")
			return err
		}
		req.Header = headers
		req.SetBasicAuth(env.H1Username, env.H1APIKey)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return err
		}

		// Parse JSON response into the struct
		var result types.HackerOneResponse
		err = json.Unmarshal(body, &result)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return err
		}

		// Update the baseURL to the next page URL
		baseURL = result.Links.Next

		// Iterate over each program
		for _, program := range result.Data {

			if config.FindTarget.HackerOne.MaxPrograms != 0 && limit >= config.FindTarget.HackerOne.MaxPrograms {
				return nil
			}

			Handle := program.Attributes.Handle
			programURL := fmt.Sprintf("https://api.hackerone.com/v1/hackers/programs/%s/structured_scopes?page[size]=100", Handle)

			programRequest, err := http.NewRequest("GET", programURL, nil)
			if err != nil {
				fmt.Println("Error creating request")
				return err
			}
			programRequest.Header = headers
			programRequest.SetBasicAuth(env.H1Username, env.H1APIKey)

			programResponse, err := client.Do(programRequest)
			if err != nil {
				fmt.Println("Error sending request:", err)
				return err
			}
			defer programResponse.Body.Close()

			programBody, err := ioutil.ReadAll(programResponse.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				return err
			}

			// Define type of response
			var programResult types.H1ProgramStruct

			// Check if the programResultString contains "URL" or "WILDCARD"
			if !strings.Contains(string(programBody), "URL") && !strings.Contains(string(programBody), "WILDCARD") {
				continue
			}

			err = json.Unmarshal(programBody, &programResult)
			if err != nil {
				fmt.Println("Error parsing JSON:", err)
				return err
			}

			// Print the ID of each H1ScopeData
			for _, scopeData := range programResult.Data {

				if config.FindTarget.HackerOne.Scope == "wide" && strings.ToLower(scopeData.Attributes.AssetType) == "wildcard" {
					if strings.Contains(scopeData.Attributes.AssetIdentifier, ",") {
						hosts := strings.Split(scopeData.Attributes.AssetIdentifier, ",")
						host := strings.Replace(hosts[0], "*.", "", 1)
						fmt.Println(host)
						has_host = true
						continue
					}

					if strings.Contains(scopeData.Attributes.AssetIdentifier, "*") {
						host := strings.Replace(scopeData.Attributes.AssetIdentifier, "*.", "", 1)
						fmt.Println(host)
						has_host = true
					}
				} else if config.FindTarget.HackerOne.Scope == "narrow" && strings.ToLower(scopeData.Attributes.AssetType) == "url" {

					if strings.Contains(scopeData.Attributes.AssetIdentifier, ",") {
						hosts := strings.Split(scopeData.Attributes.AssetIdentifier, ",")
						for _, host := range hosts {
							if !strings.Contains(host, "*") {
								fmt.Println(host)
								has_host = true
							}
						}
						continue
					}

					if !strings.Contains(scopeData.Attributes.AssetIdentifier, "*") {
						fmt.Println(scopeData.Attributes.AssetIdentifier)
						has_host = true
					}
				}
			}
			if config.FindTarget.HackerOne.MaxPrograms != 0 && has_host {
				limit++
				has_host = false
			}
		}
	}
	return nil
}
