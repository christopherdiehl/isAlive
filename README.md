# isAlive

isAlive is a simple http monitoring cli application that sends an email if the specified endpoint returns a non okay status code.

## Config

Configuration files are stored in `/home/$USER/.cache/isalive`. Everything is stored in JSON for easy manipulation.

## Usage

```
usage: isalive [<flags>] <command> [<args> ...]

A command-line monitoring application.

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

Commands:
  help [<command>...]
    Show help.

  add <endpoint>
    Add a new endpoint to monitoring.

  remove <endpoint>
    Add a new endpoint to monitoring.

  scan [<alert>]
    Scans the stored hosts
```

## Cron

To run this command via cron job for more consistent monitoring and notifications, I suggest the following.
```
go install && go build && cp isAlive /usr/local/bin/isalive && chmod +x /usr/local/bin/isalive
crontab -e
0 * * * * /usr/local/bin/isalive scan true //run job every hour
```