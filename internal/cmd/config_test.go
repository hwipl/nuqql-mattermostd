package cmd

import (
	"encoding/json"
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

func writeTestConfigFile(dir, content string) {
	// create new config file
	file := filepath.Join(dir, configFile)
	log.Println(file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// write content to config file
	_, err = f.WriteString(content)
	if err != nil {
		log.Fatal(err)
	}
}

func TestReadFromFile(t *testing.T) {
	c := NewConfig("testConfig")
	dir := createTestWorkDir()
	defer removeTestWorkDir(dir)
	c.Dir = dir

	// test with no config file
	want := c
	got := NewConfig("testConfig")
	got.Dir = dir
	got.ReadFromFile()

	if *got != *want {
		t.Errorf("got %v, wanted %v", got, want)
	}

	// test with empty config file
	writeTestConfigFile(dir, "")

	want = c
	got = NewConfig("testConfig")
	got.Dir = dir
	got.ReadFromFile()

	if *got != *want {
		t.Errorf("got %v, wanted %v", got, want)
	}

	// test with partial config file
	content := `{
		"loglevel": "debug"
	}`
	writeTestConfigFile(dir, content)

	want = c
	want.Loglevel = "debug"

	got = NewConfig("testConfig")
	got.Dir = dir
	got.ReadFromFile()

	if *got != *want {
		t.Errorf("got %v, wanted %v", got, want)
	}

	// test with full config file
	want = c
	want.Name = "otherTest"
	want.AF = "unit"
	want.Address = "192.168.1.1"
	want.Port = 12345
	want.Sockfile = "test.sock"
	want.Loglevel = "debug"
	want.DisableHistory = true
	want.PushAccounts = true
	want.FilterOwn = true
	want.DisableEncryption = true

	b, err := json.Marshal(want)
	if err != nil {
		log.Fatal(err)
	}
	writeTestConfigFile(dir, string(b))

	got = NewConfig("testConfig")
	got.Dir = dir
	got.ReadFromFile()

	if *got != *want {
		t.Errorf("got %v, wanted %v", got, want)
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
	filterOwn := false
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
	if c.FilterOwn != filterOwn {
		t.Errorf("got %t, wanted %t", c.FilterOwn, filterOwn)
	}
	if c.DisableEncryption != disableEncryption {
		t.Errorf("got %t, wanted %t", c.DisableEncryption,
			disableEncryption)
	}
}
