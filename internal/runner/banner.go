package runner

import "github.com/projectdiscovery/gologger"

const banner = `
  _____ _           _   _____                    _   
 |  ___(_)_ __   __| | |_   _|_ _ _ __ __ _  ___| |_ 
 | |_  | | '_ \ / _' |   | |/ _' | '__/ _' |/ _ \ __|
 |  _| | | | | | (_| |   | | (_| | | | (_| |  __/ |_ 
 |_|   |_|_| |_|\__,_|   |_|\__,_|_|  \__, |\___|\__|
                                       |___/
`

const version = `1.0.0`

// ShowBanner prints the banner (renamed to start with an uppercase letter)
func ShowBanner() {
	gologger.Print().Msgf("%s\n", banner)
	gologger.Print().Msgf("\t\t   e1l1ya\n\n")
}
