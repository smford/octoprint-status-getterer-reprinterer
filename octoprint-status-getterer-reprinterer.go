package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const applicationName string = "octoprint-status-getterer-reprinterer"
const applicationVersion string = "v0.4"

type OctoprintStatus struct {
	State struct {
		Error string `json:"error"`
		Flags struct {
			Cancelling    bool `json:"cancelling"`
			ClosedOrError bool `json:"closedOrError"`
			Error         bool `json:"error"`
			Finishing     bool `json:"finishing"`
			Operational   bool `json:"operational"`
			Paused        bool `json:"paused"`
			Pausing       bool `json:"pausing"`
			Printing      bool `json:"printing"`
			Ready         bool `json:"ready"`
			Resuming      bool `json:"resuming"`
			SdReady       bool `json:"sdReady"`
		} `json:"flags"`
		Text string `json:"text"`
	} `json:"state"`
}

type GettererPrinterList struct {
	Printers []struct {
		Name string `json:"name"`
		Desc string `json:"desc"`
	} `json:"printers"`
}

func init() {
	flag.String("statustoken", "", "Getterer API token")
	flag.String("gettererurl", "http://127.0.0.1:54038", "Getterer URL")
	flag.Int("ttl", 10, "TTL")
	flag.Int("padding", 2, "Column padding")
	flag.Bool("displayconfig", false, "Display configuration")
	flag.Bool("help", false, "Display help")
	flag.Bool("version", false, "Display version information")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if viper.GetBool("help") {
		displayHelp()
		os.Exit(0)
	}

	if viper.GetBool("version") {
		fmt.Println(applicationName + " " + applicationVersion)
		os.Exit(0)
	}

	configdir, configfile := filepath.Split(viper.GetString("config"))

	// set default configuration directory to current directory
	if configdir == "" {
		configdir = "."
	}

	viper.SetConfigType("yaml")
	viper.AddConfigPath(configdir)

	config := strings.TrimSuffix(configfile, ".yaml")
	config = strings.TrimSuffix(config, ".yml")

	viper.SetConfigName(config)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Config file not found")
		} else {
			log.Fatal("Config file was found but another error was discovered: ", err)
		}
	}

	if viper.GetBool("displayconfig") {
		displayConfig()
		os.Exit(0)
	}
}

func main() {

	tempPrinterList := getURL(viper.GetString("gettererurl") + "/printers?json=y")

	allPrinters := GettererPrinterList{}
	err := json.Unmarshal([]byte(tempPrinterList), &allPrinters)

	if err != nil {
		log.Fatal(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, viper.GetInt("padding"), ' ', 0)
	columnheadings := "Printer\tDesc\tStatus\tOther Status\tError\n"
	fmt.Fprint(w, columnheadings)

	url := viper.GetString("gettererurl") + "/status/"

	for _, something := range allPrinters.Printers {
		result := getURL(url + something.Name + "?json=y")

		name := something.Name
		desc := something.Desc
		state := getURL(url + something.Name)

		currentPrinterState := OctoprintStatus{}
		err = json.Unmarshal([]byte(result), &currentPrinterState)

		// fmt.Println(prettyPrint(currentPrinterState))

		if err != nil {
			log.Fatal(err)
		}

		stateOther := currentPrinterState.State.Text
		printerError := currentPrinterState.State.Error

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, desc, state, stateOther, printerError)

	}

	w.Flush()

}

func getURL(url string) string {
	client := http.Client{
		Timeout: time.Duration(viper.GetInt("ttl")) * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error: Could not create client connection to %s\n", url)
		return "error"
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			fmt.Printf("Error: Cannot take getSingle URL \"%s\", Message:%s\n", url, err)
			return "Could not take snapshot"
		}

		return string(body)
	} else {
		fmt.Println("Error: Could not getSingle: " + url + " HTTPStatus: " + string(resp.StatusCode))
		return "Could not getSingle"
	}
}

// prints out structs or json prettily
func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

// display help
func displayHelp() {
	message := `
      --config [file]       Configuration file: /path/to/file.yaml (default: "./config.yaml")
      --displayconfig       Display configuration
      --help                Display help
      --version             Display version`
	fmt.Println(applicationName + " " + applicationVersion)
	fmt.Println(message)
}

// display configuration
func displayConfig() {
	allmysettings := viper.AllSettings()
	var keys []string
	for k := range allmysettings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println("CONFIG:", k, ":", allmysettings[k])
	}
}
