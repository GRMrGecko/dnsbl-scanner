package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
)

// Basic application info.
const (
	appName        = "dnsbl-scanner"
	appDescription = "Scans DNSBL formatted files for IP ranges."
	appVersion     = "0.1"
)

// Structure for the App.
type App struct {
	flags  *Flags
	config *Config
}

var app *App

// Main program function.
func main() {
	// Parse flags and config.
	app = new(App)
	app.ParseFlags()
	app.ReadConfig()

	// Confirm blocklist files are provided.
	if len(app.config.DNSBLFiles) == 0 {
		log.Fatalln("no DNSBL files to scan, please provide in either a config file or via flags.")
	}
	// Confirm IP addresses to check are provided.
	if len(app.config.IPAddresses) == 0 {
		log.Fatalln("no IP addresses to search, please provide in either a config file or via flags.")
	}

	// Parse provided IP addresses into networks.
	var networks []*IPAddr
	for _, ipAddr := range app.config.IPAddresses {
		ip, err := ParseIPAddr(ipAddr)
		if err != nil {
			log.Fatal("Unable to parse provided IP address:", ipAddr)
		}
		networks = append(networks, ip)
	}

	// Print CSV header.
	fmt.Println("IP Address,Network,DNSBL File")

	// Read each DNS blocklist file, parse and check networks.
	for _, dnsblFile := range app.config.DNSBLFiles {
		// Open file, continue to next file if failure opening.
		file, err := os.Open(dnsblFile)
		if err != nil {
			log.Println("Unable to open file:", dnsblFile, err)
			continue
		}

		// Networks to exclude.
		var excluded []*IPAddr

		// Scan each line of the file and check if IP is in networks.
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			// Remove comments.
			commentI := strings.Index(line, "#")
			if commentI != -1 {
				line = line[:commentI]
			}

			// Trim whitespace.
			line = strings.TrimSpace(line)

			// Ignore variables.
			if strings.HasPrefix(line, "$") {
				continue
			}

			// Ignore descriptions.
			if strings.HasPrefix(line, ":") {
				continue
			}

			// Ignore empty lines.
			if line == "" {
				continue
			}

			// Only need first field.
			spaceI := strings.IndexFunc(line, unicode.IsSpace)
			if spaceI != -1 {
				line = line[:spaceI]
			}

			// If excluded, add to exclude list.
			if strings.HasPrefix(line, "!") {
				// Remove exclamation mark.
				line = line[1:]

				// Parse IP, failures should move to next line.
				ipAddr, err := ParseIPAddr(line)
				if err != nil {
					continue
				}

				// Add to excluded list and move to next line.
				excluded = append(excluded, ipAddr)
				continue
			}

			// This should be an IP address that is block listed.
			// Parse the IP address, and move to next line on parse failures.
			ipAddr, err := ParseIPAddr(line)
			if err != nil {
				continue
			}

			// If excluded, move to next line.
			for _, exclude := range excluded {
				if exclude.Contains(ipAddr) {
					continue
				}
			}

			// Check networks to see if IP block list is intercepted.
			for _, network := range networks {
				if network.Intercepts(ipAddr) {
					// If intercepts, print IP address, network, and block list file name in CSV format.
					fmt.Printf("%s,%s,%s\n", ipAddr, network, dnsblFile)
					// We can move on to next line now.
					continue
				}
			}
		}
	}
}
