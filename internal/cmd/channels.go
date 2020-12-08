package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// channels stores information about joined channels
type channels struct {
	postIDs  map[string]string
	fileName string
}

// getPostID returns the last known postID in the channel identified by chanID
func (c *channels) getPostID(chanID string) string {
	return c.postIDs[chanID]
}

// updatePostID updates the last known postID in the channel identified by
// its chanID
func (c *channels) updatePostID(chanID, postID string) {
	c.postIDs[chanID] = postID
	c.writeToFile()
}

// deleteChannel removes the channel identified by its chanID
func (c *channels) deleteChannel(chanID string) {
	delete(c.postIDs, chanID)
	c.writeToFile()
}

// readFromFile reads channels from file
func (c *channels) readFromFile() {
	file := filepath.Join(conf.Dir, c.fileName)

	// open file for reading
	f, err := os.Open(file)
	if err != nil {
		logError(err)
		return
	}

	// read accounts from file
	dec := json.NewDecoder(f)
	for {
		err := dec.Decode(&c.postIDs)
		if err == io.EOF {
			break
		}
		if err != nil {
			logFatal(err)
		}
	}
}

// writeToFile writes channels to file
func (c *channels) writeToFile() {
	file := filepath.Join(conf.Dir, c.fileName)

	// open file for writing
	f, err := os.Create(file)
	if err != nil {
		logFatal(err)
	}

	// make sure file is only readable and writable by the current user
	err = os.Chmod(file, 0600)
	if err != nil {
		logFatal(err)
	}

	// write accounts to file
	enc := json.NewEncoder(f)
	err = enc.Encode(c.postIDs)
	if err != nil {
		logFatal(err)
	}
}

func newChannels(accountID int) *channels {
	c := channels{
		postIDs:  make(map[string]string),
		fileName: fmt.Sprintf("channels%d.json", accountID),
	}
	c.readFromFile()
	return &c
}
