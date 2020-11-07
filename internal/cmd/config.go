package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	conf = NewConfig("nuqql-mattermostd")
)

// Config stores the configuration
type Config struct {
	// Name is the Name of this configuration, e.g., "nuqql-mattermostd"
	Name string
	// Dir is the working directory
	Dir string
	// AF is the address family of the server socket:
	// inet (for AF_INET) or unix (for AF_UNIX)
	AF string
	// Address is the AF_INET listen Address
	Address string
	// Port is the AF_INET listen Port
	Port uint16
	// Sockfile is the AF_UNIX socket file in the working directory
	Sockfile string
	// Loglevel is the logging level: debug, info, warn, error
	Loglevel string
	// DisableHistory disables the message history
	DisableHistory bool
	// PushAccounts toggles pushing accounts to the client on connect
	PushAccounts bool
	// DisableFilterOwn disables filtering of own messages
	DisableFilterOwn bool
	// DisableEncryption disables TLS encryption
	DisableEncryption bool
}

// GetListenNetwork returns the listen network string based on the configured
// address family
func (c *Config) GetListenNetwork() string {
	if c.AF == "unix" {
		return "unix"
	}

	// treat everything else as inet
	return "tcp"
}

// GetListenAddress returns the listen address string based on the configured
// address family
func (c *Config) GetListenAddress() string {
	if c.AF == "unix" {
		return filepath.Join(c.Dir, c.Sockfile)
	}

	// treat everything else as inet
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

// NewConfig creates a new configuration identified by the program name
func NewConfig(name string) *Config {
	// construct default directory
	confDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.Join(confDir, name)

	// create and return config
	c := Config{
		Name:     name,
		Dir:      dir,
		AF:       "inet",
		Address:  "localhost",
		Port:     32000,
		Sockfile: name + ".sock",
		Loglevel: "warn",
	}
	return &c
}
