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
func parseFlags() (*types.Config, bool) {
	config := &types.Config{}
	pflag.StringVarP(&config.Template, "template", "t", "", "Path to the template YAML file")
	envFlag := pflag.BoolP("env", "e", false, "Load .env file from current directory if available")
	pflag.Parse()
	return config, *envFlag
}

// loadEnv loads environment variables into a struct
func loadEnv(envFlag bool) *types.EnvConfig {
	_, err := os.Stat(".env")
	envExists := err == nil

	if envFlag || envExists {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: .env file not found or could not be loaded: %v", err)
		} else {
			log.Println("Loaded .env file")
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
		fmt.Println("Bugcrowd configuration exists.")
		err := platform.Bugcrowd(config)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	if config.FindTarget.HackerOne != nil {
		fmt.Println("HackerOne configuration exists.")
		err := platform.HackerOne(config, env)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func main() {
	// Show banner (assuming a runner function exists)
	runner.ShowBanner()

	// Parse the flags
	switches, envFlag := parseFlags()

	// Load environment variables
	env := loadEnv(envFlag)
	fmt.Println("Loaded Environment Variables:", env.H1APIKey, env.H1Username)

	// Load template
	config, err := runner.LoadTemplate(switches.Template)
	if err != nil {
		fmt.Printf("Error loading template: %v\n", err)
		return
	}

	loadPlatform(config, env)
}
