package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	// conf is the global configuration
	conf = NewConfig("nuqql-mattermostd")

	// configFile is the json file that contains the config
	configFile = "config.json"
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

// ReadFromFile reads the config from the configuration file "config.json" in
// the working directory
func (c *Config) ReadFromFile() {
	file := filepath.Join(c.Dir, configFile)

	// open file for reading
	f, err := os.Open(file)
	if err != nil {
		return
	}

	// read config from file
	dec := json.NewDecoder(f)
	for {
		err := dec.Decode(c)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}
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
