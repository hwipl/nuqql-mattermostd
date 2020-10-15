package cmd

import (
	"fmt"
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
	// af is the address family of the server socket:
	// inet (for AF_INET) or unix (for AF_UNIX)
	af string
	// address is the AF_INET listen address
	address string
	// port is the AF_INET listen port
	port uint16
	// sockfile is the AF_UNIX socket file in the working directory
	sockfile string
	// loglevel is the logging level: debug, info, warn, error
	loglevel string
	// pushAccounts toggles pushing accounts to the client on connect
	pushAccounts bool
}

// getListenNetwork returns the listen network string based on the configured
// address family
func (c *config) getListenNetwork() string {
	if c.af == "unix" {
		return "unix"
	}

	// treat everything else as inet
	return "tcp"
}

// getListenAddress returns the listen address string based on the configured
// address family
func (c *config) getListenAddress() string {
	if c.af == "unix" {
		return filepath.Join(c.dir, c.sockfile)
	}

	// treat everything else as inet
	return fmt.Sprintf("%s:%d", c.address, c.port)
}

// newConfig creates a new configuration identified by the program name
func newConfig(name string) *config {
	// construct default directory
	confDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.Join(confDir, name)

	// create and return config
	c := config{
		name:     name,
		dir:      dir,
		af:       "inet",
		address:  "localhost",
		port:     32000,
		sockfile: name + ".sock",
		loglevel: "warn",
	}
	return &c
}
