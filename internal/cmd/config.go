package cmd

import (
	"log"
	"os"
	"path/filepath"
)

var (
	conf = newConfig("nuqql-mattermostd")
)

// config stores the configuration
type config struct {
	// name is the name of this configuration, e.g., "nuqql-mattermostd"
	name string
	// dir is the working directory
	dir string
}

// newConfig creates a new configuration identified by the program name
func newConfig(name string) *config {
	// construct default directory
	confDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.FromSlash(confDir + "/" + name)

	// create and return config
	c := config{
		name: name,
		dir:  dir,
	}
	return &c
}
