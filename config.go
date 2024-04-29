package main

import (
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/kkyr/fig"
)

// Main configuration structure.
type Config struct {
	DNSBLFiles  []string `fig:"dnsbl_files"`
	IPAddresses []string `fig:"ip_addresses"`
}

// Read the configuration file.
func (a *App) ReadConfig() {
	// Set defaults.
	config := &Config{}

	// Gets the current user for getting the home directory.
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Configuration paths.
	localConfig, _ := filepath.Abs("./config.yaml")
	homeDirConfig := usr.HomeDir + "/.config/dnsbl-scanner/config.yaml"
	etcConfig := "/etc/dnsbl-scanner/config.yaml"

	// Determine which configuration to use.
	var configFile string
	if _, err := os.Stat(app.flags.Config); err == nil && app.flags.Config != "" {
		configFile = app.flags.Config
	} else if _, err := os.Stat(localConfig); err == nil {
		configFile = localConfig
	} else if _, err := os.Stat(homeDirConfig); err == nil {
		configFile = homeDirConfig
	} else if _, err := os.Stat(etcConfig); err == nil {
		configFile = etcConfig
	}

	// Load configurations from file if exists.
	if configFile != "" {
		filePath, fileName := path.Split(configFile)
		err = fig.Load(config,
			fig.File(fileName),
			fig.Dirs(filePath),
		)
		if err != nil {
			log.Printf("Error parsing configuration: %s\n", err)
			return
		}
	}

	// Set overrides from flags.
	if len(app.flags.DNSBLFiles) != 0 {
		config.DNSBLFiles = app.flags.DNSBLFiles
	}
	if len(app.flags.IPAddresses) != 0 {
		config.IPAddresses = app.flags.IPAddresses
	}

	// Set global config structure.
	app.config = config
}
