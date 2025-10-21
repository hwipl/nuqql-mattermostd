/*
nuqql-mattermost is a network daemon that implements the nuqql interface and
uses the Mattermost Public Server API to connect to Mattermost servers. It can
be used as a backend for nuqql or as a standalone chat client daemon.
*/
package main

import "github.com/hwipl/nuqql-mattermostd/internal/cmd"

func main() {
	cmd.Run()
}
