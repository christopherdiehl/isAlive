package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	unknown = 0
	healthy = 1
	timeout = 2
	failed  = 3
)

type Host struct {
	Endpoint string
	Status   int
}

func CreateHost(endpoint string) *Host {
	return &Host{
		Endpoint: endpoint,
		Status:   unknown,
	}
}

var (
	// debug   = kingpin.Flag("debug", "Enable debug mode.").Bool()
	// timeout = kingpin.Flag("timeout", "Timeout waiting for monitoring.").Default("5s").OverrideDefaultFromEnvar("PING_TIMEOUT").Short('t').Duration()
	app = kingpin.New("isalive", "A command-line monitoring application.")

	add             = app.Command("add", "Add a new endpoint to monitoring.")
	addEndpointName = add.Arg("endpoint", "Endpoint to monitoring.").Required().String()

	remove             = app.Command("remove", "Add a new endpoint to monitoring.")
	removeEndpointName = remove.Arg("endpoint", "Endpoint to remove from monitoring list.").Required().String()

	run      = app.Command("run", "Runs the monitoring system")
	runAlert = run.Arg("alert", "Set to true to enable email alerts for failed host").Bool()
)

/* getConfigDirectory returns the directory where the configuration files are stored for isalive
 * /home/${user}/.cache/isalive
 */

func getConfigDirectory() string {
	user, err := user.Current()
	if err != nil {
		fmt.Println("User is not configured properly. Unable to retrieve home directory")
		os.Exit(1)
	}
	return user.HomeDir + "/.cache/isalive"
}
func retrieveHosts() []*Host {
	var hosts []*Host
	hostPath := getConfigDirectory() + "/hosts.txt"
	endpoints, err := ioutil.ReadFile(hostPath)
	if err != nil {
		fmt.Println("Unable to open file")
		return nil
	}
	err = json.Unmarshal([]byte(endpoints), &hosts)
	if err != nil {
		return nil
	}
	return hosts
}
func overwriteHosts(hosts []*Host) bool {
	hostPath := getConfigDirectory() + "/hosts.txt"
	hostsJson, err := json.Marshal(hosts)
	if err != nil {
		fmt.Println("Unable to marshal json")
		return false
	}
	err = ioutil.WriteFile(hostPath, hostsJson, 0700)
	if err != nil {
		fmt.Println("Unable to write file")
		return false
	}
	return true
}
func initalize() {
	path := getConfigDirectory()
	os.MkdirAll(path, 0700) // if already created nothing is done
}
func main() {
	initalize()
	kingpin.Version("0.0.1")
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// Register user
	case add.FullCommand():
		addEndpoint(*addEndpointName)

	// Post message
	case remove.FullCommand():
		removeEndpoint(*removeEndpointName)
	case run.FullCommand():
		if runAlert == nil {
			*runAlert = true
		}
		println(runAlert)
	}
}

/* addEndpoint creates a host and appends it's data to the endpoints file
 */
func addEndpoint(endpoint string) bool {
	// hostPath := getConfigDirectory() + "/hosts.txt"
	// f, err := os.OpenFile(hostPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0700)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return false
	// }
	// defer f.Close()
	// if _, err = f.WriteString(endpoint + "\n"); err != nil {
	// 	fmt.Println(err)
	// 	fmt.Println("Unable to append host to datastore")
	// 	return false
	// }
	// fmt.Println("Addition successful")
	// return true
	hosts := retrieveHosts()
	if hosts == nil {
		fmt.Println("Unable to remove endpoint")
	}
	host := CreateHost(endpoint)
	hosts = append(hosts, host)
	overwriteHosts(hosts)
	return true
}

// removeEndpoint removes the host from the string file
func removeEndpoint(endpoint string) bool {
	hosts := retrieveHosts()
	if hosts == nil {
		fmt.Println("Unable to remove endpoint")
	}
	hostIndex := -1
	for i, host := range hosts {
		if host.Endpoint == endpoint {
			hostIndex = i
			break
		}
	}
	if hostIndex != -1 {
		hosts = append(hosts[:hostIndex], hosts[hostIndex+1:]...)
	}
	overwriteHosts(hosts)
	// hostPath := getConfigDirectory() + "/hosts.txt"
	// endpoints, err := ioutil.ReadFile(hostPath)
	// if err != nil {
	// 	fmt.Println("Unable to open file")
	// 	return false
	// }
	// updatedEndpoints := strings.Replace(string(endpoints), endpoint+"\n", "", -1)
	// fmt.Println(updatedEndpoints)
	// err = ioutil.WriteFile(hostPath, []byte(updatedEndpoints), 0700)
	// if err != nil {
	// 	fmt.Println("Unable to write file")
	// 	return false
	// }
	return true
}
