package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestGetListenNetwork(t *testing.T) {
	c := NewConfig("testConfig")

	// test unix
	want := "unix"
	c.AF = want
	got := c.GetListenNetwork()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}

	// test inet
	c.AF = "inet"
	want = "tcp"
	got = c.GetListenNetwork()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}

func TestGetListenAddress(t *testing.T) {
	c := NewConfig("testConfig")

	// test unix
	c.AF = "unix"
	want := filepath.Join(c.Dir, c.Sockfile)
	got := c.GetListenAddress()
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}

	// test inet
	c.AF = "inet"
	want = fmt.Sprintf("%s:%d", c.Address, c.Port)
	got = c.GetListenAddress()
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
	disableHistory := false
	pushAccounts := false
	disableFilterOwn := false
	disableEncryption := false

	c := NewConfig(name)
	if c.Name != name {
		t.Errorf("got %s, wanted %s", c.Name, name)
	}
	if c.Dir != dir {
		t.Errorf("got %s, wanted %s", c.Dir, dir)
	}
	if c.AF != af {
		t.Errorf("got %s, wanted %s", c.AF, af)
	}
	if c.Address != address {
		t.Errorf("got %s, wanted %s", c.Address, address)
	}
	if c.Port != port {
		t.Errorf("got %d, wanted %d", c.Port, port)
	}
	if c.Sockfile != sockfile {
		t.Errorf("got %s, wanted %s", c.Sockfile, sockfile)
	}
	if c.Loglevel != loglevel {
		t.Errorf("got %s, wanted %s", c.Loglevel, loglevel)
	}
	if c.DisableHistory != disableHistory {
		t.Errorf("got %t, wanted %t", c.DisableHistory, disableHistory)
	}
	if c.PushAccounts != pushAccounts {
		t.Errorf("got %t, wanted %t", c.PushAccounts, pushAccounts)
	}
	if c.DisableFilterOwn != disableFilterOwn {
		t.Errorf("got %t, wanted %t", c.DisableFilterOwn,
			disableFilterOwn)
	}
	if c.DisableEncryption != disableEncryption {
		t.Errorf("got %t, wanted %t", c.DisableEncryption,
			disableEncryption)
	}
}
