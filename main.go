package main

import (
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	// debug   = kingpin.Flag("debug", "Enable debug mode.").Bool()
	// timeout = kingpin.Flag("timeout", "Timeout waiting for monitoring.").Default("5s").OverrideDefaultFromEnvar("PING_TIMEOUT").Short('t').Duration()
	app = kingpin.New("isalive", "A command-line monitoring application.")

	add         = app.Command("add", "Add a new endpoint to monitoring.")
	addEndpoint = add.Arg("endpoint", "Endpoint to monitoring.").Required().String()

	remove         = app.Command("remove", "Add a new endpoint to monitoring.")
	removeEndpoint = remove.Arg("endpoint", "Endpoint to remove from monitoring list.").Required().String()

	run      = app.Command("run", "Runs the monitoring system")
	runAlert = run.Arg("alert", "Set to true to enable email alerts for failed host").Bool()
)

func main() {
	kingpin.Version("0.0.1")
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// Register user
	case add.FullCommand():
		println(*addEndpoint)

	// Post message
	case remove.FullCommand():
		println(*removeEndpoint)
	case run.FullCommand():
		if runAlert == nil {
			*runAlert = true
		}
		println(runAlert)
	}

}
