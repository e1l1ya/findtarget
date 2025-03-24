package main

import (
	"fmt"
	"log"
	"os"

	"github.com/e1l1ya/findtarget/internal/platform"
	"github.com/e1l1ya/findtarget/internal/runner"
	"github.com/e1l1ya/findtarget/pkg/types"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
)

// parseFlags parses the flags and sets defaults if necessary.
func parseFlags() (*types.Config, bool, bool) {
	config := &types.Config{}
	pflag.StringVarP(&config.Template, "template", "t", "", "Path to the template YAML file")
	envFlag := pflag.BoolP("env", "e", false, "Load .env file from current directory if available")
	silentFlag := pflag.BoolP("silent", "s", false, "Run the script in silent mode without verbose output")
	pflag.Parse()

	// Validate that the --template flag is provided
	if config.Template == "" {
		log.Fatal("Error: The --template or -t flag is required. Please provide a path to the template YAML file.")
	}

	return config, *envFlag, *silentFlag
}

// loadEnv loads environment variables into a struct
func loadEnv(envFlag bool) *types.EnvConfig {
	_, err := os.Stat(".env")
	envExists := err == nil

	if envFlag || envExists {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: .env file not found or could not be loaded: %v", err)
			os.Exit(1)
		} else {
		}
	}

	envConfig := &types.EnvConfig{
		H1APIKey:   os.Getenv("H1_API_KEY"),
		H1Username: os.Getenv("H1_USERNAME"),
	}

	// Validate required environment variables
	if envConfig.H1APIKey == "" {
		log.Fatal("Error: H1_API_KEY not found in environment variables")
	}
	if envConfig.H1Username == "" {
		log.Fatal("Error: H1_USERNAME not found in environment variables")
	}

	return envConfig
}

// Load configs and decide which platform must scan
func loadPlatform(config *types.Config, env *types.EnvConfig) error {
	// Select target from Bugcrowd
	if config.FindTarget.BugCrowd != nil {
		err := platform.Bugcrowd(config)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	if config.FindTarget.HackerOne != nil {
		err := platform.HackerOne(config, env)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func main() {

	// Parse the flags
	switches, envFlag, silentFlag := parseFlags()

	// Set the silent flag
	if !silentFlag {
		runner.ShowBanner()
	}

	// Load environment variables
	env := loadEnv(envFlag)

	// Load template
	config, err := runner.LoadTemplate(switches.Template)
	if err != nil {
		fmt.Printf("Error loading template: %v\n", err)
		return
	}

	loadPlatform(config, env)
}
