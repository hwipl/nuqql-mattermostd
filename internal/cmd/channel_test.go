package cmd

import (
	"testing"
)

func TestChannelsReadFromFile(t *testing.T) {
	// configure working directory
	dir := createTestWorkDir()
	defer removeTestWorkDir(dir)
	conf.Dir = dir

	// create channels
	c := newChannels(0)

	// test read empty
	c.readFromFile()
	want := ""
	got := c.getPostID("channelID")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	// write entry and test again
	c.updatePostID("channelID", "postID")
	c.postIDs = make(map[string]string)
	c.readFromFile()
	want = "postID"
	got = c.getPostID("channelID")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
