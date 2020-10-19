package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestGetListenNetwork(t *testing.T) {
	c := newConfig("testConfig")

	// test unix
	want := "unix"
	c.af = want
	got := c.getListenNetwork()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}

	// test inet
	c.af = "inet"
	want = "tcp"
	got = c.getListenNetwork()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}

func TestGetListenAddress(t *testing.T) {
	c := newConfig("testConfig")

	// test unix
	c.af = "unix"
	want := filepath.Join(c.dir, c.sockfile)
	got := c.getListenAddress()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}

	// test inet
	c.af = "inet"
	want = fmt.Sprintf("%s:%d", c.address, c.port)
	got = c.getListenAddress()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}

func TestNewConfig(t *testing.T) {
	name := "testConfig"
	confDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.Join(confDir, name)
	af := "inet"
	address := "localhost"
	port := uint16(32000)
	sockfile := name + ".sock"
	loglevel := "warn"

	c := newConfig(name)
	if c.name != name {
		t.Errorf("got %s, wanted %s", c.name, name)
	}
	if c.dir != dir {
		t.Errorf("got %s, wanted %s", c.dir, dir)
	}
	if c.af != af {
		t.Errorf("got %s, wanted %s", c.af, af)
	}
	if c.address != address {
		t.Errorf("got %s, wanted %s", c.address, address)
	}
	if c.port != port {
		t.Errorf("got %d, wanted %d", c.port, port)
	}
	if c.sockfile != sockfile {
		t.Errorf("got %s, wanted %s", c.sockfile, sockfile)
	}
	if c.loglevel != loglevel {
		t.Errorf("got %s, wanted %s", c.loglevel, loglevel)
	}
}
