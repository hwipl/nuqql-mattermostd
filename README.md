# nuqql-mattermostd

nuqql-mattermostd is a network daemon that implements the nuqql interface and
uses the [Mattermost Golang
Driver](https://github.com/mattermost/mattermost-server/blob/master/model/client4.go)
to connect to Mattermost servers. It can be used as a backend for
[nuqql](https://github.com/hwipl/nuqql) or as a standalone chat client daemon.

## Quick Start

You can download and install nuqql-mattermostd with its dependencies to your
GOPATH or GOBIN with the go tool:

```console
$ go get github.com/hwipl/nuqql-mattermostd/cmd/nuqql-mattermostd
```

Make sure your GOPATH or GOBIN (e.g., `~/go/bin`) is in your PATH.

After the installation, you can run nuqql-mattermostd by running the
`nuqql-mattermostd` command:

```console
$ nuqql-mattermostd
```

By default, it listens on TCP port 32000 on your local host. So, you can
connect with, e.g., telnet to it with the following command:

```console
$ telnet localhost 32000
```

In the telnet session you can:
* add Mattermost accounts with: `account add mattermost <account> <password>`.
  * Note: the format of `<account>` is `<username>@<server>`, e.g.,
    `dummy_user@yourserver.org:8065`.
* retrieve the list of accounts and their numbers/IDs with `account list`.
* retrieve your buddy/channel list with `account <id> buddies` or `account <id>
  chat list`
* send a message to a channel with `account <id> chat send <channel> <message>`
* get a list of commands with `help`

##  Usage

You can run `nuqql-mattermostd` with the following command line arguments:

```
Usage of ./nuqql-mattermostd:
  -address address
        set AF_INET listen address (default "localhost")
  -af family
	set socket address family: "inet" for AF_INET, "unix" for AF_UNIX
        (default "inet")
  -dir directory
        set working directory (default "/home/user/.config/nuqql-mattermostd")
  -disable-encryption
        disable TLS encryption
  -disable-filterown
        disable filtering of own messages
  -disable-history
        disable message history
  -loglevel level
        set logging level: debug, info, warn, error (default "warn")
  -port port
        set AF_INET listen port (default 32000)
  -push-accounts
        push accounts to client
  -sockfile file
	set AF_UNIX socket file in working directory (default
        "nuqql-mattermostd.sock")
  -v    show version and exit
```

## Changes

* v0.1.0:
  * First/initial release.
