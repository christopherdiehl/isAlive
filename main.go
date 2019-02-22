package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"os/user"
	"strings"
	"sync"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	// debug   = kingpin.Flag("debug", "Enable debug mode.").Bool()
	// timeout = kingpin.Flag("timeout", "Timeout waiting for monitoring.").Default("5s").OverrideDefaultFromEnvar("PING_TIMEOUT").Short('t').Duration()
	app = kingpin.New("isalive", "A command-line monitoring application.")

	add             = app.Command("add", "Add a new endpoint to monitoring.")
	addEndpointName = add.Arg("endpoint", "Endpoint to monitoring.").Required().String()

	remove             = app.Command("remove", "Add a new endpoint to monitoring.")
	removeEndpointName = remove.Arg("endpoint", "Endpoint to remove from monitoring list.").Required().String()

	scan          = app.Command("scan", "Runs the monitoring system")
	scanAlertFlag = scan.Arg("alert", "Set to true to enable email alerts for failed host").Bool()

	configure            = app.Command("configure", "Configure the email server. Please create new gmail account")
	configureFromAddress = configure.Flag("fromAddress", "The email used to send notifications").Required().String()
	configurePassword    = configure.Flag("password", "The password for the email used to send notifications").Required().String()
	configureToAddress   = configure.Flag("toAddress", "The email to recieve notifications").Required().String()
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
	case scan.FullCommand():
		if scanAlertFlag == nil {
			*scanAlertFlag = true
		}
		scanHosts(*scanAlertFlag)
	case configure.FullCommand():
		credentials := &EmailCredentials{
			ToAddress:   *configureToAddress,
			FromAddress: *configureFromAddress,
			Password:    *configurePassword,
		}
		SetEmailCredentials(credentials)
	}

}

/* Command Logic */
// addEndpoint creates a host and appends it's data to the endpoints file
func addEndpoint(endpoint string) bool {
	hosts := retrieveHosts()
	host := CreateHost(endpoint)
	hosts = append(hosts, host)
	overwriteHosts(hosts)
	return true
}

// removeEndpoint removes the host from the string file
func removeEndpoint(endpoint string) bool {
	hosts := retrieveHosts()
	hostIndex := -1
	//find location of host
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
	return true
}

func scanHosts(sendAlert bool) bool {
	emailHandler := CreateEmail()
	hosts := retrieveHosts()
	if hosts == nil {
		fmt.Println("No hosts defined")
		return true
	}
	var wg sync.WaitGroup
	for _, host := range hosts {
		wg.Add(1)
		go func(host *Host) {
			defer wg.Done()
			host.scan()
		}(host)
	}
	wg.Wait()
	for _, host := range hosts {
		fmt.Println(host.Endpoint, host.Status)
		if host.Status >= 300 {
			emailHandler.AppendFailedEndpoint(host)
		}
	}
	if sendAlert {
		emailHandler.SendEmail()
	}
	return true
}

/* Email logic */

// EmailHandler contains the email metadata
type EmailHandler struct {
	FromAddress string // account used to send emails
	Password    string // password for email to use
	ToAddress   string // email to send alerts to
	Status      string // status of email
	Body        string // body of email
}

// EmailCredentials is the struct that is actually written to disc
type EmailCredentials struct {
	FromAddress string // account used to send emails
	Password    string // password for email to use
	ToAddress   string // email to send alerts to
}

// GetEmailCredentials returns stored credentials in the config file
func GetEmailCredentials() *EmailCredentials {
	var credentials *EmailCredentials
	path := getConfigDirectory() + "/email.json"
	fileContents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	err = json.Unmarshal([]byte(fileContents), &credentials)
	if err != nil {
		return nil
	}
	return credentials
}

// SetEmailCredentials stores the credentials in the config file
func SetEmailCredentials(credentials *EmailCredentials) bool {
	path := getConfigDirectory() + "/email.json"
	credentialsJSON, err := json.Marshal(credentials)
	if err != nil {
		return false
	}
	err = ioutil.WriteFile(path, credentialsJSON, 0700)
	return true
}

// CreateEmail is used to setup the emailBody for future usage
func CreateEmail() *EmailHandler {
	credentials := GetEmailCredentials()
	if credentials == nil {
		return nil
	}
	emailBody := "To: " + credentials.ToAddress + "\r\nSubject: Health Check Failed!\r\n \r\nThe following hosts failed with the following status:\r\nhost,status\r\n"
	return &EmailHandler{
		FromAddress: credentials.FromAddress,
		Password:    credentials.Password,
		ToAddress:   credentials.ToAddress,
		Status:      "Unsent",
		Body:        emailBody,
	}
}

// SendEmail sends email with specified defaults
func (email *EmailHandler) SendEmail() bool {
	if email == nil || email.Status == "Sent" {
		return false
	}
	// TODO support more than gmail
	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", email.FromAddress, email.Password, "smtp.gmail.com"),
		email.FromAddress, []string{email.ToAddress}, []byte(email.Body))
	if err != nil {
		fmt.Println("Unable to send email due to ", err)
		email.Status = "Failed"
		return false
	}
	email.Status = "Sent"
	return true
}

// AppendFailedEndpoint adds endpoint to email Body
func (email *EmailHandler) AppendFailedEndpoint(host *Host) bool {
	if email == nil || host == nil {
		return false
	}
	email.Body = email.Body + host.Endpoint + "," + string(host.Status) + "\r\n"
	return true
}

/*Host Information*/
type Host struct {
	Endpoint string
	Status   int // see https://golang.org/src/net/http/status.go
}

//CreateHost returns pointer to Host struct and prepends https:// if invalid prefix supplied
func CreateHost(endpoint string) *Host {
	if !strings.HasPrefix(endpoint, "https://") && !strings.HasPrefix(endpoint, "http://") {
		endpoint = "https://" + endpoint
	}
	return &Host{
		Endpoint: endpoint,
		Status:   0,
	}
}

//scanHost scans host, sets status, and returns host
func (host *Host) scan() bool {
	resp, err := http.Get(host.Endpoint)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()
	host.Status = resp.StatusCode
	return true
}

// retrieveHosts retrieves the host information from the config file and returns a slice of Host struct pointers
func retrieveHosts() []*Host {
	var hosts []*Host
	hostPath := getConfigDirectory() + "/hosts.json"
	fileContents, err := ioutil.ReadFile(hostPath)
	if err != nil {
		return nil
	}
	err = json.Unmarshal([]byte(fileContents), &hosts)
	if err != nil {
		return nil
	}
	return hosts
}

// overwriteHosts overwrites the host metadata file.
// TODO: test concurrent async calls to the cli could result in faulty data.
func overwriteHosts(hosts []*Host) bool {
	hostPath := getConfigDirectory() + "/hosts.json"
	hostsJSON, err := json.Marshal(hosts)
	if err != nil {
		fmt.Println("Unable to marshal json")
		return false
	}
	err = ioutil.WriteFile(hostPath, hostsJSON, 0700)
	if err != nil {
		fmt.Println("Unable to write file")
		return false
	}
	return true
}
