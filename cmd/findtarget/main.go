package main

import (
	"fmt"
	"log"

	"github.com/e1l1ya/findtarget/internal/platform"
	"github.com/e1l1ya/findtarget/internal/runner"
	"github.com/e1l1ya/findtarget/pkg/types"
	"github.com/spf13/pflag"
)

// parseFlags parses the flags and sets defaults if necessary.
func parseFlags() (*types.Config, bool) {
	config := &types.Config{}
	pflag.StringVarP(&config.Template, "template", "t", "", "Path to the template YAML file")
	silentFlag := pflag.BoolP("silent", "s", false, "Run the script in silent mode without verbose output")
	pflag.Parse()

	// Validate that the --template flag is provided
	if config.Template == "" {
		log.Fatal("Error: The --template or -t flag is required. Please provide a path to the template YAML file.")
	}

	return config, *silentFlag
}

// Load configs and decide which platform must scan
func loadPlatform(config *types.Config) error {
	// Select target from Bugcrowd
	if config.FindTarget.BugCrowd != nil {
		err := platform.Bugcrowd(config)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	if config.FindTarget.HackerOne != nil {
		err := platform.HackerOne(config)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func main() {

	// Parse the flags
	switches, silentFlag := parseFlags()

	// Set the silent flag
	if !silentFlag {
		runner.ShowBanner()
	}

	// Load environment variables

	// Load template
	config, err := runner.LoadTemplate(switches.Template)
	if err != nil {
		fmt.Printf("Error loading template: %v\n", err)
		return
	}

	loadPlatform(config)
}
